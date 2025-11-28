package dnsforward

import (
	"log/slog"
	"net"
	"os"
	"testing"
	"time"

	"github.com/AdguardTeam/golibs/timeutil"
	"github.com/stretchr/testify/assert"
)

// TestPrefetchManager_TimeWindow tests the time window functionality.
func TestPrefetchManager_TimeWindow(t *testing.T) {
	s := &Server{
		logger: slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})),
		conf: ServerConfig{
			UDPListenAddrs: []*net.UDPAddr{{Port: 53}},
			Config: Config{
				PrefetchEnabled:                true,
				PrefetchThreshold:              3,
				PrefetchTimeWindow:             timeutil.Duration(2 * time.Second),
				PrefetchMaxEntries:             10000,
				PrefetchCleanupInterval:        timeutil.Duration(1 * time.Hour),
				PrefetchMaxConcurrentRefresh:   50,
			},
		},
	}

	pm := NewPrefetchManager(s)

	domain := "example.com"
	ttl := uint32(300)

	// Record 2 hits
	pm.Record(domain, ttl)
	pm.Record(domain, ttl)

	// Check that it's not hot yet (threshold is 3)
	shard := pm.getShard(domain + ".")
	shard.mu.RLock()
	_, isHot := shard.domains[domain+"."]
	counter := shard.hitCounters[domain+"."]
	hitCount := 0
	if counter != nil {
		hitCount = counter.count
	}
	shard.mu.RUnlock()

	assert.False(t, isHot, "Domain should not be hot with only 2 hits")
	assert.Equal(t, 2, hitCount, "Hit count should be 2")

	// Record 1 more hit (total 3, should become hot)
	pm.Record(domain, ttl)

	shard.mu.RLock()
	_, isHot = shard.domains[domain+"."]
	counter = shard.hitCounters[domain+"."]
	hitCount = 0
	if counter != nil {
		hitCount = counter.count
	}
	shard.mu.RUnlock()

	assert.True(t, isHot, "Domain should be hot with 3 hits")
	assert.Equal(t, 3, hitCount, "Hit count should be 3")

	// Wait for time window to expire (2 seconds + buffer)
	time.Sleep(2500 * time.Millisecond)

	// Record 1 hit (old hits should be expired, count should be 1)
	pm.Record(domain, ttl)

	shard.mu.RLock()
	_, isHot = shard.domains[domain+"."]
	counter = shard.hitCounters[domain+"."]
	hitCount = 0
	if counter != nil {
		hitCount = counter.count
	}
	shard.mu.RUnlock()

	assert.False(t, isHot, "Domain should not be hot after time window expiration")
	assert.Equal(t, 1, hitCount, "Hit count should be 1 after time window expiration")

	// Record 2 more hits quickly (total 3 within window)
	pm.Record(domain, ttl)
	pm.Record(domain, ttl)

	shard.mu.RLock()
	_, isHot = shard.domains[domain+"."]
	counter = shard.hitCounters[domain+"."]
	hitCount = 0
	if counter != nil {
		hitCount = counter.count
	}
	shard.mu.RUnlock()

	assert.True(t, isHot, "Domain should be hot again with 3 hits in new window")
	assert.Equal(t, 3, hitCount, "Hit count should be 3")
}

// TestPrefetchManager_TimeWindowCleanup tests that cleanup removes old timestamps.
func TestPrefetchManager_TimeWindowCleanup(t *testing.T) {
	s := &Server{
		logger: slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})),
		conf: ServerConfig{
			UDPListenAddrs: []*net.UDPAddr{{Port: 53}},
			Config: Config{
				PrefetchEnabled:                true,
				PrefetchThreshold:              5,
				PrefetchTimeWindow:             timeutil.Duration(1 * time.Second),
				PrefetchMaxEntries:             10000,
				PrefetchCleanupInterval:        timeutil.Duration(1 * time.Hour),
				PrefetchMaxConcurrentRefresh:   50,
			},
		},
	}

	pm := NewPrefetchManager(s)

	// Record hits for multiple domains
	for i := 0; i < 5; i++ {
		pm.Record("domain1.com", 300)
		pm.Record("domain2.com", 300)
	}

	// Both domains should be hot
	shard1 := pm.getShard("domain1.com.")
	shard2 := pm.getShard("domain2.com.")

	shard1.mu.RLock()
	_, isHot1 := shard1.domains["domain1.com."]
	shard1.mu.RUnlock()

	shard2.mu.RLock()
	_, isHot2 := shard2.domains["domain2.com."]
	shard2.mu.RUnlock()

	assert.True(t, isHot1, "domain1 should be hot")
	assert.True(t, isHot2, "domain2 should be hot")

	// Wait for time window to expire
	time.Sleep(1500 * time.Millisecond)

	// Record 1 hit for domain1 (should update and remove old timestamps)
	pm.Record("domain1.com", 300)

	shard1.mu.RLock()
	_, isHot1 = shard1.domains["domain1.com."]
	counter1 := shard1.hitCounters["domain1.com."]
	hitCount1 := 0
	if counter1 != nil {
		hitCount1 = counter1.count
	}
	shard1.mu.RUnlock()

	assert.False(t, isHot1, "domain1 should not be hot after window expiration")
	assert.Equal(t, 1, hitCount1, "domain1 hit count should be 1")

	// domain2 should still have old data (not accessed)
	shard2.mu.RLock()
	counter2 := shard2.hitCounters["domain2.com."]
	hitCount2 := 0
	if counter2 != nil {
		hitCount2 = counter2.count
	}
	shard2.mu.RUnlock()

	assert.Equal(t, 5, hitCount2, "domain2 should still have 5 hits (not cleaned yet)")
}

// TestPrefetchManager_TimeWindowDifferentDurations tests different time window durations.
func TestPrefetchManager_TimeWindowDifferentDurations(t *testing.T) {
	testCases := []struct {
		name       string
		timeWindow time.Duration
		threshold  int
		sleepTime  time.Duration
		expectHot  bool
	}{
		{
			name:       "Short window - hits expire",
			timeWindow: 500 * time.Millisecond,
			threshold:  3,
			sleepTime:  600 * time.Millisecond,
			expectHot:  false,
		},
		{
			name:       "Long window - hits persist",
			timeWindow: 5 * time.Second,
			threshold:  3,
			sleepTime:  600 * time.Millisecond,
			expectHot:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s := &Server{
				logger: slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})),
				conf: ServerConfig{
					UDPListenAddrs: []*net.UDPAddr{{Port: 53}},
					Config: Config{
						PrefetchEnabled:                true,
						PrefetchThreshold:              tc.threshold,
						PrefetchTimeWindow:             timeutil.Duration(tc.timeWindow),
						PrefetchMaxEntries:             10000,
						PrefetchCleanupInterval:        timeutil.Duration(1 * time.Hour),
						PrefetchMaxConcurrentRefresh:   50,
					},
				},
			}

			pm := NewPrefetchManager(s)

			domain := "test.com"

			// Record threshold-1 hits
			for i := 0; i < tc.threshold-1; i++ {
				pm.Record(domain, 300)
			}

			// Wait
			time.Sleep(tc.sleepTime)

			// Record 1 more hit
			pm.Record(domain, 300)

			shard := pm.getShard(domain + ".")
			shard.mu.RLock()
			_, isHot := shard.domains[domain+"."]
			shard.mu.RUnlock()

			assert.Equal(t, tc.expectHot, isHot, "Hot status should match expectation")
		})
	}
}
