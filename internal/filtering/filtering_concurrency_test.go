package filtering

import (
	"context"
	"sync"
	"testing"

	"github.com/AdguardTeam/golibs/logutil/slogutil"
	"github.com/miekg/dns"
)

// BenchmarkMatchHost_Sequential measures sequential performance
func BenchmarkMatchHost_Sequential(b *testing.B) {
	d := createTestDNSFilter(b)
	defer d.Close()

	setts := &Settings{
		FilteringEnabled:  true,
		ProtectionEnabled: true,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = d.matchHost("example.com", dns.TypeA, setts)
	}
}

// BenchmarkMatchHost_Parallel measures parallel performance with read locks
func BenchmarkMatchHost_Parallel(b *testing.B) {
	d := createTestDNSFilter(b)
	defer d.Close()

	setts := &Settings{
		FilteringEnabled:  true,
		ProtectionEnabled: true,
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = d.matchHost("example.com", dns.TypeA, setts)
		}
	})
}

// BenchmarkBlockedResponseTTL_Sequential measures config read performance
func BenchmarkBlockedResponseTTL_Sequential(b *testing.B) {
	d := createTestDNSFilter(b)
	defer d.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = d.BlockedResponseTTL()
	}
}

// BenchmarkBlockedResponseTTL_Parallel measures parallel config read
func BenchmarkBlockedResponseTTL_Parallel(b *testing.B) {
	d := createTestDNSFilter(b)
	defer d.Close()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = d.BlockedResponseTTL()
		}
	})
}

// BenchmarkProtectionStatus_Sequential measures protection status read
func BenchmarkProtectionStatus_Sequential(b *testing.B) {
	d := createTestDNSFilter(b)
	defer d.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = d.ProtectionStatus()
	}
}

// BenchmarkProtectionStatus_Parallel measures parallel protection status read
func BenchmarkProtectionStatus_Parallel(b *testing.B) {
	d := createTestDNSFilter(b)
	defer d.Close()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = d.ProtectionStatus()
		}
	})
}

// BenchmarkMixedOperations_Parallel simulates real-world mixed operations
func BenchmarkMixedOperations_Parallel(b *testing.B) {
	d := createTestDNSFilter(b)
	defer d.Close()

	setts := &Settings{
		FilteringEnabled:  true,
		ProtectionEnabled: true,
	}

	domains := []string{
		"example.com",
		"test.com",
		"google.com",
		"github.com",
		"stackoverflow.com",
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			// 80% DNS queries
			if i%10 < 8 {
				domain := domains[i%len(domains)]
				_, _ = d.matchHost(domain, dns.TypeA, setts)
			} else if i%10 == 8 {
				// 10% config reads
				_ = d.BlockedResponseTTL()
			} else {
				// 10% protection status reads
				_, _ = d.ProtectionStatus()
			}
			i++
		}
	})
}

// Helper function to create a test DNSFilter
func createTestDNSFilter(b *testing.B) *DNSFilter {
	b.Helper()

	logger := slogutil.NewDiscardLogger()

	d := &DNSFilter{
		logger:     logger,
		idGen:      newIDGenerator(0, logger),
		engineLock: sync.RWMutex{},
		confMu:     &sync.RWMutex{},
		conf: &Config{
			BlockedResponseTTL: 300,
			ProtectionEnabled:  true,
			FilteringEnabled:   true,
			filtersMu:          &sync.RWMutex{},
		},
	}

	// Initialize with empty engines to avoid nil pointer
	ctx := context.Background()
	_ = d.initFiltering(ctx, nil, nil, nil)

	return d
}
