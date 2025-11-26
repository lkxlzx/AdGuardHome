package dnsforward

import (
	"fmt"
	"log/slog"
	"net"
	"os"
	"testing"
	"time"

	"github.com/AdguardTeam/golibs/timeutil"
)

// BenchmarkPrefetch_Record benchmarks the Record operation
func BenchmarkPrefetch_Record(b *testing.B) {
	s := &Server{
		logger: slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError})),
		conf: ServerConfig{
			UDPListenAddrs: []*net.UDPAddr{{Port: 53}},
			Config: Config{
				PrefetchEnabled:                true,
				PrefetchThreshold:              5,
				PrefetchTimeWindow:             timeutil.Duration(1 * time.Hour),
				PrefetchMaxEntries:             10000,
				PrefetchCleanupInterval:        timeutil.Duration(1 * time.Hour),
				PrefetchSoftLimit:              50,
				PrefetchHardLimit:              150,
			},
		},
	}

	pm := NewPrefetchManager(s)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		domain := fmt.Sprintf("example%d.com", i%1000)
		pm.Record(domain, 300)
	}
}

// BenchmarkPrefetch_RecordHotDomains benchmarks recording hot domains
func BenchmarkPrefetch_RecordHotDomains(b *testing.B) {
	s := &Server{
		logger: slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError})),
		conf: ServerConfig{
			UDPListenAddrs: []*net.UDPAddr{{Port: 53}},
			Config: Config{
				PrefetchEnabled:                true,
				PrefetchThreshold:              5,
				PrefetchTimeWindow:             timeutil.Duration(1 * time.Hour),
				PrefetchMaxEntries:             10000,
				PrefetchCleanupInterval:        timeutil.Duration(1 * time.Hour),
				PrefetchSoftLimit:              50,
				PrefetchHardLimit:              150,
			},
		},
	}

	pm := NewPrefetchManager(s)

	// Pre-populate with hot domains
	for i := 0; i < 100; i++ {
		domain := fmt.Sprintf("hot%d.com", i)
		for j := 0; j < 10; j++ {
			pm.Record(domain, 300)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		domain := fmt.Sprintf("hot%d.com", i%100)
		pm.Record(domain, 300)
	}
}

// BenchmarkPrefetch_Cleanup benchmarks the cleanup operation
func BenchmarkPrefetch_Cleanup(b *testing.B) {
	scenarios := []struct {
		name    string
		entries int
	}{
		{"1K_entries", 1000},
		{"5K_entries", 5000},
		{"10K_entries", 10000},
		{"20K_entries", 20000},
	}

	for _, sc := range scenarios {
		b.Run(sc.name, func(b *testing.B) {
			s := &Server{
				logger: slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError})),
				conf: ServerConfig{
					UDPListenAddrs: []*net.UDPAddr{{Port: 53}},
					Config: Config{
						PrefetchEnabled:                true,
						PrefetchThreshold:              5,
						PrefetchTimeWindow:             timeutil.Duration(1 * time.Hour),
						PrefetchMaxEntries:             sc.entries,
						PrefetchCleanupInterval:        timeutil.Duration(1 * time.Hour),
						PrefetchSoftLimit:              50,
						PrefetchHardLimit:              150,
					},
				},
			}

			pm := NewPrefetchManager(s)

			// Populate with entries directly to avoid triggering cleanup
			now := time.Now()
			for i := 0; i < sc.entries; i++ {
				domain := fmt.Sprintf("domain%d.com.", i)
				shard := pm.getShard(domain)
				
				shard.mu.Lock()
				shard.hits[domain] = 3
				shard.lastAccess[domain] = now
				shard.hitTimestamps[domain] = []time.Time{now, now, now}
				
				// Make some entries old
				if i%3 == 0 {
					shard.lastAccess[domain] = now.Add(-25 * time.Hour)
				}
				shard.mu.Unlock()
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				pm.cleanup()
			}
		})
	}
}

// BenchmarkPrefetch_CleanupSmart benchmarks smart cleanup (only when needed)
func BenchmarkPrefetch_CleanupSmart(b *testing.B) {
	s := &Server{
		logger: slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError})),
		conf: ServerConfig{
			UDPListenAddrs: []*net.UDPAddr{{Port: 53}},
			Config: Config{
				PrefetchEnabled:                true,
				PrefetchThreshold:              5,
				PrefetchTimeWindow:             timeutil.Duration(1 * time.Hour),
				PrefetchMaxEntries:             10000,
				PrefetchCleanupInterval:        timeutil.Duration(1 * time.Hour),
				PrefetchSoftLimit:              50,
				PrefetchHardLimit:              150,
			},
		},
	}

	pm := NewPrefetchManager(s)

	// Populate with 5000 entries (under limit)
	for i := 0; i < 5000; i++ {
		domain := fmt.Sprintf("domain%d.com", i)
		pm.Record(domain, 300)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pm.cleanupInternal(false) // Smart cleanup
	}
}

// BenchmarkPrefetch_GetTotalEntries benchmarks counting total entries
func BenchmarkPrefetch_GetTotalEntries(b *testing.B) {
	s := &Server{
		logger: slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError})),
		conf: ServerConfig{
			UDPListenAddrs: []*net.UDPAddr{{Port: 53}},
			Config: Config{
				PrefetchEnabled:                true,
				PrefetchThreshold:              5,
				PrefetchTimeWindow:             timeutil.Duration(1 * time.Hour),
				PrefetchMaxEntries:             10000,
				PrefetchCleanupInterval:        timeutil.Duration(1 * time.Hour),
				PrefetchSoftLimit:              50,
				PrefetchHardLimit:              150,
			},
		},
	}

	pm := NewPrefetchManager(s)

	// Populate with 10000 entries
	for i := 0; i < 10000; i++ {
		domain := fmt.Sprintf("domain%d.com", i)
		pm.Record(domain, 300)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = pm.getTotalEntries()
	}
}
