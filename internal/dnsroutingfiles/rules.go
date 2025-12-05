package dnsroutingfiles

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"time"

	"github.com/AdguardTeam/AdGuardHome/internal/dnsrouting"
)

// convertToAdGuardFormat converts parsed rules to AdGuard format.
func convertToAdGuardFormat(rules []dnsrouting.ParsedRule) []byte {
	var buf bytes.Buffer
	for _, rule := range rules {
		// Convert to AdGuard format based on match type
		switch rule.MatchType {
		case "DOMAIN-SUFFIX":
			// DOMAIN-SUFFIX: example.com -> ||example.com^
			buf.WriteString("||")
			buf.WriteString(rule.Domain)
			buf.WriteString("^\n")
		case "DOMAIN":
			// DOMAIN: example.com -> |example.com|
			buf.WriteString("|")
			buf.WriteString(rule.Domain)
			buf.WriteString("|\n")
		case "DOMAIN-KEYWORD":
			// DOMAIN-KEYWORD: keyword -> keyword (as substring match)
			buf.WriteString(rule.Domain)
			buf.WriteString("\n")
		default:
			// Default: treat as DOMAIN-SUFFIX
			buf.WriteString("||")
			buf.WriteString(rule.Domain)
			buf.WriteString("^\n")
		}
	}
	return buf.Bytes()
}

// AddDomainListRule adds a new domain list rule from a URL.
// It downloads, parses, and persists the rule file.
// Note: Configuration management is handled by the caller.
func (m *fileManager) AddDomainListRule(ctx context.Context, rule *DomainListRule) error {
	// Validate rule
	if rule.ID == 0 {
		return fmt.Errorf("rule ID is required")
	}

	if rule.Name == "" {
		return fmt.Errorf("rule name is required")
	}

	if err := m.validateURL(rule.URL); err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	if rule.UpstreamGroup == "" {
		return fmt.Errorf("upstream group is required")
	}

	// Download rule file
	data, err := m.downloadRuleFile(ctx, rule.URL)
	if err != nil {
		return fmt.Errorf("downloading rule file: %w", err)
	}

	// Parse rule file
	parser := dnsrouting.NewParser()
	parseResult, err := parser.Parse(bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("parsing rule file: %w", err)
	}

	m.config.Logger.InfoContext(ctx, "parsed rule file",
		"id", rule.ID,
		"format", parseResult.Format,
		"valid_rules", parseResult.ValidRules,
	)

	// Update rule with parse results
	rule.RulesCount = parseResult.ValidRules
	rule.LastUpdated = time.Now()
	rule.FilePath = m.getRuleFilePath(rule.ID)

	// Convert parsed rules to AdGuard format
	parsedData := convertToAdGuardFormat(parseResult.Rules)

	// Save parsed rules to disk in AdGuard format
	if err := m.writeFileAtomic(rule.FilePath, parsedData); err != nil {
		return fmt.Errorf("saving rule file: %w", err)
	}

	m.config.Logger.InfoContext(ctx, "added domain list rule file",
		"id", rule.ID,
		"name", rule.Name,
		"rules_count", rule.RulesCount,
	)

	// Notify router
	m.scheduleRouterUpdate(ctx)

	// Schedule auto-update for this rule
	m.scheduleRuleUpdate(rule.ID)

	return nil
}

// UpdateDomainListRule updates an existing domain list rule file.
// It re-downloads and re-parses the rule file.
// Note: Configuration management is handled by the caller.
func (m *fileManager) UpdateDomainListRule(ctx context.Context, rule *DomainListRule) error {
	// Validate rule
	if rule.ID == 0 {
		return fmt.Errorf("rule ID is required")
	}

	if rule.Name == "" {
		return fmt.Errorf("rule name is required")
	}

	if err := m.validateURL(rule.URL); err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	if rule.UpstreamGroup == "" {
		return fmt.Errorf("upstream group is required")
	}

	// Download rule file
	data, err := m.downloadRuleFile(ctx, rule.URL)
	if err != nil {
		return fmt.Errorf("downloading rule file: %w", err)
	}

	// Parse rule file
	parser := dnsrouting.NewParser()
	parseResult, err := parser.Parse(bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("parsing rule file: %w", err)
	}

	m.config.Logger.InfoContext(ctx, "parsed updated rule file",
		"id", rule.ID,
		"format", parseResult.Format,
		"valid_rules", parseResult.ValidRules,
	)

	// Update rule info
	rule.RulesCount = parseResult.ValidRules
	rule.LastUpdated = time.Now()
	rule.FilePath = m.getRuleFilePath(rule.ID)

	// Convert parsed rules to AdGuard format
	parsedData := convertToAdGuardFormat(parseResult.Rules)

	// Save parsed rules to disk in AdGuard format
	if err := m.writeFileAtomic(rule.FilePath, parsedData); err != nil {
		return fmt.Errorf("saving rule file: %w", err)
	}

	m.config.Logger.InfoContext(ctx, "updated domain list rule file",
		"id", rule.ID,
		"name", rule.Name,
		"rules_count", rule.RulesCount,
	)

	// Notify router
	m.scheduleRouterUpdate(ctx)

	// Reschedule auto-update for this rule
	m.scheduleRuleUpdate(rule.ID)

	return nil
}

// DeleteDomainListRule deletes a domain list rule file.
// Note: This only deletes the file. Configuration management is handled by the caller.
func (m *fileManager) DeleteDomainListRule(ctx context.Context, ruleID int64) error {
	// Stop and remove the auto-update timer for this rule
	m.mu.Lock()
	if timer, exists := m.ruleTimers[ruleID]; exists && timer != nil {
		timer.Stop()
		delete(m.ruleTimers, ruleID)
	}
	m.mu.Unlock()

	// Get file path
	filePath := m.getRuleFilePath(ruleID)

	// Remove rule file
	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("removing rule file: %w", err)
	}

	m.config.Logger.InfoContext(ctx, "deleted domain list rule file",
		"id", ruleID,
		"path", filePath,
	)

	// Notify router
	m.scheduleRouterUpdate(ctx)

	return nil
}

// RefreshDomainListRule re-downloads and re-parses a rule file.
// Note: Configuration management is handled by the caller.
func (m *fileManager) RefreshDomainListRule(ctx context.Context, rule *DomainListRule) error {
	// Validate rule
	if rule.ID == 0 {
		return fmt.Errorf("rule ID is required")
	}

	if rule.URL == "" {
		return fmt.Errorf("rule URL is required")
	}

	// Download rule file
	data, err := m.downloadRuleFile(ctx, rule.URL)
	if err != nil {
		return fmt.Errorf("downloading rule file: %w", err)
	}

	// Parse rule file
	parser := dnsrouting.NewParser()
	parseResult, err := parser.Parse(bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("parsing rule file: %w", err)
	}

	m.config.Logger.InfoContext(ctx, "parsed refreshed rule file",
		"id", rule.ID,
		"format", parseResult.Format,
		"valid_rules", parseResult.ValidRules,
	)

	// Update rule info
	rule.RulesCount = parseResult.ValidRules
	rule.LastUpdated = time.Now()
	rule.FilePath = m.getRuleFilePath(rule.ID)

	// Convert parsed rules to AdGuard format
	parsedData := convertToAdGuardFormat(parseResult.Rules)

	// Save parsed rules to disk in AdGuard format
	if err := m.writeFileAtomic(rule.FilePath, parsedData); err != nil {
		return fmt.Errorf("saving rule file: %w", err)
	}

	m.config.Logger.InfoContext(ctx, "refreshed domain list rule file",
		"id", rule.ID,
		"name", rule.Name,
		"rules_count", rule.RulesCount,
	)

	// Notify router
	m.scheduleRouterUpdate(ctx)

	// Reschedule auto-update for this rule
	m.scheduleRuleUpdate(rule.ID)

	return nil
}

// scheduleRouterUpdate schedules a router update with debouncing.
func (m *fileManager) scheduleRouterUpdate(ctx context.Context) {
	m.updateMu.Lock()

	// Mark update as pending
	m.updatePending = true

	// Reset timer if it exists
	if m.updateTimer != nil {
		m.updateTimer.Stop()
	}

	// Schedule update after 100ms
	m.updateTimer = time.AfterFunc(100*time.Millisecond, func() {
		// Use a new background context to avoid using a potentially cancelled context
		updateCtx := context.Background()
		
		m.updateMu.Lock()

		if !m.updatePending {
			m.updateMu.Unlock()
			return
		}

		m.updatePending = false
		m.updateMu.Unlock()

		// Call router update WITHOUT holding any locks to avoid deadlock
		if m.config.OnRouterUpdate != nil {
			if err := m.config.OnRouterUpdate(updateCtx); err != nil {
				m.config.Logger.ErrorContext(updateCtx, "router update failed", "error", err)
			}
		}
	})

	m.updateMu.Unlock()
}
