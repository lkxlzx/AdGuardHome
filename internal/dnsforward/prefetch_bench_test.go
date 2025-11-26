package dnsforward

import (
	"log/slog"
	"net"
	"os"
	"testing"
)

// BenchmarkPrefetchRecord_Sequential tests sequential Record calls
func BenchmarkPrefetchRecord_Sequential(b *testing.B) {
	s := &Server{
		logger: slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError})),
		conf: ServerConfig{
			UDPListenAddrs: []*net.UDPAddr{{Port: 53}},
		},
	}

	pm := NewPrefetchManager(s)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		domain := "example.com."
		pm.Record(domain, 3600)
	}
}

// BenchmarkPrefetchRecord_Parallel tests parallel Record calls (simulates high concurrency)
func BenchmarkPrefetchRecord_Parallel(b *testing.B) {
	s := &Server{
		logger: slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError})),
		conf: ServerConfig{
			UDPListenAddrs: []*net.UDPAddr{{Port: 53}},
		},
	}

	pm := NewPrefetchManager(s)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			// Use different domains to test shard distribution
			domain := "example" + string(rune(i%100)) + ".com."
			pm.Record(domain, 3600)
			i++
		}
	})
}

// BenchmarkPrefetchRecord_SameDomain tests contention on same domain
func BenchmarkPrefetchRecord_SameDomain(b *testing.B) {
	s := &Server{
		logger: slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError})),
		conf: ServerConfig{
			UDPListenAddrs: []*net.UDPAddr{{Port: 53}},
		},
	}

	pm := NewPrefetchManager(s)
	domain := "example.com."

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			pm.Record(domain, 3600)
		}
	})
}

// BenchmarkPrefetchGetStats tests GetStats performance
func BenchmarkPrefetchGetStats(b *testing.B) {
	s := &Server{
		logger: slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError})),
		conf: ServerConfig{
			UDPListenAddrs: []*net.UDPAddr{{Port: 53}},
		},
	}

	pm := NewPrefetchManager(s)

	// Add some test data
	for i := 0; i < 1000; i++ {
		domain := "example" + string(rune(i)) + ".com."
		pm.Record(domain, 3600)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pm.GetStats()
	}
}

// BenchmarkPrefetchCleanup tests cleanup performance
func BenchmarkPrefetchCleanup(b *testing.B) {
	s := &Server{
		logger: slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError})),
		conf: ServerConfig{
			UDPListenAddrs: []*net.UDPAddr{{Port: 53}},
		},
	}

	pm := NewPrefetchManager(s)

	// Add test data before each cleanup
	for i := 0; i < b.N; i++ {
		// Add 1000 domains
		for j := 0; j < 1000; j++ {
			domain := "example" + string(rune(j)) + ".com."
			pm.Record(domain, 3600)
		}

		b.StartTimer()
		pm.cleanup()
		b.StopTimer()
	}
}
