package dnsrouting

import (
	"context"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRouter_Match(t *testing.T) {
	logger := slog.Default()
	router := NewRouter(logger)

	// Add custom rules
	router.SetCustomRules([]Rule{
		{
			Domain:        "example.com",
			MatchType:     MatchTypeDomain,
			UpstreamGroup: "custom-dns",
			Enabled:       true,
		},
		{
			Domain:        "google.com",
			MatchType:     MatchTypeDomainSuffix,
			UpstreamGroup: "international-dns",
			Enabled:       true,
		},
	})

	// Add source rules
	source := &RuleSource{
		ID:            1,
		Name:          "China Domains",
		UpstreamGroup: "china-dns",
		Priority:      10,
		Enabled:       true,
		Rules: []Rule{
			{
				Domain:        "baidu.com",
				MatchType:     MatchTypeDomainSuffix,
				UpstreamGroup: "china-dns",
				Enabled:       true,
			},
			{
				Domain:        "taobao",
				MatchType:     MatchTypeDomainKeyword,
				UpstreamGroup: "china-dns",
				Enabled:       true,
			},
		},
		RulesCount: 2,
	}
	err := router.AddSource(source)
	require.NoError(t, err)

	ctx := context.Background()

	testCases := []struct {
		name          string
		domain        string
		expectMatch   bool
		expectGroup   string
	}{
		{
			name:        "exact_match_custom",
			domain:      "example.com",
			expectMatch: true,
			expectGroup: "custom-dns",
		},
		{
			name:        "suffix_match_custom",
			domain:      "www.google.com",
			expectMatch: true,
			expectGroup: "international-dns",
		},
		{
			name:        "suffix_match_source",
			domain:      "www.baidu.com",
			expectMatch: true,
			expectGroup: "china-dns",
		},
		{
			name:        "keyword_match",
			domain:      "www.taobao.com",
			expectMatch: true,
			expectGroup: "china-dns",
		},
		{
			name:        "no_match",
			domain:      "unknown.com",
			expectMatch: false,
			expectGroup: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			group, matched := router.Match(ctx, tc.domain)
			assert.Equal(t, tc.expectMatch, matched)
			if matched {
				assert.Equal(t, tc.expectGroup, group)
			}
		})
	}
}

func TestRouter_Priority(t *testing.T) {
	logger := slog.Default()
	router := NewRouter(logger)

	// Add two sources with different priorities
	source1 := &RuleSource{
		ID:            1,
		Name:          "Low Priority",
		UpstreamGroup: "low-priority-dns",
		Priority:      20,
		Enabled:       true,
		Rules: []Rule{
			{
				Domain:        "example.com",
				MatchType:     MatchTypeDomainSuffix,
				UpstreamGroup: "low-priority-dns",
				Enabled:       true,
			},
		},
		RulesCount: 1,
	}

	source2 := &RuleSource{
		ID:            2,
		Name:          "High Priority",
		UpstreamGroup: "high-priority-dns",
		Priority:      10,
		Enabled:       true,
		Rules: []Rule{
			{
				Domain:        "example.com",
				MatchType:     MatchTypeDomainSuffix,
				UpstreamGroup: "high-priority-dns",
				Enabled:       true,
			},
		},
		RulesCount: 1,
	}

	err := router.AddSource(source1)
	require.NoError(t, err)
	err = router.AddSource(source2)
	require.NoError(t, err)

	ctx := context.Background()

	// Should match high priority source
	group, matched := router.Match(ctx, "www.example.com")
	assert.True(t, matched)
	assert.Equal(t, "high-priority-dns", group)
}

func TestRouter_CustomRulesPriority(t *testing.T) {
	logger := slog.Default()
	router := NewRouter(logger)

	// Add source rule
	source := &RuleSource{
		ID:            1,
		Name:          "Source",
		UpstreamGroup: "source-dns",
		Priority:      10,
		Enabled:       true,
		Rules: []Rule{
			{
				Domain:        "example.com",
				MatchType:     MatchTypeDomainSuffix,
				UpstreamGroup: "source-dns",
				Enabled:       true,
			},
		},
		RulesCount: 1,
	}
	err := router.AddSource(source)
	require.NoError(t, err)

	// Add custom rule (should have higher priority)
	router.SetCustomRules([]Rule{
		{
			Domain:        "example.com",
			MatchType:     MatchTypeDomain,
			UpstreamGroup: "custom-dns",
			Enabled:       true,
		},
	})

	ctx := context.Background()

	// Should match custom rule (higher priority)
	group, matched := router.Match(ctx, "example.com")
	assert.True(t, matched)
	assert.Equal(t, "custom-dns", group)
}
