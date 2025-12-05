package home

import (
	"encoding/json"
	"net/http"
	"sync"
	"sync/atomic"

	"github.com/AdguardTeam/AdGuardHome/internal/aghhttp"
	"github.com/AdguardTeam/AdGuardHome/internal/dnsroutingfiles"
	"github.com/AdguardTeam/AdGuardHome/internal/filtering"
	"github.com/AdguardTeam/urlfilter/rules"
)

// nextDnsRoutingRuleID is an atomic counter for generating unique rule IDs
var nextDnsRoutingRuleID atomic.Int64

// handleGetDnsRoutingRules handles GET /control/dns_routing/rules
func (web *webAPI) handleGetDnsRoutingRules(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	// Get rules from config file (source of truth)
	config.RLock()
	filters := config.Filtering.Filters
	config.RUnlock()
	
	// Convert to response format
	rules := make([]DnsRoutingRule, 0)
	for _, filter := range filters {
		if !filter.DnsRouting {
			continue
		}
		
		rules = append(rules, DnsRoutingRule{
			ID:             int64(filter.ID),
			Enabled:        filter.Enabled,
			URL:            filter.URL,
			Name:           filter.Name,
			UpstreamGroup:  filter.UpstreamGroup,
			UpdateInterval: filter.UpdateInterval,
			Priority:       filter.Priority,
			RulesCount:     filter.RulesCount,
			LastUpdated:    filter.LastUpdated.Format("2006-01-02T15:04:05Z07:00"),
		})
	}
	
	aghhttp.WriteJSONResponseOK(ctx, web.logger, w, r, rules)
}

// handleAddDnsRoutingRule handles POST /control/dns_routing/add
func (web *webAPI) handleAddDnsRoutingRule(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	l := web.logger

	// Limit request body size to 1MB to prevent DoS attacks
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	type request struct {
		Name           string `json:"name"`
		URL            string `json:"url"`
		UpstreamGroup  string `json:"upstream_group"`
		UpdateInterval int    `json:"update_interval"`
		Priority       int    `json:"priority"`
	}

	req := request{}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		aghhttp.ErrorAndLog(ctx, l, r, w, http.StatusBadRequest, "decoding request: %s", err)
		return
	}

	// Validate
	if req.Name == "" {
		aghhttp.ErrorAndLog(ctx, l, r, w, http.StatusBadRequest, "name is required")
		return
	}
	if req.URL == "" {
		aghhttp.ErrorAndLog(ctx, l, r, w, http.StatusBadRequest, "url is required")
		return
	}
	if req.UpstreamGroup == "" {
		aghhttp.ErrorAndLog(ctx, l, r, w, http.StatusBadRequest, "upstream_group is required")
		return
	}

	// Generate new ID using atomic counter to prevent race condition
	newID := nextDnsRoutingRuleID.Add(1)

	// Create new rule for File Manager
	rule := &dnsroutingfiles.DomainListRule{
		ID:            newID,
		Name:          req.Name,
		URL:           req.URL,
		UpstreamGroup: req.UpstreamGroup,
		Priority:      req.Priority,
		Enabled:       true,
	}

	// Add rule using File Manager (handles download, parse, persist)
	err = globalContext.dnsRoutingFileManager.AddDomainListRule(ctx, rule)
	if err != nil {
		aghhttp.ErrorAndLog(ctx, l, r, w, http.StatusInternalServerError, "adding dns routing rule: %s", err)
		return
	}

	// Add to config file
	config.Lock()
	newFilter := filtering.FilterYAML{
		Enabled:        true,
		URL:            req.URL,
		Name:           req.Name,
		RulesCount:     rule.RulesCount,
		LastUpdated:    rule.LastUpdated,
		DnsRouting:     true,
		UpstreamGroup:  req.UpstreamGroup,
		UpdateInterval: req.UpdateInterval,
		Priority:       req.Priority,
	}
	newFilter.ID = rules.ListID(newID)
	config.Filtering.Filters = append(config.Filtering.Filters, newFilter)
	config.Unlock()

	// Save config
	if !web.saveConfigIfNeeded(ctx, l, r, w) {
		// Config save failed, try to rollback file manager changes
		if err := globalContext.dnsRoutingFileManager.DeleteDomainListRule(ctx, newID); err != nil {
			l.ErrorContext(ctx, "failed to rollback after config save failure", "error", err)
		}
		return
	}

	// Return response
	response := DnsRoutingRule{
		ID:             rule.ID,
		Enabled:        true,
		URL:            req.URL,
		Name:           req.Name,
		UpstreamGroup:  req.UpstreamGroup,
		UpdateInterval: req.UpdateInterval,
		Priority:       req.Priority,
		RulesCount:     rule.RulesCount,
		LastUpdated:    rule.LastUpdated.Format("2006-01-02T15:04:05Z07:00"),
	}

	aghhttp.WriteJSONResponseOK(ctx, l, w, r, response)
}

// handleUpdateDnsRoutingRule handles POST /control/dns_routing/update
func (web *webAPI) handleUpdateDnsRoutingRule(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	l := web.logger

	// Limit request body size to 1MB to prevent DoS attacks
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	type request struct {
		ID             int64  `json:"id"`
		Name           string `json:"name"`
		URL            string `json:"url"`
		UpstreamGroup  string `json:"upstream_group"`
		UpdateInterval int    `json:"update_interval"`
		Priority       int    `json:"priority"`
		Enabled        bool   `json:"enabled"`
	}

	req := request{}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		aghhttp.ErrorAndLog(ctx, l, r, w, http.StatusBadRequest, "decoding request: %s", err)
		return
	}

	// Validate
	if req.ID == 0 {
		aghhttp.ErrorAndLog(ctx, l, r, w, http.StatusBadRequest, "id is required")
		return
	}
	if req.Name == "" {
		aghhttp.ErrorAndLog(ctx, l, r, w, http.StatusBadRequest, "name is required")
		return
	}
	if req.URL == "" {
		aghhttp.ErrorAndLog(ctx, l, r, w, http.StatusBadRequest, "url is required")
		return
	}
	if req.UpstreamGroup == "" {
		aghhttp.ErrorAndLog(ctx, l, r, w, http.StatusBadRequest, "upstream_group is required")
		return
	}

	// Create updated rule for File Manager
	rule := &dnsroutingfiles.DomainListRule{
		ID:            req.ID,
		Name:          req.Name,
		URL:           req.URL,
		UpstreamGroup: req.UpstreamGroup,
		Priority:      req.Priority,
		Enabled:       req.Enabled,
	}

	// Update rule using File Manager (handles re-download if URL changed)
	err = globalContext.dnsRoutingFileManager.UpdateDomainListRule(ctx, rule)
	if err != nil {
		aghhttp.ErrorAndLog(ctx, l, r, w, http.StatusInternalServerError, "updating dns routing rule: %s", err)
		return
	}

	// Update config file
	config.Lock()
	for i := range config.Filtering.Filters {
		if int64(config.Filtering.Filters[i].ID) == req.ID && config.Filtering.Filters[i].DnsRouting {
			config.Filtering.Filters[i].Name = req.Name
			config.Filtering.Filters[i].URL = req.URL
			config.Filtering.Filters[i].UpstreamGroup = req.UpstreamGroup
			config.Filtering.Filters[i].UpdateInterval = req.UpdateInterval
			config.Filtering.Filters[i].Priority = req.Priority
			config.Filtering.Filters[i].Enabled = req.Enabled
			config.Filtering.Filters[i].RulesCount = rule.RulesCount
			config.Filtering.Filters[i].LastUpdated = rule.LastUpdated
			break
		}
	}
	config.Unlock()

	// Save config
	if !web.saveConfigIfNeeded(ctx, l, r, w) {
		return
	}

	aghhttp.WriteJSONResponseOK(ctx, l, w, r, struct{}{})
}

// handleDeleteDnsRoutingRule handles POST /control/dns_routing/delete
func (web *webAPI) handleDeleteDnsRoutingRule(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	l := web.logger

	// Limit request body size to 1MB to prevent DoS attacks
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	type request struct {
		ID int64 `json:"id"`
	}

	req := request{}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		aghhttp.ErrorAndLog(ctx, l, r, w, http.StatusBadRequest, "decoding request: %s", err)
		return
	}

	// Validate
	if req.ID == 0 {
		aghhttp.ErrorAndLog(ctx, l, r, w, http.StatusBadRequest, "id is required")
		return
	}

	// Delete rule using File Manager (handles file cleanup)
	err = globalContext.dnsRoutingFileManager.DeleteDomainListRule(ctx, req.ID)
	if err != nil {
		aghhttp.ErrorAndLog(ctx, l, r, w, http.StatusInternalServerError, "deleting dns routing rule: %s", err)
		return
	}

	// Remove from config file
	config.Lock()
	newFilters := make([]filtering.FilterYAML, 0, len(config.Filtering.Filters))
	for _, filter := range config.Filtering.Filters {
		if int64(filter.ID) != req.ID || !filter.DnsRouting {
			newFilters = append(newFilters, filter)
		}
	}
	config.Filtering.Filters = newFilters
	config.Unlock()

	// Save config
	if !web.saveConfigIfNeeded(ctx, l, r, w) {
		return
	}

	aghhttp.WriteJSONResponseOK(ctx, l, w, r, struct{}{})
}

// handleRefreshDnsRoutingRule handles POST /control/dns_routing/refresh
func (web *webAPI) handleRefreshDnsRoutingRule(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	l := web.logger

	// Limit request body size to 1MB to prevent DoS attacks
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	type request struct {
		ID int64 `json:"id"` // 0 means refresh all
	}

	req := request{}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		aghhttp.ErrorAndLog(ctx, l, r, w, http.StatusBadRequest, "decoding request: %s", err)
		return
	}

	if req.ID == 0 {
		// Refresh all rules with proper synchronization
		config.RLock()
		filters := config.Filtering.Filters
		config.RUnlock()
		
		// Use sync.WaitGroup to wait for all refreshes to complete
		var wg sync.WaitGroup
		// Limit concurrent refreshes to avoid resource exhaustion
		semaphore := make(chan struct{}, 5) // Max 5 concurrent refreshes
		
		for _, filter := range filters {
			if !filter.DnsRouting || !filter.Enabled {
				continue
			}
			
			wg.Add(1)
			go func(f filtering.FilterYAML) {
				defer wg.Done()
				
				// Acquire semaphore
				semaphore <- struct{}{}
				defer func() { <-semaphore }()
				
				rule := &dnsroutingfiles.DomainListRule{
					ID:            int64(f.ID),
					Name:          f.Name,
					URL:           f.URL,
					UpstreamGroup: f.UpstreamGroup,
					Priority:      f.Priority,
					Enabled:       f.Enabled,
				}
				
				err := globalContext.dnsRoutingFileManager.RefreshDomainListRule(ctx, rule)
				if err != nil {
					l.ErrorContext(ctx, "refreshing dns routing rule", "id", f.ID, "error", err)
					return
				}
				
				// Update config with new rules count and last updated
				config.Lock()
				for i := range config.Filtering.Filters {
					if config.Filtering.Filters[i].ID == f.ID && config.Filtering.Filters[i].DnsRouting {
						config.Filtering.Filters[i].RulesCount = rule.RulesCount
						config.Filtering.Filters[i].LastUpdated = rule.LastUpdated
						break
					}
				}
				config.Unlock()
			}(filter)
		}
		
		// Wait for all refreshes to complete
		wg.Wait()
	} else {
		// Refresh specific rule
		config.RLock()
		var targetFilter *filtering.FilterYAML
		for _, filter := range config.Filtering.Filters {
			if int64(filter.ID) == req.ID && filter.DnsRouting {
				f := filter // Create a copy
				targetFilter = &f
				break
			}
		}
		config.RUnlock()
		
		if targetFilter == nil {
			aghhttp.ErrorAndLog(ctx, l, r, w, http.StatusNotFound, "DNS routing rule not found")
			return
		}
		
		rule := &dnsroutingfiles.DomainListRule{
			ID:            int64(targetFilter.ID),
			Name:          targetFilter.Name,
			URL:           targetFilter.URL,
			UpstreamGroup: targetFilter.UpstreamGroup,
			Priority:      targetFilter.Priority,
			Enabled:       targetFilter.Enabled,
		}
		
		err = globalContext.dnsRoutingFileManager.RefreshDomainListRule(ctx, rule)
		if err != nil {
			aghhttp.ErrorAndLog(ctx, l, r, w, http.StatusInternalServerError, "refreshing dns routing rule: %s", err)
			return
		}
		
		// Update config with new rules count and last updated
		config.Lock()
		for i := range config.Filtering.Filters {
			if int64(config.Filtering.Filters[i].ID) == req.ID && config.Filtering.Filters[i].DnsRouting {
				config.Filtering.Filters[i].RulesCount = rule.RulesCount
				config.Filtering.Filters[i].LastUpdated = rule.LastUpdated
				break
			}
		}
		config.Unlock()
	}
	
	// Save config after refresh
	if !web.saveConfigIfNeeded(ctx, l, r, w) {
		return
	}

	aghhttp.WriteJSONResponseOK(ctx, l, w, r, struct{}{})
}

// handleGetCustomDomainRules handles GET /control/dns_routing/custom_rules
func (web *webAPI) handleGetCustomDomainRules(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	config.RLock()
	rules := make([]CustomDomainRule, len(config.DNS.CustomDomainRules))
	copy(rules, config.DNS.CustomDomainRules)
	config.RUnlock()
	
	aghhttp.WriteJSONResponseOK(ctx, web.logger, w, r, rules)
}

// handleAddCustomDomainRule handles POST /control/dns_routing/custom_rules/add
func (web *webAPI) handleAddCustomDomainRule(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	l := web.logger

	type request struct {
		Domain        string `json:"domain"`
		MatchType     string `json:"matchType"`
		UpstreamGroup string `json:"upstreamGroup"`
		Enabled       bool   `json:"enabled"`
	}

	req := request{}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		aghhttp.ErrorAndLog(ctx, l, r, w, http.StatusBadRequest, "decoding request: %s", err)
		return
	}

	// Validate
	if req.Domain == "" {
		aghhttp.ErrorAndLog(ctx, l, r, w, http.StatusBadRequest, "domain is required")
		return
	}
	if req.MatchType == "" {
		aghhttp.ErrorAndLog(ctx, l, r, w, http.StatusBadRequest, "matchType is required")
		return
	}
	if req.UpstreamGroup == "" {
		aghhttp.ErrorAndLog(ctx, l, r, w, http.StatusBadRequest, "upstreamGroup is required")
		return
	}

	// Validate match type
	validMatchTypes := map[string]bool{
		"DOMAIN":         true,
		"DOMAIN-SUFFIX":  true,
		"DOMAIN-KEYWORD": true,
	}
	if !validMatchTypes[req.MatchType] {
		aghhttp.ErrorAndLog(ctx, l, r, w, http.StatusBadRequest, "invalid matchType: must be DOMAIN, DOMAIN-SUFFIX, or DOMAIN-KEYWORD")
		return
	}

	// Create new rule for File Manager
	rule := &dnsroutingfiles.CustomRule{
		Domain:        req.Domain,
		MatchType:     req.MatchType,
		UpstreamGroup: req.UpstreamGroup,
		Enabled:       req.Enabled,
	}

	// Add rule using File Manager
	err = globalContext.dnsRoutingFileManager.AddCustomRule(ctx, rule)
	if err != nil {
		aghhttp.ErrorAndLog(ctx, l, r, w, http.StatusInternalServerError, "adding custom rule: %s", err)
		return
	}

	newRule := CustomDomainRule{
		Domain:        rule.Domain,
		MatchType:     rule.MatchType,
		UpstreamGroup: rule.UpstreamGroup,
		Enabled:       rule.Enabled,
	}

	aghhttp.WriteJSONResponseOK(ctx, l, w, r, newRule)
}

// handleUpdateCustomDomainRule handles POST /control/dns_routing/custom_rules/update
func (web *webAPI) handleUpdateCustomDomainRule(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	l := web.logger

	type request struct {
		OldDomain     string `json:"oldDomain"`
		OldMatchType  string `json:"oldMatchType"`
		Domain        string `json:"domain"`
		MatchType     string `json:"matchType"`
		UpstreamGroup string `json:"upstreamGroup"`
		Enabled       bool   `json:"enabled"`
	}

	req := request{}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		aghhttp.ErrorAndLog(ctx, l, r, w, http.StatusBadRequest, "decoding request: %s", err)
		return
	}

	// Validate
	if req.OldDomain == "" {
		aghhttp.ErrorAndLog(ctx, l, r, w, http.StatusBadRequest, "oldDomain is required")
		return
	}
	if req.OldMatchType == "" {
		aghhttp.ErrorAndLog(ctx, l, r, w, http.StatusBadRequest, "oldMatchType is required")
		return
	}
	if req.Domain == "" {
		aghhttp.ErrorAndLog(ctx, l, r, w, http.StatusBadRequest, "domain is required")
		return
	}
	if req.MatchType == "" {
		aghhttp.ErrorAndLog(ctx, l, r, w, http.StatusBadRequest, "matchType is required")
		return
	}
	if req.UpstreamGroup == "" {
		aghhttp.ErrorAndLog(ctx, l, r, w, http.StatusBadRequest, "upstreamGroup is required")
		return
	}

	// Validate match type
	validMatchTypes := map[string]bool{
		"DOMAIN":         true,
		"DOMAIN-SUFFIX":  true,
		"DOMAIN-KEYWORD": true,
	}
	if !validMatchTypes[req.MatchType] {
		aghhttp.ErrorAndLog(ctx, l, r, w, http.StatusBadRequest, "invalid matchType")
		return
	}

	// Create updated rule for File Manager
	rule := &dnsroutingfiles.CustomRule{
		Domain:        req.Domain,
		MatchType:     req.MatchType,
		UpstreamGroup: req.UpstreamGroup,
		Enabled:       req.Enabled,
	}

	// Update rule using File Manager
	err = globalContext.dnsRoutingFileManager.UpdateCustomRule(ctx, req.OldDomain, req.OldMatchType, rule)
	if err != nil {
		aghhttp.ErrorAndLog(ctx, l, r, w, http.StatusInternalServerError, "updating custom rule: %s", err)
		return
	}

	aghhttp.WriteJSONResponseOK(ctx, l, w, r, struct{}{})
}

// handleDeleteCustomDomainRule handles POST /control/dns_routing/custom_rules/delete
func (web *webAPI) handleDeleteCustomDomainRule(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	l := web.logger

	type request struct {
		Domain    string `json:"domain"`
		MatchType string `json:"matchType"`
	}

	req := request{}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		aghhttp.ErrorAndLog(ctx, l, r, w, http.StatusBadRequest, "decoding request: %s", err)
		return
	}

	// Validate
	if req.Domain == "" {
		aghhttp.ErrorAndLog(ctx, l, r, w, http.StatusBadRequest, "domain is required")
		return
	}
	if req.MatchType == "" {
		aghhttp.ErrorAndLog(ctx, l, r, w, http.StatusBadRequest, "matchType is required")
		return
	}

	// Delete rule using File Manager
	err = globalContext.dnsRoutingFileManager.DeleteCustomRule(ctx, req.Domain, req.MatchType)
	if err != nil {
		aghhttp.ErrorAndLog(ctx, l, r, w, http.StatusInternalServerError, "deleting custom rule: %s", err)
		return
	}

	aghhttp.WriteJSONResponseOK(ctx, l, w, r, struct{}{})
}

