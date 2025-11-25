package dnsforward

import (
	"fmt"
	"strings"

	"github.com/AdguardTeam/dnsproxy/proxy"
	"github.com/AdguardTeam/dnsproxy/upstream"
)

// GetUpstreamGroupByID returns the upstream group for a given group ID.
// This function is used by DNS routing rules to select upstreams based on domain matching.
// If the group is not found or disabled, it returns nil.
func (s *Server) GetUpstreamGroupByID(groupID string) *UpstreamGroup {
	s.serverLock.RLock()
	defer s.serverLock.RUnlock()

	// Find the group by ID
	for i := range s.conf.UpstreamGroups {
		if s.conf.UpstreamGroups[i].ID == groupID {
			group := &s.conf.UpstreamGroups[i]
			// Only return if enabled
			if group.Enabled {
				return group
			}
			return nil
		}
	}

	return nil
}

// GetUpstreamStrings returns the upstream server addresses for a given group ID.
// Returns a slice of upstream server addresses (e.g., ["8.8.8.8", "1.1.1.1"]).
// Empty entries and comments (entries starting with #) are filtered out.
func (s *Server) GetUpstreamStrings(groupID string) []string {
	group := s.GetUpstreamGroupByID(groupID)
	if group == nil {
		return nil
	}

	// Filter out empty entries and comments
	var validUpstreams []string
	for _, u := range group.Upstreams {
		u = strings.TrimSpace(u)
		if u != "" && !strings.HasPrefix(u, "#") {
			validUpstreams = append(validUpstreams, u)
		}
	}

	return validUpstreams
}

// GetDefaultUpstreamGroup returns the default upstream group.
// This is used when no routing rules match the DNS query.
func (s *Server) GetDefaultUpstreamGroup() *UpstreamGroup {
	s.serverLock.RLock()
	defer s.serverLock.RUnlock()

	for i := range s.conf.UpstreamGroups {
		if s.conf.UpstreamGroups[i].IsDefault && s.conf.UpstreamGroups[i].Enabled {
			return &s.conf.UpstreamGroups[i]
		}
	}

	return nil
}

// Note: GetEnabledUpstreamGroups and GetUpstreamGroupByName were removed as they were unused.
// If needed in the future, they can be re-added from v1 branch.

// GetUpstreamGroupForDomain returns the upstream group ID for a given domain
// based on DNS routing rules (whitelist filters with upstream groups).
// Returns empty string if no matching rule is found.
func (s *Server) GetUpstreamGroupForDomain(domain string) string {
	// Remove trailing dot from FQDN
	domain = strings.TrimSuffix(domain, ".")
	domain = strings.ToLower(domain)

	s.serverLock.RLock()
	defer s.serverLock.RUnlock()

	// Debug: log the number of custom rules
	s.logger.Debug("checking custom domain rules", "domain", domain, "rules_count", len(s.conf.CustomDomainRules))

	// First, check custom domain rules (higher priority)
	for i, rule := range s.conf.CustomDomainRules {
		// Skip disabled rules
		if !rule.Enabled {
			continue
		}
		
		pattern := rule.MatchType + "," + rule.Domain
		s.logger.Debug("checking custom rule", "index", i, "pattern", pattern, "upstream_group", rule.UpstreamGroup)
		
		if matchDomainPattern(domain, pattern) {
			s.logger.Debug("matched custom domain rule", "domain", domain, "pattern", pattern, "upstream_group", rule.UpstreamGroup)
			return rule.UpstreamGroup
		}
	}

	// Then, check DNS routing rules from filter lists
	for _, rule := range s.conf.DnsRoutingRules {
		if !rule.Enabled {
			continue
		}

		// Check each domain pattern in the rule
		for _, pattern := range rule.Domains {
			if matchDomainPattern(domain, pattern) {
				s.logger.Debug("matched dns routing rule", "domain", domain, "pattern", pattern, "group_id", rule.GroupID)
				return rule.GroupID
			}
		}
	}

	s.logger.Debug("no matching rule found", "domain", domain)
	return ""
}

// Note: matchDomainPattern is now defined in domain_match.go to avoid code duplication

// createCustomUpstreamConfig creates a CustomUpstreamConfig from upstream server addresses.
func (s *Server) createCustomUpstreamConfig(upstreamAddrs []string) (*proxy.CustomUpstreamConfig, error) {
	if len(upstreamAddrs) == 0 {
		return nil, fmt.Errorf("no upstream addresses provided")
	}

	// Parse upstream addresses
	upstreams := make([]upstream.Upstream, 0, len(upstreamAddrs))
	for _, addr := range upstreamAddrs {
		// Use nil options for simplicity - will use default settings
		ups, err := upstream.AddressToUpstream(addr, nil)
		if err != nil {
			s.logger.Error("failed to parse upstream", "address", addr, "error", err)
			continue
		}

		upstreams = append(upstreams, ups)
	}

	if len(upstreams) == 0 {
		return nil, fmt.Errorf("no valid upstreams")
	}

	// Create upstream config
	upsConf := &proxy.UpstreamConfig{
		Upstreams: upstreams,
	}

	// Create custom upstream config
	// Convert CacheSize from uint32 to int
	cacheSize := int(s.conf.CacheSize)
	customConf := proxy.NewCustomUpstreamConfig(
		upsConf,
		true, // use parallel queries
		cacheSize,
		s.conf.EDNSClientSubnet.Enabled,
	)

	return customConf, nil
}


// getUpstreamGroupByID is an internal function to get upstream group by ID.
// It's used by DNS routing logic in process.go.
func (s *Server) getUpstreamGroupByID(groupID string) *UpstreamGroup {
	return s.GetUpstreamGroupByID(groupID)
}

// createUpstreamConfigFromGroup creates a CustomUpstreamConfig from an UpstreamGroup.
// This is used by DNS routing to create upstream configuration for matched domains.
func (s *Server) createUpstreamConfigFromGroup(group *UpstreamGroup) *proxy.CustomUpstreamConfig {
	if group == nil || len(group.Upstreams) == 0 {
		return nil
	}

	// Parse upstream addresses
	upstreams := make([]upstream.Upstream, 0, len(group.Upstreams))
	for _, addr := range group.Upstreams {
		addr = strings.TrimSpace(addr)
		if addr == "" || strings.HasPrefix(addr, "#") {
			continue
		}

		u, err := upstream.AddressToUpstream(addr, nil)
		if err != nil {
			s.logger.Error("failed to parse upstream address", "addr", addr, "error", err)
			continue
		}
		upstreams = append(upstreams, u)
	}

	if len(upstreams) == 0 {
		return nil
	}

	// Create upstream config
	upsConf := &proxy.UpstreamConfig{
		Upstreams: upstreams,
	}

	// Create custom upstream config
	cacheSize := int(s.conf.CacheSize)
	customConf := proxy.NewCustomUpstreamConfig(
		upsConf,
		true, // use parallel queries
		cacheSize,
		s.conf.EDNSClientSubnet.Enabled,
	)

	return customConf
}
