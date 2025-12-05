package dnsroutingfiles

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"testing/quick"
	"time"
)

// **Feature: dns-routing-file-manager, Property 19: Metadata persistence**
// **Validates: Requirements 10.1, 10.3**
// For any rule operation that modifies metadata (add, update), all metadata fields
// (ID, name, URL, upstream group, priority, enabled state) should be persisted to the metadata file.
func TestMetadataRoundTrip(t *testing.T) {
	t.Parallel()

	// Property test function
	f := func(numRules uint8, numCustomRules uint8) bool {
		// Create temporary directory for test
		tmpDir := t.TempDir()

		// Create file manager
		logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
		config := Config{
			DataDir: tmpDir,
			Logger:  logger,
		}

		m, err := NewManager(config)
		if err != nil {
			t.Logf("failed to create manager: %v", err)
			return false
		}
		defer m.Close()

		fm := m.(*fileManager)
		ctx := context.Background()

		// Generate random domain list rules
		for i := uint8(0); i < numRules && i < 10; i++ {
			rule := &DomainListRule{
				ID:            int64(i + 1),
				Name:          "Test Rule " + string(rune('A'+i)),
				URL:           "https://example.com/rule" + string(rune('0'+i)),
				UpstreamGroup: "group" + string(rune('0'+i)),
				Priority:      int(i),
				Enabled:       i%2 == 0,
				RulesCount:    int(i * 10),
				LastUpdated:   time.Now().Add(-time.Duration(i) * time.Hour),
				FilePath:      filepath.Join(tmpDir, "dns_routing_rules", string(rune('0'+i))+".txt"),
			}
			fm.domainListRules[rule.ID] = rule
		}

		// Generate random custom rules
		for i := uint8(0); i < numCustomRules && i < 10; i++ {
			matchTypes := []string{"DOMAIN", "DOMAIN-SUFFIX", "DOMAIN-KEYWORD"}
			rule := &CustomRule{
				Domain:        "example" + string(rune('0'+i)) + ".com",
				MatchType:     matchTypes[i%3],
				UpstreamGroup: "group" + string(rune('0'+i)),
				Enabled:       i%2 == 0,
			}
			fm.customRules = append(fm.customRules, rule)
		}

		// Save metadata
		if err := fm.saveMetadata(ctx); err != nil {
			t.Logf("failed to save metadata: %v", err)
			return false
		}

		// Create a new manager to load the metadata
		m2, err := NewManager(config)
		if err != nil {
			t.Logf("failed to create second manager: %v", err)
			return false
		}
		defer m2.Close()

		fm2 := m2.(*fileManager)

		// Load metadata
		if err := fm2.loadMetadata(ctx); err != nil {
			t.Logf("failed to load metadata: %v", err)
			return false
		}

		// Verify domain list rules
		if len(fm2.domainListRules) != len(fm.domainListRules) {
			t.Logf("domain list rules count mismatch: expected %d, got %d",
				len(fm.domainListRules), len(fm2.domainListRules))
			return false
		}

		for id, originalRule := range fm.domainListRules {
			loadedRule, exists := fm2.domainListRules[id]
			if !exists {
				t.Logf("rule with ID %d not found after loading", id)
				return false
			}

			// Compare all fields
			if loadedRule.ID != originalRule.ID ||
				loadedRule.Name != originalRule.Name ||
				loadedRule.URL != originalRule.URL ||
				loadedRule.UpstreamGroup != originalRule.UpstreamGroup ||
				loadedRule.Priority != originalRule.Priority ||
				loadedRule.Enabled != originalRule.Enabled ||
				loadedRule.RulesCount != originalRule.RulesCount ||
				loadedRule.FilePath != originalRule.FilePath {
				t.Logf("rule fields mismatch for ID %d", id)
				return false
			}

			// Check timestamp (allow small difference due to JSON serialization)
			timeDiff := loadedRule.LastUpdated.Sub(originalRule.LastUpdated)
			if timeDiff < -time.Second || timeDiff > time.Second {
				t.Logf("timestamp mismatch for ID %d: diff=%v", id, timeDiff)
				return false
			}
		}

		// Verify custom rules
		if len(fm2.customRules) != len(fm.customRules) {
			t.Logf("custom rules count mismatch: expected %d, got %d",
				len(fm.customRules), len(fm2.customRules))
			return false
		}

		for i, originalRule := range fm.customRules {
			loadedRule := fm2.customRules[i]

			if loadedRule.Domain != originalRule.Domain ||
				loadedRule.MatchType != originalRule.MatchType ||
				loadedRule.UpstreamGroup != originalRule.UpstreamGroup ||
				loadedRule.Enabled != originalRule.Enabled {
				t.Logf("custom rule fields mismatch at index %d", i)
				return false
			}
		}

		return true
	}

	// Run property test with 100 iterations
	config := &quick.Config{MaxCount: 100}
	if err := quick.Check(f, config); err != nil {
		t.Errorf("metadata round-trip property failed: %v", err)
	}
}

// **Feature: dns-routing-file-manager, Property 20: Metadata loading**
// **Validates: Requirements 10.2**
// For any File Manager restart, all rule metadata should be loaded from the metadata file
// and match what was persisted.
func TestMetadataLoading(t *testing.T) {
	t.Parallel()

	// Create temporary directory for test
	tmpDir := t.TempDir()

	// Create file manager
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	config := Config{
		DataDir: tmpDir,
		Logger:  logger,
	}

	m, err := NewManager(config)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	fm := m.(*fileManager)
	ctx := context.Background()

	// Add some test data
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
	fm.domainListRules[testRule.ID] = testRule

	testCustomRule := &CustomRule{
		Domain:        "example.com",
		MatchType:     "DOMAIN-SUFFIX",
		UpstreamGroup: "test-group",
		Enabled:       true,
	}
	fm.customRules = append(fm.customRules, testCustomRule)

	// Save metadata
	if err := fm.saveMetadata(ctx); err != nil {
		t.Fatalf("failed to save metadata: %v", err)
	}

	m.Close()

	// Create a new manager (simulating restart)
	m2, err := NewManager(config)
	if err != nil {
		t.Fatalf("failed to create second manager: %v", err)
	}
	defer m2.Close()

	fm2 := m2.(*fileManager)

	// Load metadata
	if err := fm2.loadMetadata(ctx); err != nil {
		t.Fatalf("failed to load metadata: %v", err)
	}

	// Verify domain list rule was loaded
	loadedRule, exists := fm2.domainListRules[testRule.ID]
	if !exists {
		t.Fatal("test rule not found after loading")
	}

	if loadedRule.Name != testRule.Name ||
		loadedRule.URL != testRule.URL ||
		loadedRule.UpstreamGroup != testRule.UpstreamGroup ||
		loadedRule.Priority != testRule.Priority ||
		loadedRule.Enabled != testRule.Enabled ||
		loadedRule.RulesCount != testRule.RulesCount {
		t.Error("loaded rule fields do not match original")
	}

	// Verify custom rule was loaded
	if len(fm2.customRules) != 1 {
		t.Fatalf("expected 1 custom rule, got %d", len(fm2.customRules))
	}

	loadedCustomRule := fm2.customRules[0]
	if loadedCustomRule.Domain != testCustomRule.Domain ||
		loadedCustomRule.MatchType != testCustomRule.MatchType ||
		loadedCustomRule.UpstreamGroup != testCustomRule.UpstreamGroup ||
		loadedCustomRule.Enabled != testCustomRule.Enabled {
		t.Error("loaded custom rule fields do not match original")
	}
}

// Test metadata loading with non-existent file
func TestMetadataLoadingNonExistent(t *testing.T) {
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

	// Load metadata (file doesn't exist)
	if err := fm.loadMetadata(ctx); err != nil {
		t.Errorf("loading non-existent metadata should not error: %v", err)
	}

	// Verify empty state
	if len(fm.domainListRules) != 0 {
		t.Errorf("expected 0 domain list rules, got %d", len(fm.domainListRules))
	}

	if len(fm.customRules) != 0 {
		t.Errorf("expected 0 custom rules, got %d", len(fm.customRules))
	}
}
