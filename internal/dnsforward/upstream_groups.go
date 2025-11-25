package dnsforward

import (
	"sort"
	"strings"
	"sync"

	"github.com/AdguardTeam/dnsproxy/proxy"
	"github.com/AdguardTeam/dnsproxy/upstream"
)

// upstreamConfigCache caches CustomUpstreamConfig objects for each upstream group
// to avoid recreating them (and their internal caches) on every DNS query
type upstreamConfigCache struct {
	mu      sync.RWMutex
	configs map[string]*proxy.CustomUpstreamConfig
}

// newUpstreamConfigCache creates a new upstream config cache
func newUpstreamConfigCache() *upstreamConfigCache {
	return &upstreamConfigCache{
		configs: make(map[string]*proxy.CustomUpstreamConfig),
	}
}

// Get retrieves a cached config or creates a new one
func (c *upstreamConfigCache) Get(groupID string, createFunc func() *proxy.CustomUpstreamConfig) *proxy.CustomUpstreamConfig {
	// Try read lock first
	c.mu.RLock()
	config, exists := c.configs[groupID]
	c.mu.RUnlock()

	if exists {
		return config
	}

	// Need to create, use write lock
	c.mu.Lock()
	defer c.mu.Unlock()

	// Double-check after acquiring write lock
	if config, exists := c.configs[groupID]; exists {
		return config
	}

	// Create new config
	config = createFunc()
	if config != nil {
		c.configs[groupID] = config
	}

	return config
}

// Clear removes all cached configs
func (c *upstreamConfigCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.configs = make(map[string]*proxy.CustomUpstreamConfig)
}

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

// Note: The following functions were removed as they were unused:
// - GetUpstreamStrings() - can be re-added from v2 branch if needed
// - GetDefaultUpstreamGroup() - can be re-added from v2 branch if needed
// - GetEnabledUpstreamGroups() - removed (unused)
// - GetUpstreamGroupByName() - removed (unused)

// GetUpstreamGroupForDomain returns the upstream group ID for a given domain
// based on DNS routing rules (whitelist filters with upstream groups).
// Returns empty string if no matching rule is found.
// Uses LRU cache to improve performance for frequently queried domains.
func (s *Server) GetUpstreamGroupForDomain(domain string) string {
	// Remove trailing dot from FQDN
	domain = strings.TrimSuffix(domain, ".")
	domain = strings.ToLower(domain)

	// Check cache first
	if s.domainCache != nil {
		if upstreamID, found := s.domainCache.Get(domain); found {
			s.logger.Debug("cache hit", "domain", domain, "upstream_group", upstreamID)
			return upstreamID
		}
	}

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
			// Cache the result
			if s.domainCache != nil {
				s.domainCache.Set(domain, rule.UpstreamGroup)
			}
			return rule.UpstreamGroup
		}
	}

	// Sort rules by priority (descending)
	sortedRules := make([]DNSRoutingRule, len(s.conf.DNSRoutingRules))
	copy(sortedRules, s.conf.DNSRoutingRules)
	sort.Slice(sortedRules, func(i, j int) bool {
		return sortedRules[i].Priority > sortedRules[j].Priority
	})

	// Check rules in priority order
	for _, rule := range sortedRules {
		// Check each domain pattern in the rule
		for _, pattern := range rule.Domains {
			if matchDomainPattern(domain, pattern) {
				s.logger.Debug("matched dns routing rule", "domain", domain, "pattern", pattern, "group_id", rule.GroupID, "priority", rule.Priority)
				// Cache the result
				if s.domainCache != nil {
					s.domainCache.Set(domain, rule.GroupID)
				}
				return rule.GroupID
			}
		}
	}

	s.logger.Debug("no matching rule found", "domain", domain)
	// Cache negative result (no match)
	if s.domainCache != nil {
		s.domainCache.Set(domain, "")
	}
	return ""
}

// Note: matchDomainPattern is now defined in domain_match.go to avoid code duplication

// ClearDomainCache clears the domain routing cache.
// This should be called when DNS routing rules or custom domain rules are updated.
func (s *Server) ClearDomainCache() {
	if s.domainCache != nil {
		s.domainCache.Clear()
		s.logger.Debug("domain cache cleared")
	}
}

// GetDomainCacheStats returns statistics about the domain routing cache.
func (s *Server) GetDomainCacheStats() CacheStats {
	if s.domainCache != nil {
		return s.domainCache.GetStats()
	}
	return CacheStats{}
}
