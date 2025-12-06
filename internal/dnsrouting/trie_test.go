package dnsrouting

import (
	"testing"
)

func TestDomainTrie_Basic(t *testing.T) {
	trie := NewDomainTrie()

	// Insert some domains
	trie.Insert("example.com", "group1", 1)
	trie.Insert("test.com", "group2", 1)

	// Test exact match
	group, found := trie.Search("example.com")
	if !found || group != "group1" {
		t.Errorf("Expected group1 for example.com, got %s, found=%v", group, found)
	}

	// Test suffix match
	group, found = trie.Search("www.example.com")
	if !found || group != "group1" {
		t.Errorf("Expected group1 for www.example.com, got %s, found=%v", group, found)
	}

	// Test no match
	_, found = trie.Search("notfound.com")
	if found {
		t.Error("Expected no match for notfound.com")
	}
}

func TestDomainTrie_SuffixMatching(t *testing.T) {
	trie := NewDomainTrie()

	trie.Insert("example.com", "group1", 1)

	tests := []struct {
		domain   string
		expected string
		found    bool
	}{
		{"example.com", "group1", true},
		{"www.example.com", "group1", true},
		{"api.example.com", "group1", true},
		{"sub.api.example.com", "group1", true},
		{"example.org", "", false},
		{"notexample.com", "", false},
	}

	for _, tt := range tests {
		group, found := trie.Search(tt.domain)
		if found != tt.found {
			t.Errorf("Domain %s: expected found=%v, got %v", tt.domain, tt.found, found)
		}
		if found && group != tt.expected {
			t.Errorf("Domain %s: expected group=%s, got %s", tt.domain, tt.expected, group)
		}
	}
}

func TestDomainTrie_Priority(t *testing.T) {
	trie := NewDomainTrie()

	// Insert with different priorities
	trie.Insert("example.com", "group1", 10)
	trie.Insert("api.example.com", "group2", 5) // Higher priority (lower number)

	// Should match the more specific rule with higher priority
	group, found := trie.Search("api.example.com")
	if !found {
		t.Fatal("Expected to find match for api.example.com")
	}
	if group != "group2" {
		t.Errorf("Expected group2 (higher priority), got %s", group)
	}
}

func TestCompiledRules_ExactMatch(t *testing.T) {
	compiled := NewCompiledRules()

	compiled.AddRule(Rule{
		Domain:        "example.com",
		MatchType:     MatchTypeDomain,
		UpstreamGroup: "group1",
		Enabled:       true,
	})

	// Test exact match
	group, found := compiled.Match("example.com")
	if !found || group != "group1" {
		t.Errorf("Expected group1, got %s, found=%v", group, found)
	}

	// Should not match subdomain
	_, found = compiled.Match("www.example.com")
	if found {
		t.Error("Exact match should not match subdomain")
	}
}

func TestCompiledRules_SuffixMatch(t *testing.T) {
	compiled := NewCompiledRules()

	compiled.AddRule(Rule{
		Domain:        "example.com",
		MatchType:     MatchTypeDomainSuffix,
		UpstreamGroup: "group1",
		Priority:      1,
		Enabled:       true,
	})

	tests := []struct {
		domain   string
		expected string
		found    bool
	}{
		{"example.com", "group1", true},
		{"www.example.com", "group1", true},
		{"api.example.com", "group1", true},
		{"example.org", "", false},
	}

	for _, tt := range tests {
		group, found := compiled.Match(tt.domain)
		if found != tt.found {
			t.Errorf("Domain %s: expected found=%v, got %v", tt.domain, tt.found, found)
		}
		if found && group != tt.expected {
			t.Errorf("Domain %s: expected group=%s, got %s", tt.domain, tt.expected, group)
		}
	}
}

func TestCompiledRules_KeywordMatch(t *testing.T) {
	compiled := NewCompiledRules()

	compiled.AddRule(Rule{
		Domain:        "google",
		MatchType:     MatchTypeDomainKeyword,
		UpstreamGroup: "group1",
		Enabled:       true,
	})

	tests := []struct {
		domain   string
		expected string
		found    bool
	}{
		{"google.com", "group1", true},
		{"www.google.com", "group1", true},
		{"google.co.uk", "group1", true},
		{"mygooglesite.com", "group1", true},
		{"facebook.com", "", false},
	}

	for _, tt := range tests {
		group, found := compiled.Match(tt.domain)
		if found != tt.found {
			t.Errorf("Domain %s: expected found=%v, got %v", tt.domain, tt.found, found)
		}
		if found && group != tt.expected {
			t.Errorf("Domain %s: expected group=%s, got %s", tt.domain, tt.expected, group)
		}
	}
}

func TestCompiledRules_Mixed(t *testing.T) {
	compiled := NewCompiledRules()

	// Add different types of rules
	compiled.AddRule(Rule{
		Domain:        "exact.com",
		MatchType:     MatchTypeDomain,
		UpstreamGroup: "exact-group",
		Enabled:       true,
	})

	compiled.AddRule(Rule{
		Domain:        "suffix.com",
		MatchType:     MatchTypeDomainSuffix,
		UpstreamGroup: "suffix-group",
		Priority:      1,
		Enabled:       true,
	})

	compiled.AddRule(Rule{
		Domain:        "keyword",
		MatchType:     MatchTypeDomainKeyword,
		UpstreamGroup: "keyword-group",
		Enabled:       true,
	})

	tests := []struct {
		domain   string
		expected string
		found    bool
	}{
		{"exact.com", "exact-group", true},
		{"www.exact.com", "suffix-group", false}, // No suffix rule for exact.com
		{"suffix.com", "suffix-group", true},
		{"www.suffix.com", "suffix-group", true},
		{"keyword.com", "keyword-group", true},
		{"mykeywordsite.com", "keyword-group", true},
	}

	for _, tt := range tests {
		group, found := compiled.Match(tt.domain)
		if found != tt.found {
			t.Errorf("Domain %s: expected found=%v, got %v", tt.domain, tt.found, found)
		}
		if found && group != tt.expected {
			t.Errorf("Domain %s: expected group=%s, got %s", tt.domain, tt.expected, group)
		}
	}
}

func BenchmarkDomainTrie_Search(b *testing.B) {
	trie := NewDomainTrie()

	// Pre-populate with many domains
	domains := []string{
		"google.com", "facebook.com", "twitter.com", "github.com",
		"stackoverflow.com", "reddit.com", "amazon.com", "youtube.com",
	}

	for _, domain := range domains {
		trie.Insert(domain, "group1", 1)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		trie.Search("www.google.com")
	}
}

func BenchmarkCompiledRules_Match(b *testing.B) {
	compiled := NewCompiledRules()

	// Add various rules
	domains := []string{
		"google.com", "facebook.com", "twitter.com", "github.com",
		"stackoverflow.com", "reddit.com", "amazon.com", "youtube.com",
	}

	for _, domain := range domains {
		compiled.AddRule(Rule{
			Domain:        domain,
			MatchType:     MatchTypeDomainSuffix,
			UpstreamGroup: "group1",
			Priority:      1,
			Enabled:       true,
		})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		compiled.Match("www.google.com")
	}
}
