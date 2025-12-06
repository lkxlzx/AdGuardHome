// Package dnsroutingfiles provides centralized management of DNS routing rule files.
// It handles downloading, parsing, persisting, and loading of both domain list rules
// (from URLs) and custom domain rules (user-defined).
package dnsroutingfiles

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/AdguardTeam/AdGuardHome/internal/dnsrouting"
)

// migrationChecked is a global flag to avoid repeated migration checks
// This improves startup performance by skipping migration logic after first check
var migrationChecked bool
var migrationCheckMu sync.Mutex

// Manager handles all DNS routing rule file operations.
type Manager interface {
	// AddDomainListRule adds a new domain list rule from a URL.
	// It downloads, parses, and persists the rule file.
	AddDomainListRule(ctx context.Context, rule *DomainListRule) error

	// UpdateDomainListRule updates an existing domain list rule.
	// It re-downloads and re-parses the rule file if the URL changed.
	UpdateDomainListRule(ctx context.Context, rule *DomainListRule) error

	// DeleteDomainListRule deletes a domain list rule and its file.
	DeleteDomainListRule(ctx context.Context, ruleID int64) error

	// RefreshDomainListRule re-downloads and re-parses a rule file.
	RefreshDomainListRule(ctx context.Context, rule *DomainListRule) error

	// GetDomainListRule returns a domain list rule by ID.
	GetDomainListRule(ruleID int64) (*DomainListRule, error)

	// GetAllDomainListRules returns all domain list rules.
	GetAllDomainListRules() []*DomainListRule

	// GetRuleFilePath returns the file path for a rule.
	GetRuleFilePath(ruleID int64) string

	// AddCustomRule adds a new custom domain rule.
	AddCustomRule(ctx context.Context, rule *CustomRule) error

	// UpdateCustomRule updates an existing custom domain rule.
	UpdateCustomRule(ctx context.Context, oldDomain, oldMatchType string, rule *CustomRule) error

	// DeleteCustomRule deletes a custom domain rule.
	DeleteCustomRule(ctx context.Context, domain, matchType string) error

	// GetAllCustomRules returns all custom domain rules.
	GetAllCustomRules() []*CustomRule

	// LoadAll loads all persisted rules from disk on startup.
	LoadAll(ctx context.Context) error

	// Close closes the file manager and releases resources.
	Close() error
}

// DomainListRule represents a DNS routing rule from a URL source.
type DomainListRule struct {
	ID             int64     `json:"id"`
	Name           string    `json:"name"`
	URL            string    `json:"url"`
	UpstreamGroup  string    `json:"upstream_group"`
	Priority       int       `json:"priority"`
	Enabled        bool      `json:"enabled"`
	RulesCount     int       `json:"rules_count"`
	LastUpdated    time.Time `json:"last_updated"`
	FilePath       string    `json:"file_path"`       // Path to the downloaded rule file
	UpdateInterval int       `json:"update_interval"` // Update interval in minutes (0 = use default)
}

// CustomRule represents a user-defined DNS routing rule.
type CustomRule struct {
	Domain        string `json:"domain"`
	MatchType     string `json:"match_type"` // DOMAIN, DOMAIN-SUFFIX, DOMAIN-KEYWORD
	UpstreamGroup string `json:"upstream_group"`
	Enabled       bool   `json:"enabled"`
}

// RouterUpdateCallback is called when rules are updated and the router needs to reload.
type RouterUpdateCallback func(ctx context.Context) error

// GetRuleConfigCallback is called to get the current configuration for a rule.
// Returns the rule config including update interval, or nil if rule not found.
type GetRuleConfigCallback func(ruleID int64) *DomainListRule

// SaveRuleConfigCallback is called to save the updated rule configuration.
type SaveRuleConfigCallback func(ctx context.Context, rule *DomainListRule) error

// Config contains configuration for the File Manager.
type Config struct {
	// DataDir is the base data directory (e.g., "data")
	DataDir string

	// Logger is the structured logger instance
	Logger *slog.Logger

	// HTTPClient is used for downloading rule files from URLs
	HTTPClient *http.Client

	// OnRouterUpdate is called when rules change and the router needs to reload
	OnRouterUpdate RouterUpdateCallback

	// GetRuleConfig is called to get the current configuration for a rule
	GetRuleConfig GetRuleConfigCallback

	// SaveRuleConfig is called to save the updated rule configuration
	SaveRuleConfig SaveRuleConfigCallback
}

// fileManager is the concrete implementation of Manager.
type fileManager struct {
	// config holds the manager configuration
	config Config

	// mu protects domainListRules and customRules
	mu sync.RWMutex

	// domainListRules stores all domain list rules indexed by ID
	domainListRules map[int64]*DomainListRule

	// customRules stores all custom domain rules
	customRules []*CustomRule

	// updateTimer is used for batching router updates
	updateTimer *time.Timer

	// updatePending indicates if a router update is pending
	updatePending bool

	// updateMu protects updateTimer and updatePending
	updateMu sync.Mutex

	// ruleTimers stores individual timers for each rule's auto-update
	ruleTimers map[int64]*time.Timer

	// ruleRetryCount tracks retry attempts for each rule
	ruleRetryCount map[int64]int

	// stopAutoUpdate is used to stop all auto-update timers
	stopAutoUpdate chan struct{}
	
	// closeOnce ensures Close is only executed once
	closeOnce sync.Once
}

// NewManager creates a new DNS routing file manager.
func NewManager(config Config) (Manager, error) {
	if config.DataDir == "" {
		return nil, fmt.Errorf("data directory is required")
	}

	if config.Logger == nil {
		return nil, fmt.Errorf("logger is required")
	}

	if config.HTTPClient == nil {
		config.HTTPClient = http.DefaultClient
	}

	m := &fileManager{
		config:          config,
		domainListRules: make(map[int64]*DomainListRule),
		customRules:     make([]*CustomRule, 0),
		ruleTimers:      make(map[int64]*time.Timer),
		ruleRetryCount:  make(map[int64]int),
		stopAutoUpdate:  make(chan struct{}),
	}

	// Ensure the rule directory exists
	if err := m.ensureDirectory(); err != nil {
		return nil, fmt.Errorf("ensuring directory: %w", err)
	}

	return m, nil
}

// Close closes the file manager and releases resources.
// It is safe to call Close multiple times.
func (m *fileManager) Close() error {
	var closeErr error
	
	m.closeOnce.Do(func() {
		// Stop auto-update goroutine safely
		m.mu.Lock()
		if m.stopAutoUpdate != nil {
			close(m.stopAutoUpdate)
			m.stopAutoUpdate = nil
		}

		// Stop all rule timers
		for _, timer := range m.ruleTimers {
			if timer != nil {
				timer.Stop()
			}
		}
		m.ruleTimers = make(map[int64]*time.Timer)
		m.mu.Unlock()

		m.updateMu.Lock()
		// Stop any pending timer
		if m.updateTimer != nil {
			m.updateTimer.Stop()
		}

		// Flush any pending router update
		if m.updatePending {
			if m.config.OnRouterUpdate != nil {
				ctx := context.Background()
				if err := m.config.OnRouterUpdate(ctx); err != nil {
					m.config.Logger.ErrorContext(ctx, "flushing pending router update", "error", err)
					closeErr = err
				}
			}
			m.updatePending = false
		}
		m.updateMu.Unlock()
	})
	
	return closeErr
}

// oldClose is kept for reference - the old implementation
func (m *fileManager) oldClose() error {
	// Stop auto-update goroutine safely
	m.mu.Lock()
	if m.stopAutoUpdate != nil {
		select {
		case <-m.stopAutoUpdate:
			// Already closed
		default:
			close(m.stopAutoUpdate)
		}
		m.stopAutoUpdate = nil
	}

	// Stop all rule timers
	for _, timer := range m.ruleTimers {
		if timer != nil {
			timer.Stop()
		}
	}
	m.ruleTimers = make(map[int64]*time.Timer)
	m.mu.Unlock()

	m.updateMu.Lock()
	defer m.updateMu.Unlock()

	// Stop any pending timer
	if m.updateTimer != nil {
		m.updateTimer.Stop()
	}

	// Flush any pending router update
	if m.updatePending {
		if m.config.OnRouterUpdate != nil {
			ctx := context.Background()
			if err := m.config.OnRouterUpdate(ctx); err != nil {
				m.config.Logger.ErrorContext(ctx, "flushing pending router update", "error", err)
			}
		}
		m.updatePending = false
	}

	return nil
}

// Domain list rule operations are implemented in rules.go

// GetDomainListRule returns a domain list rule by ID.
func (m *fileManager) GetDomainListRule(ruleID int64) (*DomainListRule, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	rule, exists := m.domainListRules[ruleID]
	if !exists {
		return nil, fmt.Errorf("rule with ID %d not found", ruleID)
	}

	// Return a copy to prevent external modification
	ruleCopy := *rule
	return &ruleCopy, nil
}

// GetAllDomainListRules returns all domain list rules.
func (m *fileManager) GetAllDomainListRules() []*DomainListRule {
	m.mu.RLock()
	defer m.mu.RUnlock()

	rules := make([]*DomainListRule, 0, len(m.domainListRules))
	for _, rule := range m.domainListRules {
		// Return copies to prevent external modification
		ruleCopy := *rule
		rules = append(rules, &ruleCopy)
	}

	return rules
}

// GetRuleFilePath returns the file path for a rule.
func (m *fileManager) GetRuleFilePath(ruleID int64) string {
	return m.getRuleFilePath(ruleID)
}

// AddCustomRule adds a new custom domain rule.
func (m *fileManager) AddCustomRule(ctx context.Context, rule *CustomRule) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Validate rule
	if rule.Domain == "" {
		return fmt.Errorf("domain is required")
	}

	if rule.MatchType == "" {
		return fmt.Errorf("match type is required")
	}

	// Validate match type
	validMatchTypes := map[string]bool{
		"DOMAIN":         true,
		"DOMAIN-SUFFIX":  true,
		"DOMAIN-KEYWORD": true,
	}
	if !validMatchTypes[rule.MatchType] {
		return fmt.Errorf("invalid match type: must be DOMAIN, DOMAIN-SUFFIX, or DOMAIN-KEYWORD")
	}

	if rule.UpstreamGroup == "" {
		return fmt.Errorf("upstream group is required")
	}

	// Check for duplicate
	for _, existingRule := range m.customRules {
		if existingRule.Domain == rule.Domain && existingRule.MatchType == rule.MatchType {
			return fmt.Errorf("rule with domain %s and match type %s already exists", rule.Domain, rule.MatchType)
		}
	}

	// Add to internal state
	m.customRules = append(m.customRules, rule)

	// Save custom rules file
	if err := m.saveCustomRules(ctx); err != nil {
		// Rollback
		m.customRules = m.customRules[:len(m.customRules)-1]
		return fmt.Errorf("saving custom rules: %w", err)
	}

	// Save metadata
	if err := m.saveMetadata(ctx); err != nil {
		return fmt.Errorf("saving metadata: %w", err)
	}

	m.config.Logger.InfoContext(ctx, "added custom rule",
		"domain", rule.Domain,
		"match_type", rule.MatchType,
	)

	// Unlock is handled by defer, no manual unlock needed
	// Notify router after lock is released by defer
	// Use a goroutine to avoid any potential deadlock
	go m.scheduleRouterUpdate(ctx)

	return nil
}

// UpdateCustomRule updates an existing custom domain rule.
func (m *fileManager) UpdateCustomRule(ctx context.Context, oldDomain, oldMatchType string, rule *CustomRule) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Validate new rule
	if rule.Domain == "" {
		return fmt.Errorf("domain is required")
	}

	if rule.MatchType == "" {
		return fmt.Errorf("match type is required")
	}

	validMatchTypes := map[string]bool{
		"DOMAIN":         true,
		"DOMAIN-SUFFIX":  true,
		"DOMAIN-KEYWORD": true,
	}
	if !validMatchTypes[rule.MatchType] {
		return fmt.Errorf("invalid match type")
	}

	if rule.UpstreamGroup == "" {
		return fmt.Errorf("upstream group is required")
	}

	// Find existing rule
	var foundIndex = -1
	var oldRule *CustomRule
	for i, existingRule := range m.customRules {
		if existingRule.Domain == oldDomain && existingRule.MatchType == oldMatchType {
			foundIndex = i
			oldRule = existingRule
			break
		}
	}

	if foundIndex == -1 {
		return fmt.Errorf("rule with domain %s and match type %s not found", oldDomain, oldMatchType)
	}

	// Update rule
	m.customRules[foundIndex] = rule

	// Save custom rules file
	if err := m.saveCustomRules(ctx); err != nil {
		// Rollback
		m.customRules[foundIndex] = oldRule
		return fmt.Errorf("saving custom rules: %w", err)
	}

	// Save metadata
	if err := m.saveMetadata(ctx); err != nil {
		return fmt.Errorf("saving metadata: %w", err)
	}

	m.config.Logger.InfoContext(ctx, "updated custom rule",
		"old_domain", oldDomain,
		"new_domain", rule.Domain,
	)

	// Unlock is handled by defer, no manual unlock needed
	// Notify router after lock is released by defer
	// Use a goroutine to avoid any potential deadlock
	go m.scheduleRouterUpdate(ctx)

	return nil
}

// DeleteCustomRule deletes a custom domain rule.
func (m *fileManager) DeleteCustomRule(ctx context.Context, domain, matchType string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Find rule
	var foundIndex = -1
	var oldRule *CustomRule
	for i, existingRule := range m.customRules {
		if existingRule.Domain == domain && existingRule.MatchType == matchType {
			foundIndex = i
			oldRule = existingRule
			break
		}
	}

	if foundIndex == -1 {
		return fmt.Errorf("rule with domain %s and match type %s not found", domain, matchType)
	}

	// Remove from slice
	m.customRules = append(m.customRules[:foundIndex], m.customRules[foundIndex+1:]...)

	// Save custom rules file
	if err := m.saveCustomRules(ctx); err != nil {
		// Rollback
		m.customRules = append(m.customRules[:foundIndex], append([]*CustomRule{oldRule}, m.customRules[foundIndex:]...)...)
		return fmt.Errorf("saving custom rules: %w", err)
	}

	// Save metadata
	if err := m.saveMetadata(ctx); err != nil {
		return fmt.Errorf("saving metadata: %w", err)
	}

	m.config.Logger.InfoContext(ctx, "deleted custom rule",
		"domain", domain,
		"match_type", matchType,
	)

	// Unlock is handled by defer, no manual unlock needed
	// Notify router after lock is released by defer
	// Use a goroutine to avoid any potential deadlock
	go m.scheduleRouterUpdate(ctx)

	return nil
}

// GetAllCustomRules returns all custom domain rules.
func (m *fileManager) GetAllCustomRules() []*CustomRule {
	m.mu.RLock()
	defer m.mu.RUnlock()

	rules := make([]*CustomRule, len(m.customRules))
	for i, rule := range m.customRules {
		// Return copies to prevent external modification
		ruleCopy := *rule
		rules[i] = &ruleCopy
	}

	return rules
}

// LoadAll loads all persisted rules from disk on startup.
// NOTE: Configuration is now loaded from AdGuardHome.yaml by the config system.
// This method is deprecated and kept for compatibility.
func (m *fileManager) LoadAll(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Load metadata (deprecated - does nothing now)
	if err := m.loadMetadata(ctx); err != nil {
		return fmt.Errorf("loading metadata: %w", err)
	}

	// Try to migrate from old implementation (migrate rule files only)
	if err := m.migrateFromOldImplementation(ctx); err != nil {
		m.config.Logger.WarnContext(ctx, "migration from old implementation failed", "error", err)
		// Continue anyway
	}

	// Load custom rules
	if err := m.loadCustomRules(ctx); err != nil {
		return fmt.Errorf("loading custom rules: %w", err)
	}

	// Note: Domain list rules are now loaded from AdGuardHome.yaml by reloadDnsRoutingRules()
	// This only loads custom rules from data/dns_routing_rules/custom_rules.txt
	if len(m.customRules) > 0 {
		m.config.Logger.InfoContext(ctx, "loaded custom rules from file",
			"custom_rules", len(m.customRules),
		)
	} else {
		m.config.Logger.DebugContext(ctx, "no custom rules file found or file is empty")
	}

	return nil
}

// migrateFromOldImplementation migrates DNS routing rules from the old filtering system.
// This only migrates rule files and runtime state. Configuration is always read from AdGuardHome.yaml.
func (m *fileManager) migrateFromOldImplementation(ctx context.Context) error {
	// Fast path: check if migration was already checked in this process
	migrationCheckMu.Lock()
	if migrationChecked {
		migrationCheckMu.Unlock()
		return nil
	}
	migrationCheckMu.Unlock()
	
	// Check if metadata file exists (new system already initialized)
	metadataPath := m.getMetadataFilePath()
	m.config.Logger.DebugContext(ctx, "checking for migration", "metadata_path", metadataPath)
	
	if _, err := os.Stat(metadataPath); err == nil {
		m.config.Logger.InfoContext(ctx, "migration skipped: metadata file exists")
		// Mark as checked to avoid future checks
		migrationCheckMu.Lock()
		migrationChecked = true
		migrationCheckMu.Unlock()
		return nil
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("checking metadata file: %w", err)
	}

	// Check if old rule files exist
	oldFiltersDir := filepath.Join("data", "filters")
	if _, err := os.Stat(oldFiltersDir); os.IsNotExist(err) {
		m.config.Logger.InfoContext(ctx, "no old filters directory found, skipping migration")
		return nil
	}

	// Configuration is already loaded from YAML in loadMetadata
	// We just need to migrate the rule files
	var migratedCount int
	for _, rule := range m.domainListRules {
		// Check if old rule file exists
		oldFilePath := filepath.Join(oldFiltersDir, fmt.Sprintf("%d.txt", rule.ID))
		if _, err := os.Stat(oldFilePath); os.IsNotExist(err) {
			m.config.Logger.DebugContext(ctx, "old rule file not found",
				"id", rule.ID,
				"path", oldFilePath)
			continue
		}

		m.config.Logger.InfoContext(ctx, "migrating DNS routing rule file",
			"id", rule.ID,
			"name", rule.Name,
			"old_path", oldFilePath)

		// Read old rule file
		oldData, err := os.ReadFile(oldFilePath)
		if err != nil {
			m.config.Logger.ErrorContext(ctx, "failed to read old rule file",
				"id", rule.ID,
				"path", oldFilePath,
				"error", err)
			continue
		}

		// Parse the rule file to get rule count
		parser := dnsrouting.NewParser()
		parseResult, err := parser.Parse(bytes.NewReader(oldData))
		if err != nil {
			m.config.Logger.ErrorContext(ctx, "failed to parse old rule file",
				"id", rule.ID,
				"error", err)
			continue
		}

		// Update runtime state
		rule.RulesCount = parseResult.ValidRules
		rule.LastUpdated = time.Now()

		// Copy rule file to new location
		if err := m.writeFileAtomic(rule.FilePath, oldData); err != nil {
			m.config.Logger.ErrorContext(ctx, "failed to copy rule file",
				"id", rule.ID,
				"old_path", oldFilePath,
				"new_path", rule.FilePath,
				"error", err)
			continue
		}

		migratedCount++

		m.config.Logger.InfoContext(ctx, "migrated DNS routing rule file",
			"id", rule.ID,
			"name", rule.Name,
			"rules_count", rule.RulesCount,
			"old_path", oldFilePath,
			"new_path", rule.FilePath)
	}

	if migratedCount > 0 {
		// Save metadata for migrated rules
		if err := m.saveMetadata(ctx); err != nil {
			return fmt.Errorf("saving migrated metadata: %w", err)
		}

		m.config.Logger.InfoContext(ctx, "migration completed",
			"migrated_rules", migratedCount)
	} else {
		m.config.Logger.DebugContext(ctx, "no DNS routing rule files found to migrate")
	}

	// Mark migration as checked to avoid future checks
	migrationCheckMu.Lock()
	migrationChecked = true
	migrationCheckMu.Unlock()

	return nil
}


// scheduleRuleUpdate schedules an auto-update for a specific rule.
func (m *fileManager) scheduleRuleUpdate(ruleID int64) {
	if m.config.GetRuleConfig == nil || m.config.SaveRuleConfig == nil {
		return
	}

	// Get rule config
	ruleConfig := m.config.GetRuleConfig(ruleID)
	if ruleConfig == nil || !ruleConfig.Enabled {
		return
	}

	// Calculate update interval
	updateInterval := time.Duration(ruleConfig.UpdateInterval) * time.Minute
	if updateInterval == 0 {
		// Use default interval (24 hours) if not specified
		updateInterval = 24 * time.Hour
	}

	// Calculate time until next update
	now := time.Now()
	nextUpdate := ruleConfig.LastUpdated.Add(updateInterval)
	timeUntilUpdate := nextUpdate.Sub(now)

	// If already past update time, update immediately
	if timeUntilUpdate < 0 {
		timeUntilUpdate = 1 * time.Second
	}

	m.config.Logger.InfoContext(context.Background(), "scheduled auto-update for DNS routing rule",
		"id", ruleID,
		"name", ruleConfig.Name,
		"update_interval", updateInterval,
		"time_until_update", timeUntilUpdate)

	// Stop existing timer if any
	m.mu.Lock()
	if existingTimer, exists := m.ruleTimers[ruleID]; exists && existingTimer != nil {
		existingTimer.Stop()
	}

	// Create new timer with panic recovery
	timer := time.AfterFunc(timeUntilUpdate, func() {
		defer func() {
			if r := recover(); r != nil {
				m.config.Logger.ErrorContext(context.Background(), "panic in auto-update timer",
					"id", ruleID,
					"panic", r)
			}
		}()
		m.autoUpdateRule(ruleID)
	})
	m.ruleTimers[ruleID] = timer
	m.mu.Unlock()
}

// autoUpdateRule performs auto-update for a specific rule with retry limit.
func (m *fileManager) autoUpdateRule(ruleID int64) {
	ctx := context.Background()

	// Get current config
	ruleConfig := m.config.GetRuleConfig(ruleID)
	if ruleConfig == nil {
		m.config.Logger.DebugContext(ctx, "rule not found in config, skipping auto-update",
			"id", ruleID)
		return
	}

	// Skip disabled rules
	if !ruleConfig.Enabled {
		m.config.Logger.DebugContext(ctx, "rule disabled, skipping auto-update",
			"id", ruleID)
		return
	}

	m.config.Logger.InfoContext(ctx, "auto-updating DNS routing rule",
		"id", ruleID,
		"name", ruleConfig.Name)

	// Update the rule
	err := m.RefreshDomainListRule(ctx, ruleConfig)
	if err != nil {
		// Get current retry count
		m.mu.Lock()
		retryCount := m.ruleRetryCount[ruleID]
		m.mu.Unlock()

		m.config.Logger.ErrorContext(ctx, "failed to auto-update DNS routing rule",
			"id", ruleID,
			"name", ruleConfig.Name,
			"error", err,
			"retry_count", retryCount)

		// Maximum 5 retries
		const maxRetries = 5
		if retryCount >= maxRetries {
			m.config.Logger.ErrorContext(ctx, "max retries reached for DNS routing rule, giving up",
				"id", ruleID,
				"name", ruleConfig.Name,
				"max_retries", maxRetries)
			
			// Reset retry count and schedule next regular update
			m.mu.Lock()
			m.ruleRetryCount[ruleID] = 0
			m.mu.Unlock()
			m.scheduleRuleUpdate(ruleID)
			return
		}

		// Increment retry count
		m.mu.Lock()
		m.ruleRetryCount[ruleID] = retryCount + 1
		m.mu.Unlock()

		// Exponential backoff: 1h, 2h, 4h, 8h, 16h
		retryDelay := time.Duration(1<<uint(retryCount)) * time.Hour
		
		m.config.Logger.InfoContext(ctx, "scheduling retry for DNS routing rule",
			"id", ruleID,
			"name", ruleConfig.Name,
			"retry_count", retryCount+1,
			"retry_delay", retryDelay)

		// Schedule retry with exponential backoff
		m.mu.Lock()
		timer := time.AfterFunc(retryDelay, func() {
			m.autoUpdateRule(ruleID)
		})
		m.ruleTimers[ruleID] = timer
		m.mu.Unlock()
		return
	}

	// Success - reset retry count
	m.mu.Lock()
	m.ruleRetryCount[ruleID] = 0
	m.mu.Unlock()

	// Save the updated config
	if m.config.SaveRuleConfig != nil {
		err = m.config.SaveRuleConfig(ctx, ruleConfig)
		if err != nil {
			m.config.Logger.ErrorContext(ctx, "failed to save updated rule config",
				"id", ruleID,
				"name", ruleConfig.Name,
				"error", err)
		}
	}

	m.config.Logger.InfoContext(ctx, "auto-updated DNS routing rule",
		"id", ruleID,
		"name", ruleConfig.Name,
		"rules_count", ruleConfig.RulesCount)

	// Schedule next update
	m.scheduleRuleUpdate(ruleID)
}

// StartAutoUpdates starts auto-update timers for all rules.
func (m *fileManager) StartAutoUpdates() {
	if m.config.GetRuleConfig == nil {
		return
	}

	ctx := context.Background()
	m.config.Logger.InfoContext(ctx, "starting auto-updates for DNS routing rules")

	// Get all rule IDs from config
	m.mu.RLock()
	ruleIDs := make([]int64, 0, len(m.domainListRules))
	for id := range m.domainListRules {
		ruleIDs = append(ruleIDs, id)
	}
	m.mu.RUnlock()

	// Schedule updates for each rule
	for _, ruleID := range ruleIDs {
		m.scheduleRuleUpdate(ruleID)
	}
}
