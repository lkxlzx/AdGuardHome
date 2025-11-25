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
	
	// Parse Clash-style patterns (check before lowercasing to preserve prefix)
	patternUpper := strings.ToUpper(pattern)
	
	if strings.HasPrefix(patternUpper, "DOMAIN,") {
		// Exact match
		targetDomain := strings.ToLower(pattern[7:]) // Skip "DOMAIN,"
		return domain == targetDomain
	}

	if strings.HasPrefix(patternUpper, "DOMAIN-SUFFIX,") {
		// Suffix match
		suffix := strings.ToLower(pattern[14:]) // Skip "DOMAIN-SUFFIX,"
		return domain == suffix || strings.HasSuffix(domain, "."+suffix)
	}

	if strings.HasPrefix(patternUpper, "DOMAIN-KEYWORD,") {
		// Keyword match
		keyword := strings.ToLower(pattern[15:]) // Skip "DOMAIN-KEYWORD,"
		return strings.Contains(domain, keyword)
	}

	// Default: treat as exact match
	pattern = strings.ToLower(pattern)
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
