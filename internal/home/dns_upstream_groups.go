package home

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/AdguardTeam/AdGuardHome/internal/aghhttp"
	"github.com/AdguardTeam/dnsproxy/upstream"
	"github.com/AdguardTeam/golibs/errors"
	"github.com/google/uuid"
	"github.com/miekg/dns"
)

// upstreamGroupRequest is the request body for creating/updating upstream groups.
type upstreamGroupRequest struct {
	Name         string   `json:"name"`
	Enabled      bool     `json:"enabled"`
	IsDefault    bool     `json:"is_default"`
	UpstreamDNS  []string `json:"upstream_dns"`
	FallbackDNS  []string `json:"fallback_dns,omitempty"`
	BootstrapDNS []string `json:"bootstrap_dns,omitempty"`
}

// testUpstreamResult represents the result of testing a single upstream server.
type testUpstreamResult struct {
	Upstream string `json:"upstream"`
	Success  bool   `json:"success"`
	Error    string `json:"error,omitempty"`
	RTT      int64  `json:"rtt,omitempty"`
}

// testGroupResult represents the result of testing an upstream group.
type testGroupResult struct {
	GroupID string               `json:"group_id"`
	Results []testUpstreamResult `json:"results"`
}

// handleGetUpstreamGroups handles GET /control/dns/upstream_groups
func (web *webAPI) handleGetUpstreamGroups(w http.ResponseWriter, r *http.Request) {
	config.RLock()
	groups := config.DNS.UpstreamGroups
	config.RUnlock()

	_ = json.NewEncoder(w).Encode(groups)
}


// handleAddUpstreamGroup handles POST /control/dns/upstream_groups
func (web *webAPI) handleAddUpstreamGroup(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	l := web.logger

	var req upstreamGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		aghhttp.ErrorAndLog(ctx, l, r, w, http.StatusBadRequest, "invalid request body: %s", err)
		return
	}

	// Validate request
	if err := validateUpstreamGroupRequest(&req); err != nil {
		aghhttp.ErrorAndLog(ctx, l, r, w, http.StatusBadRequest, "%s", err)
		return
	}

	// Check for duplicate name
	config.RLock()
	for _, g := range config.DNS.UpstreamGroups {
		if g.Name == req.Name {
			config.RUnlock()
			aghhttp.ErrorAndLog(ctx, l, r, w, http.StatusConflict, "group with name %q already exists", req.Name)
			return
		}
	}
	config.RUnlock()

	// Create new group
	now := time.Now().UTC().Format(time.RFC3339)
	group := UpstreamGroup{
		ID:           uuid.New().String(),
		Name:         req.Name,
		Enabled:      req.Enabled,
		IsDefault:    req.IsDefault,
		UpstreamDNS:  req.UpstreamDNS,
		FallbackDNS:  req.FallbackDNS,
		BootstrapDNS: req.BootstrapDNS,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	// Modify config - config.write will handle locking
	config.Lock()
	// If this is set as default, unset other defaults
	if group.IsDefault {
		for i := range config.DNS.UpstreamGroups {
			config.DNS.UpstreamGroups[i].IsDefault = false
		}
	}
	// Add to config
	config.DNS.UpstreamGroups = append(config.DNS.UpstreamGroups, group)
	config.Unlock()

	// Save config (config.write handles its own locking)
	if err := config.write(ctx, l, web.tlsManager, web.auth, web.conf.workDir, web.conf.confPath); err != nil {
		aghhttp.ErrorAndLog(ctx, l, r, w, http.StatusInternalServerError, "failed to save configuration: %s", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(group)
}


// handleUpdateUpstreamGroup handles PUT /control/dns/upstream_groups/{id}
func (web *webAPI) handleUpdateUpstreamGroup(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	l := web.logger

	id := r.PathValue("id")
	if id == "" {
		aghhttp.ErrorAndLog(ctx, l, r, w, http.StatusBadRequest, "group id is required")
		return
	}

	var req upstreamGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		aghhttp.ErrorAndLog(ctx, l, r, w, http.StatusBadRequest, "invalid request body: %s", err)
		return
	}

	// Validate request
	if err := validateUpstreamGroupRequest(&req); err != nil {
		aghhttp.ErrorAndLog(ctx, l, r, w, http.StatusBadRequest, "%s", err)
		return
	}

	config.Lock()
	// Find the group
	groupIndex := -1
	for i, g := range config.DNS.UpstreamGroups {
		if g.ID == id {
			groupIndex = i
			break
		}
	}

	if groupIndex == -1 {
		config.Unlock()
		aghhttp.ErrorAndLog(ctx, l, r, w, http.StatusNotFound, "group not found")
		return
	}

	// Check for duplicate name (excluding current group)
	for i, g := range config.DNS.UpstreamGroups {
		if i != groupIndex && g.Name == req.Name {
			config.Unlock()
			aghhttp.ErrorAndLog(ctx, l, r, w, http.StatusConflict, "group with name %q already exists", req.Name)
			return
		}
	}

	// Update group
	group := &config.DNS.UpstreamGroups[groupIndex]
	
	// Cannot disable default group
	if group.IsDefault && !req.Enabled {
		config.Unlock()
		aghhttp.ErrorAndLog(ctx, l, r, w, http.StatusBadRequest, "cannot disable default group")
		return
	}
	
	group.Name = req.Name
	group.Enabled = req.Enabled
	group.UpstreamDNS = req.UpstreamDNS
	group.FallbackDNS = req.FallbackDNS
	group.BootstrapDNS = req.BootstrapDNS
	group.UpdatedAt = time.Now().UTC().Format(time.RFC3339)

	// If this is set as default, unset other defaults
	if req.IsDefault && !group.IsDefault {
		for i := range config.DNS.UpstreamGroups {
			config.DNS.UpstreamGroups[i].IsDefault = false
		}
		group.IsDefault = true
	} else if !req.IsDefault && group.IsDefault {
		// Cannot unset default without setting another
		config.Unlock()
		aghhttp.ErrorAndLog(ctx, l, r, w, http.StatusBadRequest, "cannot unset default group without setting another")
		return
	}
	config.Unlock()

	// Save config (config.write handles its own locking)
	if err := config.write(ctx, l, web.tlsManager, web.auth, web.conf.workDir, web.conf.confPath); err != nil {
		aghhttp.ErrorAndLog(ctx, l, r, w, http.StatusInternalServerError, "failed to save configuration: %s", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(group)
}


// handleDeleteUpstreamGroup handles DELETE /control/dns/upstream_groups/{id}
func (web *webAPI) handleDeleteUpstreamGroup(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	l := web.logger

	id := r.PathValue("id")
	if id == "" {
		aghhttp.ErrorAndLog(ctx, l, r, w, http.StatusBadRequest, "group id is required")
		return
	}

	config.Lock()
	// Find the group
	groupIndex := -1
	for i, g := range config.DNS.UpstreamGroups {
		if g.ID == id {
			groupIndex = i
			break
		}
	}

	if groupIndex == -1 {
		config.Unlock()
		aghhttp.ErrorAndLog(ctx, l, r, w, http.StatusNotFound, "group not found")
		return
	}

	// Cannot delete default group
	if config.DNS.UpstreamGroups[groupIndex].IsDefault {
		config.Unlock()
		aghhttp.ErrorAndLog(ctx, l, r, w, http.StatusConflict, "cannot delete default group")
		return
	}

	// Remove group
	config.DNS.UpstreamGroups = append(
		config.DNS.UpstreamGroups[:groupIndex],
		config.DNS.UpstreamGroups[groupIndex+1:]...,
	)
	config.Unlock()

	// Save config (config.write handles its own locking)
	if err := config.write(ctx, l, web.tlsManager, web.auth, web.conf.workDir, web.conf.confPath); err != nil {
		aghhttp.ErrorAndLog(ctx, l, r, w, http.StatusInternalServerError, "failed to save configuration: %s", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}


// handleSetDefaultGroup handles POST /control/dns/upstream_groups/{id}/default
func (web *webAPI) handleSetDefaultGroup(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	l := web.logger

	id := r.PathValue("id")
	if id == "" {
		aghhttp.ErrorAndLog(ctx, l, r, w, http.StatusBadRequest, "group id is required")
		return
	}

	config.Lock()
	// Find the group
	groupIndex := -1
	for i, g := range config.DNS.UpstreamGroups {
		if g.ID == id {
			groupIndex = i
			break
		}
	}

	if groupIndex == -1 {
		config.Unlock()
		aghhttp.ErrorAndLog(ctx, l, r, w, http.StatusNotFound, "group not found")
		return
	}

	// Unset all defaults
	for i := range config.DNS.UpstreamGroups {
		config.DNS.UpstreamGroups[i].IsDefault = false
	}

	// Set new default
	config.DNS.UpstreamGroups[groupIndex].IsDefault = true
	config.DNS.UpstreamGroups[groupIndex].UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	config.Unlock()

	// Save config (config.write handles its own locking)
	if err := config.write(ctx, l, web.tlsManager, web.auth, web.conf.workDir, web.conf.confPath); err != nil {
		aghhttp.ErrorAndLog(ctx, l, r, w, http.StatusInternalServerError, "failed to save configuration: %s", err)
		return
	}

	// Reconfigure DNS server to use the new default group
	if globalContext.dnsServer != nil {
		dnsConf, err := newServerConfig(
			&config.DNS,
			config.Clients.Sources,
			web.tlsManager.config(),
			web.tlsManager,
			web.httpReg,
			globalContext.clients.storage,
			web.confModifier,
		)
		if err != nil {
			aghhttp.ErrorAndLog(ctx, l, r, w, http.StatusInternalServerError, "failed to create DNS config: %s", err)
			return
		}

		err = globalContext.dnsServer.Reconfigure(ctx, dnsConf)
		if err != nil {
			aghhttp.ErrorAndLog(ctx, l, r, w, http.StatusInternalServerError, "failed to reconfigure DNS server: %s", err)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}


// handleTestUpstreamGroup handles POST /control/dns/upstream_groups/{id}/test
func (web *webAPI) handleTestUpstreamGroup(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	l := web.logger

	id := r.PathValue("id")
	if id == "" {
		aghhttp.ErrorAndLog(ctx, l, r, w, http.StatusBadRequest, "group id is required")
		return
	}

	// Copy group data to avoid holding lock during HTTP call
	config.RLock()
	var upstreamDNS []string
	found := false
	for i := range config.DNS.UpstreamGroups {
		if config.DNS.UpstreamGroups[i].ID == id {
			upstreamDNS = make([]string, len(config.DNS.UpstreamGroups[i].UpstreamDNS))
			copy(upstreamDNS, config.DNS.UpstreamGroups[i].UpstreamDNS)
			found = true
			break
		}
	}
	config.RUnlock()

	if !found {
		aghhttp.ErrorAndLog(ctx, l, r, w, http.StatusNotFound, "group not found")
		return
	}

	// Test each upstream (similar to reference project implementation)
	type upstreamTestResult struct {
		Upstream string `json:"upstream"`
		Success  bool   `json:"success"`
		RTT      int64  `json:"rtt,omitempty"` // milliseconds
		Error    string `json:"error,omitempty"`
	}

	results := make([]upstreamTestResult, 0, len(upstreamDNS))
	var totalTime time.Duration
	successCount := 0

	for _, upstreamAddr := range upstreamDNS {
		internalResult := testSingleUpstream(upstreamAddr)
		
		// Convert to frontend format
		result := upstreamTestResult{
			Upstream: internalResult.Upstream,
			Success:  internalResult.Status == "ok",
			Error:    internalResult.Error,
		}
		
		if result.Success && internalResult.ResponseTime != "" {
			if duration, err := time.ParseDuration(internalResult.ResponseTime); err == nil {
				result.RTT = duration.Milliseconds()
				totalTime += duration
				successCount++
			}
		}
		
		results = append(results, result)
	}

	// Calculate average response time
	avgResponseTime := "N/A"
	if successCount > 0 {
		avgDuration := totalTime / time.Duration(successCount)
		avgResponseTime = avgDuration.Round(time.Millisecond).String()
	}

	// Build response
	response := map[string]interface{}{
		"status":  "ok",
		"results": results,
		"summary": map[string]interface{}{
			"total":             len(upstreamDNS),
			"success":           successCount,
			"failed":            len(upstreamDNS) - successCount,
			"avg_response_time": avgResponseTime,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}

const (
	// testUpstreamTimeout is the timeout for testing upstream servers
	testUpstreamTimeout = 5 * time.Second
)

// testSingleUpstream tests a single upstream server
func testSingleUpstream(upstreamAddr string) struct {
	Upstream     string `json:"upstream"`
	Status       string `json:"status"`
	ResponseTime string `json:"response_time,omitempty"`
	Error        string `json:"error,omitempty"`
} {
	result := struct {
		Upstream     string `json:"upstream"`
		Status       string `json:"status"`
		ResponseTime string `json:"response_time,omitempty"`
		Error        string `json:"error,omitempty"`
	}{
		Upstream: upstreamAddr,
		Status:   "error",
	}

	// Create upstream
	u, err := upstream.AddressToUpstream(upstreamAddr, &upstream.Options{
		Timeout: testUpstreamTimeout,
	})
	if err != nil {
		result.Error = simplifyError(err)
		return result
	}
	defer u.Close()

	// Test with a simple DNS query (google.com A record)
	testReq := &dns.Msg{
		MsgHdr: dns.MsgHdr{
			Id:               dns.Id(),
			RecursionDesired: true,
		},
		Question: []dns.Question{{
			Name:   "google.com.",
			Qtype:  dns.TypeA,
			Qclass: dns.ClassINET,
		}},
	}

	start := time.Now()
	_, err = u.Exchange(testReq)
	elapsed := time.Since(start)

	if err != nil {
		result.Error = simplifyError(err)
		return result
	}

	result.Status = "ok"
	result.ResponseTime = elapsed.Round(time.Millisecond).String()
	return result
}

// simplifyError converts technical error messages to user-friendly error keys
// These keys should match the translation keys in the frontend (error_*)
func simplifyError(err error) string {
	if err == nil {
		return ""
	}

	// Check for standard error types first
	if errors.Is(err, context.DeadlineExceeded) {
		return "timeout"
	}
	if errors.Is(err, context.Canceled) {
		return "canceled"
	}

	errMsg := err.Error()
	
	// Return error keys that match frontend translation keys (error_*)
	// Remove verbose technical details like IP addresses and ports
	switch {
	case strings.Contains(errMsg, "i/o timeout"):
		return "timeout"
	case strings.Contains(errMsg, "connection refused"):
		return "connection_refused"
	case strings.Contains(errMsg, "no such host"):
		return "host_not_found"
	case strings.Contains(errMsg, "network is unreachable"):
		return "network_unreachable"
	case strings.Contains(errMsg, "connection reset"):
		return "connection_reset"
	case strings.Contains(errMsg, "invalid"):
		return "invalid_address"
	case strings.Contains(errMsg, "unsupported"):
		return "unsupported_protocol"
	case strings.Contains(errMsg, "tls") || strings.Contains(errMsg, "certificate"):
		return "tls"
	default:
		// Return a generic error key
		return "connection_failed"
	}
}

// validateUpstreamGroupRequest validates the upstream group request
func validateUpstreamGroupRequest(req *upstreamGroupRequest) error {
	if req.Name == "" {
		return errors.Error("group name is required")
	}

	if len(req.Name) > 50 {
		return errors.Error("group name must not exceed 50 characters")
	}

	if len(req.UpstreamDNS) == 0 {
		return errors.Error("at least one upstream DNS server is required")
	}

	return nil
}
