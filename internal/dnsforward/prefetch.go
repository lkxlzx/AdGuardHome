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

// PrefetchManager handles active cache warming for hot domains.
type PrefetchManager struct {
	logger *slog.Logger
	server *Server

	// domains stores the expiration time for tracked domains.
	// Key: domain name (lowercase, FQDN)
	// Value: expiration time
	domains map[string]time.Time

	// hits counts the number of accesses for each domain.
	// Key: domain name
	// Value: hit count
	hits map[string]int

	// lastAccess tracks the last access time for each domain.
	// Key: domain name
	// Value: last access time
	// This is used for cleanup to prevent memory leaks.
	lastAccess map[string]time.Time

	// mu protects domains, hits, and lastAccess.
	mu sync.RWMutex

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
	return &PrefetchManager{
		logger:     s.logger.With(slogutil.KeyPrefix, "prefetch"),
		server:     s,
		domains:    make(map[string]time.Time),
		hits:       make(map[string]int),
		lastAccess: make(map[string]time.Time),
		threshold:  5,     // Default threshold: 5 hits to be considered "hot"
		maxEntries: 10000, // Maximum 10,000 tracked domains to prevent memory leaks
		stopCh:     make(chan struct{}),
		refreshSem: make(chan struct{}, 50), // Limit to 50 concurrent refresh operations
	}
}

// Start starts the background worker.
func (pm *PrefetchManager) Start() {
	go pm.worker()
}

// Stop stops the background worker.
func (pm *PrefetchManager) Stop() {
	close(pm.stopCh)
}

// Record updates the hit count and expiration time for a domain.
func (pm *PrefetchManager) Record(domain string, ttl uint32) {
	if domain == "" || ttl == 0 {
		return
	}

	domain = dns.Fqdn(domain)
	now := time.Now()

	pm.mu.Lock()
	defer pm.mu.Unlock()

	// Update hit count and last access time
	pm.hits[domain]++
	pm.lastAccess[domain] = now

	// Only track if it exceeds the threshold
	if pm.hits[domain] >= pm.threshold {
		// Calculate expiration time
		pm.domains[domain] = now.Add(time.Duration(ttl) * time.Second)
	}

	// Trigger cleanup if we exceed max entries
	if len(pm.hits) > pm.maxEntries {
		pm.logger.Info("prefetch cache exceeded max entries, triggering cleanup",
			"current", len(pm.hits),
			"max", pm.maxEntries)
		// Trigger async cleanup to avoid blocking
		go pm.cleanup()
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

	pm.mu.Lock()
	for domain, expiry := range pm.domains {
		// If expired or about to expire (within 5 seconds)
		if now.After(expiry) || now.Add(5*time.Second).After(expiry) {
			toRefresh = append(toRefresh, domain)
			// Remove from domains map to prevent immediate re-querying
			delete(pm.domains, domain)
		}
	}
	pm.mu.Unlock()

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
	pm.mu.Lock()
	defer pm.mu.Unlock()

	now := time.Now()
	cleanupThreshold := 24 * time.Hour // Remove entries not accessed in 24 hours
	lowHitThreshold := pm.threshold    // Remove entries below threshold

	var removed int

	// Clean up entries that haven't been accessed recently or have low hit counts
	for domain, lastAccess := range pm.lastAccess {
		shouldRemove := false

		// Remove if not accessed in 24 hours AND not in active domains
		// (keep hot domains even if old)
		if now.Sub(lastAccess) > cleanupThreshold {
			if _, inDomains := pm.domains[domain]; !inDomains {
				shouldRemove = true
			}
		}

		// Remove if hit count is below threshold and not in active domains
		if _, inDomains := pm.domains[domain]; !inDomains {
			if hits, exists := pm.hits[domain]; exists && hits < lowHitThreshold {
				shouldRemove = true
			}
		}

		if shouldRemove {
			delete(pm.hits, domain)
			delete(pm.domains, domain)
			delete(pm.lastAccess, domain)
			removed++
		}
	}

	if removed > 0 {
		pm.logger.Info("prefetch cleanup completed",
			"removed", removed,
			"remaining_hits", len(pm.hits),
			"remaining_domains", len(pm.domains))
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
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	
	return len(pm.hits), len(pm.domains), len(pm.lastAccess)
}
