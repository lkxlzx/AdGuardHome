package dnsforward

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid_http_url",
			url:     "http://example.com/rules.txt",
			wantErr: false,
		},
		{
			name:    "valid_https_url",
			url:     "https://example.com/rules.txt",
			wantErr: false,
		},
		{
			name:    "valid_file_url",
			url:     "file:///path/to/rules.txt",
			wantErr: false,
		},
		{
			name:    "empty_url",
			url:     "",
			wantErr: true,
			errMsg:  "url is required",
		},
		{
			name:    "url_too_long",
			url:     "https://example.com/" + string(make([]byte, 2100)),
			wantErr: true,
			errMsg:  "url too long",
		},
		{
			name:    "invalid_scheme",
			url:     "ftp://example.com/rules.txt",
			wantErr: true,
			errMsg:  "url scheme must be http, https, or file",
		},
		{
			name:    "no_scheme",
			url:     "example.com/rules.txt",
			wantErr: true,
			errMsg:  "url must have a scheme",
		},
		{
			name:    "http_without_host",
			url:     "http:///rules.txt",
			wantErr: true,
			errMsg:  "url must have a host",
		},
		{
			name:    "https_without_host",
			url:     "https:///rules.txt",
			wantErr: true,
			errMsg:  "url must have a host",
		},
		{
			name:    "invalid_url_format",
			url:     "ht tp://example.com",
			wantErr: true,
			errMsg:  "invalid url format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateURL(tt.url)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid_name",
			input:   "My DNS Rules",
			wantErr: false,
		},
		{
			name:    "empty_name",
			input:   "",
			wantErr: false, // Name is optional
		},
		{
			name:    "name_with_special_chars",
			input:   "Rules-2024_v1.0",
			wantErr: false,
		},
		{
			name:    "name_too_long",
			input:   string(make([]byte, 300)),
			wantErr: true,
			errMsg:  "name too long",
		},
		{
			name:    "name_with_control_chars",
			input:   "Rules\x00Test",
			wantErr: true,
			errMsg:  "name contains invalid control characters",
		},
		{
			name:    "name_only_whitespace",
			input:   "   ",
			wantErr: true,
			errMsg:  "name cannot be only whitespace",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateName(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateGroupID(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid_group_id",
			input:   "china",
			wantErr: false,
		},
		{
			name:    "valid_group_id_with_underscore",
			input:   "china_telecom",
			wantErr: false,
		},
		{
			name:    "valid_group_id_with_hyphen",
			input:   "china-mobile",
			wantErr: false,
		},
		{
			name:    "valid_group_id_with_numbers",
			input:   "group123",
			wantErr: false,
		},
		{
			name:    "empty_group_id",
			input:   "",
			wantErr: true,
			errMsg:  "group_id is required",
		},
		{
			name:    "group_id_too_long",
			input:   string(make([]byte, 150)),
			wantErr: true,
			errMsg:  "group_id too long",
		},
		{
			name:    "group_id_with_space",
			input:   "china group",
			wantErr: true,
			errMsg:  "group_id can only contain letters, numbers, underscore, and hyphen",
		},
		{
			name:    "group_id_with_special_chars",
			input:   "china@group",
			wantErr: true,
			errMsg:  "group_id can only contain letters, numbers, underscore, and hyphen",
		},
		{
			name:    "group_id_with_dot",
			input:   "china.group",
			wantErr: true,
			errMsg:  "group_id can only contain letters, numbers, underscore, and hyphen",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateGroupID(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateRuleID(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid_rule_id",
			input:   "routing_1234567890",
			wantErr: false,
		},
		{
			name:    "empty_rule_id",
			input:   "",
			wantErr: true,
			errMsg:  "id is required",
		},
		{
			name:    "rule_id_too_long",
			input:   string(make([]byte, 300)),
			wantErr: true,
			errMsg:  "id too long",
		},
		{
			name:    "rule_id_with_control_chars",
			input:   "routing\x00123",
			wantErr: true,
			errMsg:  "id contains invalid control characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateRuleID(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Benchmark tests
func BenchmarkValidateURL(b *testing.B) {
	url := "https://example.com/rules.txt"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = validateURL(url)
	}
}

func BenchmarkValidateName(b *testing.B) {
	name := "My DNS Rules"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = validateName(name)
	}
}

func BenchmarkValidateGroupID(b *testing.B) {
	groupID := "china_telecom"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = validateGroupID(groupID)
	}
}
