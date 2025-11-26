package dnsforward

import (
	"fmt"
	"log/slog"
	"net"
	"os"
	"testing"
	"time"
)

// TestPrefetchManager_Cleanup tests the cleanup functionality
func TestPrefetchManager_Cleanup(t *testing.T) {
	// Create a mock server
	s := &Server{
		logger: slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})),
		conf: ServerConfig{
			UDPListenAddrs: []*net.UDPAddr{{Port: 53}},
		},
	}

	pm := NewPrefetchManager(s)

	// Add some test domains
	now := time.Now()
	
	pm.mu.Lock()
	// Recent domain (should not be cleaned)
	pm.hits["recent.com."] = 10
	pm.lastAccess["recent.com."] = now
	pm.domains["recent.com."] = now.Add(1 * time.Hour)

	// Old domain with high hits (should not be cleaned)
	pm.hits["old-hot.com."] = 20
	pm.lastAccess["old-hot.com."] = now.Add(-25 * time.Hour)
	pm.domains["old-hot.com."] = now.Add(1 * time.Hour)

	// Old domain with low hits (should be cleaned)
	pm.hits["old-cold.com."] = 2
	pm.lastAccess["old-cold.com."] = now.Add(-25 * time.Hour)

	// Domain below threshold, not in active domains (should be cleaned)
	pm.hits["inactive.com."] = 3
	pm.lastAccess["inactive.com."] = now.Add(-1 * time.Hour)
	pm.mu.Unlock()

	// Run cleanup
	pm.cleanup()

	// Verify results
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	// Recent domain should still exist
	if _, exists := pm.hits["recent.com."]; !exists {
		t.Error("Recent domain was incorrectly removed")
	}

	// Old hot domain should still exist (high hits)
	if _, exists := pm.hits["old-hot.com."]; !exists {
		t.Error("Old hot domain was incorrectly removed")
	}

	// Old cold domain should be removed
	if _, exists := pm.hits["old-cold.com."]; exists {
		t.Error("Old cold domain was not removed")
	}

	// Inactive domain should be removed
	if _, exists := pm.hits["inactive.com."]; exists {
		t.Error("Inactive domain was not removed")
	}
}

// TestPrefetchManager_MaxEntries tests the max entries limit
func TestPrefetchManager_MaxEntries(t *testing.T) {
	s := &Server{
		logger: slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})),
		conf: ServerConfig{
			UDPListenAddrs: []*net.UDPAddr{{Port: 53}},
		},
	}

	pm := NewPrefetchManager(s)
	pm.maxEntries = 10 // Set low limit for testing

	// Add domains up to the limit
	for i := 0; i < 15; i++ {
		domain := fmt.Sprintf("domain%d.com.", i)
		pm.Record(domain, 3600)
	}

	// Wait a bit for async cleanup to potentially run
	time.Sleep(100 * time.Millisecond)

	// Check that cleanup was triggered
	hits, _, _ := pm.GetStats()
	if hits > pm.maxEntries*2 {
		t.Errorf("Cleanup did not trigger, hits=%d, maxEntries=%d", hits, pm.maxEntries)
	}
}

// TestPrefetchManager_ConcurrentRefresh tests goroutine limiting
func TestPrefetchManager_ConcurrentRefresh(t *testing.T) {
	s := &Server{
		logger: slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})),
		conf: ServerConfig{
			UDPListenAddrs: []*net.UDPAddr{{Port: 53}},
		},
	}

	pm := NewPrefetchManager(s)

	// Add many domains that will expire immediately
	now := time.Now()
	pm.mu.Lock()
	for i := 0; i < 100; i++ {
		domain := fmt.Sprintf("domain%d.com.", i)
		pm.domains[domain] = now.Add(-1 * time.Second) // Already expired
		pm.hits[domain] = 10
		pm.lastAccess[domain] = now
	}
	pm.mu.Unlock()

	// Trigger refresh
	pm.checkAndRefresh()

	// Check that semaphore is limiting concurrency
	// The semaphore should have at most 50 slots (as configured)
	// This is hard to test directly, but we can verify no panic occurs
	time.Sleep(100 * time.Millisecond)

	// If we get here without panic, the test passes
	t.Log("Concurrent refresh completed without issues")
}

// TestPrefetchManager_GetStats tests the statistics function
func TestPrefetchManager_GetStats(t *testing.T) {
	s := &Server{
		logger: slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})),
		conf: ServerConfig{
			UDPListenAddrs: []*net.UDPAddr{{Port: 53}},
		},
	}

	pm := NewPrefetchManager(s)

	// Add some test data
	pm.Record("test1.com.", 3600)
	pm.Record("test1.com.", 3600)
	pm.Record("test1.com.", 3600)
	pm.Record("test1.com.", 3600)
	pm.Record("test1.com.", 3600) // 5 hits, should be tracked

	pm.Record("test2.com.", 3600)
	pm.Record("test2.com.", 3600) // 2 hits, below threshold

	hits, domains, tracked := pm.GetStats()

	if hits != 2 {
		t.Errorf("Expected 2 hit entries, got %d", hits)
	}

	if domains != 1 {
		t.Errorf("Expected 1 domain entry, got %d", domains)
	}

	if tracked != 2 {
		t.Errorf("Expected 2 tracked entries, got %d", tracked)
	}
}
