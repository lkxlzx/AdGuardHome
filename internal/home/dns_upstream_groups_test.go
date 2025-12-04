package home

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/AdguardTeam/golibs/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestWebAPI creates a webAPI instance for testing
func newTestWebAPI() *webAPI {
	return &webAPI{
		logger: slog.Default(),
	}
}

func TestValidateUpstreamGroupRequest(t *testing.T) {
	testCases := []struct {
		name    string
		req     *upstreamGroupRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid_request",
			req: &upstreamGroupRequest{
				Name:        "Test Group",
				Enabled:     true,
				IsDefault:   false,
				UpstreamDNS: []string{"8.8.8.8", "1.1.1.1"},
			},
			wantErr: false,
		},
		{
			name: "empty_name",
			req: &upstreamGroupRequest{
				Name:        "",
				Enabled:     true,
				UpstreamDNS: []string{"8.8.8.8"},
			},
			wantErr: true,
			errMsg:  "name",
		},
		{
			name: "name_too_long",
			req: &upstreamGroupRequest{
				Name:        string(make([]byte, 51)),
				Enabled:     true,
				UpstreamDNS: []string{"8.8.8.8"},
			},
			wantErr: true,
			errMsg:  "name",
		},
		{
			name: "empty_upstream_dns",
			req: &upstreamGroupRequest{
				Name:        "Test",
				Enabled:     true,
				UpstreamDNS: []string{},
			},
			wantErr: true,
			errMsg:  "upstream",
		},
		{
			name: "nil_upstream_dns",
			req: &upstreamGroupRequest{
				Name:        "Test",
				Enabled:     true,
				UpstreamDNS: nil,
			},
			wantErr: true,
			errMsg:  "upstream",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateUpstreamGroupRequest(tc.req)
			if tc.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestHandleGetUpstreamGroups(t *testing.T) {
	// Setup test config
	config.Lock()
	config.DNS.UpstreamGroups = []UpstreamGroup{
		{
			ID:          "group-1",
			Name:        "Test Group 1",
			Enabled:     true,
			IsDefault:   true,
			UpstreamDNS: []string{"8.8.8.8"},
			CreatedAt:   time.Now().UTC().Format(time.RFC3339),
			UpdatedAt:   time.Now().UTC().Format(time.RFC3339),
		},
		{
			ID:          "group-2",
			Name:        "Test Group 2",
			Enabled:     false,
			IsDefault:   false,
			UpstreamDNS: []string{"1.1.1.1"},
			CreatedAt:   time.Now().UTC().Format(time.RFC3339),
			UpdatedAt:   time.Now().UTC().Format(time.RFC3339),
		},
	}
	config.Unlock()

	// Cleanup after test
	testutil.CleanupAndRequireSuccess(t, func() (err error) {
		config.Lock()
		defer config.Unlock()
		config.DNS.UpstreamGroups = nil
		return nil
	})

	// Create test request
	req := httptest.NewRequest(http.MethodGet, "/control/dns/upstream_groups", nil)
	w := httptest.NewRecorder()

	// Create web API instance
	web := newTestWebAPI()

	// Call handler
	web.handleGetUpstreamGroups(w, req)

	// Check response
	assert.Equal(t, http.StatusOK, w.Code)

	var groups []UpstreamGroup
	err := json.NewDecoder(w.Body).Decode(&groups)
	require.NoError(t, err)

	assert.Len(t, groups, 2)
	assert.Equal(t, "Test Group 1", groups[0].Name)
	assert.Equal(t, "Test Group 2", groups[1].Name)
}

func TestHandleAddUpstreamGroup(t *testing.T) {
	testCases := []struct {
		name           string
		requestBody    upstreamGroupRequest
		existingGroups []UpstreamGroup
		wantStatus     int
		wantErr        bool
	}{
		{
			name: "success",
			requestBody: upstreamGroupRequest{
				Name:        "New Group",
				Enabled:     true,
				IsDefault:   false,
				UpstreamDNS: []string{"8.8.8.8"},
			},
			existingGroups: []UpstreamGroup{},
			wantStatus:     http.StatusOK,
			wantErr:        false,
		},
		{
			name: "duplicate_name",
			requestBody: upstreamGroupRequest{
				Name:        "Existing Group",
				Enabled:     true,
				IsDefault:   false,
				UpstreamDNS: []string{"8.8.8.8"},
			},
			existingGroups: []UpstreamGroup{
				{
					ID:          "existing-id",
					Name:        "Existing Group",
					Enabled:     true,
					IsDefault:   false,
					UpstreamDNS: []string{"1.1.1.1"},
					CreatedAt:   time.Now().UTC().Format(time.RFC3339),
					UpdatedAt:   time.Now().UTC().Format(time.RFC3339),
				},
			},
			wantStatus: http.StatusConflict,
			wantErr:    true,
		},
		{
			name: "invalid_empty_name",
			requestBody: upstreamGroupRequest{
				Name:        "",
				Enabled:     true,
				UpstreamDNS: []string{"8.8.8.8"},
			},
			existingGroups: []UpstreamGroup{},
			wantStatus:     http.StatusBadRequest,
			wantErr:        true,
		},
		{
			name: "invalid_empty_upstream",
			requestBody: upstreamGroupRequest{
				Name:        "Test",
				Enabled:     true,
				UpstreamDNS: []string{},
			},
			existingGroups: []UpstreamGroup{},
			wantStatus:     http.StatusBadRequest,
			wantErr:        true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup test config
			testutil.CleanupAndRequireSuccess(t, func() (err error) {
				config.Lock()
				defer config.Unlock()

				config.DNS.UpstreamGroups = tc.existingGroups

				return nil
			})

			// Create request
			body, err := json.Marshal(tc.requestBody)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/control/dns/upstream_groups", bytes.NewReader(body))
			w := httptest.NewRecorder()

			// Create web API instance
			web := newTestWebAPI()

			// Call handler
			web.handleAddUpstreamGroup(w, req)

			// Check response
			assert.Equal(t, tc.wantStatus, w.Code)

			if !tc.wantErr {
				var group UpstreamGroup
				err := json.NewDecoder(w.Body).Decode(&group)
				require.NoError(t, err)

				assert.NotEmpty(t, group.ID)
				assert.Equal(t, tc.requestBody.Name, group.Name)
				assert.Equal(t, tc.requestBody.Enabled, group.Enabled)
				assert.Equal(t, tc.requestBody.IsDefault, group.IsDefault)
				assert.Equal(t, tc.requestBody.UpstreamDNS, group.UpstreamDNS)
				assert.NotEmpty(t, group.CreatedAt)
				assert.NotEmpty(t, group.UpdatedAt)
			}
		})
	}
}

func TestHandleUpdateUpstreamGroup(t *testing.T) {
	existingGroup := UpstreamGroup{
		ID:          "test-id",
		Name:        "Original Name",
		Enabled:     true,
		IsDefault:   false,
		UpstreamDNS: []string{"8.8.8.8"},
		CreatedAt:   time.Now().UTC().Format(time.RFC3339),
		UpdatedAt:   time.Now().UTC().Format(time.RFC3339),
	}

	testCases := []struct {
		name        string
		groupID     string
		requestBody upstreamGroupRequest
		wantStatus  int
		wantErr     bool
	}{
		{
			name:    "success",
			groupID: "test-id",
			requestBody: upstreamGroupRequest{
				Name:        "Updated Name",
				Enabled:     false,
				IsDefault:   false,
				UpstreamDNS: []string{"1.1.1.1"},
			},
			wantStatus: http.StatusOK,
			wantErr:    false,
		},
		{
			name:    "not_found",
			groupID: "non-existent-id",
			requestBody: upstreamGroupRequest{
				Name:        "Test",
				Enabled:     true,
				UpstreamDNS: []string{"8.8.8.8"},
			},
			wantStatus: http.StatusNotFound,
			wantErr:    true,
		},
		{
			name:    "invalid_request",
			groupID: "test-id",
			requestBody: upstreamGroupRequest{
				Name:        "",
				Enabled:     true,
				UpstreamDNS: []string{"8.8.8.8"},
			},
			wantStatus: http.StatusBadRequest,
			wantErr:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup test config
			testutil.CleanupAndRequireSuccess(t, func() (err error) {
				config.Lock()
				defer config.Unlock()

				config.DNS.UpstreamGroups = []UpstreamGroup{existingGroup}

				return nil
			})

			// Create request
			body, err := json.Marshal(tc.requestBody)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPut, "/control/dns/upstream_groups/"+tc.groupID, bytes.NewReader(body))
			w := httptest.NewRecorder()

			// Create web API instance
			web := newTestWebAPI()

			// Call handler
			web.handleUpdateUpstreamGroup(w, req)

			// Check response
			assert.Equal(t, tc.wantStatus, w.Code)

			if !tc.wantErr {
				var group UpstreamGroup
				err := json.NewDecoder(w.Body).Decode(&group)
				require.NoError(t, err)

				assert.Equal(t, tc.groupID, group.ID)
				assert.Equal(t, tc.requestBody.Name, group.Name)
				assert.Equal(t, tc.requestBody.Enabled, group.Enabled)
				assert.NotEqual(t, existingGroup.UpdatedAt, group.UpdatedAt)
			}
		})
	}
}

func TestHandleDeleteUpstreamGroup(t *testing.T) {
	testCases := []struct {
		name           string
		groupID        string
		existingGroups []UpstreamGroup
		wantStatus     int
		wantErr        bool
	}{
		{
			name:    "success",
			groupID: "test-id",
			existingGroups: []UpstreamGroup{
				{
					ID:          "test-id",
					Name:        "Test Group",
					Enabled:     true,
					IsDefault:   false,
					UpstreamDNS: []string{"8.8.8.8"},
					CreatedAt:   time.Now().UTC().Format(time.RFC3339),
					UpdatedAt:   time.Now().UTC().Format(time.RFC3339),
				},
			},
			wantStatus: http.StatusOK,
			wantErr:    false,
		},
		{
			name:    "cannot_delete_default",
			groupID: "default-id",
			existingGroups: []UpstreamGroup{
				{
					ID:          "default-id",
					Name:        "Default Group",
					Enabled:     true,
					IsDefault:   true,
					UpstreamDNS: []string{"8.8.8.8"},
					CreatedAt:   time.Now().UTC().Format(time.RFC3339),
					UpdatedAt:   time.Now().UTC().Format(time.RFC3339),
				},
			},
			wantStatus: http.StatusConflict,
			wantErr:    true,
		},
		{
			name:           "not_found",
			groupID:        "non-existent-id",
			existingGroups: []UpstreamGroup{},
			wantStatus:     http.StatusNotFound,
			wantErr:        true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup test config
			testutil.CleanupAndRequireSuccess(t, func() (err error) {
				config.Lock()
				defer config.Unlock()

				config.DNS.UpstreamGroups = tc.existingGroups

				return nil
			})

			// Create request
			req := httptest.NewRequest(http.MethodDelete, "/control/dns/upstream_groups/"+tc.groupID, nil)
			w := httptest.NewRecorder()

			// Create web API instance
			web := newTestWebAPI()

			// Call handler
			web.handleDeleteUpstreamGroup(w, req)

			// Check response
			assert.Equal(t, tc.wantStatus, w.Code)

			if !tc.wantErr {
				// Verify group was deleted
				config.RLock()
				defer config.RUnlock()

				for _, g := range config.DNS.UpstreamGroups {
					assert.NotEqual(t, tc.groupID, g.ID)
				}
			}
		})
	}
}

func TestHandleSetDefaultGroup(t *testing.T) {
	testCases := []struct {
		name           string
		groupID        string
		existingGroups []UpstreamGroup
		wantStatus     int
		wantErr        bool
	}{
		{
			name:    "success",
			groupID: "new-default-id",
			existingGroups: []UpstreamGroup{
				{
					ID:          "old-default-id",
					Name:        "Old Default",
					Enabled:     true,
					IsDefault:   true,
					UpstreamDNS: []string{"8.8.8.8"},
					CreatedAt:   time.Now().UTC().Format(time.RFC3339),
					UpdatedAt:   time.Now().UTC().Format(time.RFC3339),
				},
				{
					ID:          "new-default-id",
					Name:        "New Default",
					Enabled:     true,
					IsDefault:   false,
					UpstreamDNS: []string{"1.1.1.1"},
					CreatedAt:   time.Now().UTC().Format(time.RFC3339),
					UpdatedAt:   time.Now().UTC().Format(time.RFC3339),
				},
			},
			wantStatus: http.StatusOK,
			wantErr:    false,
		},
		{
			name:           "not_found",
			groupID:        "non-existent-id",
			existingGroups: []UpstreamGroup{},
			wantStatus:     http.StatusNotFound,
			wantErr:        true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup test config
			testutil.CleanupAndRequireSuccess(t, func() (err error) {
				config.Lock()
				defer config.Unlock()

				config.DNS.UpstreamGroups = tc.existingGroups

				return nil
			})

			// Create request
			req := httptest.NewRequest(http.MethodPost, "/control/dns/upstream_groups/"+tc.groupID+"/default", nil)
			w := httptest.NewRecorder()

			// Create web API instance
			web := newTestWebAPI()

			// Call handler
			web.handleSetDefaultGroup(w, req)

			// Check response
			assert.Equal(t, tc.wantStatus, w.Code)

			if !tc.wantErr {
				// Verify only one default group exists
				config.RLock()
				defer config.RUnlock()

				defaultCount := 0
				var defaultID string
				for _, g := range config.DNS.UpstreamGroups {
					if g.IsDefault {
						defaultCount++
						defaultID = g.ID
					}
				}

				assert.Equal(t, 1, defaultCount, "should have exactly one default group")
				assert.Equal(t, tc.groupID, defaultID, "default group should be the requested one")
			}
		})
	}
}

func TestDefaultGroupUniqueness(t *testing.T) {
	// Property: 系统中始终只有一个默认分组
	
	// Setup initial state with one default group
	testutil.CleanupAndRequireSuccess(t, func() (err error) {
		config.Lock()
		defer config.Unlock()

		config.DNS.UpstreamGroups = []UpstreamGroup{
			{
				ID:          "group-1",
				Name:        "Group 1",
				Enabled:     true,
				IsDefault:   true,
				UpstreamDNS: []string{"8.8.8.8"},
				CreatedAt:   time.Now().UTC().Format(time.RFC3339),
				UpdatedAt:   time.Now().UTC().Format(time.RFC3339),
			},
			{
				ID:          "group-2",
				Name:        "Group 2",
				Enabled:     true,
				IsDefault:   false,
				UpstreamDNS: []string{"1.1.1.1"},
				CreatedAt:   time.Now().UTC().Format(time.RFC3339),
				UpdatedAt:   time.Now().UTC().Format(time.RFC3339),
			},
		}

		return nil
	})

	// Set group-2 as default
	req := httptest.NewRequest(http.MethodPost, "/control/dns/upstream_groups/group-2/default", nil)
	w := httptest.NewRecorder()

	web := newTestWebAPI()
	web.handleSetDefaultGroup(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify only one default group
	config.RLock()
	defer config.RUnlock()

	defaultCount := 0
	for _, g := range config.DNS.UpstreamGroups {
		if g.IsDefault {
			defaultCount++
		}
	}

	assert.Equal(t, 1, defaultCount, "should have exactly one default group after setting new default")
}

func TestEnsureDefaultGroup(t *testing.T) {
	testCases := []struct {
		name           string
		initialGroups  []UpstreamGroup
		expectedResult func(t *testing.T, groups []UpstreamGroup)
	}{
		{
			name: "default_group_exists",
			initialGroups: []UpstreamGroup{
				{
					ID:          "group-1",
					Name:        "Default Group",
					Enabled:     true,
					IsDefault:   true,
					UpstreamDNS: []string{"8.8.8.8"},
					CreatedAt:   time.Now().UTC().Format(time.RFC3339),
					UpdatedAt:   time.Now().UTC().Format(time.RFC3339),
				},
			},
			expectedResult: func(t *testing.T, groups []UpstreamGroup) {
				require.Len(t, groups, 1)
				assert.True(t, groups[0].IsDefault)
				assert.True(t, groups[0].Enabled)
			},
		},
		{
			name: "default_group_disabled_should_enable",
			initialGroups: []UpstreamGroup{
				{
					ID:          "group-1",
					Name:        "Default Group",
					Enabled:     false,
					IsDefault:   true,
					UpstreamDNS: []string{"8.8.8.8"},
					CreatedAt:   time.Now().UTC().Format(time.RFC3339),
					UpdatedAt:   time.Now().UTC().Format(time.RFC3339),
				},
			},
			expectedResult: func(t *testing.T, groups []UpstreamGroup) {
				require.Len(t, groups, 1)
				assert.True(t, groups[0].IsDefault)
				assert.True(t, groups[0].Enabled, "default group should be enabled")
			},
		},
		{
			name: "no_default_set_first_enabled",
			initialGroups: []UpstreamGroup{
				{
					ID:          "group-1",
					Name:        "Group 1",
					Enabled:     true,
					IsDefault:   false,
					UpstreamDNS: []string{"8.8.8.8"},
					CreatedAt:   time.Now().UTC().Format(time.RFC3339),
					UpdatedAt:   time.Now().UTC().Format(time.RFC3339),
				},
				{
					ID:          "group-2",
					Name:        "Group 2",
					Enabled:     true,
					IsDefault:   false,
					UpstreamDNS: []string{"1.1.1.1"},
					CreatedAt:   time.Now().UTC().Format(time.RFC3339),
					UpdatedAt:   time.Now().UTC().Format(time.RFC3339),
				},
			},
			expectedResult: func(t *testing.T, groups []UpstreamGroup) {
				require.Len(t, groups, 2)
				assert.True(t, groups[0].IsDefault, "first enabled group should be set as default")
				assert.False(t, groups[1].IsDefault)
			},
		},
		{
			name: "no_enabled_groups_create_default",
			initialGroups: []UpstreamGroup{
				{
					ID:          "group-1",
					Name:        "Disabled Group",
					Enabled:     false,
					IsDefault:   false,
					UpstreamDNS: []string{"8.8.8.8"},
					CreatedAt:   time.Now().UTC().Format(time.RFC3339),
					UpdatedAt:   time.Now().UTC().Format(time.RFC3339),
				},
			},
			expectedResult: func(t *testing.T, groups []UpstreamGroup) {
				require.Len(t, groups, 2, "should create a new default group")
				
				// Find the newly created default group
				var defaultGroup *UpstreamGroup
				for i := range groups {
					if groups[i].IsDefault {
						defaultGroup = &groups[i]
						break
					}
				}
				
				require.NotNil(t, defaultGroup, "should have a default group")
				assert.True(t, defaultGroup.Enabled)
				assert.Equal(t, "Default", defaultGroup.Name)
				assert.NotEmpty(t, defaultGroup.UpstreamDNS)
				assert.NotEmpty(t, defaultGroup.ID)
			},
		},
		{
			name:          "empty_groups_create_default",
			initialGroups: []UpstreamGroup{},
			expectedResult: func(t *testing.T, groups []UpstreamGroup) {
				require.Len(t, groups, 1, "should create a default group")
				assert.True(t, groups[0].IsDefault)
				assert.True(t, groups[0].Enabled)
				assert.Equal(t, "Default", groups[0].Name)
				assert.NotEmpty(t, groups[0].UpstreamDNS)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup - clear any previous state first
			config.Lock()
			config.DNS.UpstreamGroups = make([]UpstreamGroup, len(tc.initialGroups))
			copy(config.DNS.UpstreamGroups, tc.initialGroups)
			
			// Set some default upstream DNS for testing
			config.DNS.UpstreamDNS = []string{"8.8.8.8", "1.1.1.1"}
			config.DNS.BootstrapDNS = []string{"9.9.9.10"}
			config.Unlock()

			// Cleanup after test
			testutil.CleanupAndRequireSuccess(t, func() (err error) {
				config.Lock()
				defer config.Unlock()
				config.DNS.UpstreamGroups = nil
				return nil
			})

			// Execute
			ctx := testutil.ContextWithTimeout(t, testTimeout)
			ensureDefaultGroup(ctx, slog.Default())

			// Verify
			config.RLock()
			groups := make([]UpstreamGroup, len(config.DNS.UpstreamGroups))
			copy(groups, config.DNS.UpstreamGroups)
			config.RUnlock()

			tc.expectedResult(t, groups)

			// Verify only one default group exists
			defaultCount := 0
			for _, g := range groups {
				if g.IsDefault {
					defaultCount++
				}
			}
			assert.Equal(t, 1, defaultCount, "should have exactly one default group")
		})
	}
}
