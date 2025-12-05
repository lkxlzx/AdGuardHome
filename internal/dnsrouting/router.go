package dnsrouting

import (
	"context"
	"fmt"
	"log/slog"
	"sort"
	"strings"
	"sync"
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

	r.sources[source.ID] = source
	
	// Rebuild sorted sources for performance
	r.rebuildSortedSources()
	
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

	r.sources[source.ID] = source
	
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

	r.customRules = rules
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
func (r *Router) Match(ctx context.Context, domain string) (upstreamGroup string, matched bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	domain = strings.ToLower(strings.TrimSuffix(domain, "."))

	// First, check custom rules (highest priority)
	for _, rule := range r.customRules {
		if !rule.Enabled {
			continue
		}

		if r.matchRule(domain, rule) {
			// Use Debug level for match logging to avoid log spam at high QPS
			// Set log level to DEBUG to see these messages
			r.logger.DebugContext(ctx, "matched custom rule",
				"domain", domain,
				"rule_domain", rule.Domain,
				"match_type", rule.MatchType,
				"upstream_group", rule.UpstreamGroup,
			)
			return rule.UpstreamGroup, true
		}
	}

	// Then, check source rules by priority (use pre-sorted list)
	for _, source := range r.sortedSources {
		if !source.Enabled {
			continue
		}
		for _, rule := range source.Rules {
			if !rule.Enabled {
				continue
			}

			if r.matchRule(domain, rule) {
				// Use Debug level for match logging to avoid log spam at high QPS
				// Set log level to DEBUG to see these messages
				r.logger.DebugContext(ctx, "matched source rule",
					"domain", domain,
					"source_id", source.ID,
					"source_name", source.Name,
					"rule_domain", rule.Domain,
					"match_type", rule.MatchType,
					"upstream_group", rule.UpstreamGroup,
				)
				return rule.UpstreamGroup, true
			}
		}
	}

	return "", false
}

// matchRule checks if a domain matches a rule.
func (r *Router) matchRule(domain string, rule Rule) bool {
	ruleDomain := strings.ToLower(strings.TrimSuffix(rule.Domain, "."))

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
// Must be called with write lock held.
func (r *Router) rebuildSortedSources() {
	r.sortedSources = make([]*RuleSource, 0, len(r.sources))
	for _, source := range r.sources {
		r.sortedSources = append(r.sortedSources, source)
	}
	
	// Sort by priority (lower number = higher priority)
	sort.Slice(r.sortedSources, func(i, j int) bool {
		return r.sortedSources[i].Priority < r.sortedSources[j].Priority
	})
}
