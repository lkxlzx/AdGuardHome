package filtering

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"io"
	"regexp"
	"strings"
)

// GFWListStats contains statistics about parsed GFWList rules
type GFWListStats struct {
	TotalRules   int `json:"total_rules"`
	ValidDomains int `json:"valid_domains"`
	SkippedRules int `json:"skipped_rules"`
	Comments     int `json:"comments"`
}

// IsGFWListURL checks if a URL points to a GFWList file.
func IsGFWListURL(url string) bool {
	return strings.Contains(url, "gfwlist") ||
		strings.Contains(url, "GFWList") ||
		strings.Contains(url, "gfw.txt") ||
		strings.Contains(url, "gfwlist.txt")
}

// ProcessGFWListFile processes a GFWList file and writes only domain rules to the output.
// GFWList format is base64 encoded and contains various rule types.
func ProcessGFWListFile(input io.Reader, output io.Writer) (*GFWListStats, error) {
	domains, stats, err := ParseGFWListFromReader(input)
	if err != nil {
		return nil, err
	}

	// Write domain rules to output (one per line)
	// Use AdGuard Home format
	for _, domain := range domains {
		_, err := fmt.Fprintln(output, domain)
		if err != nil {
			return nil, fmt.Errorf("writing domain: %w", err)
		}
	}

	return stats, nil
}

// ParseGFWListFromReader parses GFWList rules from an io.Reader
func ParseGFWListFromReader(reader io.Reader) ([]string, *GFWListStats, error) {
	// Read all content
	content, err := io.ReadAll(reader)
	if err != nil {
		return nil, nil, fmt.Errorf("reading content: %w", err)
	}

	// Try to decode base64 (GFWList is typically base64 encoded)
	decoded, err := base64.StdEncoding.DecodeString(string(content))
	if err != nil {
		// If decode fails, assume it's already plain text
		decoded = content
	}

	return parseGFWListContent(string(decoded))
}

// parseGFWListContent parses the decoded GFWList content
func parseGFWListContent(content string) ([]string, *GFWListStats, error) {
	scanner := bufio.NewScanner(strings.NewReader(content))
	domainMap := make(map[string]bool) // Use map to deduplicate
	stats := &GFWListStats{}

	// Regex for extracting domains
	domainRegex := regexp.MustCompile(`^(?:\|\|)?([a-zA-Z0-9][-a-zA-Z0-9]*(?:\.[a-zA-Z0-9][-a-zA-Z0-9]*)+)`)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		stats.TotalRules++

		// Skip comments and empty lines
		if line == "" || strings.HasPrefix(line, "!") || strings.HasPrefix(line, "[") {
			if strings.HasPrefix(line, "!") || strings.HasPrefix(line, "[") {
				stats.Comments++
			}
			continue
		}

		// Extract domain from the rule
		domain := extractDomainFromGFWRule(line, domainRegex)
		if domain != "" {
			domainMap[domain] = true
		} else {
			stats.SkippedRules++
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, nil, fmt.Errorf("scanning content: %w", err)
	}

	// Convert map to slice and format for AdGuard Home
	domains := make([]string, 0, len(domainMap))
	for domain := range domainMap {
		// Use AdGuard Home format: ||domain^
		domains = append(domains, "||"+domain+"^")
		stats.ValidDomains++
	}

	return domains, stats, nil
}

// extractDomainFromGFWRule extracts a domain from a GFWList rule
func extractDomainFromGFWRule(rule string, domainRegex *regexp.Regexp) string {
	// Remove common prefixes
	rule = strings.TrimPrefix(rule, "||")
	rule = strings.TrimPrefix(rule, "|")
	rule = strings.TrimPrefix(rule, ".")
	rule = strings.TrimPrefix(rule, "http://")
	rule = strings.TrimPrefix(rule, "https://")
	rule = strings.TrimPrefix(rule, "@@||") // Whitelist marker
	rule = strings.TrimPrefix(rule, "@@")

	// Remove path and query
	if idx := strings.Index(rule, "/"); idx != -1 {
		rule = rule[:idx]
	}
	if idx := strings.Index(rule, "?"); idx != -1 {
		rule = rule[:idx]
	}
	if idx := strings.Index(rule, "^"); idx != -1 {
		rule = rule[:idx]
	}
	if idx := strings.Index(rule, "*"); idx != -1 {
		// Skip wildcard rules for now
		return ""
	}

	// Extract domain using regex
	matches := domainRegex.FindStringSubmatch(rule)
	if len(matches) > 1 {
		domain := matches[1]
		// Validate domain
		if isValidDomain(domain) {
			return domain
		}
	}

	// If regex didn't match, try direct validation
	if isValidDomain(rule) {
		return rule
	}

	return ""
}

// isValidDomain validates if a string is a valid domain name
func isValidDomain(domain string) bool {
	// Basic domain validation
	if len(domain) == 0 || len(domain) > 253 {
		return false
	}

	// Check for valid characters
	validDomain := regexp.MustCompile(`^[a-zA-Z0-9][-a-zA-Z0-9]*(\.[a-zA-Z0-9][-a-zA-Z0-9]*)+$`)
	return validDomain.MatchString(domain)
}
