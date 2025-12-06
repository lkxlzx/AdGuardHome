package dnsrouting

import (
	"context"
	"fmt"
	"log/slog"
	"sort"
	"strings"
	"sync"
	"time"
)

// MatchType represents the type of domain matching.
type MatchType string

const (
	// MatchTypeDomain matches exact domain.
	MatchTypeDomain MatchType = "DOMAIN"
	// MatchTypeDomainSuffix matches domain and all its subdomains.
	MatchTypeDomainSuffix MatchType = "DOMAIN-SUFFIX"
	// MatchTypeDomainKeyword matches domains containing the keyword.
	MatchTypeDomainKeyword MatchType = "DOMAIN-KEYWORD"
)

// Rule represents a DNS routing rule.
type Rule struct {
	// Domain is the domain pattern to match.
	Domain string
	// MatchType is the type of matching.
	MatchType MatchType
	// UpstreamGroup is the ID of the upstream group to use.
	UpstreamGroup string
	// Priority is the rule priority (lower = higher priority).
	Priority int
	// Enabled indicates if the rule is active.
	Enabled bool
	
	// normalizedDomain is the pre-processed domain (lowercase, no trailing dot).
	// This is computed once when the rule is created to avoid repeated string operations.
	normalizedDomain string
}

// RuleSource represents a source of routing rules (from URL or custom).
type RuleSource struct {
	// ID is the unique identifier.
	ID int64
	// Name is the human-readable name.
	Name string
	// URL is the source URL (empty for custom rules).
	URL string
	// UpstreamGroup is the target upstream group.
	UpstreamGroup string
	// Priority is the source priority.
	Priority int
	// Enabled indicates if the source is active.
	Enabled bool
	// Rules are the parsed rules from this source.
	Rules []Rule
	// RulesCount is the number of rules.
	RulesCount int
	// LastUpdated is the last update time.
	LastUpdated string
}

// Router manages DNS routing rules and performs domain matching.
type Router struct {
	// mu protects the rules.
	mu sync.RWMutex
	
	// sources are the rule sources (from URLs).
	sources map[int64]*RuleSource
	
	// sortedSources are sources sorted by priority (cached for performance).
	sortedSources []*RuleSource
	
	// customRules are user-defined rules.
	customRules []Rule
	
	// cache is the LRU cache for routing results.
	cache *LRUCache
	
	// metrics tracks cache performance.
	metrics *cacheMetrics
	
	// logger is the logger instance.
	logger *slog.Logger
}

// NewRouter creates a new DNS router.
// The router is initialized with empty rule sets.
// Use AddSource() to add rule sources, which will automatically
// build the sorted sources list for optimal matching performance.
func NewRouter(logger *slog.Logger) *Router {
	return &Router{
		sources:       make(map[int64]*RuleSource),
		sortedSources: make([]*RuleSource, 0), // Will be populated by AddSource/UpdateSource
		customRules:   make([]Rule, 0),
		cache:         NewLRUCache(10000, 5*time.Minute), // Cache 10000 domains for 5 minutes
		metrics:       &cacheMetrics{},
		logger:        logger,
	}
}

// NewRouterWithCache creates a new DNS router with custom cache settings.
func NewRouterWithCache(logger *slog.Logger, cacheSize int, cacheTTL time.Duration) *Router {
	return &Router{
		sources:       make(map[int64]*RuleSource),
		sortedSources: make([]*RuleSource, 0),
		customRules:   make([]Rule, 0),
		cache:         NewLRUCache(cacheSize, cacheTTL),
		metrics:       &cacheMetrics{},
		logger:        logger,
	}
}

// AddSource adds a new rule source.
func (r *Router) AddSource(source *RuleSource) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.sources[source.ID]; exists {
		return fmt.Errorf("source with ID %d already exists", source.ID)
	}

	// Pre-normalize all rule domains for performance
	normalizeRules(source.Rules)
	
	r.sources[source.ID] = source
	
	// Rebuild sorted sources for performance
	r.rebuildSortedSources()
	
	// Clear cache when sources change
	if r.cache != nil {
		r.cache.Clear()
	}
	
	r.logger.InfoContext(context.TODO(), "added routing source",
		"id", source.ID,
		"name", source.Name,
		"rules_count", source.RulesCount,
	)

	return nil
}

// UpdateSource updates an existing rule source.
func (r *Router) UpdateSource(source *RuleSource) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.sources[source.ID]; !exists {
		return fmt.Errorf("source with ID %d not found", source.ID)
	}

	// Pre-normalize all rule domains for performance
	normalizeRules(source.Rules)
	
	r.sources[source.ID] = source
	
	// Clear cache when sources change
	if r.cache != nil {
		r.cache.Clear()
	}
	
	// Rebuild sorted sources for performance
	r.rebuildSortedSources()
	
	r.logger.InfoContext(context.TODO(), "updated routing source",
		"id", source.ID,
		"name", source.Name,
		"rules_count", source.RulesCount,
	)

	return nil
}

// RemoveSource removes a rule source.
func (r *Router) RemoveSource(id int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.sources[id]; !exists {
		return fmt.Errorf("source with ID %d not found", id)
	}

	delete(r.sources, id)
	
	// Rebuild sorted sources for performance
	r.rebuildSortedSources()
	
	r.logger.InfoContext(context.TODO(), "removed routing source", "id", id)

	return nil
}

// GetSource returns a rule source by ID.
func (r *Router) GetSource(id int64) (*RuleSource, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	source, exists := r.sources[id]
	return source, exists
}

// GetAllSources returns all rule sources.
func (r *Router) GetAllSources() []*RuleSource {
	r.mu.RLock()
	defer r.mu.RUnlock()

	sources := make([]*RuleSource, 0, len(r.sources))
	for _, source := range r.sources {
		sources = append(sources, source)
	}

	return sources
}

// SetCustomRules sets the custom routing rules.
func (r *Router) SetCustomRules(rules []Rule) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Pre-normalize all rule domains for performance
	normalizeRules(rules)
	
	r.customRules = rules
	
	// Clear cache when rules change
	if r.cache != nil {
		r.cache.Clear()
	}
	
	r.logger.InfoContext(context.TODO(), "updated custom routing rules", "count", len(rules))
}

// GetCustomRules returns the custom routing rules.
func (r *Router) GetCustomRules() []Rule {
	r.mu.RLock()
	defer r.mu.RUnlock()

	rules := make([]Rule, len(r.customRules))
	copy(rules, r.customRules)
	return rules
}

// Match finds the upstream group for a given domain.
// Returns the upstream group ID and true if a match is found.
// 
// Note: The ctx parameter is kept for API compatibility but is not currently used
// in the hot path for performance reasons. All debug logging has been removed from
// the matching logic to maximize throughput.
func (r *Router) Match(ctx context.Context, domain string) (upstreamGroup string, matched bool) {
	// Normalize domain once at the entry point
	domain = normalizeDomain(domain)

	// Check cache first (without holding the lock)
	if r.cache != nil {
		if entry, found := r.cache.Get(domain); found {
			r.metrics.RecordHit()
			// Removed debug logging from hot path for performance
			// Cache hit is the most common case and logging here impacts QPS significantly
			return entry.UpstreamGroup, entry.Matched
		}
		r.metrics.RecordMiss()
	}

	// Cache miss, perform full match
	r.mu.RLock()
	defer r.mu.RUnlock()

	// First, check custom rules (highest priority)
	for _, rule := range r.customRules {
		if !rule.Enabled {
			continue
		}

		if r.matchRule(domain, rule) {
			// Removed debug logging from hot path for performance
			// Logging on every match significantly impacts QPS in high-traffic scenarios
			
			// Cache the result
			if r.cache != nil {
				r.cache.Put(domain, rule.UpstreamGroup, true)
			}
			
			return rule.UpstreamGroup, true
		}
	}

	// Then, check source rules by priority (use pre-sorted list)
	// Note: sortedSources only contains enabled sources, so no need to check source.Enabled
	for _, source := range r.sortedSources {
		for _, rule := range source.Rules {
			if !rule.Enabled {
				continue
			}

			if r.matchRule(domain, rule) {
				// Removed debug logging from hot path for performance
				// Logging on every match significantly impacts QPS in high-traffic scenarios
				
				// Cache the result
				if r.cache != nil {
					r.cache.Put(domain, rule.UpstreamGroup, true)
				}
				
				return rule.UpstreamGroup, true
			}
		}
	}

	// No match found, cache the negative result
	if r.cache != nil {
		r.cache.Put(domain, "", false)
	}

	return "", false
}

// matchRule checks if a domain matches a rule.
// Uses pre-normalized domain from rule to avoid repeated string operations.
func (r *Router) matchRule(domain string, rule Rule) bool {
	// Use pre-normalized domain (computed once when rule was created)
	ruleDomain := rule.normalizedDomain

	switch rule.MatchType {
	case MatchTypeDomain:
		// Exact match
		return domain == ruleDomain

	case MatchTypeDomainSuffix:
		// Match domain and all subdomains
		if domain == ruleDomain {
			return true
		}
		// Check if it's a subdomain
		return strings.HasSuffix(domain, "."+ruleDomain)

	case MatchTypeDomainKeyword:
		// Contains keyword
		return strings.Contains(domain, ruleDomain)

	default:
		return false
	}
}

// normalizeDomain normalizes a domain name for matching.
// It converts to lowercase and removes trailing dot.
func normalizeDomain(domain string) string {
	return strings.ToLower(strings.TrimSuffix(domain, "."))
}

// normalizeRules pre-processes all rules by normalizing their domains.
// This is done once when rules are loaded to avoid repeated string operations during matching.
func normalizeRules(rules []Rule) {
	for i := range rules {
		rules[i].normalizedDomain = normalizeDomain(rules[i].Domain)
	}
}

// Stats returns routing statistics.
func (r *Router) Stats() map[string]interface{} {
	r.mu.RLock()
	defer r.mu.RUnlock()

	totalRules := len(r.customRules)
	enabledSources := 0
	for _, source := range r.sources {
		if source.Enabled {
			enabledSources++
			totalRules += source.RulesCount
		}
	}

	return map[string]interface{}{
		"total_sources":   len(r.sources),
		"enabled_sources": enabledSources,
		"custom_rules":    len(r.customRules),
		"total_rules":     totalRules,
	}
}

// rebuildSortedSources rebuilds the sorted sources list.
// Only includes enabled sources to avoid checking disabled sources during matching.
// Must be called with write lock held.
func (r *Router) rebuildSortedSources() {
	r.sortedSources = make([]*RuleSource, 0, len(r.sources))
	for _, source := range r.sources {
		// Only include enabled sources in the sorted list
		// This avoids checking source.Enabled on every match
		if source.Enabled {
			r.sortedSources = append(r.sortedSources, source)
		}
	}
	
	// Sort by priority (lower number = higher priority)
	sort.Slice(r.sortedSources, func(i, j int) bool {
		return r.sortedSources[i].Priority < r.sortedSources[j].Priority
	})
}

// ClearCache clears the routing cache.
func (r *Router) ClearCache() {
	if r.cache != nil {
		r.cache.Clear()
		r.logger.InfoContext(context.TODO(), "cleared routing cache")
	}
}

// GetCacheStats returns cache statistics.
func (r *Router) GetCacheStats() map[string]interface{} {
	if r.cache == nil {
		return map[string]interface{}{
			"enabled": false,
		}
	}

	return map[string]interface{}{
		"enabled":  true,
		"size":     r.cache.Len(),
		"capacity": r.cache.capacity,
		"hit_rate": r.metrics.GetHitRate(),
	}
}

// ResetCacheMetrics resets the cache performance metrics.
func (r *Router) ResetCacheMetrics() {
	if r.metrics != nil {
		r.metrics.Reset()
		r.logger.InfoContext(context.TODO(), "reset cache metrics")
	}
}
