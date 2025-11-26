package dnsforward

import (
	"testing"
)

// Benchmark the old implementation (with string operations on every call)
func BenchmarkMatchDomainPattern_Old(b *testing.B) {
	domain := "www.example.com"
	patterns := []string{
		"DOMAIN,www.example.com",
		"DOMAIN-SUFFIX,example.com",
		"DOMAIN-KEYWORD,example",
		"example.com",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, pattern := range patterns {
			matchDomainPattern(domain, pattern)
		}
	}
}

// Benchmark the new implementation (with pre-parsed patterns)
func BenchmarkMatchDomainPattern_New(b *testing.B) {
	domain := "www.example.com"
	patterns := []string{
		"DOMAIN,www.example.com",
		"DOMAIN-SUFFIX,example.com",
		"DOMAIN-KEYWORD,example",
		"example.com",
	}

	// Pre-parse patterns (this would be done once at startup)
	parsed := make([]*ParsedPattern, len(patterns))
	for i, pattern := range patterns {
		parsed[i] = ParsePattern(pattern)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, p := range parsed {
			MatchParsedPattern(domain, p)
		}
	}
}

// Benchmark exact match
func BenchmarkMatchDomainPattern_Exact(b *testing.B) {
	domain := "www.example.com"
	pattern := ParsePattern("DOMAIN,www.example.com")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		MatchParsedPattern(domain, pattern)
	}
}

// Benchmark suffix match
func BenchmarkMatchDomainPattern_Suffix(b *testing.B) {
	domain := "www.example.com"
	pattern := ParsePattern("DOMAIN-SUFFIX,example.com")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		MatchParsedPattern(domain, pattern)
	}
}

// Benchmark keyword match
func BenchmarkMatchDomainPattern_Keyword(b *testing.B) {
	domain := "www.example.com"
	pattern := ParsePattern("DOMAIN-KEYWORD,example")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		MatchParsedPattern(domain, pattern)
	}
}

// Benchmark cache hit rate
func BenchmarkParsePattern_CacheHit(b *testing.B) {
	pattern := "DOMAIN-SUFFIX,example.com"

	// First call to populate cache
	ParsePattern(pattern)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ParsePattern(pattern)
	}
}

// Benchmark cache miss
func BenchmarkParsePattern_CacheMiss(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Use unique pattern each time to force cache miss
		pattern := "DOMAIN-SUFFIX,example" + string(rune(i%1000)) + ".com"
		ParsePattern(pattern)
	}
}

// Benchmark matchDomainWithType (optimized version)
func BenchmarkMatchDomainWithType_Optimized(b *testing.B) {
	domain := "www.example.com"
	pattern := "example.com"
	matchType := "DOMAIN-SUFFIX"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		matchDomainWithType(domain, pattern, matchType)
	}
}
