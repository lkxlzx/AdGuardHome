package dnsroutingfiles

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/AdguardTeam/AdGuardHome/internal/aghos"
	"github.com/AdguardTeam/AdGuardHome/internal/aghrenameio"
)

const (
	// rulesDirName is the subdirectory name for DNS routing rules
	// This is relative to DataDir (workDir), so it's "data/dns_routing_rules"
	rulesDirName = "data/dns_routing_rules"

	// customRulesFileName is the name of the custom rules file
	customRulesFileName = "custom_rules.txt"
)

// Helper functions for string operations

// Deprecated: Use strings.Join instead
// joinStrings is kept for backward compatibility with old format parsing
func joinStrings(strs []string, sep string) string {
	return strings.Join(strs, sep)
}

// Deprecated: Use strings.Split instead
// splitString is kept for backward compatibility with old format parsing
func splitString(s, sep string) []string {
	if s == "" {
		return []string{}
	}
	return strings.Split(s, sep)
}

// Deprecated: Use strings.TrimSpace instead
// trimString is kept for backward compatibility with old format parsing
func trimString(s string) string {
	return strings.TrimSpace(s)
}

// ensureDirectory creates the DNS routing rules directory if it doesn't exist.
func (m *fileManager) ensureDirectory() error {
	dirPath := m.getRulesDir()

	// Check if directory exists
	if _, err := os.Stat(dirPath); err == nil {
		return nil // Directory already exists
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("checking directory: %w", err)
	}

	// Create directory with appropriate permissions
	if err := os.MkdirAll(dirPath, aghos.DefaultPermDir); err != nil {
		return fmt.Errorf("creating directory: %w", err)
	}

	m.config.Logger.InfoContext(context.TODO(), "created DNS routing rules directory", "path", dirPath)

	return nil
}

// getRulesDir returns the path to the DNS routing rules directory.
func (m *fileManager) getRulesDir() string {
	return filepath.Join(m.config.DataDir, rulesDirName)
}

// getRuleFilePath returns the file path for a domain list rule.
func (m *fileManager) getRuleFilePath(ruleID int64) string {
	fileName := strconv.FormatInt(ruleID, 10) + ".txt"
	return filepath.Join(m.getRulesDir(), fileName)
}

// getCustomRulesFilePath returns the file path for custom rules.
func (m *fileManager) getCustomRulesFilePath() string {
	return filepath.Join(m.getRulesDir(), customRulesFileName)
}

// getMetadataFilePath returns the file path for metadata.
func (m *fileManager) getMetadataFilePath() string {
	return filepath.Join(m.getRulesDir(), metadataFileName)
}

// writeFileAtomic writes data to a file atomically using a temporary file.
func (m *fileManager) writeFileAtomic(filePath string, data []byte) error {
	// Create pending file for atomic write
	pendingFile, err := aghrenameio.NewPendingFile(filePath, aghos.DefaultPermFile)
	if err != nil {
		return fmt.Errorf("creating pending file: %w", err)
	}

	// Write data
	if _, err := pendingFile.Write(data); err != nil {
		// Cleanup on error
		_ = pendingFile.Cleanup()
		return fmt.Errorf("writing data: %w", err)
	}

	// Commit the write atomically
	if err := pendingFile.CloseReplace(); err != nil {
		return fmt.Errorf("committing write: %w", err)
	}

	return nil
}

// saveMetadata is deprecated - configuration is now saved by AdGuardHome's config system.
// This function is kept for compatibility but does nothing.
func (m *fileManager) saveMetadata(ctx context.Context) error {
	// Configuration is saved by AdGuardHome's config system
	// File Manager no longer manages metadata
	m.config.Logger.DebugContext(ctx, "saveMetadata called - configuration managed by AdGuardHome config system")
	return nil
}

// loadMetadata is deprecated - configuration is now loaded from AdGuardHome.yaml.
// This function is kept for compatibility but does nothing.
func (m *fileManager) loadMetadata(ctx context.Context) error {
	// Configuration is loaded from AdGuardHome.yaml
	// File Manager no longer manages metadata
	m.config.Logger.DebugContext(ctx, "loadMetadata called - configuration loaded from AdGuardHome.yaml")
	return nil
}



// saveCustomRules saves custom rules to a JSON file.
func (m *fileManager) saveCustomRules(ctx context.Context) error {
	customRulesPath := m.getCustomRulesFilePath()

	// Use JSON format for better structure and escaping
	data, err := json.MarshalIndent(m.customRules, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling custom rules: %w", err)
	}

	// Write atomically
	if err := m.writeFileAtomic(customRulesPath, data); err != nil {
		return fmt.Errorf("writing custom rules file: %w", err)
	}

	m.config.Logger.DebugContext(ctx, "saved custom rules",
		"path", customRulesPath,
		"count", len(m.customRules),
	)

	return nil
}

// formatCustomRule formats a custom rule as a text line.
// Format: domain|match_type|upstream_group|enabled
func (m *fileManager) formatCustomRule(rule *CustomRule) string {
	enabled := "false"
	if rule.Enabled {
		enabled = "true"
	}
	return fmt.Sprintf("%s|%s|%s|%s", rule.Domain, rule.MatchType, rule.UpstreamGroup, enabled)
}

// parseCustomRule parses a custom rule from a text line.
func (m *fileManager) parseCustomRule(line string) (*CustomRule, error) {
	// Split by pipe character
	parts := splitString(line, "|")
	if len(parts) != 4 {
		return nil, fmt.Errorf("invalid custom rule format: expected 4 parts, got %d", len(parts))
	}

	enabled := parts[3] == "true"

	return &CustomRule{
		Domain:        parts[0],
		MatchType:     parts[1],
		UpstreamGroup: parts[2],
		Enabled:       enabled,
	}, nil
}

// loadCustomRules loads custom rules from the JSON file with backward compatibility.
func (m *fileManager) loadCustomRules(ctx context.Context) error {
	customRulesPath := m.getCustomRulesFilePath()

	// Check if custom rules file exists
	if _, err := os.Stat(customRulesPath); os.IsNotExist(err) {
		m.config.Logger.DebugContext(ctx, "custom rules file does not exist", "path", customRulesPath)
		return nil
	}

	// Read custom rules file
	data, err := os.ReadFile(customRulesPath)
	if err != nil {
		return fmt.Errorf("reading custom rules file: %w", err)
	}

	// Try to parse as JSON first (new format)
	var rules []*CustomRule
	err = json.Unmarshal(data, &rules)
	if err == nil {
		m.customRules = rules
		if len(m.customRules) > 0 {
			m.config.Logger.InfoContext(ctx, "loaded custom rules from JSON",
				"path", customRulesPath,
				"count", len(m.customRules),
			)
		} else {
			m.config.Logger.DebugContext(ctx, "loaded custom rules from JSON",
				"path", customRulesPath,
				"count", 0,
			)
		}
		return nil
	}

	// Fall back to old pipe-separated format for backward compatibility
	m.config.Logger.InfoContext(ctx, "JSON parse failed, trying legacy format", "error", err)
	
	lines := splitString(string(data), "\n")
	m.customRules = make([]*CustomRule, 0, len(lines))

	for i, line := range lines {
		line = trimString(line)
		if line == "" {
			continue
		}

		rule, err := m.parseCustomRule(line)
		if err != nil {
			m.config.Logger.WarnContext(ctx, "skipping invalid custom rule",
				"line", i+1,
				"error", err,
			)
			continue
		}

		m.customRules = append(m.customRules, rule)
	}

	if len(m.customRules) > 0 {
		m.config.Logger.InfoContext(ctx, "loaded custom rules from legacy format",
			"path", customRulesPath,
			"count", len(m.customRules),
		)
		// Migrate to JSON format on next save
		m.config.Logger.InfoContext(ctx, "will migrate custom rules to JSON format on next save")
	} else {
		m.config.Logger.DebugContext(ctx, "loaded custom rules from legacy format",
			"path", customRulesPath,
			"count", 0,
		)
	}

	return nil
}
