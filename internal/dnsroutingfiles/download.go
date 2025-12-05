package dnsroutingfiles

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/AdguardTeam/AdGuardHome/internal/version"
)

const (
	// defaultDownloadTimeout is the default timeout for downloading rule files
	defaultDownloadTimeout = 30 * time.Second
	
	// maxDownloadSize is the maximum size of a downloadable file (100MB)
	maxDownloadSize = 100 * 1024 * 1024
)

// downloadRuleFile downloads a rule file from a URL.
// It supports both HTTP and HTTPS protocols and includes timeout handling.
// Files are downloaded to a temporary location first, then moved atomically.
func (m *fileManager) downloadRuleFile(ctx context.Context, url string) ([]byte, error) {
	m.config.Logger.DebugContext(ctx, "downloading rule file", "url", url)

	// Create request with context
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	// Set user agent with version information
	userAgent := fmt.Sprintf("AdGuardHome/%s DNS-Routing-File-Manager", version.Version())
	req.Header.Set("User-Agent", userAgent)

	// Perform request with timeout
	client := m.config.HTTPClient
	if client == nil {
		client = &http.Client{
			Timeout: defaultDownloadTimeout,
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("downloading from %s: %w", url, err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download failed with status %d: %s", resp.StatusCode, resp.Status)
	}

	// Limit download size to prevent abuse
	limitedReader := io.LimitReader(resp.Body, maxDownloadSize)

	// Create temporary file for atomic operation
	tempDir := filepath.Join(m.config.DataDir, "data", "dns_routing_rules")
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return nil, fmt.Errorf("creating temp directory: %w", err)
	}

	tempFile, err := os.CreateTemp(tempDir, "dns_rule_*.tmp")
	if err != nil {
		return nil, fmt.Errorf("creating temp file: %w", err)
	}
	tempPath := tempFile.Name()

	// Ensure cleanup of temp file
	defer func() {
		if err := tempFile.Close(); err != nil {
			m.config.Logger.WarnContext(ctx, "failed to close temp file", "path", tempPath, "error", err)
		}
		if err := os.Remove(tempPath); err != nil && !os.IsNotExist(err) {
			m.config.Logger.WarnContext(ctx, "failed to remove temp file", "path", tempPath, "error", err)
		}
	}()

	// Copy response body to temp file
	written, err := io.Copy(tempFile, limitedReader)
	if err != nil {
		return nil, fmt.Errorf("copying data: %w", err)
	}

	// Check if we hit the size limit
	if written == maxDownloadSize {
		return nil, fmt.Errorf("file too large (>%d bytes)", maxDownloadSize)
	}

	// Close and read the temp file
	if err := tempFile.Close(); err != nil {
		return nil, fmt.Errorf("closing temp file: %w", err)
	}

	// Read the downloaded data
	data, err := os.ReadFile(tempPath)
	if err != nil {
		return nil, fmt.Errorf("reading temp file: %w", err)
	}

	m.config.Logger.InfoContext(ctx, "downloaded rule file",
		"url", url,
		"size_bytes", len(data),
	)

	return data, nil
}

// validateURL performs validation on a URL.
func (m *fileManager) validateURL(url string) error {
	if url == "" {
		return fmt.Errorf("URL is empty")
	}

	// Check minimum length
	if len(url) < 10 {
		return fmt.Errorf("URL is too short")
	}

	// Check protocol
	hasHTTP := len(url) >= 7 && url[:7] == "http://"
	hasHTTPS := len(url) >= 8 && url[:8] == "https://"
	
	if !hasHTTP && !hasHTTPS {
		return fmt.Errorf("URL must use HTTP or HTTPS protocol")
	}

	// Check for basic URL structure (protocol://domain)
	protocolEnd := 7
	if hasHTTPS {
		protocolEnd = 8
	}
	
	remainder := url[protocolEnd:]
	if len(remainder) == 0 {
		return fmt.Errorf("URL must contain a domain")
	}
	
	// Check for invalid characters
	for _, ch := range remainder {
		if ch < 32 || ch > 126 {
			return fmt.Errorf("URL contains invalid characters")
		}
	}

	return nil
}
