package dnsforward

import (
	"fmt"
	"log/slog"
	"net"
	"os"
	"testing"
	"time"
)

// TestCleanupIncremental tests the incremental LRU cleanup functionality
func TestCleanupIncremental(t *testing.T) {
	// Create a test server
	s := &Server{
		logger: slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})),
		conf: ServerConfig{
			UDPListenAddrs: []*net.UDPAddr{{Port: 53}},
		},
	}

	pm := NewPrefetchManager(s)

	// Add test entries exceeding maxEntries
	maxEntries := pm.maxEntries
	testEntries := maxEntries + 1000 // Exceed by 1000

	t.Logf("Adding %d test entries (maxEntries: %d)", testEntries, maxEntries)

	// Add entries to multiple shards
	now := time.Now()
	oldTime := now.Add(-48 * time.Hour) // Make them old enough for cleanup
	
	for i := 0; i < testEntries; i++ {
		domain := fmt.Sprintf("test-%d.example.com.", i)
		pm.Record(domain, 300) // 5 minute TTL
		
		// Manually set lastAccess to old time to trigger cleanup
		shard := pm.getShard(domain)
		shard.mu.Lock()
		shard.lastAccess[domain] = oldTime
		shard.mu.Unlock()
	}

	// Wait a bit for records to be processed
	time.Sleep(100 * time.Millisecond)

	// Get initial count
	initialCount := pm.getTotalEntries()
	t.Logf("Initial entries: %d", initialCount)

	if initialCount < maxEntries {
		t.Fatalf("Expected at least %d entries, got %d", maxEntries, initialCount)
	}

	// Run incremental cleanup
	t.Log("Running incremental cleanup...")
	pm.cleanupIncremental()

	// Wait for cleanup to complete
	time.Sleep(100 * time.Millisecond)

	// Get count after cleanup
	afterCount := pm.getTotalEntries()
	t.Logf("Entries after cleanup: %d", afterCount)

	// Verify cleanup happened
	if afterCount >= initialCount {
		t.Errorf("Cleanup did not reduce entries: before=%d, after=%d", initialCount, afterCount)
	}

	// Verify nextCleanupShard was updated
	nextShard := pm.nextCleanupShard.Load()
	expectedShard := int32(len(pm.shards) / 4) // Should be 4 for 16 shards
	if nextShard != expectedShard {
		t.Errorf("nextCleanupShard not updated correctly: got %d, expected %d", nextShard, expectedShard)
	}

	t.Logf("Cleanup removed %d entries", initialCount-afterCount)
	t.Logf("Next cleanup shard: %d", nextShard)
}

// TestCleanupIncrementalRotation tests that incremental cleanup rotates through shards
func TestCleanupIncrementalRotation(t *testing.T) {
	s := &Server{
		logger: slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})),
		conf: ServerConfig{
			UDPListenAddrs: []*net.UDPAddr{{Port: 53}},
		},
	}

	pm := NewPrefetchManager(s)

	// Add entries exceeding maxEntries
	testEntries := pm.maxEntries + 2000
	for i := 0; i < testEntries; i++ {
		domain := fmt.Sprintf("test-%d.example.com.", i)
		pm.Record(domain, 300)
	}

	time.Sleep(100 * time.Millisecond)

	// Run cleanup 4 times to complete a full rotation
	shardsPerRound := len(pm.shards) / 4
	t.Logf("Running 4 cleanup rounds (shards per round: %d)", shardsPerRound)

	for round := 0; round < 4; round++ {
		beforeShard := pm.nextCleanupShard.Load()
		pm.cleanupIncremental()
		afterShard := pm.nextCleanupShard.Load()

		expectedShard := int32((int(beforeShard) + shardsPerRound) % len(pm.shards))
		if afterShard != expectedShard {
			t.Errorf("Round %d: shard rotation incorrect: got %d, expected %d",
				round, afterShard, expectedShard)
		}

		t.Logf("Round %d: shard %d -> %d", round, beforeShard, afterShard)
	}

	// After 4 rounds, should be back to 0
	finalShard := pm.nextCleanupShard.Load()
	if finalShard != 0 {
		t.Errorf("After 4 rounds, expected shard 0, got %d", finalShard)
	}

	t.Log("Shard rotation completed successfully")
}

// TestCleanupIncrementalLRU tests that LRU (oldest entries) are removed first
func TestCleanupIncrementalLRU(t *testing.T) {
	s := &Server{
		logger: slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})),
		conf: ServerConfig{
			UDPListenAddrs: []*net.UDPAddr{{Port: 53}},
		},
	}

	pm := NewPrefetchManager(s)

	// Add old entries
	oldDomains := make([]string, 100)
	for i := 0; i < 100; i++ {
		domain := fmt.Sprintf("old-%d.example.com.", i)
		oldDomains[i] = domain
		pm.Record(domain, 300)
	}

	// Wait and make them old
	time.Sleep(200 * time.Millisecond)

	// Manually set lastAccess to old time for these domains
	oldTime := time.Now().Add(-48 * time.Hour)
	for _, domain := range oldDomains {
		shard := pm.getShard(domain)
		shard.mu.Lock()
		shard.lastAccess[domain] = oldTime
		shard.mu.Unlock()
	}

	// Add new entries to exceed maxEntries
	newDomains := make([]string, pm.maxEntries+500)
	for i := 0; i < len(newDomains); i++ {
		domain := fmt.Sprintf("new-%d.example.com.", i)
		newDomains[i] = domain
		pm.Record(domain, 300)
	}

	time.Sleep(100 * time.Millisecond)

	// Run cleanup
	t.Log("Running cleanup to remove old entries...")
	pm.cleanupIncremental()
	time.Sleep(100 * time.Millisecond)

	// Check that old domains were removed
	oldRemoved := 0
	newRemoved := 0

	for _, domain := range oldDomains {
		shard := pm.getShard(domain)
		shard.mu.RLock()
		_, exists := shard.lastAccess[domain]
		shard.mu.RUnlock()
		if !exists {
			oldRemoved++
		}
	}

	for _, domain := range newDomains[:100] { // Check first 100 new domains
		shard := pm.getShard(domain)
		shard.mu.RLock()
		_, exists := shard.lastAccess[domain]
		shard.mu.RUnlock()
		if !exists {
			newRemoved++
		}
	}

	t.Logf("Old domains removed: %d/%d", oldRemoved, len(oldDomains))
	t.Logf("New domains removed: %d/100", newRemoved)

	// Old domains should be removed more than new domains (LRU behavior)
	if oldRemoved <= newRemoved {
		t.Errorf("LRU not working correctly: old removed=%d, new removed=%d",
			oldRemoved, newRemoved)
	}

	t.Log("LRU cleanup working correctly")
}

// TestCleanupIncrementalNoOpWhenUnderLimit tests that cleanup does nothing when under limit
func TestCleanupIncrementalNoOpWhenUnderLimit(t *testing.T) {
	s := &Server{
		logger: slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})),
		conf: ServerConfig{
			UDPListenAddrs: []*net.UDPAddr{{Port: 53}},
		},
	}

	pm := NewPrefetchManager(s)

	// Add entries below maxEntries
	testEntries := pm.maxEntries / 2
	for i := 0; i < testEntries; i++ {
		domain := fmt.Sprintf("test-%d.example.com.", i)
		pm.Record(domain, 300)
	}

	time.Sleep(100 * time.Millisecond)

	beforeCount := pm.getTotalEntries()
	t.Logf("Entries before cleanup: %d (maxEntries: %d)", beforeCount, pm.maxEntries)

	// Run cleanup
	pm.cleanupIncremental()
	time.Sleep(100 * time.Millisecond)

	afterCount := pm.getTotalEntries()
	t.Logf("Entries after cleanup: %d", afterCount)

	// Should not remove anything when under limit
	if afterCount != beforeCount {
		t.Errorf("Cleanup should not remove entries when under limit: before=%d, after=%d",
			beforeCount, afterCount)
	}

	t.Log("No-op cleanup working correctly")
}
