package dnsforward

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/AdguardTeam/AdGuardHome/internal/aghhttp"
)

// Validation constants
const (
	maxURLLength     = 2048
	maxNameLength    = 256
	maxGroupIDLength = 128
	minNameLength    = 1
	minGroupIDLength = 1
)

// validateURL validates a URL string.
func validateURL(urlStr string) error {
	// Check if URL is empty
	if urlStr == "" {
		return fmt.Errorf("url is required")
	}

	// Check URL length
	if len(urlStr) > maxURLLength {
		return fmt.Errorf("url too long (max %d characters)", maxURLLength)
	}

	// Parse and validate URL format
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("invalid url format: %w", err)
	}

	// Check if URL has a scheme
	if parsedURL.Scheme == "" {
		return fmt.Errorf("url must have a scheme (http, https, or file)")
	}

	// Validate scheme
	allowedSchemes := map[string]bool{
		"http":  true,
		"https": true,
		"file":  true,
	}
	if !allowedSchemes[parsedURL.Scheme] {
		return fmt.Errorf("url scheme must be http, https, or file")
	}

	// For http/https, validate host
	if parsedURL.Scheme == "http" || parsedURL.Scheme == "https" {
		if parsedURL.Host == "" {
			return fmt.Errorf("url must have a host")
		}
	}

	return nil
}

// validateName validates a name string.
func validateName(name string) error {
	// Name can be empty (optional)
	if name == "" {
		return nil
	}

	// Check name length
	if len(name) < minNameLength {
		return fmt.Errorf("name too short (min %d character)", minNameLength)
	}

	if len(name) > maxNameLength {
		return fmt.Errorf("name too long (max %d characters)", maxNameLength)
	}

	// Check for invalid characters (control characters)
	for _, r := range name {
		if r < 32 || r == 127 {
			return fmt.Errorf("name contains invalid control characters")
		}
	}

	// Trim whitespace and check if empty
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("name cannot be only whitespace")
	}

	return nil
}

// validateGroupID validates a group ID string.
func validateGroupID(groupID string) error {
	// Check if group ID is empty
	if groupID == "" {
		return fmt.Errorf("group_id is required")
	}

	// Check group ID length
	if len(groupID) < minGroupIDLength {
		return fmt.Errorf("group_id too short (min %d character)", minGroupIDLength)
	}

	if len(groupID) > maxGroupIDLength {
		return fmt.Errorf("group_id too long (max %d characters)", maxGroupIDLength)
	}

	// Check for invalid characters (only allow alphanumeric, underscore, hyphen)
	for _, r := range groupID {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') || r == '_' || r == '-') {
			return fmt.Errorf("group_id can only contain letters, numbers, underscore, and hyphen")
		}
	}

	return nil
}

// validateRuleID validates a rule ID string.
func validateRuleID(id string) error {
	// Check if ID is empty
	if id == "" {
		return fmt.Errorf("id is required")
	}

	// Check ID length
	if len(id) > maxNameLength {
		return fmt.Errorf("id too long (max %d characters)", maxNameLength)
	}

	// Check for invalid characters
	for _, r := range id {
		if r < 32 || r == 127 {
			return fmt.Errorf("id contains invalid control characters")
		}
	}

	return nil
}

// dnsRoutingRuleJSON represents a DNS routing rule in JSON format.
type dnsRoutingRuleJSON struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	URL         string `json:"url"`
	GroupID     string `json:"group_id"`
	Enabled     bool   `json:"enabled"`
	RuleCount   int    `json:"rule_count"`
	LastUpdated string `json:"last_updated,omitempty"`
}

// dnsRoutingStatusResp represents the response for DNS routing status.
type dnsRoutingStatusResp struct {
	Rules []dnsRoutingRuleJSON `json:"rules"`
}

// handleDNSRoutingStatus returns the list of DNS routing rules.
func (s *Server) handleDNSRoutingStatus(w http.ResponseWriter, r *http.Request) {
	resp := dnsRoutingStatusResp{
		Rules: []dnsRoutingRuleJSON{},
	}

	s.serverLock.RLock()
	defer s.serverLock.RUnlock()

	for _, rule := range s.conf.DNSRoutingRules {
		rj := dnsRoutingRuleJSON{
			ID:        rule.ID,
			Name:      rule.Name,
			URL:       rule.URL,
			GroupID:   rule.GroupID,
			Enabled:   rule.Enabled,
			RuleCount: len(rule.Domains),
		}

		if rule.LastUpdated > 0 {
			rj.LastUpdated = time.Unix(rule.LastUpdated, 0).Format(time.RFC3339)
		}

		resp.Rules = append(resp.Rules, rj)
	}

	aghhttp.WriteJSONResponseOK(r.Context(), s.logger, w, r, resp)
}

// dnsRoutingAddURLReq represents a request to add a DNS routing rule.
type dnsRoutingAddURLReq struct {
	Name    string `json:"name"`
	URL     string `json:"url"`
	GroupID string `json:"group_id"`
	Enabled bool   `json:"enabled"`
}

// handleDNSRoutingAddURL adds a new DNS routing rule.
func (s *Server) handleDNSRoutingAddURL(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	l := s.logger
	req := &dnsRoutingAddURLReq{}
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		aghhttp.ErrorAndLog(ctx, l, r, w, http.StatusBadRequest, "decoding request: %s", err)
		return
	}

	// Validate URL
	if err := validateURL(req.URL); err != nil {
		aghhttp.ErrorAndLog(ctx, l, r, w, http.StatusBadRequest, "invalid url: %s", err)
		return
	}

	// Validate Name
	if err := validateName(req.Name); err != nil {
		aghhttp.ErrorAndLog(ctx, l, r, w, http.StatusBadRequest, "invalid name: %s", err)
		return
	}

	// Validate GroupID
	if err := validateGroupID(req.GroupID); err != nil {
		aghhttp.ErrorAndLog(ctx, l, r, w, http.StatusBadRequest, "invalid group_id: %s", err)
		return
	}

	s.serverLock.Lock()
	defer s.serverLock.Unlock()

	// Check if rule with same URL already exists
	for _, rule := range s.conf.DNSRoutingRules {
		if rule.URL == req.URL {
			aghhttp.ErrorAndLog(ctx, l, r, w, http.StatusBadRequest, "rule with this url already exists")
			return
		}
	}

	// Generate ID
	id := fmt.Sprintf("routing_%d", time.Now().UnixNano())

	rule := DNSRoutingRule{
		ID:          id,
		Name:        req.Name,
		URL:         req.URL,
		GroupID:     req.GroupID,
		Enabled:     req.Enabled,
		LastUpdated: time.Now().Unix(),
	}

	s.conf.DNSRoutingRules = append(s.conf.DNSRoutingRules, rule)

	// Save config
	if s.conf.ConfModifier != nil {
		s.conf.ConfModifier.Apply(ctx)
	}

	aghhttp.OK(ctx, l, w)
}

// dnsRoutingSetURLReq represents a request to update a DNS routing rule.
type dnsRoutingSetURLReq struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	URL     string `json:"url"`
	GroupID string `json:"group_id"`
	Enabled bool   `json:"enabled"`
}

// handleDNSRoutingSetURL updates an existing DNS routing rule.
func (s *Server) handleDNSRoutingSetURL(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	l := s.logger
	req := &dnsRoutingSetURLReq{}
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		aghhttp.ErrorAndLog(ctx, l, r, w, http.StatusBadRequest, "decoding request: %s", err)
		return
	}

	// Validate ID if provided
	if req.ID != "" {
		if err := validateRuleID(req.ID); err != nil {
			aghhttp.ErrorAndLog(ctx, l, r, w, http.StatusBadRequest, "invalid id: %s", err)
			return
		}
	}

	// Validate URL if provided
	if req.URL != "" {
		if err := validateURL(req.URL); err != nil {
			aghhttp.ErrorAndLog(ctx, l, r, w, http.StatusBadRequest, "invalid url: %s", err)
			return
		}
	}

	s.serverLock.Lock()
	defer s.serverLock.Unlock()

	// Find rule by ID
	var found bool
	for i, rule := range s.conf.DNSRoutingRules {
		if rule.ID == req.ID {
			s.conf.DNSRoutingRules[i].Name = req.Name
			s.conf.DNSRoutingRules[i].URL = req.URL
			s.conf.DNSRoutingRules[i].GroupID = req.GroupID
			s.conf.DNSRoutingRules[i].Enabled = req.Enabled
			s.conf.DNSRoutingRules[i].LastUpdated = time.Now().Unix()
			found = true
			break
		}
	}

	if !found {
		// Try to find by URL for backward compatibility
		for i, rule := range s.conf.DNSRoutingRules {
			if rule.URL == req.URL {
				s.conf.DNSRoutingRules[i].Name = req.Name
				s.conf.DNSRoutingRules[i].GroupID = req.GroupID
				s.conf.DNSRoutingRules[i].Enabled = req.Enabled
				s.conf.DNSRoutingRules[i].LastUpdated = time.Now().Unix()
				found = true
				break
			}
		}
	}

	if !found {
		aghhttp.ErrorAndLog(ctx, l, r, w, http.StatusNotFound, "rule not found")
		return
	}

	// Save config
	if s.conf.ConfModifier != nil {
		s.conf.ConfModifier.Apply(ctx)
	}

	aghhttp.OK(ctx, l, w)
}

// handleDNSRoutingRemoveURL removes a DNS routing rule.
func (s *Server) handleDNSRoutingRemoveURL(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	l := s.logger
	req := &dnsRoutingSetURLReq{} // Reuse struct for ID/URL
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		aghhttp.ErrorAndLog(ctx, l, r, w, http.StatusBadRequest, "decoding request: %s", err)
		return
	}

	s.serverLock.Lock()
	defer s.serverLock.Unlock()

	// Filter out the rule to remove
	newRules := make([]DNSRoutingRule, 0, len(s.conf.DNSRoutingRules))
	var removed bool

	for _, rule := range s.conf.DNSRoutingRules {
		if rule.ID == req.ID || (req.ID == "" && rule.URL == req.URL) {
			removed = true
			continue
		}
		newRules = append(newRules, rule)
	}

	if !removed {
		aghhttp.ErrorAndLog(ctx, l, r, w, http.StatusNotFound, "rule not found")
		return
	}

	s.conf.DNSRoutingRules = newRules

	// Save config
	if s.conf.ConfModifier != nil {
		s.conf.ConfModifier.Apply(ctx)
	}

	aghhttp.OK(ctx, l, w)
}

// handleDNSRoutingRefresh refreshes DNS routing rules.
func (s *Server) handleDNSRoutingRefresh(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement actual refresh logic if needed, or just return OK
	// Since we are using independent engine, maybe trigger a reload?
	aghhttp.OK(r.Context(), s.logger, w)
}
