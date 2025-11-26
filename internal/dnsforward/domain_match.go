package dnsforward

import (
	"strings"
	"sync"
)

// ParsedPattern represents a pre-processed domain pattern for efficient matching.
type ParsedPattern struct {
	MatchType string // "DOMAIN", "DOMAIN-SUFFIX", "DOMAIN-KEYWORD", or ""
	Pattern   string // The pattern in lowercase
	Original  string // Original pattern for reference
}

// patternCache caches parsed patterns to avoid repeated string operations.
var patternCache = struct {
	sync.RWMutex
	cache map[string]*ParsedPattern
}{
	cache: make(map[string]*ParsedPattern),
}

// ParsePattern pre-processes a pattern for efficient matching.
// This should be called once when loading rules, not on every match.
func ParsePattern(pattern string) *ParsedPattern {
	// Check cache first
	patternCache.RLock()
	if cached, exists := patternCache.cache[pattern]; exists {
		patternCache.RUnlock()
		return cached
	}
	patternCache.RUnlock()

	// Parse the pattern
	parsed := &ParsedPattern{
		Original: pattern,
	}

	patternUpper := strings.ToUpper(pattern)

	if strings.HasPrefix(patternUpper, "DOMAIN,") {
		parsed.MatchType = "DOMAIN"
		parsed.Pattern = strings.ToLower(pattern[7:])
	} else if strings.HasPrefix(patternUpper, "DOMAIN-SUFFIX,") {
		parsed.MatchType = "DOMAIN-SUFFIX"
		parsed.Pattern = strings.ToLower(pattern[14:])
	} else if strings.HasPrefix(patternUpper, "DOMAIN-KEYWORD,") {
		parsed.MatchType = "DOMAIN-KEYWORD"
		parsed.Pattern = strings.ToLower(pattern[15:])
	} else {
		// Default: exact match
		parsed.MatchType = ""
		parsed.Pattern = strings.ToLower(pattern)
	}

	// Cache the result
	patternCache.Lock()
	patternCache.cache[pattern] = parsed
	patternCache.Unlock()

	return parsed
}

// MatchParsedPattern matches a domain against a pre-parsed pattern.
// This is the optimized version that should be used in hot paths.
func MatchParsedPattern(domain string, parsed *ParsedPattern) bool {
	// Assume domain is already lowercase (caller's responsibility)
	// If not, uncomment the next line:
	// domain = strings.ToLower(domain)

	switch parsed.MatchType {
	case "DOMAIN":
		return domain == parsed.Pattern

	case "DOMAIN-SUFFIX":
		return domain == parsed.Pattern || strings.HasSuffix(domain, "."+parsed.Pattern)

	case "DOMAIN-KEYWORD":
		return strings.Contains(domain, parsed.Pattern)

	default:
		// Default: exact match
		return domain == parsed.Pattern
	}
}

// matchDomainPattern checks if a domain matches a pattern.
// Supports three types of patterns:
// - DOMAIN,example.com - exact match
// - DOMAIN-SUFFIX,example.com - suffix match (*.example.com)
// - DOMAIN-KEYWORD,example - keyword match (contains)
//
// This function is used by both DNS routing and upstream group selection.
// Note: This is the legacy version. For better performance, use ParsePattern + MatchParsedPattern.
func matchDomainPattern(domain, pattern string) bool {
	domain = strings.ToLower(domain)
	parsed := ParsePattern(pattern)
	return MatchParsedPattern(domain, parsed)
}

// matchDomainWithType checks if a domain matches a pattern with explicit match type.
// This is used for custom domain rules where match type is stored separately.
func matchDomainWithType(domain, pattern, matchType string) bool {
	domain = strings.ToLower(domain)
	pattern = strings.ToLower(pattern)

	switch matchType {
	case "DOMAIN":
		// Exact match
		return domain == pattern

	case "DOMAIN-SUFFIX":
		// Suffix match (including exact match)
		return domain == pattern || strings.HasSuffix(domain, "."+pattern)

	case "DOMAIN-KEYWORD":
		// Keyword match
		return strings.Contains(domain, pattern)

	default:
		// Default to exact match
		return domain == pattern
	}
}
