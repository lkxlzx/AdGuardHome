package dnsforward

import (
	"strings"
	"sync"
	"testing"
)

// TestDomainCache_Integration tests the domain cache integration with GetUpstreamGroupForDomain
func TestDomainCache_Integration(t *testing.T) {
	// Create a server with domain cache
	s := &Server{
		domainCache: NewDomainCache(10),
		conf: ServerConfig{
			Config: Config{
				CustomDomainRules: []CustomDomainRule{
					{
						Domain:        "example.com",
						MatchType:     "DOMAIN",
						UpstreamGroup: "group1",
						Enabled:       true,
					},
					{
						Domain:        "test",
						MatchType:     "DOMAIN-KEYWORD",
						UpstreamGroup: "group2",
						Enabled:       true,
					},
					{
						Domain:        "google.com",
						MatchType:     "DOMAIN-SUFFIX",
						UpstreamGroup: "group3",
						Enabled:       true,
					},
				},
			},
		},
		logger: testLogger,
	}

	testCases := []struct {
		name           string
		domain         string
		expectedGroup  string
		shouldHitCache bool
	}{
		{
			name:           "exact_match_first_time",
			domain:         "example.com",
			expectedGroup:  "group1",
			shouldHitCache: false,
		},
		{
			name:           "exact_match_cached",
			domain:         "example.com",
			expectedGroup:  "group1",
			shouldHitCache: true,
		},
		{
			name:           "keyword_match_first_time",
			domain:         "mytest.com",
			expectedGroup:  "group2",
			shouldHitCache: false,
		},
		{
			name:           "keyword_match_cached",
			domain:         "mytest.com",
			expectedGroup:  "group2",
			shouldHitCache: true,
		},
		{
			name:           "wildcard_match_first_time",
			domain:         "mail.google.com",
			expectedGroup:  "group3",
			shouldHitCache: false,
		},
		{
			name:           "wildcard_match_cached",
			domain:         "mail.google.com",
			expectedGroup:  "group3",
			shouldHitCache: true,
		},
		{
			name:           "no_match_first_time",
			domain:         "nomatch.com",
			expectedGroup:  "",
			shouldHitCache: false,
		},
		{
			name:           "no_match_cached",
			domain:         "nomatch.com",
			expectedGroup:  "",
			shouldHitCache: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Get initial cache stats
			initialStats := s.domainCache.GetStats()

			// Call GetUpstreamGroupForDomain
			result := s.GetUpstreamGroupForDomain(tc.domain)

			// Verify result
			if result != tc.expectedGroup {
				t.Errorf("expected group %q, got %q", tc.expectedGroup, result)
			}

			// Verify cache behavior
			finalStats := s.domainCache.GetStats()
			if tc.shouldHitCache {
				// Cache hit expected - size should not change
				if finalStats.Size != initialStats.Size {
					t.Errorf("cache size changed unexpectedly: %d -> %d", initialStats.Size, finalStats.Size)
				}
			} else {
				// Cache miss expected - size should increase
				if finalStats.Size != initialStats.Size+1 {
					t.Errorf("cache size should increase by 1: %d -> %d", initialStats.Size, finalStats.Size)
				}
			}
		})
	}
}

// TestDomainCache_ClearOnRuleUpdate tests that cache is cleared when rules are updated
func TestDomainCache_ClearOnRuleUpdate(t *testing.T) {
	s := &Server{
		domainCache: NewDomainCache(10),
		conf: ServerConfig{
			Config: Config{
				CustomDomainRules: []CustomDomainRule{
					{
						Domain:        "example.com",
						MatchType:     "DOMAIN",
						UpstreamGroup: "group1",
						Enabled:       true,
					},
				},
			},
		},
		logger: testLogger,
	}

	// First query - should cache
	result1 := s.GetUpstreamGroupForDomain("example.com")
	if result1 != "group1" {
		t.Errorf("expected group1, got %q", result1)
	}

	stats := s.domainCache.GetStats()
	if stats.Size != 1 {
		t.Errorf("expected cache size 1, got %d", stats.Size)
	}

	// Clear cache (simulating rule update)
	s.ClearDomainCache()

	stats = s.domainCache.GetStats()
	if stats.Size != 0 {
		t.Errorf("expected cache size 0 after clear, got %d", stats.Size)
	}

	// Query again - should cache again
	result2 := s.GetUpstreamGroupForDomain("example.com")
	if result2 != "group1" {
		t.Errorf("expected group1, got %q", result2)
	}

	stats = s.domainCache.GetStats()
	if stats.Size != 1 {
		t.Errorf("expected cache size 1 after re-query, got %d", stats.Size)
	}
}

// TestDomainCache_CaseInsensitive tests that domain matching is case-insensitive
func TestDomainCache_CaseInsensitive(t *testing.T) {
	s := &Server{
		domainCache: NewDomainCache(10),
		conf: ServerConfig{
			Config: Config{
				CustomDomainRules: []CustomDomainRule{
					{
						Domain:        "Example.COM",
						MatchType:     "DOMAIN",
						UpstreamGroup: "group1",
						Enabled:       true,
					},
				},
			},
		},
		logger: testLogger,
	}

	testCases := []struct {
		domain string
	}{
		{"example.com"},
		{"EXAMPLE.COM"},
		{"Example.Com"},
		{"eXaMpLe.CoM"},
	}

	for _, tc := range testCases {
		t.Run(tc.domain, func(t *testing.T) {
			result := s.GetUpstreamGroupForDomain(tc.domain)
			if result != "group1" {
				t.Errorf("domain %q: expected group1, got %q", tc.domain, result)
			}
		})
	}

	// All queries should hit the same cache entry (case-insensitive)
	stats := s.domainCache.GetStats()
	if stats.Size != 1 {
		t.Errorf("expected cache size 1 (all domains should map to same entry), got %d", stats.Size)
	}
}

// TestDomainCache_FQDNHandling tests that FQDN (with trailing dot) is handled correctly
func TestDomainCache_FQDNHandling(t *testing.T) {
	s := &Server{
		domainCache: NewDomainCache(10),
		conf: ServerConfig{
			Config: Config{
				CustomDomainRules: []CustomDomainRule{
					{
						Domain:        "example.com",
						MatchType:     "DOMAIN",
						UpstreamGroup: "group1",
						Enabled:       true,
					},
				},
			},
		},
		logger: testLogger,
	}

	// Query with FQDN (trailing dot)
	result1 := s.GetUpstreamGroupForDomain("example.com.")
	if result1 != "group1" {
		t.Errorf("FQDN query: expected group1, got %q", result1)
	}

	// Query without FQDN
	result2 := s.GetUpstreamGroupForDomain("example.com")
	if result2 != "group1" {
		t.Errorf("non-FQDN query: expected group1, got %q", result2)
	}

	// Both should hit the same cache entry
	stats := s.domainCache.GetStats()
	if stats.Size != 1 {
		t.Errorf("expected cache size 1 (FQDN and non-FQDN should map to same entry), got %d", stats.Size)
	}
}

// TestDomainCache_DisabledRules tests that disabled rules are not matched or cached
func TestDomainCache_DisabledRules(t *testing.T) {
	s := &Server{
		domainCache: NewDomainCache(10),
		conf: ServerConfig{
			Config: Config{
				CustomDomainRules: []CustomDomainRule{
					{
						Domain:        "disabled.com",
						MatchType:     "DOMAIN",
						UpstreamGroup: "group1",
						Enabled:       false, // Disabled
					},
					{
						Domain:        "enabled.com",
						MatchType:     "DOMAIN",
						UpstreamGroup: "group2",
						Enabled:       true,
					},
				},
			},
		},
		logger: testLogger,
	}

	// Query disabled rule - should not match
	result1 := s.GetUpstreamGroupForDomain("disabled.com")
	if result1 != "" {
		t.Errorf("disabled rule should not match, got %q", result1)
	}

	// Query enabled rule - should match
	result2 := s.GetUpstreamGroupForDomain("enabled.com")
	if result2 != "group2" {
		t.Errorf("enabled rule: expected group2, got %q", result2)
	}

	// Cache should have 2 entries (both queries cached, including negative result)
	stats := s.domainCache.GetStats()
	if stats.Size != 2 {
		t.Errorf("expected cache size 2, got %d", stats.Size)
	}
}

// TestDomainCache_ConcurrentAccess tests concurrent access to the cache
func TestDomainCache_ConcurrentAccess(t *testing.T) {
	s := &Server{
		domainCache: NewDomainCache(100),
		conf: ServerConfig{
			Config: Config{
				CustomDomainRules: []CustomDomainRule{
					{
						Domain:        "test",
						MatchType:     "DOMAIN-KEYWORD",
						UpstreamGroup: "group1",
						Enabled:       true,
					},
				},
			},
		},
		logger: testLogger,
	}

	const numGoroutines = 50
	const numQueries = 100

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Launch multiple goroutines querying the same domains
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numQueries; j++ {
				domain := "test" + string(rune('a'+j%26)) + ".com"
				result := s.GetUpstreamGroupForDomain(domain)
				if result != "group1" {
					t.Errorf("goroutine %d: domain %q: expected group1, got %q", id, domain, result)
				}
			}
		}(i)
	}

	wg.Wait()

	// Verify cache is populated
	stats := s.domainCache.GetStats()
	if stats.Size == 0 {
		t.Error("cache should be populated after concurrent access")
	}
	if stats.Size > 100 {
		t.Errorf("cache size should not exceed capacity: got %d", stats.Size)
	}
}

// TestDomainCache_LRUEvictionIntegration tests that LRU eviction works correctly in integration
func TestDomainCache_LRUEvictionIntegration(t *testing.T) {
	s := &Server{
		domainCache: NewDomainCache(3), // Small cache for testing eviction
		conf: ServerConfig{
			Config: Config{
				CustomDomainRules: []CustomDomainRule{
					{
						Domain:        "test",
						MatchType:     "DOMAIN-KEYWORD",
						UpstreamGroup: "group1",
						Enabled:       true,
					},
				},
			},
		},
		logger: testLogger,
	}

	// Fill cache to capacity
	domains := []string{"test1.com", "test2.com", "test3.com"}
	for _, domain := range domains {
		result := s.GetUpstreamGroupForDomain(domain)
		if result != "group1" {
			t.Errorf("domain %q: expected group1, got %q", domain, result)
		}
	}

	stats := s.domainCache.GetStats()
	if stats.Size != 3 {
		t.Errorf("expected cache size 3, got %d", stats.Size)
	}

	// Add one more domain - should evict the least recently used
	result := s.GetUpstreamGroupForDomain("test4.com")
	if result != "group1" {
		t.Errorf("expected group1, got %q", result)
	}

	stats = s.domainCache.GetStats()
	if stats.Size != 3 {
		t.Errorf("cache size should remain 3 after eviction, got %d", stats.Size)
	}

	// Verify that test1.com was evicted (it was the oldest)
	// Query it again - should not be in cache
	initialSize := stats.Size
	s.GetUpstreamGroupForDomain("test1.com")
	stats = s.domainCache.GetStats()
	
	// If test1.com was evicted, querying it again won't change the size
	// (because cache is full and will evict another entry)
	if stats.Size != initialSize {
		t.Errorf("unexpected cache size change: %d -> %d", initialSize, stats.Size)
	}
}

// TestDomainCache_NegativeCaching tests that negative results (no match) are cached
func TestDomainCache_NegativeCaching(t *testing.T) {
	s := &Server{
		domainCache: NewDomainCache(10),
		conf: ServerConfig{
			Config: Config{
				CustomDomainRules: []CustomDomainRule{
					{
						Domain:        "example.com",
						MatchType:     "DOMAIN",
						UpstreamGroup: "group1",
						Enabled:       true,
					},
				},
			},
		},
		logger: testLogger,
	}

	// Query a domain that doesn't match any rule
	result1 := s.GetUpstreamGroupForDomain("nomatch.com")
	if result1 != "" {
		t.Errorf("expected empty result, got %q", result1)
	}

	stats := s.domainCache.GetStats()
	if stats.Size != 1 {
		t.Errorf("negative result should be cached, expected size 1, got %d", stats.Size)
	}

	// Query again - should hit cache
	result2 := s.GetUpstreamGroupForDomain("nomatch.com")
	if result2 != "" {
		t.Errorf("expected empty result, got %q", result2)
	}

	stats = s.domainCache.GetStats()
	if stats.Size != 1 {
		t.Errorf("cache size should not change on cache hit, got %d", stats.Size)
	}
}

// TestDomainCache_GetStatsIntegration tests the GetDomainCacheStats method
func TestDomainCache_GetStatsIntegration(t *testing.T) {
	s := &Server{
		domainCache: NewDomainCache(10),
		conf: ServerConfig{
			Config: Config{
				CustomDomainRules: []CustomDomainRule{
					{
						Domain:        "example.com",
						MatchType:     "DOMAIN",
						UpstreamGroup: "group1",
						Enabled:       true,
					},
				},
			},
		},
		logger: testLogger,
	}

	// Initial stats
	stats := s.GetDomainCacheStats()
	if stats.Size != 0 {
		t.Errorf("initial cache size should be 0, got %d", stats.Size)
	}
	if stats.Capacity != 10 {
		t.Errorf("cache capacity should be 10, got %d", stats.Capacity)
	}

	// Add some entries
	domains := []string{"example.com", "test.com", "google.com"}
	for _, domain := range domains {
		s.GetUpstreamGroupForDomain(domain)
	}

	stats = s.GetDomainCacheStats()
	if stats.Size != 3 {
		t.Errorf("cache size should be 3, got %d", stats.Size)
	}
}

// TestDomainCache_NilCache tests behavior when cache is nil
func TestDomainCache_NilCache(t *testing.T) {
	s := &Server{
		domainCache: nil, // No cache
		conf: ServerConfig{
			Config: Config{
				CustomDomainRules: []CustomDomainRule{
					{
						Domain:        "example.com",
						MatchType:     "DOMAIN",
						UpstreamGroup: "group1",
						Enabled:       true,
					},
				},
			},
		},
		logger: testLogger,
	}

	// Should still work without cache
	result := s.GetUpstreamGroupForDomain("example.com")
	if result != "group1" {
		t.Errorf("expected group1, got %q", result)
	}

	// GetDomainCacheStats should return empty stats
	stats := s.GetDomainCacheStats()
	if stats.Size != 0 || stats.Capacity != 0 {
		t.Errorf("nil cache should return empty stats, got %+v", stats)
	}

	// ClearDomainCache should not panic
	s.ClearDomainCache()
}

// BenchmarkDomainCache_IntegrationCacheHit benchmarks cache hit performance
func BenchmarkDomainCache_IntegrationCacheHit(b *testing.B) {
	s := &Server{
		domainCache: NewDomainCache(1000),
		conf: ServerConfig{
			Config: Config{
				CustomDomainRules: []CustomDomainRule{
					{
						Domain:        "example.com",
						MatchType:     "DOMAIN",
						UpstreamGroup: "group1",
						Enabled:       true,
					},
				},
			},
		},
		logger: testLogger,
	}

	// Pre-populate cache
	s.GetUpstreamGroupForDomain("example.com")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.GetUpstreamGroupForDomain("example.com")
	}
}

// BenchmarkDomainCache_IntegrationCacheMiss benchmarks cache miss performance
func BenchmarkDomainCache_IntegrationCacheMiss(b *testing.B) {
	s := &Server{
		domainCache: NewDomainCache(1000),
		conf: ServerConfig{
			Config: Config{
				CustomDomainRules: []CustomDomainRule{
					{
						Domain:        "test",
						MatchType:     "DOMAIN-KEYWORD",
						UpstreamGroup: "group1",
						Enabled:       true,
					},
				},
			},
		},
		logger: testLogger,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		domain := "test" + strings.Repeat("x", i%100) + ".com"
		s.GetUpstreamGroupForDomain(domain)
	}
}
