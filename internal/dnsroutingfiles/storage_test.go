package dnsroutingfiles

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// **Feature: dns-routing-file-manager, Property 7: Directory initialization**
// **Validates: Requirements 3.5**
// For any File Manager initialization, if the rule directory does not exist,
// it should be created automatically.
func TestDirectoryInitialization(t *testing.T) {
	t.Parallel()

	// Create temporary directory for test
	tmpDir := t.TempDir()

	// Verify directory doesn't exist yet
	rulesDir := filepath.Join(tmpDir, rulesDirName)
	if _, err := os.Stat(rulesDir); !os.IsNotExist(err) {
		t.Fatal("rules directory should not exist before initialization")
	}

	// Create file manager (should create directory)
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

	// Verify directory was created
	if _, err := os.Stat(rulesDir); os.IsNotExist(err) {
		t.Fatal("rules directory should exist after initialization")
	}

	// Verify it's actually a directory
	info, err := os.Stat(rulesDir)
	if err != nil {
		t.Fatalf("failed to stat rules directory: %v", err)
	}

	if !info.IsDir() {
		t.Fatal("rules path should be a directory")
	}
}

// **Feature: dns-routing-file-manager, Property 5: File storage conventions**
// **Validates: Requirements 3.1, 3.2**
// For any domain list rule, the rule file should be stored in the dedicated directory
// with a filename matching the pattern {ID}.txt.
func TestFileStorageConventions(t *testing.T) {
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

	fm := m.(*fileManager)

	// Test various rule IDs
	testCases := []struct {
		ruleID       int64
		expectedName string
	}{
		{1, "1.txt"},
		{42, "42.txt"},
		{999, "999.txt"},
		{12345, "12345.txt"},
	}

	for _, tc := range testCases {
		t.Run(tc.expectedName, func(t *testing.T) {
			filePath := fm.getRuleFilePath(tc.ruleID)

			// Verify path is in the correct directory
			expectedDir := filepath.Join(tmpDir, rulesDirName)
			actualDir := filepath.Dir(filePath)
			if actualDir != expectedDir {
				t.Errorf("expected directory %s, got %s", expectedDir, actualDir)
			}

			// Verify filename matches pattern
			actualName := filepath.Base(filePath)
			if actualName != tc.expectedName {
				t.Errorf("expected filename %s, got %s", tc.expectedName, actualName)
			}
		})
	}

	// Test custom rules file path
	customRulesPath := fm.getCustomRulesFilePath()
	expectedCustomPath := filepath.Join(tmpDir, rulesDirName, customRulesFileName)
	if customRulesPath != expectedCustomPath {
		t.Errorf("expected custom rules path %s, got %s", expectedCustomPath, customRulesPath)
	}

	// Test metadata file path
	metadataPath := fm.getMetadataFilePath()
	expectedMetadataPath := filepath.Join(tmpDir, rulesDirName, metadataFileName)
	if metadataPath != expectedMetadataPath {
		t.Errorf("expected metadata path %s, got %s", expectedMetadataPath, metadataPath)
	}
}

// **Feature: dns-routing-file-manager, Property 9: Atomic write safety**
// **Validates: Requirements 4.5**
// For any file write operation, the file should either be completely written or not written at all,
// preventing partial writes that could corrupt data.
func TestAtomicWrites(t *testing.T) {
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

	fm := m.(*fileManager)

	// Test atomic write
	testFile := filepath.Join(tmpDir, "test_atomic.txt")
	testData := []byte("test data for atomic write")

	// Write data atomically
	if err := fm.writeFileAtomic(testFile, testData); err != nil {
		t.Fatalf("failed to write file atomically: %v", err)
	}

	// Verify file exists and contains correct data
	readData, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	if string(readData) != string(testData) {
		t.Errorf("file content mismatch: expected %q, got %q", testData, readData)
	}

	// Test overwriting existing file
	newData := []byte("new data for atomic write")
	if err := fm.writeFileAtomic(testFile, newData); err != nil {
		t.Fatalf("failed to overwrite file atomically: %v", err)
	}

	// Verify file was overwritten
	readData, err = os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("failed to read file after overwrite: %v", err)
	}

	if string(readData) != string(newData) {
		t.Errorf("file content mismatch after overwrite: expected %q, got %q", newData, readData)
	}
}

// Test custom rules save and load
func TestCustomRulesSaveLoad(t *testing.T) {
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

	fm := m.(*fileManager)
	ctx := context.Background()

	// Add test custom rules
	testRules := []*CustomRule{
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
		{
			Domain:        "keyword",
			MatchType:     "DOMAIN-KEYWORD",
			UpstreamGroup: "group3",
			Enabled:       true,
		},
	}

	fm.customRules = testRules

	// Save custom rules
	if err := fm.saveCustomRules(ctx); err != nil {
		t.Fatalf("failed to save custom rules: %v", err)
	}

	// Create new manager to load rules
	m2, err := NewManager(config)
	if err != nil {
		t.Fatalf("failed to create second manager: %v", err)
	}
	defer m2.Close()

	fm2 := m2.(*fileManager)

	// Load custom rules
	if err := fm2.loadCustomRules(ctx); err != nil {
		t.Fatalf("failed to load custom rules: %v", err)
	}

	// Verify rules were loaded correctly
	if len(fm2.customRules) != len(testRules) {
		t.Fatalf("expected %d custom rules, got %d", len(testRules), len(fm2.customRules))
	}

	for i, expectedRule := range testRules {
		actualRule := fm2.customRules[i]

		if actualRule.Domain != expectedRule.Domain ||
			actualRule.MatchType != expectedRule.MatchType ||
			actualRule.UpstreamGroup != expectedRule.UpstreamGroup ||
			actualRule.Enabled != expectedRule.Enabled {
			t.Errorf("custom rule %d mismatch: expected %+v, got %+v", i, expectedRule, actualRule)
		}
	}
}

// **Feature: dns-routing-file-manager, Property 8: Custom rules round-trip**
// **Validates: Requirements 4.3**
// For any set of custom rules, saving them to disk and then loading them back should preserve
// all rule properties (domain, match type, upstream group, enabled state).
func TestCustomRulesRoundTrip(t *testing.T) {
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

	fm := m.(*fileManager)
	ctx := context.Background()

	// Test with various custom rules
	testCases := []struct {
		name  string
		rules []*CustomRule
	}{
		{
			name: "single rule",
			rules: []*CustomRule{
				{Domain: "example.com", MatchType: "DOMAIN", UpstreamGroup: "group1", Enabled: true},
			},
		},
		{
			name: "multiple rules",
			rules: []*CustomRule{
				{Domain: "example.com", MatchType: "DOMAIN", UpstreamGroup: "group1", Enabled: true},
				{Domain: "test.org", MatchType: "DOMAIN-SUFFIX", UpstreamGroup: "group2", Enabled: false},
				{Domain: "keyword", MatchType: "DOMAIN-KEYWORD", UpstreamGroup: "group3", Enabled: true},
			},
		},
		{
			name:  "empty rules",
			rules: []*CustomRule{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fm.customRules = tc.rules

			// Save
			if err := fm.saveCustomRules(ctx); err != nil {
				t.Fatalf("failed to save custom rules: %v", err)
			}

			// Load
			fm.customRules = nil
			if err := fm.loadCustomRules(ctx); err != nil {
				t.Fatalf("failed to load custom rules: %v", err)
			}

			// Verify
			if len(fm.customRules) != len(tc.rules) {
				t.Fatalf("expected %d rules, got %d", len(tc.rules), len(fm.customRules))
			}

			for i, expectedRule := range tc.rules {
				actualRule := fm.customRules[i]

				if actualRule.Domain != expectedRule.Domain ||
					actualRule.MatchType != expectedRule.MatchType ||
					actualRule.UpstreamGroup != expectedRule.UpstreamGroup ||
					actualRule.Enabled != expectedRule.Enabled {
					t.Errorf("rule %d mismatch: expected %+v, got %+v", i, expectedRule, actualRule)
				}
			}
		})
	}
}

// Test LoadAll method
func TestLoadAll(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	config := Config{
		DataDir: tmpDir,
		Logger:  logger,
	}

	// Create first manager and add data
	m1, err := NewManager(config)
	if err != nil {
		t.Fatalf("failed to create first manager: %v", err)
	}

	fm1 := m1.(*fileManager)
	ctx := context.Background()

	// Add domain list rule
	testRule := &DomainListRule{
		ID:            1,
		Name:          "Test Rule",
		URL:           "https://example.com/rules.txt",
		UpstreamGroup: "test-group",
		Priority:      10,
		Enabled:       true,
		RulesCount:    100,
		LastUpdated:   time.Now(),
		FilePath:      filepath.Join(tmpDir, "dns_routing_rules", "1.txt"),
	}
	fm1.domainListRules[testRule.ID] = testRule

	// Add custom rule
	testCustomRule := &CustomRule{
		Domain:        "example.com",
		MatchType:     "DOMAIN-SUFFIX",
		UpstreamGroup: "test-group",
		Enabled:       true,
	}
	fm1.customRules = append(fm1.customRules, testCustomRule)

	// Save both
	if err := fm1.saveMetadata(ctx); err != nil {
		t.Fatalf("failed to save metadata: %v", err)
	}
	if err := fm1.saveCustomRules(ctx); err != nil {
		t.Fatalf("failed to save custom rules: %v", err)
	}

	m1.Close()

	// Create second manager and load all
	m2, err := NewManager(config)
	if err != nil {
		t.Fatalf("failed to create second manager: %v", err)
	}
	defer m2.Close()

	if err := m2.LoadAll(ctx); err != nil {
		t.Fatalf("failed to load all: %v", err)
	}

	fm2 := m2.(*fileManager)

	// Verify domain list rule
	if len(fm2.domainListRules) != 1 {
		t.Fatalf("expected 1 domain list rule, got %d", len(fm2.domainListRules))
	}

	loadedRule, exists := fm2.domainListRules[testRule.ID]
	if !exists {
		t.Fatal("test rule not found")
	}

	if loadedRule.Name != testRule.Name {
		t.Errorf("rule name mismatch: expected %s, got %s", testRule.Name, loadedRule.Name)
	}

	// Verify custom rule
	if len(fm2.customRules) != 1 {
		t.Fatalf("expected 1 custom rule, got %d", len(fm2.customRules))
	}

	if fm2.customRules[0].Domain != testCustomRule.Domain {
		t.Errorf("custom rule domain mismatch: expected %s, got %s",
			testCustomRule.Domain, fm2.customRules[0].Domain)
	}
}
