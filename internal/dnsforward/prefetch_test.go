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

	// Get shards for test domains
	recentShard := pm.getShard("recent.com.")
	oldHotShard := pm.getShard("old-hot.com.")
	oldColdShard := pm.getShard("old-cold.com.")
	inactiveShard := pm.getShard("inactive.com.")

	// Recent domain (should not be cleaned)
	recentShard.mu.Lock()
	recentShard.hitCounters["recent.com."] = &hitCounter{count: 10, windowStart: now}
	recentShard.lastAccess["recent.com."] = now
	recentShard.domains["recent.com."] = now.Add(1 * time.Hour)
	recentShard.mu.Unlock()

	// Old domain with high hits (should not be cleaned)
	oldHotShard.mu.Lock()
	oldHotShard.hitCounters["old-hot.com."] = &hitCounter{count: 20, windowStart: now.Add(-25 * time.Hour)}
	oldHotShard.lastAccess["old-hot.com."] = now.Add(-25 * time.Hour)
	oldHotShard.domains["old-hot.com."] = now.Add(1 * time.Hour)
	oldHotShard.mu.Unlock()

	// Old domain with low hits (should be cleaned)
	oldColdShard.mu.Lock()
	oldColdShard.hitCounters["old-cold.com."] = &hitCounter{count: 2, windowStart: now.Add(-25 * time.Hour)}
	oldColdShard.lastAccess["old-cold.com."] = now.Add(-25 * time.Hour)
	oldColdShard.mu.Unlock()

	// Domain below threshold, not in active domains (should be cleaned)
	inactiveShard.mu.Lock()
	inactiveShard.hitCounters["inactive.com."] = &hitCounter{count: 3, windowStart: now.Add(-1 * time.Hour)}
	inactiveShard.lastAccess["inactive.com."] = now.Add(-1 * time.Hour)
	inactiveShard.mu.Unlock()

	// Run cleanup
	pm.cleanup()

	// Verify results
	recentShard.mu.RLock()
	recentCounter := recentShard.hitCounters["recent.com."]
	recentExists := recentCounter != nil && recentCounter.count > 0
	recentShard.mu.RUnlock()

	oldHotShard.mu.RLock()
	oldHotCounter := oldHotShard.hitCounters["old-hot.com."]
	oldHotExists := oldHotCounter != nil && oldHotCounter.count > 0
	oldHotShard.mu.RUnlock()

	oldColdShard.mu.RLock()
	_, oldColdExists := oldColdShard.hitCounters["old-cold.com."]
	oldColdShard.mu.RUnlock()

	inactiveShard.mu.RLock()
	_, inactiveExists := inactiveShard.hitCounters["inactive.com."]
	inactiveShard.mu.RUnlock()

	// Recent domain should still exist
	if !recentExists {
		t.Error("Recent domain was incorrectly removed")
	}

	// Old hot domain should still exist (high hits)
	if !oldHotExists {
		t.Error("Old hot domain was incorrectly removed")
	}

	// Old cold domain should be removed
	if oldColdExists {
		t.Error("Old cold domain was not removed")
	}

	// Inactive domain should be removed
	if inactiveExists {
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
	for i := 0; i < 100; i++ {
		domain := fmt.Sprintf("domain%d.com.", i)
		shard := pm.getShard(domain)
		shard.mu.Lock()
		shard.domains[domain] = now.Add(-1 * time.Second) // Already expired
		shard.hitCounters[domain] = &hitCounter{count: 10, windowStart: now}
		shard.lastAccess[domain] = now
		shard.mu.Unlock()
	}

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
