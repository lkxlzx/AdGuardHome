package dnsrouting

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"
)

func TestBatchMatcher_MatchBatch(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	router := NewRouter(logger)

	// Add some rules
	router.SetCustomRules([]Rule{
		{Domain: "example.com", MatchType: MatchTypeDomainSuffix, UpstreamGroup: "group1", Enabled: true},
		{Domain: "test.com", MatchType: MatchTypeDomainSuffix, UpstreamGroup: "group2", Enabled: true},
	})

	batcher := NewBatchMatcher(router, 2)

	domains := []string{
		"www.example.com",
		"api.test.com",
		"unknown.com",
		"example.com",
	}

	ctx := context.Background()
	results := batcher.MatchBatch(ctx, domains)

	if len(results) != len(domains) {
		t.Fatalf("Expected %d results, got %d", len(domains), len(results))
	}

	// Check results
	expected := []struct {
		domain string
		group  string
		found  bool
	}{
		{"www.example.com", "group1", true},
		{"api.test.com", "group2", true},
		{"unknown.com", "", false},
		{"example.com", "group1", true},
	}

	for i, exp := range expected {
		result := results[i]
		if result.Domain != exp.domain {
			t.Errorf("Result %d: expected domain %s, got %s", i, exp.domain, result.Domain)
		}
		if result.Matched != exp.found {
			t.Errorf("Result %d: expected matched=%v, got %v", i, exp.found, result.Matched)
		}
		if result.Matched && result.UpstreamGroup != exp.group {
			t.Errorf("Result %d: expected group %s, got %s", i, exp.group, result.UpstreamGroup)
		}
	}
}

func TestBatchMatcher_MatchBatchSimple(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	router := NewRouter(logger)

	router.SetCustomRules([]Rule{
		{Domain: "example.com", MatchType: MatchTypeDomainSuffix, UpstreamGroup: "group1", Enabled: true},
	})

	batcher := NewBatchMatcher(router, 2)

	domains := []string{"www.example.com", "test.com"}
	ctx := context.Background()
	results := batcher.MatchBatchSimple(ctx, domains)

	if len(results) != 2 {
		t.Fatalf("Expected 2 results, got %d", len(results))
	}

	if !results[0].Matched {
		t.Error("Expected first result to match")
	}
	if results[1].Matched {
		t.Error("Expected second result not to match")
	}
}

func TestBatchMatcher_MatchBatchMap(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	router := NewRouter(logger)

	router.SetCustomRules([]Rule{
		{Domain: "example.com", MatchType: MatchTypeDomainSuffix, UpstreamGroup: "group1", Enabled: true},
	})

	batcher := NewBatchMatcher(router, 2)

	domains := []string{"www.example.com", "test.com"}
	ctx := context.Background()
	resultMap := batcher.MatchBatchMap(ctx, domains)

	if len(resultMap) != 2 {
		t.Fatalf("Expected 2 results in map, got %d", len(resultMap))
	}

	if result, exists := resultMap["www.example.com"]; !exists || !result.Matched {
		t.Error("Expected www.example.com to match")
	}

	if result, exists := resultMap["test.com"]; !exists || result.Matched {
		t.Error("Expected test.com not to match")
	}
}

func TestBatchMatcher_WithCallback(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	router := NewRouter(logger)

	router.SetCustomRules([]Rule{
		{Domain: "example.com", MatchType: MatchTypeDomainSuffix, UpstreamGroup: "group1", Enabled: true},
	})

	batcher := NewBatchMatcher(router, 2)

	domains := []string{"www.example.com", "test.com", "api.example.com"}
	ctx := context.Background()

	resultCount := 0
	matchedCount := 0

	batcher.BatchMatchWithCallback(ctx, domains, func(result BatchResult) {
		resultCount++
		if result.Matched {
			matchedCount++
		}
	})

	if resultCount != 3 {
		t.Errorf("Expected 3 results, got %d", resultCount)
	}

	if matchedCount != 2 {
		t.Errorf("Expected 2 matches, got %d", matchedCount)
	}
}

func TestBatchMatcher_ContextCancellation(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	router := NewRouter(logger)

	batcher := NewBatchMatcher(router, 2)

	domains := make([]string, 1000)
	for i := range domains {
		domains[i] = "example.com"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	results := batcher.MatchBatch(ctx, domains)

	// Should have some results, but may not complete all due to cancellation
	if len(results) == 0 {
		t.Error("Expected some results even with cancellation")
	}
}

func TestBatchMatcher_EmptyInput(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	router := NewRouter(logger)

	batcher := NewBatchMatcher(router, 2)

	ctx := context.Background()
	results := batcher.MatchBatch(ctx, []string{})

	if len(results) != 0 {
		t.Errorf("Expected 0 results for empty input, got %d", len(results))
	}
}

func BenchmarkBatchMatcher_Small(b *testing.B) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	router := NewRouter(logger)

	router.SetCustomRules([]Rule{
		{Domain: "example.com", MatchType: MatchTypeDomainSuffix, UpstreamGroup: "group1", Enabled: true},
		{Domain: "test.com", MatchType: MatchTypeDomainSuffix, UpstreamGroup: "group2", Enabled: true},
	})

	batcher := NewBatchMatcher(router, 4)
	domains := []string{"www.example.com", "api.test.com", "unknown.com"}
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		batcher.MatchBatch(ctx, domains)
	}
}

func BenchmarkBatchMatcher_Large(b *testing.B) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	router := NewRouter(logger)

	router.SetCustomRules([]Rule{
		{Domain: "example.com", MatchType: MatchTypeDomainSuffix, UpstreamGroup: "group1", Enabled: true},
		{Domain: "test.com", MatchType: MatchTypeDomainSuffix, UpstreamGroup: "group2", Enabled: true},
	})

	batcher := NewBatchMatcher(router, 8)
	
	// Create 100 domains
	domains := make([]string, 100)
	for i := range domains {
		if i%2 == 0 {
			domains[i] = "www.example.com"
		} else {
			domains[i] = "api.test.com"
		}
	}
	
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		batcher.MatchBatch(ctx, domains)
	}
}

func BenchmarkBatchMatcher_Sequential(b *testing.B) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	router := NewRouter(logger)

	router.SetCustomRules([]Rule{
		{Domain: "example.com", MatchType: MatchTypeDomainSuffix, UpstreamGroup: "group1", Enabled: true},
	})

	domains := make([]string, 100)
	for i := range domains {
		domains[i] = "www.example.com"
	}
	
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, domain := range domains {
			router.Match(ctx, domain)
		}
	}
}
