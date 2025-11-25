package filtering

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestDownloadWithRetry_Success(t *testing.T) {
	// Create a test server that returns valid content
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test content"))
	}))
	defer server.Close()

	content, err := downloadWithRetry(server.URL, maxRetries)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if string(content) != "test content" {
		t.Errorf("Expected 'test content', got: %s", string(content))
	}
}

func TestDownloadWithRetry_SuccessAfterRetry(t *testing.T) {
	// Create a test server that fails first, then succeeds
	var attemptCount int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := atomic.AddInt32(&attemptCount, 1)
		if count < 2 {
			// Fail on first attempt
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		// Succeed on second attempt
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success after retry"))
	}))
	defer server.Close()

	content, err := downloadWithRetry(server.URL, maxRetries)
	if err != nil {
		t.Fatalf("Expected no error after retry, got: %v", err)
	}

	if string(content) != "success after retry" {
		t.Errorf("Expected 'success after retry', got: %s", string(content))
	}

	if atomic.LoadInt32(&attemptCount) != 2 {
		t.Errorf("Expected 2 attempts, got: %d", attemptCount)
	}
}

func TestDownloadWithRetry_MaxRetriesExceeded(t *testing.T) {
	// Create a test server that always fails
	var attemptCount int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&attemptCount, 1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	_, err := downloadWithRetry(server.URL, maxRetries)
	if err == nil {
		t.Fatal("Expected error after max retries, got nil")
	}

	if atomic.LoadInt32(&attemptCount) != int32(maxRetries) {
		t.Errorf("Expected %d attempts, got: %d", maxRetries, attemptCount)
	}

	// Check error message contains attempt information
	if !strings.Contains(err.Error(), fmt.Sprintf("attempt %d/%d", maxRetries, maxRetries)) {
		t.Errorf("Error message should contain attempt information: %v", err)
	}
}

func TestDownloadWithRetry_FileTooLarge(t *testing.T) {
	// Create a test server that returns content larger than limit
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", maxClashRuleFileSize+1))
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	_, err := downloadWithRetry(server.URL, maxRetries)
	if err == nil {
		t.Fatal("Expected error for file too large, got nil")
	}

	if !strings.Contains(err.Error(), "file too large") {
		t.Errorf("Expected 'file too large' error, got: %v", err)
	}
}

func TestDownloadWithRetry_ContentSizeExceeded(t *testing.T) {
	// Create a test server that returns content larger than limit
	largeContent := make([]byte, maxClashRuleFileSize+1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(largeContent)
	}))
	defer server.Close()

	_, err := downloadWithRetry(server.URL, maxRetries)
	if err == nil {
		t.Fatal("Expected error for content size exceeded, got nil")
	}

	if !strings.Contains(err.Error(), "file too large") {
		t.Errorf("Expected 'file too large' error, got: %v", err)
	}
}

func TestDownloadWithRetry_Timeout(t *testing.T) {
	// Create a test server that delays response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(clashRuleDownloadTimeout + 1*time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	_, err := downloadWithRetry(server.URL, 1) // Only 1 retry for faster test
	if err == nil {
		t.Fatal("Expected timeout error, got nil")
	}

	if !strings.Contains(err.Error(), "attempt") {
		t.Errorf("Expected timeout error with attempt info, got: %v", err)
	}
}

func TestParseClashRules_ValidYAML(t *testing.T) {
	yamlContent := `payload:
  - DOMAIN,example.com
  - DOMAIN-SUFFIX,test.com
  - DOMAIN-KEYWORD,google
  - IP-CIDR,192.168.1.0/24
`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(yamlContent))
	}))
	defer server.Close()

	domains, stats, err := ParseClashRules(server.URL)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if stats.TotalRules != 4 {
		t.Errorf("Expected 4 total rules, got: %d", stats.TotalRules)
	}

	if stats.ValidDomains != 3 {
		t.Errorf("Expected 3 valid domains, got: %d", stats.ValidDomains)
	}

	if stats.IPRules != 1 {
		t.Errorf("Expected 1 IP rule, got: %d", stats.IPRules)
	}

	if len(domains) != 3 {
		t.Errorf("Expected 3 domains, got: %d", len(domains))
	}
}

func TestParseClashRules_InvalidYAML(t *testing.T) {
	invalidYAML := `invalid yaml content {{{`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(invalidYAML))
	}))
	defer server.Close()

	_, _, err := ParseClashRules(server.URL)
	if err == nil {
		t.Fatal("Expected error for invalid YAML, got nil")
	}

	if !strings.Contains(err.Error(), "parsing YAML") {
		t.Errorf("Expected 'parsing YAML' error, got: %v", err)
	}
}

func TestParseClashRules_NetworkError(t *testing.T) {
	// Use an invalid URL to trigger network error
	_, _, err := ParseClashRules("http://invalid-url-that-does-not-exist-12345.com")
	if err == nil {
		t.Fatal("Expected network error, got nil")
	}

	if !strings.Contains(err.Error(), "downloading rules") {
		t.Errorf("Expected 'downloading rules' error, got: %v", err)
	}
}

func TestConvertClashRuleToAdGuardFormat(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		valid    bool
	}{
		{
			name:     "DOMAIN rule",
			input:    "DOMAIN,example.com",
			expected: "example.com",
			valid:    true,
		},
		{
			name:     "DOMAIN-SUFFIX rule",
			input:    "DOMAIN-SUFFIX,test.com",
			expected: "||test.com^",
			valid:    true,
		},
		{
			name:     "DOMAIN-KEYWORD rule",
			input:    "DOMAIN-KEYWORD,google",
			expected: "*google*",
			valid:    true,
		},
		{
			name:     "IP-CIDR rule (should be skipped)",
			input:    "IP-CIDR,192.168.1.0/24",
			expected: "",
			valid:    false,
		},
		{
			name:     "Invalid format",
			input:    "INVALID",
			expected: "",
			valid:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, valid := ConvertClashRuleToAdGuardFormat(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
			if valid != tt.valid {
				t.Errorf("Expected valid=%v, got %v", tt.valid, valid)
			}
		})
	}
}

func TestIsClashRuleURL(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected bool
	}{
		{
			name:     "URL with /Clash/",
			url:      "https://example.com/Clash/rules.yaml",
			expected: true,
		},
		{
			name:     "URL with /clash/",
			url:      "https://example.com/clash/rules.yaml",
			expected: true,
		},
		{
			name:     "URL ending with .yaml",
			url:      "https://example.com/rules.yaml",
			expected: true,
		},
		{
			name:     "URL ending with .yml",
			url:      "https://example.com/rules.yml",
			expected: true,
		},
		{
			name:     "Regular URL",
			url:      "https://example.com/rules.txt",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsClashRuleURL(tt.url)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}
