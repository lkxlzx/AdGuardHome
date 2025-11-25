package dnsforward

import "strings"

// matchDomainPattern checks if a domain matches a pattern.
// Supports three types of patterns:
// - DOMAIN,example.com - exact match
// - DOMAIN-SUFFIX,example.com - suffix match (*.example.com)
// - DOMAIN-KEYWORD,example - keyword match (contains)
//
// This function is used by both DNS routing and upstream group selection.
func matchDomainPattern(domain, pattern string) bool {
	domain = strings.ToLower(domain)
	pattern = strings.ToLower(pattern)

	// Parse Clash-style patterns
	if strings.HasPrefix(pattern, "DOMAIN,") {
		// Exact match
		targetDomain := strings.TrimPrefix(pattern, "DOMAIN,")
		return domain == targetDomain
	}

	if strings.HasPrefix(pattern, "DOMAIN-SUFFIX,") {
		// Suffix match
		suffix := strings.TrimPrefix(pattern, "DOMAIN-SUFFIX,")
		return domain == suffix || strings.HasSuffix(domain, "."+suffix)
	}

	if strings.HasPrefix(pattern, "DOMAIN-KEYWORD,") {
		// Keyword match
		keyword := strings.TrimPrefix(pattern, "DOMAIN-KEYWORD,")
		return strings.Contains(domain, keyword)
	}

	// Default: treat as exact match
	return domain == pattern
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
