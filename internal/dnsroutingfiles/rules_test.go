package dnsroutingfiles

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

// Test AddDomainListRule
func TestAddDomainListRule(t *testing.T) {
	t.Parallel()

	// Create test server with rule data
	testRuleData := `payload:
  - DOMAIN,example.com
  - DOMAIN-SUFFIX,test.org
  - DOMAIN-KEYWORD,keyword`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(testRuleData))
	}))
	defer server.Close()

	tmpDir := t.TempDir()
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	config := Config{
		DataDir: tmpDir,
		Logger:  logger,
	}

	m, err := NewManager(config)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}
	defer m.Close()

	ctx := context.Background()

	// Add rule
	rule := &DomainListRule{
		ID:            1,
		Name:          "Test Rule",
		URL:           server.URL,
		UpstreamGroup: "test-group",
		Priority:      10,
		Enabled:       true,
	}

	if err := m.AddDomainListRule(ctx, rule); err != nil {
		t.Fatalf("failed to add rule: %v", err)
	}

	// Verify rule was added
	loadedRule, err := m.GetDomainListRule(1)
	if err != nil {
		t.Fatalf("failed to get rule: %v", err)
	}

	if loadedRule.Name != rule.Name {
		t.Errorf("name mismatch: expected %s, got %s", rule.Name, loadedRule.Name)
	}

	if loadedRule.RulesCount != 3 {
		t.Errorf("expected 3 rules, got %d", loadedRule.RulesCount)
	}
}

// Test UpdateDomainListRule
func TestUpdateDomainListRule(t *testing.T) {
	t.Parallel()

	testRuleData1 := `payload:
  - DOMAIN,example.com`
	testRuleData2 := `payload:
  - DOMAIN,example.com
  - DOMAIN,test.org`

	server1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(testRuleData1))
	}))
	defer server1.Close()

	server2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(testRuleData2))
	}))
	defer server2.Close()

	tmpDir := t.TempDir()
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	config := Config{
		DataDir: tmpDir,
		Logger:  logger,
	}

	m, err := NewManager(config)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}
	defer m.Close()

	ctx := context.Background()

	// Add initial rule
	rule := &DomainListRule{
		ID:            1,
		Name:          "Test Rule",
		URL:           server1.URL,
		UpstreamGroup: "test-group",
		Priority:      10,
		Enabled:       true,
	}

	if err := m.AddDomainListRule(ctx, rule); err != nil {
		t.Fatalf("failed to add rule: %v", err)
	}

	// Update rule with new URL
	updatedRule := &DomainListRule{
		ID:            1,
		Name:          "Updated Rule",
		URL:           server2.URL,
		UpstreamGroup: "new-group",
		Priority:      20,
		Enabled:       false,
	}

	if err := m.UpdateDomainListRule(ctx, updatedRule); err != nil {
		t.Fatalf("failed to update rule: %v", err)
	}

	// Verify update
	loadedRule, err := m.GetDomainListRule(1)
	if err != nil {
		t.Fatalf("failed to get rule: %v", err)
	}

	if loadedRule.Name != "Updated Rule" {
		t.Errorf("name not updated: expected 'Updated Rule', got %s", loadedRule.Name)
	}

	if loadedRule.RulesCount != 2 {
		t.Errorf("expected 2 rules after update, got %d", loadedRule.RulesCount)
	}
}

// **Feature: dns-routing-file-manager, Property 6: File cleanup on deletion**
// **Validates: Requirements 3.4**
// For any rule deletion, the corresponding rule file should be removed from disk,
// leaving no orphaned files.
func TestDeleteDomainListRule(t *testing.T) {
	t.Parallel()

	testRuleData := `payload:
  - DOMAIN,example.com`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(testRuleData))
	}))
	defer server.Close()

	tmpDir := t.TempDir()
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	config := Config{
		DataDir: tmpDir,
		Logger:  logger,
	}

	m, err := NewManager(config)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}
	defer m.Close()

	ctx := context.Background()

	// Add rule
	rule := &DomainListRule{
		ID:            1,
		Name:          "Test Rule",
		URL:           server.URL,
		UpstreamGroup: "test-group",
		Priority:      10,
		Enabled:       true,
	}

	if err := m.AddDomainListRule(ctx, rule); err != nil {
		t.Fatalf("failed to add rule: %v", err)
	}

	fm := m.(*fileManager)
	filePath := fm.getRuleFilePath(1)

	// Verify file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Fatal("rule file should exist after adding")
	}

	// Delete rule
	if err := m.DeleteDomainListRule(ctx, 1); err != nil {
		t.Fatalf("failed to delete rule: %v", err)
	}

	// Verify file was removed
	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		t.Error("rule file should be removed after deletion")
	}

	// Verify rule is gone from manager
	_, err = m.GetDomainListRule(1)
	if err == nil {
		t.Error("expected error when getting deleted rule")
	}
}

// **Feature: dns-routing-file-manager, Property 17: Refresh re-downloads**
// **Validates: Requirements 9.4**
// For any rule refresh operation, the rule file should be re-downloaded from its URL,
// replacing the existing file.
func TestRefreshDomainListRule(t *testing.T) {
	t.Parallel()

	callCount := 0
	testRuleData1 := `payload:
  - DOMAIN,example.com`
	testRuleData2 := `payload:
  - DOMAIN,example.com
  - DOMAIN,test.org
  - DOMAIN,another.com`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		callCount++
		if callCount == 1 {
			w.Write([]byte(testRuleData1))
		} else {
			w.Write([]byte(testRuleData2))
		}
	}))
	defer server.Close()

	tmpDir := t.TempDir()
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	config := Config{
		DataDir: tmpDir,
		Logger:  logger,
	}

	m, err := NewManager(config)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}
	defer m.Close()

	ctx := context.Background()

	// Add rule
	rule := &DomainListRule{
		ID:            1,
		Name:          "Test Rule",
		URL:           server.URL,
		UpstreamGroup: "test-group",
		Priority:      10,
		Enabled:       true,
	}

	if err := m.AddDomainListRule(ctx, rule); err != nil {
		t.Fatalf("failed to add rule: %v", err)
	}

	// Verify initial rules count
	loadedRule, err := m.GetDomainListRule(1)
	if err != nil {
		t.Fatalf("failed to get rule: %v", err)
	}

	if loadedRule.RulesCount != 1 {
		t.Errorf("expected 1 rule initially, got %d", loadedRule.RulesCount)
	}

	initialTime := loadedRule.LastUpdated

	// Wait a bit to ensure timestamp changes
	time.Sleep(10 * time.Millisecond)

	// Refresh rule
	if err := m.RefreshDomainListRule(ctx, 1); err != nil {
		t.Fatalf("failed to refresh rule: %v", err)
	}

	// Verify rules count updated
	loadedRule, err = m.GetDomainListRule(1)
	if err != nil {
		t.Fatalf("failed to get rule after refresh: %v", err)
	}

	if loadedRule.RulesCount != 3 {
		t.Errorf("expected 3 rules after refresh, got %d", loadedRule.RulesCount)
	}

	// Verify timestamp updated
	if !loadedRule.LastUpdated.After(initialTime) {
		t.Error("last updated timestamp should be updated after refresh")
	}
}

// **Feature: dns-routing-file-manager, Property 21: Timestamp updates**
// **Validates: Requirements 10.4**
// For any rule refresh operation, the last_updated timestamp in the metadata should be
// updated to reflect the refresh time.
func TestTimestampUpdates(t *testing.T) {
	t.Parallel()

	testRuleData := `payload:
  - DOMAIN,example.com`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(testRuleData))
	}))
	defer server.Close()

	tmpDir := t.TempDir()
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	config := Config{
		DataDir: tmpDir,
		Logger:  logger,
	}

	m, err := NewManager(config)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}
	defer m.Close()

	ctx := context.Background()

	// Add rule
	rule := &DomainListRule{
		ID:            1,
		Name:          "Test Rule",
		URL:           server.URL,
		UpstreamGroup: "test-group",
		Priority:      10,
		Enabled:       true,
	}

	if err := m.AddDomainListRule(ctx, rule); err != nil {
		t.Fatalf("failed to add rule: %v", err)
	}

	// Get initial timestamp
	loadedRule, err := m.GetDomainListRule(1)
	if err != nil {
		t.Fatalf("failed to get rule: %v", err)
	}

	initialTime := loadedRule.LastUpdated

	// Wait to ensure timestamp difference
	time.Sleep(10 * time.Millisecond)

	// Refresh
	if err := m.RefreshDomainListRule(ctx, 1); err != nil {
		t.Fatalf("failed to refresh rule: %v", err)
	}

	// Verify timestamp updated
	loadedRule, err = m.GetDomainListRule(1)
	if err != nil {
		t.Fatalf("failed to get rule after refresh: %v", err)
	}

	if !loadedRule.LastUpdated.After(initialTime) {
		t.Errorf("timestamp should be updated: initial=%v, after=%v",
			initialTime, loadedRule.LastUpdated)
	}

	// Verify metadata was saved with new timestamp
	m2, err := NewManager(config)
	if err != nil {
		t.Fatalf("failed to create second manager: %v", err)
	}
	defer m2.Close()

	if err := m2.LoadAll(ctx); err != nil {
		t.Fatalf("failed to load all: %v", err)
	}

	reloadedRule, err := m2.GetDomainListRule(1)
	if err != nil {
		t.Fatalf("failed to get rule from reloaded manager: %v", err)
	}

	timeDiff := reloadedRule.LastUpdated.Sub(loadedRule.LastUpdated)
	if timeDiff < -time.Second || timeDiff > time.Second {
		t.Errorf("timestamp mismatch after reload: diff=%v", timeDiff)
	}
}
