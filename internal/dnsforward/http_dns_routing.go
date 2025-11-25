package dnsforward

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/AdguardTeam/AdGuardHome/internal/aghhttp"
)

// dnsRoutingRuleJSON represents a DNS routing rule in JSON format.
type dnsRoutingRuleJSON struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	URL           string `json:"url"`
	GroupID       string `json:"group_id"`
	Enabled       bool   `json:"enabled"`
	RuleCount     int    `json:"rule_count"`
	LastUpdated   string `json:"last_updated,omitempty"`
}

// dnsRoutingStatusResp represents the response for DNS routing status.
type dnsRoutingStatusResp struct {
	Rules []dnsRoutingRuleJSON `json:"rules"`
}

// handleDnsRoutingStatus returns the list of DNS routing rules.
func (s *Server) handleDnsRoutingStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	l := s.logger

	s.serverLock.RLock()
	defer s.serverLock.RUnlock()

	resp := dnsRoutingStatusResp{
		Rules: make([]dnsRoutingRuleJSON, 0, len(s.conf.DnsRoutingRules)),
	}

	for _, rule := range s.conf.DnsRoutingRules {
		rj := dnsRoutingRuleJSON{
			ID:        rule.ID,
			Name:      rule.Name,
			URL:       rule.URL,
			GroupID:   rule.GroupID,
			Enabled:   rule.Enabled,
			RuleCount: rule.RuleCount,
		}

		if rule.LastUpdated > 0 {
			rj.LastUpdated = time.Unix(rule.LastUpdated, 0).Format(time.RFC3339)
		}

		resp.Rules = append(resp.Rules, rj)
	}

	aghhttp.WriteJSONResponseOK(ctx, l, w, r, resp)
}

// dnsRoutingAddURLReq represents a request to add a DNS routing rule.
type dnsRoutingAddURLReq struct {
	Name    string `json:"name"`
	URL     string `json:"url"`
	GroupID string `json:"group_id"`
	Enabled bool   `json:"enabled"`
}

// handleDnsRoutingAddURL adds a new DNS routing rule.
func (s *Server) handleDnsRoutingAddURL(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	l := s.logger

	req := dnsRoutingAddURLReq{}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		aghhttp.ErrorAndLog(ctx, l, r, w, http.StatusBadRequest, "decoding request: %s", err)
		return
	}

	// Validate
	if req.URL == "" {
		aghhttp.ErrorAndLog(ctx, l, r, w, http.StatusBadRequest, "url is required")
		return
	}

	if req.GroupID == "" {
		aghhttp.ErrorAndLog(ctx, l, r, w, http.StatusBadRequest, "group_id is required")
		return
	}

	// Generate ID
	id := fmt.Sprintf("routing_%d", time.Now().UnixNano())

	rule := DnsRoutingRule{
		ID:          id,
		Name:        req.Name,
		URL:         req.URL,
		GroupID:     req.GroupID,
		Enabled:     req.Enabled,
		RuleCount:   0,
		LastUpdated: 0,
	}

	s.serverLock.Lock()
	s.conf.DnsRoutingRules = append(s.conf.DnsRoutingRules, rule)
	s.serverLock.Unlock()

	// TODO: Trigger download and parsing of the rule file

	aghhttp.OK(ctx, l, w)
}

// dnsRoutingSetURLReq represents a request to update a DNS routing rule.
type dnsRoutingSetURLReq struct {
	Data struct {
		Name    string `json:"name"`
		URL     string `json:"url"`
		GroupID string `json:"group_id"`
		Enabled bool   `json:"enabled"`
	} `json:"data"`
	URL string `json:"url"` // Original URL to identify the rule
}

// handleDnsRoutingSetURL updates an existing DNS routing rule.
func (s *Server) handleDnsRoutingSetURL(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	l := s.logger

	req := dnsRoutingSetURLReq{}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		aghhttp.ErrorAndLog(ctx, l, r, w, http.StatusBadRequest, "decoding request: %s", err)
		return
	}

	s.serverLock.Lock()
	defer s.serverLock.Unlock()

	// Find the rule by URL
	found := false
	for i := range s.conf.DnsRoutingRules {
		if s.conf.DnsRoutingRules[i].URL == req.URL {
			s.conf.DnsRoutingRules[i].Name = req.Data.Name
			s.conf.DnsRoutingRules[i].URL = req.Data.URL
			s.conf.DnsRoutingRules[i].GroupID = req.Data.GroupID
			s.conf.DnsRoutingRules[i].Enabled = req.Data.Enabled
			found = true
			break
		}
	}

	if !found {
		aghhttp.ErrorAndLog(ctx, l, r, w, http.StatusNotFound, "rule not found")
		return
	}

	aghhttp.OK(ctx, l, w)
}

// dnsRoutingRemoveURLReq represents a request to remove a DNS routing rule.
type dnsRoutingRemoveURLReq struct {
	URL string `json:"url"`
}

// handleDnsRoutingRemoveURL removes a DNS routing rule.
func (s *Server) handleDnsRoutingRemoveURL(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	l := s.logger

	req := dnsRoutingRemoveURLReq{}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		aghhttp.ErrorAndLog(ctx, l, r, w, http.StatusBadRequest, "decoding request: %s", err)
		return
	}

	s.serverLock.Lock()
	defer s.serverLock.Unlock()

	// Find and remove the rule
	found := false
	for i := range s.conf.DnsRoutingRules {
		if s.conf.DnsRoutingRules[i].URL == req.URL {
			s.conf.DnsRoutingRules = append(
				s.conf.DnsRoutingRules[:i],
				s.conf.DnsRoutingRules[i+1:]...,
			)
			found = true
			break
		}
	}

	if !found {
		aghhttp.ErrorAndLog(ctx, l, r, w, http.StatusNotFound, "rule not found")
		return
	}

	aghhttp.OK(ctx, l, w)
}

// handleDnsRoutingRefresh refreshes all DNS routing rules.
func (s *Server) handleDnsRoutingRefresh(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	l := s.logger

	// TODO: Implement refresh logic for DNS routing rules
	l.InfoContext(ctx, "dns routing refresh requested")

	aghhttp.OK(ctx, l, w)
}
