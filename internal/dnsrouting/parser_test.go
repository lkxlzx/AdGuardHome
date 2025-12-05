package dnsrouting

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParser_DetectFormat(t *testing.T) {
	testCases := []struct {
		name     string
		content  string
		expected RuleFormat
	}{
		{
			name: "clash_format",
			content: `payload:
  - DOMAIN,example.com
  - DOMAIN-SUFFIX,google.com`,
			expected: RuleFormatClash,
		},
		{
			name: "gfwlist_format",
			content: `[AutoProxy 0.2.9]
||google.com
||youtube.com`,
			expected: RuleFormatGFWList,
		},
		{
			name: "adguard_format",
			content: `! Title: Test List
||example.com^
||google.com^`,
			expected: RuleFormatAdGuard,
		},
	}

	p := NewParser()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			format, err := p.DetectFormat(strings.NewReader(tc.content))
			require.NoError(t, err)
			assert.Equal(t, tc.expected, format)
		})
	}
}

func TestParser_ParseClash(t *testing.T) {
	content := `payload:
  - DOMAIN,example.com
  - DOMAIN-SUFFIX,google.com
  - DOMAIN-KEYWORD,youtube
  # Comment line
  - DOMAIN-SUFFIX,facebook.com`

	p := NewParser()
	result, err := p.Parse(strings.NewReader(content))

	require.NoError(t, err)
	assert.Equal(t, RuleFormatClash, result.Format)
	assert.Equal(t, 4, result.ValidRules)
	assert.Len(t, result.Rules, 4)

	// Check first rule
	assert.Equal(t, "example.com", result.Rules[0].Domain)
	assert.Equal(t, "DOMAIN", result.Rules[0].MatchType)

	// Check second rule
	assert.Equal(t, "google.com", result.Rules[1].Domain)
	assert.Equal(t, "DOMAIN-SUFFIX", result.Rules[1].MatchType)
}

func TestParser_ParseGFWList(t *testing.T) {
	content := `[AutoProxy 0.2.9]
! Comment
||google.com
||youtube.com^
.facebook.com
|https://twitter.com/`

	p := NewParser()
	result, err := p.Parse(strings.NewReader(content))

	require.NoError(t, err)
	assert.Equal(t, RuleFormatGFWList, result.Format)
	assert.Greater(t, result.ValidRules, 0)
	assert.Greater(t, len(result.Rules), 0)
}

func TestParser_ParseAdGuard(t *testing.T) {
	content := `! Title: Test List
! Description: Test
||example.com^
||google.com^
|youtube.com^`

	p := NewParser()
	result, err := p.Parse(strings.NewReader(content))

	require.NoError(t, err)
	assert.Equal(t, RuleFormatAdGuard, result.Format)
	assert.Equal(t, 3, result.ValidRules)
	assert.Len(t, result.Rules, 3)
}

func TestConvertToAdGuardFormat(t *testing.T) {
	rules := []ParsedRule{
		{Domain: "example.com", MatchType: "DOMAIN"},
		{Domain: "google.com", MatchType: "DOMAIN-SUFFIX"},
		{Domain: "youtube", MatchType: "DOMAIN-KEYWORD"},
	}

	result := ConvertToAdGuardFormat(rules)

	assert.Len(t, result, 3)
	assert.Contains(t, result, "|example.com^")
	assert.Contains(t, result, "||google.com^")
	assert.Contains(t, result, "||youtube^")
}

func TestConvertToAdGuardFormat_Deduplication(t *testing.T) {
	rules := []ParsedRule{
		{Domain: "example.com", MatchType: "DOMAIN-SUFFIX"},
		{Domain: "example.com", MatchType: "DOMAIN-SUFFIX"}, // Duplicate
		{Domain: "google.com", MatchType: "DOMAIN-SUFFIX"},
	}

	result := ConvertToAdGuardFormat(rules)

	assert.Len(t, result, 2)
}
