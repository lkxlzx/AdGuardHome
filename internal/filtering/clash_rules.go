package filtering

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// ClashRule represents a single rule from Clash format
type ClashRule struct {
	Type  string // DOMAIN, DOMAIN-SUFFIX, DOMAIN-KEYWORD, IP-CIDR, etc.
	Value string // The actual domain or IP
}

// ClashRuleFile represents the structure of a Clash rule file
type ClashRuleFile struct {
	Payload []string `yaml:"payload"`
}

// ClashRuleStats contains statistics about parsed rules
type ClashRuleStats struct {
	TotalRules      int `json:"total_rules"`
	DomainRules     int `json:"domain_rules"`
	DomainSuffix    int `json:"domain_suffix"`
	DomainKeyword   int `json:"domain_keyword"`
	IPRules         int `json:"ip_rules"`
	OtherRules      int `json:"other_rules"`
	ValidDomains    int `json:"valid_domains"` // Total valid domain rules
}

const (
	// clashRuleDownloadTimeout is the timeout for downloading Clash rule files
	clashRuleDownloadTimeout = 30 * time.Second
	// maxClashRuleFileSize is the maximum size of a Clash rule file (10MB)
	maxClashRuleFileSize = 10 * 1024 * 1024
	// maxRetries is the maximum number of retry attempts for downloading rules
	maxRetries = 3
	// retryDelay is the delay between retry attempts
	retryDelay = 2 * time.Second
)

// downloadWithRetry downloads content from URL with retry mechanism
func downloadWithRetry(url string, maxRetries int) ([]byte, error) {
	client := &http.Client{
		Timeout: clashRuleDownloadTimeout,
	}

	var lastErr error
	for attempt := 1; attempt <= maxRetries; attempt++ {
		resp, err := client.Get(url)
		if err != nil {
			lastErr = fmt.Errorf("attempt %d/%d failed: %w", attempt, maxRetries, err)
			if attempt < maxRetries {
				time.Sleep(retryDelay)
				continue
			}
			return nil, lastErr
		}
		defer resp.Body.Close()

		// Check status code
		if resp.StatusCode != http.StatusOK {
			lastErr = fmt.Errorf("attempt %d/%d: unexpected status code %d", attempt, maxRetries, resp.StatusCode)
			if attempt < maxRetries {
				time.Sleep(retryDelay)
				continue
			}
			return nil, lastErr
		}

		// Check content length if available
		if resp.ContentLength > maxClashRuleFileSize {
			return nil, fmt.Errorf("file too large: %d bytes (max: %d)", resp.ContentLength, maxClashRuleFileSize)
		}

		// Read content with size limit
		limitedReader := io.LimitReader(resp.Body, maxClashRuleFileSize+1)
		content, err := io.ReadAll(limitedReader)
		if err != nil {
			lastErr = fmt.Errorf("attempt %d/%d: reading response: %w", attempt, maxRetries, err)
			if attempt < maxRetries {
				time.Sleep(retryDelay)
				continue
			}
			return nil, lastErr
		}

		// Check if content exceeds size limit
		if len(content) > maxClashRuleFileSize {
			return nil, fmt.Errorf("file too large: exceeds %d bytes", maxClashRuleFileSize)
		}

		// Success
		return content, nil
	}

	return nil, lastErr
}

// ParseClashRules parses a Clash rule file from a URL and returns domain rules
func ParseClashRules(url string) ([]string, *ClashRuleStats, error) {
	// Download the rule file with retry mechanism
	content, err := downloadWithRetry(url, maxRetries)
	if err != nil {
		return nil, nil, fmt.Errorf("downloading rules from %s: %w", url, err)
	}

	// Parse YAML
	var ruleFile ClashRuleFile
	err = yaml.Unmarshal(content, &ruleFile)
	if err != nil {
		return nil, nil, fmt.Errorf("parsing YAML: %w", err)
	}

	// Parse and filter rules
	domains := make([]string, 0)
	stats := &ClashRuleStats{}

	for _, rule := range ruleFile.Payload {
		stats.TotalRules++

		// Parse rule format: "TYPE,value"
		parts := strings.SplitN(rule, ",", 2)
		if len(parts) != 2 {
			stats.OtherRules++
			continue
		}

		ruleType := strings.TrimSpace(parts[0])
		ruleValue := strings.TrimSpace(parts[1])

		switch ruleType {
		case "DOMAIN":
			stats.DomainRules++
			stats.ValidDomains++
			domains = append(domains, ruleValue)

		case "DOMAIN-SUFFIX":
			stats.DomainSuffix++
			stats.ValidDomains++
			// Convert to AdGuard Home format: ||example.com^
			domains = append(domains, "||"+ruleValue+"^")

		case "DOMAIN-KEYWORD":
			stats.DomainKeyword++
			stats.ValidDomains++
			// Convert to AdGuard Home format: *keyword*
			domains = append(domains, "*"+ruleValue+"*")

		case "IP-CIDR", "IP-CIDR6":
			stats.IPRules++
			// Skip IP rules

		default:
			stats.OtherRules++
		}
	}

	return domains, stats, nil
}

// ParseClashRulesFromReader parses Clash rules from an io.Reader
func ParseClashRulesFromReader(reader io.Reader) ([]string, *ClashRuleStats, error) {
	content, err := io.ReadAll(reader)
	if err != nil {
		return nil, nil, fmt.Errorf("reading content: %w", err)
	}

	var ruleFile ClashRuleFile
	err = yaml.Unmarshal(content, &ruleFile)
	if err != nil {
		return nil, nil, fmt.Errorf("parsing YAML: %w", err)
	}

	domains := make([]string, 0)
	stats := &ClashRuleStats{}

	for _, rule := range ruleFile.Payload {
		stats.TotalRules++

		parts := strings.SplitN(rule, ",", 2)
		if len(parts) != 2 {
			stats.OtherRules++
			continue
		}

		ruleType := strings.TrimSpace(parts[0])
		ruleValue := strings.TrimSpace(parts[1])

		switch ruleType {
		case "DOMAIN":
			stats.DomainRules++
			stats.ValidDomains++
			domains = append(domains, ruleValue)

		case "DOMAIN-SUFFIX":
			stats.DomainSuffix++
			stats.ValidDomains++
			domains = append(domains, "||"+ruleValue+"^")

		case "DOMAIN-KEYWORD":
			stats.DomainKeyword++
			stats.ValidDomains++
			domains = append(domains, "*"+ruleValue+"*")

		case "IP-CIDR", "IP-CIDR6":
			stats.IPRules++

		default:
			stats.OtherRules++
		}
	}

	return domains, stats, nil
}

// ValidateClashRuleURL validates and fetches statistics for a Clash rule URL
func ValidateClashRuleURL(url string) (*ClashRuleStats, error) {
	_, stats, err := ParseClashRules(url)
	return stats, err
}

// ConvertClashRuleToAdGuardFormat converts a single Clash rule to AdGuard Home format
func ConvertClashRuleToAdGuardFormat(rule string) (string, bool) {
	parts := strings.SplitN(rule, ",", 2)
	if len(parts) != 2 {
		return "", false
	}

	ruleType := strings.TrimSpace(parts[0])
	ruleValue := strings.TrimSpace(parts[1])

	switch ruleType {
	case "DOMAIN":
		return ruleValue, true

	case "DOMAIN-SUFFIX":
		return "||" + ruleValue + "^", true

	case "DOMAIN-KEYWORD":
		return "*" + ruleValue + "*", true

	case "IP-CIDR", "IP-CIDR6":
		// Skip IP rules
		return "", false

	default:
		return "", false
	}
}

// ExtractDomainRulesFromText extracts domain rules from plain text (line by line)
func ExtractDomainRulesFromText(text string) ([]string, *ClashRuleStats, error) {
	scanner := bufio.NewScanner(strings.NewReader(text))
	domains := make([]string, 0)
	stats := &ClashRuleStats{}

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		
		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		stats.TotalRules++

		// Try to parse as Clash format
		if strings.Contains(line, ",") {
			domain, ok := ConvertClashRuleToAdGuardFormat(line)
			if ok {
				domains = append(domains, domain)
				stats.ValidDomains++
				
				// Update specific counters
				if strings.HasPrefix(line, "DOMAIN-SUFFIX") {
					stats.DomainSuffix++
				} else if strings.HasPrefix(line, "DOMAIN-KEYWORD") {
					stats.DomainKeyword++
				} else if strings.HasPrefix(line, "DOMAIN,") {
					stats.DomainRules++
				}
			} else {
				if strings.HasPrefix(line, "IP-CIDR") {
					stats.IPRules++
				} else {
					stats.OtherRules++
				}
			}
		} else {
			// Treat as plain domain
			domains = append(domains, line)
			stats.ValidDomains++
			stats.DomainRules++
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, nil, fmt.Errorf("scanning text: %w", err)
	}

	return domains, stats, nil
}


// ProcessClashRuleFile processes a Clash rule file and writes only domain rules to the output.
// This function is used when downloading Clash rule files to filter out IP rules.
func ProcessClashRuleFile(input io.Reader, output io.Writer) (*ClashRuleStats, error) {
	domains, stats, err := ParseClashRulesFromReader(input)
	if err != nil {
		return nil, err
	}

	// Write domain rules to output (one per line)
	for _, domain := range domains {
		_, err := fmt.Fprintln(output, domain)
		if err != nil {
			return nil, fmt.Errorf("writing domain: %w", err)
		}
	}

	return stats, nil
}

// IsClashRuleURL checks if a URL points to a Clash rule file.
// It checks for common Clash rule URL patterns.
func IsClashRuleURL(url string) bool {
	// Check if URL contains common Clash rule indicators
	return strings.Contains(url, "/Clash/") ||
		strings.Contains(url, "/clash/") ||
		strings.HasSuffix(url, ".yaml") ||
		strings.HasSuffix(url, ".yml")
}
