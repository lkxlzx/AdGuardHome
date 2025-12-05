package dnsrouting

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"io"
	"strings"

	"github.com/AdguardTeam/golibs/errors"
	"gopkg.in/yaml.v3"
)

// RuleFormat represents the format of the rule file.
type RuleFormat int

const (
	// RuleFormatUnknown is the unknown format.
	RuleFormatUnknown RuleFormat = iota
	// RuleFormatClash is the Clash rule format.
	RuleFormatClash
	// RuleFormatGFWList is the GFWList format.
	RuleFormatGFWList
	// RuleFormatAdGuard is the AdGuard Home format.
	RuleFormatAdGuard
)

// ParsedRule represents a parsed domain rule.
type ParsedRule struct {
	// Domain is the domain name.
	Domain string
	// MatchType is the type of matching (DOMAIN, DOMAIN-SUFFIX, DOMAIN-KEYWORD).
	MatchType string
}

// ParseResult contains the result of parsing a rule file.
type ParseResult struct {
	// Rules is the list of parsed rules.
	Rules []ParsedRule
	// Format is the detected format of the rule file.
	Format RuleFormat
	// TotalLines is the total number of lines processed.
	TotalLines int
	// ValidRules is the number of valid rules parsed.
	ValidRules int
}

// Parser is the interface for rule file parsers.
type Parser interface {
	// Parse parses the rule file and returns the result.
	Parse(r io.Reader) (*ParseResult, error)
	// DetectFormat detects the format of the rule file.
	DetectFormat(r io.Reader) (RuleFormat, error)
}

// parser is the default implementation of Parser.
type parser struct{}

// NewParser creates a new rule file parser.
func NewParser() Parser {
	return &parser{}
}

// Parse parses the rule file and returns the result.
func (p *parser) Parse(r io.Reader) (*ParseResult, error) {
	// Read all content first to detect format
	content, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("reading content: %w", err)
	}

	// Detect format
	format := p.detectFormatFromContent(content)

	// Parse based on format
	switch format {
	case RuleFormatClash:
		return p.parseClash(content)
	case RuleFormatGFWList:
		return p.parseGFWList(content)
	case RuleFormatAdGuard:
		return p.parseAdGuard(content)
	default:
		return nil, errors.Error("unknown rule format")
	}
}

// DetectFormat detects the format of the rule file.
func (p *parser) DetectFormat(r io.Reader) (RuleFormat, error) {
	content, err := io.ReadAll(r)
	if err != nil {
		return RuleFormatUnknown, fmt.Errorf("reading content: %w", err)
	}

	return p.detectFormatFromContent(content), nil
}

// detectFormatFromContent detects the format from content.
func (p *parser) detectFormatFromContent(content []byte) RuleFormat {
	contentStr := string(content)

	// Check for Clash format (YAML with payload field)
	if strings.Contains(contentStr, "payload:") || strings.HasPrefix(contentStr, "# Clash") {
		return RuleFormatClash
	}

	// Check for GFWList format (base64 encoded or starts with [AutoProxy])
	if strings.HasPrefix(contentStr, "[AutoProxy") {
		return RuleFormatGFWList
	}

	// Try to decode as base64 (GFWList is often base64 encoded)
	if _, err := base64.StdEncoding.DecodeString(strings.TrimSpace(contentStr)); err == nil {
		if len(content) > 100 { // Reasonable size for base64 encoded list
			return RuleFormatGFWList
		}
	}

	// Check for AdGuard format (starts with ! or ||)
	lines := strings.Split(contentStr, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "!") || strings.HasPrefix(line, "||") {
			return RuleFormatAdGuard
		}
		break
	}

	return RuleFormatUnknown
}

// parseClash parses Clash format rules.
func (p *parser) parseClash(content []byte) (*ParseResult, error) {
	var clashConfig struct {
		Payload []string `yaml:"payload"`
	}

	err := yaml.Unmarshal(content, &clashConfig)
	if err != nil {
		return nil, fmt.Errorf("parsing clash yaml: %w", err)
	}

	result := &ParseResult{
		Format:     RuleFormatClash,
		Rules:      make([]ParsedRule, 0, len(clashConfig.Payload)),
		TotalLines: len(clashConfig.Payload),
	}

	for _, line := range clashConfig.Payload {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		rule := p.parseClashRule(line)
		if rule != nil {
			result.Rules = append(result.Rules, *rule)
			result.ValidRules++
		}
	}

	return result, nil
}

// parseClashRule parses a single Clash rule line.
func (p *parser) parseClashRule(line string) *ParsedRule {
	// Clash rule format: DOMAIN,example.com or DOMAIN-SUFFIX,example.com
	parts := strings.Split(line, ",")
	if len(parts) < 2 {
		return nil
	}

	ruleType := strings.TrimSpace(parts[0])
	domain := strings.TrimSpace(parts[1])

	if domain == "" {
		return nil
	}

	var matchType string
	switch ruleType {
	case "DOMAIN":
		matchType = "DOMAIN"
	case "DOMAIN-SUFFIX":
		matchType = "DOMAIN-SUFFIX"
	case "DOMAIN-KEYWORD":
		matchType = "DOMAIN-KEYWORD"
	default:
		return nil
	}

	return &ParsedRule{
		Domain:    domain,
		MatchType: matchType,
	}
}

// parseGFWList parses GFWList format rules.
func (p *parser) parseGFWList(content []byte) (*ParseResult, error) {
	// Try to decode base64 first
	decoded, err := base64.StdEncoding.DecodeString(string(content))
	if err != nil {
		// Not base64, use original content
		decoded = content
	}

	result := &ParseResult{
		Format: RuleFormatGFWList,
		Rules:  make([]ParsedRule, 0),
	}

	scanner := bufio.NewScanner(strings.NewReader(string(decoded)))
	for scanner.Scan() {
		result.TotalLines++
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "!") || strings.HasPrefix(line, "[") {
			continue
		}

		rule := p.parseGFWListRule(line)
		if rule != nil {
			result.Rules = append(result.Rules, *rule)
			result.ValidRules++
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scanning lines: %w", err)
	}

	return result, nil
}

// parseGFWListRule parses a single GFWList rule line.
func (p *parser) parseGFWListRule(line string) *ParsedRule {
	// Remove common prefixes
	line = strings.TrimPrefix(line, "||")
	line = strings.TrimPrefix(line, "|")
	line = strings.TrimPrefix(line, ".")
	line = strings.TrimSuffix(line, "^")
	line = strings.TrimSuffix(line, "/")

	// Skip if contains wildcards or regex
	if strings.Contains(line, "*") || strings.Contains(line, "?") {
		return nil
	}

	// Extract domain from URL
	if strings.Contains(line, "://") {
		parts := strings.Split(line, "://")
		if len(parts) > 1 {
			line = parts[1]
		}
	}

	// Remove path
	if idx := strings.Index(line, "/"); idx != -1 {
		line = line[:idx]
	}

	// Remove port
	if idx := strings.Index(line, ":"); idx != -1 {
		line = line[:idx]
	}

	line = strings.TrimSpace(line)
	if line == "" {
		return nil
	}

	// GFWList rules are typically domain suffixes
	return &ParsedRule{
		Domain:    line,
		MatchType: "DOMAIN-SUFFIX",
	}
}

// parseAdGuard parses AdGuard Home format rules.
func (p *parser) parseAdGuard(content []byte) (*ParseResult, error) {
	result := &ParseResult{
		Format: RuleFormatAdGuard,
		Rules:  make([]ParsedRule, 0),
	}

	scanner := bufio.NewScanner(strings.NewReader(string(content)))
	for scanner.Scan() {
		result.TotalLines++
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "!") || strings.HasPrefix(line, "#") {
			continue
		}

		rule := p.parseAdGuardRule(line)
		if rule != nil {
			result.Rules = append(result.Rules, *rule)
			result.ValidRules++
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scanning lines: %w", err)
	}

	return result, nil
}

// parseAdGuardRule parses a single AdGuard rule line.
func (p *parser) parseAdGuardRule(line string) *ParsedRule {
	// AdGuard format: ||example.com^ or |example.com^
	line = strings.TrimPrefix(line, "||")
	line = strings.TrimPrefix(line, "|")
	line = strings.TrimSuffix(line, "^")
	line = strings.TrimSuffix(line, "$")

	// Remove modifiers
	if idx := strings.Index(line, "$"); idx != -1 {
		line = line[:idx]
	}

	line = strings.TrimSpace(line)
	if line == "" {
		return nil
	}

	// AdGuard rules are typically domain suffixes
	return &ParsedRule{
		Domain:    line,
		MatchType: "DOMAIN-SUFFIX",
	}
}

// ConvertToAdGuardFormat converts parsed rules to AdGuard Home format.
func ConvertToAdGuardFormat(rules []ParsedRule) []string {
	result := make([]string, 0, len(rules))
	seen := make(map[string]bool)

	for _, rule := range rules {
		var line string
		switch rule.MatchType {
		case "DOMAIN":
			line = fmt.Sprintf("|%s^", rule.Domain)
		case "DOMAIN-SUFFIX":
			line = fmt.Sprintf("||%s^", rule.Domain)
		case "DOMAIN-KEYWORD":
			// AdGuard doesn't have direct keyword support, use as suffix
			line = fmt.Sprintf("||%s^", rule.Domain)
		default:
			continue
		}

		// Deduplicate
		if !seen[line] {
			seen[line] = true
			result = append(result, line)
		}
	}

	return result
}
