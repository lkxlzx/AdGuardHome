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

// **Feature: dns-routing-file-manager, Property 15: Download functionality**
// **Validates: Requirements 9.1**
// For any domain list rule with a valid URL, adding the rule should result in
// the rule file being downloaded from the URL and saved to disk.
func TestDownloadFunctionality(t *testing.T) {
	t.Parallel()

	// Create test HTTP server
	testData := []byte("test rule data\nline 2\nline 3")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(testData)
	}))
	defer server.Close()

	// Create file manager
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

	// Download file
	data, err := fm.downloadRuleFile(ctx, server.URL)
	if err != nil {
		t.Fatalf("failed to download file: %v", err)
	}

	// Verify data
	if string(data) != string(testData) {
		t.Errorf("downloaded data mismatch: expected %q, got %q", testData, data)
	}
}

// **Feature: dns-routing-file-manager, Property 18: Protocol support**
// **Validates: Requirements 9.5**
// For any rule URL using HTTP or HTTPS protocol, the File Manager should successfully
// download the rule file.
func TestProtocolSupport(t *testing.T) {
	t.Parallel()

	testData := []byte("test data")

	// Test HTTP
	t.Run("HTTP", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write(testData)
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

		fm := m.(*fileManager)
		ctx := context.Background()

		data, err := fm.downloadRuleFile(ctx, server.URL)
		if err != nil {
			t.Fatalf("failed to download via HTTP: %v", err)
		}

		if string(data) != string(testData) {
			t.Errorf("data mismatch: expected %q, got %q", testData, data)
		}
	})

	// Test HTTPS
	t.Run("HTTPS", func(t *testing.T) {
		server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write(testData)
		}))
		defer server.Close()

		tmpDir := t.TempDir()
		logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
		config := Config{
			DataDir:    tmpDir,
			Logger:     logger,
			HTTPClient: server.Client(), // Use test server's client which trusts its certificate
		}

		m, err := NewManager(config)
		if err != nil {
			t.Fatalf("failed to create manager: %v", err)
		}
		defer m.Close()

		fm := m.(*fileManager)
		ctx := context.Background()

		data, err := fm.downloadRuleFile(ctx, server.URL)
		if err != nil {
			t.Fatalf("failed to download via HTTPS: %v", err)
		}

		if string(data) != string(testData) {
			t.Errorf("data mismatch: expected %q, got %q", testData, data)
		}
	})
}

// **Feature: dns-routing-file-manager, Property 16: Download error handling**
// **Validates: Requirements 9.2**
// For any download operation that fails (timeout, network error, invalid response),
// the File Manager should return a clear error and not create a partial or corrupted file.
func TestDownloadErrorHandling(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	// Test 404 error
	t.Run("404 Not Found", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer server.Close()

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

		_, err = fm.downloadRuleFile(ctx, server.URL)
		if err == nil {
			t.Error("expected error for 404 response, got nil")
		}
	})

	// Test 500 error
	t.Run("500 Internal Server Error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

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

		_, err = fm.downloadRuleFile(ctx, server.URL)
		if err == nil {
			t.Error("expected error for 500 response, got nil")
		}
	})

	// Test timeout
	t.Run("Timeout", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Simulate slow response
			time.Sleep(2 * time.Second)
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		// Create client with very short timeout
		client := &http.Client{
			Timeout: 100 * time.Millisecond,
		}

		config := Config{
			DataDir:    tmpDir,
			Logger:     logger,
			HTTPClient: client,
		}

		m, err := NewManager(config)
		if err != nil {
			t.Fatalf("failed to create manager: %v", err)
		}
		defer m.Close()

		fm := m.(*fileManager)
		ctx := context.Background()

		_, err = fm.downloadRuleFile(ctx, server.URL)
		if err == nil {
			t.Error("expected timeout error, got nil")
		}
	})

	// Test invalid URL
	t.Run("Invalid URL", func(t *testing.T) {
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

		_, err = fm.downloadRuleFile(ctx, "not-a-valid-url")
		if err == nil {
			t.Error("expected error for invalid URL, got nil")
		}
	})

	// Test connection refused
	t.Run("Connection Refused", func(t *testing.T) {
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

		// Use a port that's unlikely to be in use
		_, err = fm.downloadRuleFile(ctx, "http://localhost:54321/nonexistent")
		if err == nil {
			t.Error("expected connection error, got nil")
		}
	})
}

// Test URL validation
func TestURLValidation(t *testing.T) {
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

	testCases := []struct {
		name      string
		url       string
		shouldErr bool
	}{
		{"Valid HTTP", "http://example.com/rules.txt", false},
		{"Valid HTTPS", "https://example.com/rules.txt", false},
		{"Empty URL", "", true},
		{"FTP Protocol", "ftp://example.com/rules.txt", true},
		{"No Protocol", "example.com/rules.txt", true},
		{"Too Short", "http", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := fm.validateURL(tc.url)
			if tc.shouldErr && err == nil {
				t.Errorf("expected error for URL %q, got nil", tc.url)
			}
			if !tc.shouldErr && err != nil {
				t.Errorf("unexpected error for URL %q: %v", tc.url, err)
			}
		})
	}
}
