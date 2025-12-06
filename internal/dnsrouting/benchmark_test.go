package dnsrouting

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand"
	"os"
	"testing"
	"time"
)

// Benchmark test domains
var benchmarkDomains = []string{
	"google.com",
	"github.com",
	"cloudflare.com",
	"microsoft.com",
	"amazon.com",
	"facebook.com",
	"twitter.com",
	"youtube.com",
	"wikipedia.org",
	"reddit.com",
	"stackoverflow.com",
	"linkedin.com",
	"netflix.com",
	"apple.com",
	"adobe.com",
	"baidu.com",
	"qq.com",
	"taobao.com",
	"tmall.com",
	"jd.com",
	"example.com",
	"test.com",
	"demo.com",
	"sample.com",
	"localhost.com",
}

// setupBenchmarkRouter creates a router with test rules
func setupBenchmarkRouter(withCache bool, cacheSize int) *Router {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelError, // Reduce logging overhead
	}))

	var router *Router
	if withCache {
		router = NewRouterWithCache(logger, cacheSize, 5*time.Minute)
	} else {
		router = NewRouter(logger)
	}

	// Add test rule sources
	sources := []*RuleSource{
		{
			ID:            1,
			Name:          "Test Rules 1",
			UpstreamGroup: "group1",
			Priority:      1,
			Enabled:       true,
			Rules: []Rule{
				{Domain: "google.com", MatchType: MatchTypeDomain, UpstreamGroup: "group1", Enabled: true},
				{Domain: "google.com", MatchType: MatchTypeDomainSuffix, UpstreamGroup: "group1", Enabled: true},
				{Domain: "github.com", MatchType: MatchTypeDomain, UpstreamGroup: "group1", Enabled: true},
				{Domain: "github.com", MatchType: MatchTypeDomainSuffix, UpstreamGroup: "group1", Enabled: true},
			},
			RulesCount: 4,
		},
		{
			ID:            2,
			Name:          "Test Rules 2",
			UpstreamGroup: "group2",
			Priority:      2,
			Enabled:       true,
			Rules: []Rule{
				{Domain: "cloudflare.com", MatchType: MatchTypeDomain, UpstreamGroup: "group2", Enabled: true},
				{Domain: "cloudflare.com", MatchType: MatchTypeDomainSuffix, UpstreamGroup: "group2", Enabled: true},
				{Domain: "microsoft.com", MatchType: MatchTypeDomain, UpstreamGroup: "group2", Enabled: true},
				{Domain: "microsoft.com", MatchType: MatchTypeDomainSuffix, UpstreamGroup: "group2", Enabled: true},
			},
			RulesCount: 4,
		},
		{
			ID:            3,
			Name:          "Test Rules 3",
			UpstreamGroup: "group3",
			Priority:      3,
			Enabled:       true,
			Rules: []Rule{
				{Domain: "amazon.com", MatchType: MatchTypeDomain, UpstreamGroup: "group3", Enabled: true},
				{Domain: "amazon.com", MatchType: MatchTypeDomainSuffix, UpstreamGroup: "group3", Enabled: true},
				{Domain: "facebook.com", MatchType: MatchTypeDomain, UpstreamGroup: "group3", Enabled: true},
				{Domain: "facebook.com", MatchType: MatchTypeDomainSuffix, UpstreamGroup: "group3", Enabled: true},
			},
			RulesCount: 4,
		},
	}

	for _, source := range sources {
		if err := router.AddSource(source); err != nil {
			panic(fmt.Sprintf("Failed to add source: %v", err))
		}
	}

	return router
}

// BenchmarkRouterMatch_NoCache tests routing without cache
func BenchmarkRouterMatch_NoCache(b *testing.B) {
	router := setupBenchmarkRouter(false, 0)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		domain := benchmarkDomains[i%len(benchmarkDomains)]
		_, _ = router.Match(ctx, domain)
	}
}

// BenchmarkRouterMatch_WithCache tests routing with cache
func BenchmarkRouterMatch_WithCache(b *testing.B) {
	router := setupBenchmarkRouter(true, 10000)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		domain := benchmarkDomains[i%len(benchmarkDomains)]
		_, _ = router.Match(ctx, domain)
	}
}

// BenchmarkRouterMatch_ColdCache tests first-time queries (cache misses)
func BenchmarkRouterMatch_ColdCache(b *testing.B) {
	router := setupBenchmarkRouter(true, 10000)
	ctx := context.Background()

	// Generate unique domains to ensure cache misses
	domains := make([]string, b.N)
	for i := 0; i < b.N; i++ {
		domains[i] = fmt.Sprintf("domain%d.example.com", i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = router.Match(ctx, domains[i])
	}
}

// BenchmarkRouterMatch_HotCache tests repeated queries (cache hits)
func BenchmarkRouterMatch_HotCache(b *testing.B) {
	router := setupBenchmarkRouter(true, 10000)
	ctx := context.Background()

	// Warm up cache
	for _, domain := range benchmarkDomains {
		_, _ = router.Match(ctx, domain)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		domain := benchmarkDomains[i%len(benchmarkDomains)]
		_, _ = router.Match(ctx, domain)
	}
}

// BenchmarkRouterMatch_Parallel tests concurrent routing
func BenchmarkRouterMatch_Parallel(b *testing.B) {
	router := setupBenchmarkRouter(true, 10000)
	ctx := context.Background()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			domain := benchmarkDomains[i%len(benchmarkDomains)]
			_, _ = router.Match(ctx, domain)
			i++
		}
	})
}

// BenchmarkRouterMatch_ParallelNoCache tests concurrent routing without cache
func BenchmarkRouterMatch_ParallelNoCache(b *testing.B) {
	router := setupBenchmarkRouter(false, 0)
	ctx := context.Background()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			domain := benchmarkDomains[i%len(benchmarkDomains)]
			_, _ = router.Match(ctx, domain)
			i++
		}
	})
}

// BenchmarkCache_Get tests cache get performance
func BenchmarkCache_Get(b *testing.B) {
	cache := NewLRUCache(10000, 5*time.Minute)

	// Populate cache
	for i, domain := range benchmarkDomains {
		cache.Put(domain, fmt.Sprintf("group%d", i%3), true)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		domain := benchmarkDomains[i%len(benchmarkDomains)]
		_, _ = cache.Get(domain)
	}
}

// BenchmarkCache_Put tests cache put performance
func BenchmarkCache_Put(b *testing.B) {
	cache := NewLRUCache(10000, 5*time.Minute)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		domain := fmt.Sprintf("domain%d.example.com", i)
		cache.Put(domain, "group1", true)
	}
}

// BenchmarkCache_Concurrent tests concurrent cache access
func BenchmarkCache_Concurrent(b *testing.B) {
	cache := NewLRUCache(10000, 5*time.Minute)

	// Populate cache
	for i, domain := range benchmarkDomains {
		cache.Put(domain, fmt.Sprintf("group%d", i%3), true)
	}

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			domain := benchmarkDomains[i%len(benchmarkDomains)]
			if i%2 == 0 {
				_, _ = cache.Get(domain)
			} else {
				cache.Put(domain, "group1", true)
			}
			i++
		}
	})
}

// BenchmarkRouterMatch_LargeRuleSet tests with many rules
func BenchmarkRouterMatch_LargeRuleSet(b *testing.B) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelError,
	}))

	router := NewRouterWithCache(logger, 10000, 5*time.Minute)

	// Create large rule set
	for i := 0; i < 100; i++ {
		rules := make([]Rule, 0, 200)
		for j := 0; j < 100; j++ {
			rules = append(rules, Rule{
				Domain:        fmt.Sprintf("domain%d-%d.example.com", i, j),
				MatchType:     MatchTypeDomain,
				UpstreamGroup: fmt.Sprintf("group%d", i%10),
				Enabled:       true,
			})
			rules = append(rules, Rule{
				Domain:        fmt.Sprintf("domain%d-%d.example.com", i, j),
				MatchType:     MatchTypeDomainSuffix,
				UpstreamGroup: fmt.Sprintf("group%d", i%10),
				Enabled:       true,
			})
		}
		source := &RuleSource{
			ID:            int64(i + 1),
			Name:          fmt.Sprintf("Rule Set %d", i+1),
			UpstreamGroup: fmt.Sprintf("group%d", i%10),
			Priority:      i + 1,
			Enabled:       true,
			Rules:         rules,
			RulesCount:    len(rules),
		}
		if err := router.AddSource(source); err != nil {
			b.Fatalf("Failed to add source: %v", err)
		}
	}

	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		domain := fmt.Sprintf("domain%d-%d.example.com", rand.Intn(100), rand.Intn(100))
		_, _ = router.Match(ctx, domain)
	}
}

// BenchmarkRouterMatch_CacheSizes tests different cache sizes
func BenchmarkRouterMatch_CacheSizes(b *testing.B) {
	cacheSizes := []int{1000, 5000, 10000, 50000, 100000}

	for _, size := range cacheSizes {
		b.Run(fmt.Sprintf("CacheSize_%d", size), func(b *testing.B) {
			router := setupBenchmarkRouter(true, size)
			ctx := context.Background()

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				domain := benchmarkDomains[i%len(benchmarkDomains)]
				_, _ = router.Match(ctx, domain)
			}
		})
	}
}
