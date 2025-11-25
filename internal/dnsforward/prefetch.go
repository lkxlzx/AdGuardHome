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

	// mu protects domains and hits.
	mu sync.RWMutex

	// threshold is the number of hits required to consider a domain "hot".
	threshold int

	// stopCh is used to signal the worker to stop.
	stopCh chan struct{}
}

// NewPrefetchManager creates a new PrefetchManager.
func NewPrefetchManager(s *Server) *PrefetchManager {
	return &PrefetchManager{
		logger:    s.logger.With(slogutil.KeyPrefix, "prefetch"),
		server:    s,
		domains:   make(map[string]time.Time),
		hits:      make(map[string]int),
		threshold: 5, // Default threshold
		stopCh:    make(chan struct{}),
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

	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.hits[domain]++

	// Only track if it exceeds the threshold
	if pm.hits[domain] >= pm.threshold {
		// Calculate expiration time, adding a small buffer (e.g., 1 second)
		// to ensure we query *after* it expires in the cache.
		// However, for prefetch, we actually want to query *slightly before* or *right at* expiry
		// if we could update the cache in place.
		// But since we rely on dnsproxy's behavior, querying right after expiry is safest
		// to trigger a fresh upstream query.
		// Let's set it to Now + TTL.
		pm.domains[domain] = time.Now().Add(time.Duration(ttl) * time.Second)
	}
}

// worker periodically checks for expired domains and refreshes them.
func (pm *PrefetchManager) worker() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-pm.stopCh:
			return
		case <-ticker.C:
			pm.checkAndRefresh()
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
			// Update expiry to prevent immediate re-querying until we get a new Record call
			// We'll assume it will be refreshed and get a new TTL soon.
			// If not accessed again, it will eventually be dropped from hits/domains (TODO: cleanup logic)
			delete(pm.domains, domain)
		}
	}
	pm.mu.Unlock()

	if len(toRefresh) == 0 {
		return
	}

	pm.logger.Debug("refreshing hot domains", "count", len(toRefresh))

	for _, domain := range toRefresh {
		go pm.refresh(domain)
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
		pm.logger.Debug("failed to refresh domain", "domain", domain, "err", err)
	} else {
		pm.logger.Debug("refreshed domain", "domain", domain)
	}
}
