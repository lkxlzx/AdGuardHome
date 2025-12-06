package home

import (
	"fmt"
	"testing"
	"time"

	"github.com/AdguardTeam/AdGuardHome/internal/dnsforward"
	"github.com/google/uuid"
)

// setupUpstreamGroupsForBenchmark creates a test configuration with multiple upstream groups
func setupUpstreamGroupsForBenchmark(groupCount int) *dnsConfig {
	config := &dnsConfig{
		UpstreamGroups: make([]UpstreamGroup, groupCount),
	}
	
	// Create test upstream groups
	for i := 0; i < groupCount; i++ {
		config.UpstreamGroups[i] = UpstreamGroup{
			ID:           uuid.New().String(),
			Name:         fmt.Sprintf("Group %d", i+1),
			Enabled:      true,
			IsDefault:    i == 0,
			UpstreamDNS:  []string{"https://dns.cloudflare.com/dns-query"},
			BootstrapDNS: []string{"1.1.1.1", "1.0.0.1"},
			CreatedAt:    time.Now().UTC().Format(time.RFC3339),
			UpdatedAt:    time.Now().UTC().Format(time.RFC3339),
		}
	}
	
	return config
}

// BenchmarkUpstreamGroupLookup_Linear tests the original O(n) linear search
func BenchmarkUpstreamGroupLookup_Linear(b *testing.B) {
	groupCounts := []int{5, 10, 20, 50, 100}
	
	for _, count := range groupCounts {
		b.Run(fmt.Sprintf("Groups_%d", count), func(b *testing.B) {
			config := setupUpstreamGroupsForBenchmark(count)
			
			// Get the last group ID (worst case for linear search)
			targetID := config.UpstreamGroups[count-1].ID
			
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				// Simulate original linear search
				var result *dnsforward.UpstreamGroupConfig
				for _, group := range config.UpstreamGroups {
					if group.ID == targetID && group.Enabled {
						result = &dnsforward.UpstreamGroupConfig{
							Name:         group.Name,
							UpstreamDNS:  group.UpstreamDNS,
							BootstrapDNS: group.BootstrapDNS,
							FallbackDNS:  group.FallbackDNS,
						}
						break
					}
				}
				_ = result
			}
		})
	}
}

// BenchmarkUpstreamGroupLookup_Cached tests the new O(1) cached lookup
func BenchmarkUpstreamGroupLookup_Cached(b *testing.B) {
	groupCounts := []int{5, 10, 20, 50, 100}
	
	for _, count := range groupCounts {
		b.Run(fmt.Sprintf("Groups_%d", count), func(b *testing.B) {
			config := setupUpstreamGroupsForBenchmark(count)
			
			// Build cache
			config.buildUpstreamGroupCache()
			
			// Get the last group ID (same as linear search for fair comparison)
			targetID := config.UpstreamGroups[count-1].ID
			
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				// Use O(1) cached lookup
				result := config.getUpstreamGroupConfig(targetID)
				_ = result
			}
		})
	}
}

// BenchmarkUpstreamGroupLookup_ConcurrentCached tests concurrent access to cached lookup
func BenchmarkUpstreamGroupLookup_ConcurrentCached(b *testing.B) {
	config := setupUpstreamGroupsForBenchmark(50)
	config.buildUpstreamGroupCache()
	
	// Get multiple target IDs for concurrent access
	targetIDs := make([]string, 10)
	for i := 0; i < 10; i++ {
		targetIDs[i] = config.UpstreamGroups[i*5].ID
	}
	
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			targetID := targetIDs[i%len(targetIDs)]
			result := config.getUpstreamGroupConfig(targetID)
			_ = result
			i++
		}
	})
}

// BenchmarkUpstreamGroupCacheBuild tests cache building performance
func BenchmarkUpstreamGroupCacheBuild(b *testing.B) {
	groupCounts := []int{10, 50, 100, 500}
	
	for _, count := range groupCounts {
		b.Run(fmt.Sprintf("Groups_%d", count), func(b *testing.B) {
			config := setupUpstreamGroupsForBenchmark(count)
			
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				// Reset cache state
				config.invalidateUpstreamGroupCache()
				
				// Build cache
				config.buildUpstreamGroupCache()
			}
		})
	}
}
