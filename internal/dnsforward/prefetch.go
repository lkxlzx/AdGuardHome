package dnsforward

import (
	"fmt"
	"log/slog"
	"net"
	"sync"
	"time"

	"github.com/AdguardTeam/golibs/logutil/slogutil"
	"github.com/miekg/dns"
)

// domainShard represents a shard of the domain tracking maps.
// Using sharded maps reduces lock contention in high-concurrency scenarios.
type domainShard struct {
	mu         sync.RWMutex
	domains    map[string]time.Time
	hits       map[string]int
	lastAccess map[string]time.Time
}

// PrefetchManager handles active cache warming for hot domains.
type PrefetchManager struct {
	logger *slog.Logger
	server *Server

	// shards is an array of domain shards for reduced lock contention.
	// Using 16 shards provides good balance between memory and concurrency.
	shards [16]*domainShard

	// threshold is the number of hits required to consider a domain "hot".
	threshold int

	// maxEntries is the maximum number of entries to track.
	// When exceeded, cleanup will be triggered.
	maxEntries int

	// stopCh is used to signal the worker to stop.
	stopCh chan struct{}

	// refreshSem limits concurrent refresh operations.
	refreshSem chan struct{}
}

// NewPrefetchManager creates a new PrefetchManager.
func NewPrefetchManager(s *Server) *PrefetchManager {
	pm := &PrefetchManager{
		logger:     s.logger.With(slogutil.KeyPrefix, "prefetch"),
		server:     s,
		threshold:  5,     // Default threshold: 5 hits to be considered "hot"
		maxEntries: 10000, // Maximum 10,000 tracked domains to prevent memory leaks
		stopCh:     make(chan struct{}),
		refreshSem: make(chan struct{}, 50), // Limit to 50 concurrent refresh operations
	}

	// Initialize shards
	for i := range pm.shards {
		pm.shards[i] = &domainShard{
			domains:    make(map[string]time.Time),
			hits:       make(map[string]int),
			lastAccess: make(map[string]time.Time),
		}
	}

	return pm
}

// Start starts the background worker.
func (pm *PrefetchManager) Start() {
	go pm.worker()
}

// Stop stops the background worker.
func (pm *PrefetchManager) Stop() {
	close(pm.stopCh)
}

// getShard returns the shard for a given domain using a simple hash function.
func (pm *PrefetchManager) getShard(domain string) *domainShard {
	// Simple hash function: sum of bytes modulo number of shards
	var hash uint32
	for i := 0; i < len(domain); i++ {
		hash = hash*31 + uint32(domain[i])
	}
	return pm.shards[hash%uint32(len(pm.shards))]
}

// getTotalEntries returns the total number of entries across all shards.
func (pm *PrefetchManager) getTotalEntries() int {
	total := 0
	for _, shard := range pm.shards {
		shard.mu.RLock()
		total += len(shard.hits)
		shard.mu.RUnlock()
	}
	return total
}

// Record updates the hit count and expiration time for a domain.
func (pm *PrefetchManager) Record(domain string, ttl uint32) {
	if domain == "" || ttl == 0 {
		return
	}

	domain = dns.Fqdn(domain)
	now := time.Now()

	// Get the shard for this domain
	shard := pm.getShard(domain)

	shard.mu.Lock()
	defer shard.mu.Unlock()

	// Update hit count and last access time
	shard.hits[domain]++
	shard.lastAccess[domain] = now

	// Only track if it exceeds the threshold
	if shard.hits[domain] >= pm.threshold {
		// Calculate expiration time
		shard.domains[domain] = now.Add(time.Duration(ttl) * time.Second)
	}

	// Check if we need cleanup (check periodically, not every time)
	// Only check on shard 0 to avoid multiple cleanup triggers
	if shard == pm.shards[0] && len(shard.hits)%1000 == 0 {
		totalEntries := pm.getTotalEntries()
		if totalEntries > pm.maxEntries {
			pm.logger.Info("prefetch cache exceeded max entries, triggering cleanup",
				"current", totalEntries,
				"max", pm.maxEntries)
			// Trigger async cleanup to avoid blocking
			go pm.cleanup()
		}
	}
}

// worker periodically checks for expired domains and refreshes them.
func (pm *PrefetchManager) worker() {
	refreshTicker := time.NewTicker(10 * time.Second)
	cleanupTicker := time.NewTicker(1 * time.Hour) // Cleanup every hour
	defer refreshTicker.Stop()
	defer cleanupTicker.Stop()

	for {
		select {
		case <-pm.stopCh:
			return
		case <-refreshTicker.C:
			pm.checkAndRefresh()
		case <-cleanupTicker.C:
			pm.cleanup()
		}
	}
}

// checkAndRefresh scans for expired domains and triggers a refresh.
func (pm *PrefetchManager) checkAndRefresh() {
	now := time.Now()
	var toRefresh []string

	// Scan all shards
	for _, shard := range pm.shards {
		shard.mu.Lock()
		for domain, expiry := range shard.domains {
			// If expired or about to expire (within 5 seconds)
			if now.After(expiry) || now.Add(5*time.Second).After(expiry) {
				toRefresh = append(toRefresh, domain)
				// Remove from domains map to prevent immediate re-querying
				delete(shard.domains, domain)
			}
		}
		shard.mu.Unlock()
	}

	if len(toRefresh) == 0 {
		return
	}

	pm.logger.Debug("refreshing hot domains", "count", len(toRefresh))

	// Refresh domains with concurrency limit
	for _, domain := range toRefresh {
		domain := domain // Capture loop variable
		go func() {
			// Acquire semaphore to limit concurrent refreshes
			pm.refreshSem <- struct{}{}
			defer func() { <-pm.refreshSem }()

			pm.refresh(domain)
		}()
	}
}

// cleanup removes stale entries from hits, domains, and lastAccess maps
// to prevent memory leaks. This is called periodically and when maxEntries is exceeded.
func (pm *PrefetchManager) cleanup() {
	now := time.Now()
	cleanupThreshold := 24 * time.Hour // Remove entries not accessed in 24 hours
	lowHitThreshold := pm.threshold    // Remove entries below threshold

	var totalRemoved int
	var totalHits, totalDomains int

	// Clean up each shard
	for _, shard := range pm.shards {
		shard.mu.Lock()

		var removed int

		// Clean up entries that haven't been accessed recently or have low hit counts
		for domain, lastAccess := range shard.lastAccess {
			shouldRemove := false

			// Remove if not accessed in 24 hours AND not in active domains
			// (keep hot domains even if old)
			if now.Sub(lastAccess) > cleanupThreshold {
				if _, inDomains := shard.domains[domain]; !inDomains {
					shouldRemove = true
				}
			}

			// Remove if hit count is below threshold and not in active domains
			if _, inDomains := shard.domains[domain]; !inDomains {
				if hits, exists := shard.hits[domain]; exists && hits < lowHitThreshold {
					shouldRemove = true
				}
			}

			if shouldRemove {
				delete(shard.hits, domain)
				delete(shard.domains, domain)
				delete(shard.lastAccess, domain)
				removed++
			}
		}

		totalRemoved += removed
		totalHits += len(shard.hits)
		totalDomains += len(shard.domains)

		shard.mu.Unlock()
	}

	if totalRemoved > 0 {
		pm.logger.Info("prefetch cleanup completed",
			"removed", totalRemoved,
			"remaining_hits", totalHits,
			"remaining_domains", totalDomains)
	}
}

// refresh sends a DNS query to the server itself to trigger a cache update.
func (pm *PrefetchManager) refresh(domain string) {
	// Create a local DNS client
	c := new(dns.Client)
	c.Timeout = 5 * time.Second

	m := new(dns.Msg)
	m.SetQuestion(domain, dns.TypeA)
	m.RecursionDesired = true

	// Send query to localhost
	// We need to know the listening port. Assuming default or first configured.
	// For simplicity, we can try to use the internal Resolve method directly
	// to avoid network stack overhead and port discovery issues.
	// However, calling Server.Resolve bypasses some logic.
	// Let's try to use the network if possible, or fall back to internal processing.

	// Using network loopback is safer to simulate a real client and trigger full processing chain.
	port := "53"
	if len(pm.server.conf.UDPListenAddrs) > 0 {
		port = fmt.Sprintf("%d", pm.server.conf.UDPListenAddrs[0].Port)
	}

	// Use 127.0.0.1
	target := net.JoinHostPort("127.0.0.1", port)

	_, _, err := c.Exchange(m, target)
	if err != nil {
		// Upgrade to Warn level for production visibility
		pm.logger.Warn("failed to refresh domain", "domain", domain, "err", err)
	} else {
		pm.logger.Debug("refreshed domain", "domain", domain)
	}
}

// GetStats returns statistics about the prefetch manager.
// This is useful for monitoring and debugging.
func (pm *PrefetchManager) GetStats() (hits, domains, tracked int) {
	for _, shard := range pm.shards {
		shard.mu.RLock()
		hits += len(shard.hits)
		domains += len(shard.domains)
		tracked += len(shard.lastAccess)
		shard.mu.RUnlock()
	}
	
	return hits, domains, tracked
}
