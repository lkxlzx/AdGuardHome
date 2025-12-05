// Package dnsroutingfiles provides file management for DNS routing rules.
// It only handles downloading and parsing rule files.
package dnsroutingfiles

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/AdguardTeam/AdGuardHome/internal/aghos"
	"github.com/AdguardTeam/AdGuardHome/internal/aghrenameio"
	"github.com/AdguardTeam/AdGuardHome/internal/dnsrouting"
)

// SimpleManager handles downloading and parsing DNS routing rule files.
type SimpleManager interface {
	// DownloadAndParseRule downloads a rule file from URL and parses it.
	// Returns the parsed rules and rule count.
	DownloadAndParseRule(ctx context.Context, ruleID int64, url string) (rulesCount int, err error)

	// GetRuleFilePath returns the file path for a rule.
	GetRuleFilePath(ruleID int64) string

	// DeleteRuleFile deletes a rule file.
	DeleteRuleFile(ctx context.Context, ruleID int64) error

	// LoadRuleFile loads and parses an existing rule file.
	LoadRuleFile(ctx context.Context, ruleID int64) (rules []dnsrouting.ParsedRule, err error)
}

// Config contains configuration for the Simple Manager.
type SimpleConfig struct {
	// DataDir is the base data directory
	DataDir string

	// Logger is the structured logger instance
	Logger *slog.Logger

	// HTTPClient is used for downloading rule files
	HTTPClient *http.Client
}

// simpleManager is the implementation.
type simpleManager struct {
	config SimpleConfig
}

// NewSimpleManager creates a new simple file manager.
func NewSimpleManager(config SimpleConfig) (SimpleManager, error) {
	if config.DataDir == "" {
		return nil, fmt.Errorf("data directory is required")
	}

	if config.Logger == nil {
		return nil, fmt.Errorf("logger is required")
	}

	if config.HTTPClient == nil {
		config.HTTPClient = &http.Client{
			Timeout: 30 * time.Second,
		}
	}

	m := &simpleManager{
		config: config,
	}

	// Ensure directory exists
	if err := m.ensureDirectory(); err != nil {
		return nil, fmt.Errorf("ensuring directory: %w", err)
	}

	return m, nil
}

// ensureDirectory creates the DNS routing rules directory if it doesn't exist.
func (m *simpleManager) ensureDirectory() error {
	dirPath := m.getRulesDir()

	if _, err := os.Stat(dirPath); err == nil {
		return nil
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("checking directory: %w", err)
	}

	if err := os.MkdirAll(dirPath, aghos.DefaultPermDir); err != nil {
		return fmt.Errorf("creating directory: %w", err)
	}

	m.config.Logger.InfoContext(context.TODO(), "created DNS routing rules directory", "path", dirPath)
	return nil
}

// getRulesDir returns the path to the DNS routing rules directory.
func (m *simpleManager) getRulesDir() string {
	return filepath.Join(m.config.DataDir, "data", "dns_routing_rules")
}

// GetRuleFilePath returns the file path for a rule.
func (m *simpleManager) GetRuleFilePath(ruleID int64) string {
	fileName := strconv.FormatInt(ruleID, 10) + ".txt"
	return filepath.Join(m.getRulesDir(), fileName)
}

// DownloadAndParseRule downloads a rule file from URL and parses it.
func (m *simpleManager) DownloadAndParseRule(ctx context.Context, ruleID int64, url string) (rulesCount int, err error) {
	m.config.Logger.InfoContext(ctx, "downloading rule file", "id", ruleID, "url", url)

	// Download
	data, err := m.downloadFile(ctx, url)
	if err != nil {
		return 0, fmt.Errorf("downloading: %w", err)
	}

	// Parse
	parser := dnsrouting.NewParser()
	parseResult, err := parser.Parse(bytes.NewReader(data))
	if err != nil {
		return 0, fmt.Errorf("parsing: %w", err)
	}

	m.config.Logger.InfoContext(ctx, "parsed rule file",
		"id", ruleID,
		"format", parseResult.Format,
		"valid_rules", parseResult.ValidRules,
	)

	// Save to disk
	filePath := m.GetRuleFilePath(ruleID)
	if err := m.writeFileAtomic(filePath, data); err != nil {
		return 0, fmt.Errorf("saving file: %w", err)
	}

	return parseResult.ValidRules, nil
}

// DeleteRuleFile deletes a rule file.
func (m *simpleManager) DeleteRuleFile(ctx context.Context, ruleID int64) error {
	filePath := m.GetRuleFilePath(ruleID)
	
	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("deleting file: %w", err)
	}

	m.config.Logger.InfoContext(ctx, "deleted rule file", "id", ruleID, "path", filePath)
	return nil
}

// LoadRuleFile loads and parses an existing rule file.
func (m *simpleManager) LoadRuleFile(ctx context.Context, ruleID int64) ([]dnsrouting.ParsedRule, error) {
	filePath := m.GetRuleFilePath(ruleID)

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}

	parser := dnsrouting.NewParser()
	parseResult, err := parser.Parse(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("parsing: %w", err)
	}

	return parseResult.Rules, nil
}

// downloadFile downloads a file from URL.
func (m *simpleManager) downloadFile(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	resp, err := m.config.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	return data, nil
}

// writeFileAtomic writes data to a file atomically.
func (m *simpleManager) writeFileAtomic(filePath string, data []byte) error {
	pendingFile, err := aghrenameio.NewPendingFile(filePath, aghos.DefaultPermFile)
	if err != nil {
		return fmt.Errorf("creating pending file: %w", err)
	}

	if _, err := pendingFile.Write(data); err != nil {
		_ = pendingFile.Cleanup()
		return fmt.Errorf("writing data: %w", err)
	}

	if err := pendingFile.CloseReplace(); err != nil {
		return fmt.Errorf("committing write: %w", err)
	}

	return nil
}
