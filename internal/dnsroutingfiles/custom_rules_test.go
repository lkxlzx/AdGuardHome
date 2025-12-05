package dnsroutingfiles

import (
	"context"
	"log/slog"
	"os"
	"testing"
)

// Test AddCustomRule
func TestAddCustomRule(t *testing.T) {
	t.Parallel()

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

	// Add custom rule
	rule := &CustomRule{
		Domain:        "example.com",
		MatchType:     "DOMAIN-SUFFIX",
		UpstreamGroup: "test-group",
		Enabled:       true,
	}

	if err := m.AddCustomRule(ctx, rule); err != nil {
		t.Fatalf("failed to add custom rule: %v", err)
	}

	// Verify rule was added
	rules := m.GetAllCustomRules()
	if len(rules) != 1 {
		t.Fatalf("expected 1 custom rule, got %d", len(rules))
	}

	if rules[0].Domain != rule.Domain {
		t.Errorf("domain mismatch: expected %s, got %s", rule.Domain, rules[0].Domain)
	}
}

// Test UpdateCustomRule
func TestUpdateCustomRule(t *testing.T) {
	t.Parallel()

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
	rule := &CustomRule{
		Domain:        "example.com",
		MatchType:     "DOMAIN",
		UpstreamGroup: "test-group",
		Enabled:       true,
	}

	if err := m.AddCustomRule(ctx, rule); err != nil {
		t.Fatalf("failed to add custom rule: %v", err)
	}

	// Update rule
	updatedRule := &CustomRule{
		Domain:        "updated.com",
		MatchType:     "DOMAIN-SUFFIX",
		UpstreamGroup: "new-group",
		Enabled:       false,
	}

	if err := m.UpdateCustomRule(ctx, "example.com", "DOMAIN", updatedRule); err != nil {
		t.Fatalf("failed to update custom rule: %v", err)
	}

	// Verify update
	rules := m.GetAllCustomRules()
	if len(rules) != 1 {
		t.Fatalf("expected 1 custom rule, got %d", len(rules))
	}

	if rules[0].Domain != "updated.com" {
		t.Errorf("domain not updated: expected 'updated.com', got %s", rules[0].Domain)
	}

	if rules[0].UpstreamGroup != "new-group" {
		t.Errorf("upstream group not updated: expected 'new-group', got %s", rules[0].UpstreamGroup)
	}
}

// Test DeleteCustomRule
func TestDeleteCustomRule(t *testing.T) {
	t.Parallel()

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

	// Add rules
	rule1 := &CustomRule{
		Domain:        "example.com",
		MatchType:     "DOMAIN",
		UpstreamGroup: "test-group",
		Enabled:       true,
	}

	rule2 := &CustomRule{
		Domain:        "test.org",
		MatchType:     "DOMAIN-SUFFIX",
		UpstreamGroup: "test-group",
		Enabled:       true,
	}

	if err := m.AddCustomRule(ctx, rule1); err != nil {
		t.Fatalf("failed to add rule1: %v", err)
	}

	if err := m.AddCustomRule(ctx, rule2); err != nil {
		t.Fatalf("failed to add rule2: %v", err)
	}

	// Verify 2 rules
	rules := m.GetAllCustomRules()
	if len(rules) != 2 {
		t.Fatalf("expected 2 custom rules, got %d", len(rules))
	}

	// Delete first rule
	if err := m.DeleteCustomRule(ctx, "example.com", "DOMAIN"); err != nil {
		t.Fatalf("failed to delete custom rule: %v", err)
	}

	// Verify only 1 rule remains
	rules = m.GetAllCustomRules()
	if len(rules) != 1 {
		t.Fatalf("expected 1 custom rule after deletion, got %d", len(rules))
	}

	if rules[0].Domain != "test.org" {
		t.Errorf("wrong rule remained: expected 'test.org', got %s", rules[0].Domain)
	}
}

// Test custom rule validation
func TestCustomRuleValidation(t *testing.T) {
	t.Parallel()

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

	// Test empty domain
	rule := &CustomRule{
		Domain:        "",
		MatchType:     "DOMAIN",
		UpstreamGroup: "test-group",
		Enabled:       true,
	}

	if err := m.AddCustomRule(ctx, rule); err == nil {
		t.Error("expected error for empty domain")
	}

	// Test empty match type
	rule = &CustomRule{
		Domain:        "example.com",
		MatchType:     "",
		UpstreamGroup: "test-group",
		Enabled:       true,
	}

	if err := m.AddCustomRule(ctx, rule); err == nil {
		t.Error("expected error for empty match type")
	}

	// Test invalid match type
	rule = &CustomRule{
		Domain:        "example.com",
		MatchType:     "INVALID",
		UpstreamGroup: "test-group",
		Enabled:       true,
	}

	if err := m.AddCustomRule(ctx, rule); err == nil {
		t.Error("expected error for invalid match type")
	}

	// Test empty upstream group
	rule = &CustomRule{
		Domain:        "example.com",
		MatchType:     "DOMAIN",
		UpstreamGroup: "",
		Enabled:       true,
	}

	if err := m.AddCustomRule(ctx, rule); err == nil {
		t.Error("expected error for empty upstream group")
	}

	// Test duplicate rule
	rule = &CustomRule{
		Domain:        "example.com",
		MatchType:     "DOMAIN",
		UpstreamGroup: "test-group",
		Enabled:       true,
	}

	if err := m.AddCustomRule(ctx, rule); err != nil {
		t.Fatalf("failed to add first rule: %v", err)
	}

	if err := m.AddCustomRule(ctx, rule); err == nil {
		t.Error("expected error for duplicate rule")
	}
}

// Test custom rule persistence
func TestCustomRulePersistence(t *testing.T) {
	t.Parallel()

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

	ctx := context.Background()

	// Add custom rules
	rules := []*CustomRule{
		{
			Domain:        "example.com",
			MatchType:     "DOMAIN",
			UpstreamGroup: "group1",
			Enabled:       true,
		},
		{
			Domain:        "test.org",
			MatchType:     "DOMAIN-SUFFIX",
			UpstreamGroup: "group2",
			Enabled:       false,
		},
	}

	for _, rule := range rules {
		if err := m.AddCustomRule(ctx, rule); err != nil {
			t.Fatalf("failed to add rule: %v", err)
		}
	}

	m.Close()

	// Create new manager and load
	m2, err := NewManager(config)
	if err != nil {
		t.Fatalf("failed to create second manager: %v", err)
	}
	defer m2.Close()

	if err := m2.LoadAll(ctx); err != nil {
		t.Fatalf("failed to load all: %v", err)
	}

	// Verify rules were persisted
	loadedRules := m2.GetAllCustomRules()
	if len(loadedRules) != len(rules) {
		t.Fatalf("expected %d rules, got %d", len(rules), len(loadedRules))
	}

	for i, expectedRule := range rules {
		actualRule := loadedRules[i]
		if actualRule.Domain != expectedRule.Domain ||
			actualRule.MatchType != expectedRule.MatchType ||
			actualRule.UpstreamGroup != expectedRule.UpstreamGroup ||
			actualRule.Enabled != expectedRule.Enabled {
			t.Errorf("rule %d mismatch: expected %+v, got %+v", i, expectedRule, actualRule)
		}
	}
}
