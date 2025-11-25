package dnsforward

import "testing"

func TestMatchDomainPattern(t *testing.T) {
	tests := []struct {
		name    string
		domain  string
		pattern string
		want    bool
	}{
		// Exact match tests
		{
			name:    "exact_match_simple",
			domain:  "example.com",
			pattern: "DOMAIN,example.com",
			want:    true,
		},
		{
			name:    "exact_match_case_insensitive",
			domain:  "Example.COM",
			pattern: "DOMAIN,example.com",
			want:    true,
		},
		{
			name:    "exact_match_no_match",
			domain:  "test.example.com",
			pattern: "DOMAIN,example.com",
			want:    false,
		},

		// Suffix match tests
		{
			name:    "suffix_match_exact",
			domain:  "example.com",
			pattern: "DOMAIN-SUFFIX,example.com",
			want:    true,
		},
		{
			name:    "suffix_match_subdomain",
			domain:  "test.example.com",
			pattern: "DOMAIN-SUFFIX,example.com",
			want:    true,
		},
		{
			name:    "suffix_match_deep_subdomain",
			domain:  "a.b.c.example.com",
			pattern: "DOMAIN-SUFFIX,example.com",
			want:    true,
		},
		{
			name:    "suffix_match_no_match",
			domain:  "notexample.com",
			pattern: "DOMAIN-SUFFIX,example.com",
			want:    false,
		},
		{
			name:    "suffix_match_case_insensitive",
			domain:  "Test.Example.COM",
			pattern: "DOMAIN-SUFFIX,example.com",
			want:    true,
		},

		// Keyword match tests
		{
			name:    "keyword_match_contains",
			domain:  "test.example.com",
			pattern: "DOMAIN-KEYWORD,example",
			want:    true,
		},
		{
			name:    "keyword_match_at_start",
			domain:  "example.com",
			pattern: "DOMAIN-KEYWORD,example",
			want:    true,
		},
		{
			name:    "keyword_match_at_end",
			domain:  "test.example",
			pattern: "DOMAIN-KEYWORD,example",
			want:    true,
		},
		{
			name:    "keyword_match_no_match",
			domain:  "test.com",
			pattern: "DOMAIN-KEYWORD,example",
			want:    false,
		},
		{
			name:    "keyword_match_case_insensitive",
			domain:  "Test.EXAMPLE.com",
			pattern: "DOMAIN-KEYWORD,example",
			want:    true,
		},

		// Default (exact match without prefix)
		{
			name:    "default_exact_match",
			domain:  "example.com",
			pattern: "example.com",
			want:    true,
		},
		{
			name:    "default_no_match",
			domain:  "test.example.com",
			pattern: "example.com",
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := matchDomainPattern(tt.domain, tt.pattern)
			if got != tt.want {
				t.Errorf("matchDomainPattern(%q, %q) = %v, want %v",
					tt.domain, tt.pattern, got, tt.want)
			}
		})
	}
}

func TestMatchDomainWithType(t *testing.T) {
	tests := []struct {
		name      string
		domain    string
		pattern   string
		matchType string
		want      bool
	}{
		// DOMAIN type tests
		{
			name:      "domain_exact_match",
			domain:    "example.com",
			pattern:   "example.com",
			matchType: "DOMAIN",
			want:      true,
		},
		{
			name:      "domain_no_match",
			domain:    "test.example.com",
			pattern:   "example.com",
			matchType: "DOMAIN",
			want:      false,
		},
		{
			name:      "domain_case_insensitive",
			domain:    "Example.COM",
			pattern:   "example.com",
			matchType: "DOMAIN",
			want:      true,
		},

		// DOMAIN-SUFFIX type tests
		{
			name:      "suffix_exact_match",
			domain:    "example.com",
			pattern:   "example.com",
			matchType: "DOMAIN-SUFFIX",
			want:      true,
		},
		{
			name:      "suffix_subdomain_match",
			domain:    "test.example.com",
			pattern:   "example.com",
			matchType: "DOMAIN-SUFFIX",
			want:      true,
		},
		{
			name:      "suffix_no_match",
			domain:    "notexample.com",
			pattern:   "example.com",
			matchType: "DOMAIN-SUFFIX",
			want:      false,
		},

		// DOMAIN-KEYWORD type tests
		{
			name:      "keyword_contains",
			domain:    "test.example.com",
			pattern:   "example",
			matchType: "DOMAIN-KEYWORD",
			want:      true,
		},
		{
			name:      "keyword_no_match",
			domain:    "test.com",
			pattern:   "example",
			matchType: "DOMAIN-KEYWORD",
			want:      false,
		},

		// Unknown type (defaults to exact match)
		{
			name:      "unknown_type_exact_match",
			domain:    "example.com",
			pattern:   "example.com",
			matchType: "UNKNOWN",
			want:      true,
		},
		{
			name:      "unknown_type_no_match",
			domain:    "test.example.com",
			pattern:   "example.com",
			matchType: "UNKNOWN",
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := matchDomainWithType(tt.domain, tt.pattern, tt.matchType)
			if got != tt.want {
				t.Errorf("matchDomainWithType(%q, %q, %q) = %v, want %v",
					tt.domain, tt.pattern, tt.matchType, got, tt.want)
			}
		})
	}
}

// Benchmark tests
func BenchmarkMatchDomainPattern(b *testing.B) {
	domain := "test.example.com"
	pattern := "DOMAIN-SUFFIX,example.com"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		matchDomainPattern(domain, pattern)
	}
}

func BenchmarkMatchDomainWithType(b *testing.B) {
	domain := "test.example.com"
	pattern := "example.com"
	matchType := "DOMAIN-SUFFIX"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		matchDomainWithType(domain, pattern, matchType)
	}
}
