package dnsrouting

import (
	"strings"
	"sync"
)

// TrieNode represents a node in the domain trie
type TrieNode struct {
	children      map[string]*TrieNode
	upstreamGroup string
	isEnd         bool
	priority      int
}

// DomainTrie implements a trie for efficient domain matching
type DomainTrie struct {
	mu   sync.RWMutex
	root *TrieNode
}

// NewDomainTrie creates a new domain trie
func NewDomainTrie() *DomainTrie {
	return &DomainTrie{
		root: &TrieNode{
			children: make(map[string]*TrieNode),
		},
	}
}

// Insert adds a domain to the trie
// Domain should be in reverse order (com.example.www)
func (t *DomainTrie) Insert(domain string, upstreamGroup string, priority int) {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Split domain into parts (reverse order for suffix matching)
	parts := strings.Split(domain, ".")
	
	// Reverse the parts for suffix matching
	// example.com -> com.example
	for i, j := 0, len(parts)-1; i < j; i, j = i+1, j-1 {
		parts[i], parts[j] = parts[j], parts[i]
	}

	node := t.root
	for _, part := range parts {
		if node.children == nil {
			node.children = make(map[string]*TrieNode)
		}
		
		if _, exists := node.children[part]; !exists {
			node.children[part] = &TrieNode{
				children: make(map[string]*TrieNode),
			}
		}
		node = node.children[part]
	}

	// Mark as end node and store upstream group
	node.isEnd = true
	node.upstreamGroup = upstreamGroup
	node.priority = priority
}

// Search finds the best matching domain in the trie
// Returns upstream group and true if found
func (t *DomainTrie) Search(domain string) (string, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	// Split domain into parts
	parts := strings.Split(strings.ToLower(domain), ".")
	
	// Reverse for suffix matching
	for i, j := 0, len(parts)-1; i < j; i, j = i+1, j-1 {
		parts[i], parts[j] = parts[j], parts[i]
	}

	// Try to find longest suffix match
	node := t.root
	var bestMatch string
	var bestPriority int = int(^uint(0) >> 1) // Max int

	for _, part := range parts {
		if node.children == nil {
			break
		}
		
		next, exists := node.children[part]
		if !exists {
			break
		}
		
		node = next
		
		// Check if this is a valid end node
		if node.isEnd {
			// Use the match with highest priority (lower number)
			if node.priority < bestPriority {
				bestMatch = node.upstreamGroup
				bestPriority = node.priority
			}
		}
	}

	if bestMatch != "" {
		return bestMatch, true
	}
	return "", false
}

// Clear removes all entries from the trie
func (t *DomainTrie) Clear() {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.root = &TrieNode{
		children: make(map[string]*TrieNode),
	}
}

// CompiledRules holds pre-compiled rules for fast matching
type CompiledRules struct {
	mu sync.RWMutex
	
	// Trie for DOMAIN-SUFFIX rules
	suffixTrie *DomainTrie
	
	// Map for exact DOMAIN rules
	exactMap map[string]string
	
	// Slice for DOMAIN-KEYWORD rules (still need linear search)
	keywordRules []compiledKeywordRule
}

type compiledKeywordRule struct {
	keyword       string
	upstreamGroup string
	priority      int
}

// NewCompiledRules creates a new compiled rules structure
func NewCompiledRules() *CompiledRules {
	return &CompiledRules{
		suffixTrie:   NewDomainTrie(),
		exactMap:     make(map[string]string),
		keywordRules: make([]compiledKeywordRule, 0),
	}
}

// AddRule adds a rule to the compiled structure
func (c *CompiledRules) AddRule(rule Rule) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !rule.Enabled {
		return
	}

	domain := strings.ToLower(rule.Domain)

	switch rule.MatchType {
	case MatchTypeDomain:
		// Exact match - use map
		c.exactMap[domain] = rule.UpstreamGroup
		
	case MatchTypeDomainSuffix:
		// Suffix match - use trie
		c.suffixTrie.Insert(domain, rule.UpstreamGroup, rule.Priority)
		
	case MatchTypeDomainKeyword:
		// Keyword match - use slice
		c.keywordRules = append(c.keywordRules, compiledKeywordRule{
			keyword:       domain,
			upstreamGroup: rule.UpstreamGroup,
			priority:      rule.Priority,
		})
	}
}

// Match finds the upstream group for a domain using compiled rules
func (c *CompiledRules) Match(domain string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	domain = strings.ToLower(domain)

	// 1. Try exact match first (fastest - O(1))
	if group, exists := c.exactMap[domain]; exists {
		return group, true
	}

	// 2. Try suffix match (fast - O(k) where k is domain length)
	if group, found := c.suffixTrie.Search(domain); found {
		return group, true
	}

	// 3. Try keyword match (slower - O(n) where n is number of keyword rules)
	for _, rule := range c.keywordRules {
		if strings.Contains(domain, rule.keyword) {
			return rule.upstreamGroup, true
		}
	}

	return "", false
}

// Clear removes all compiled rules
func (c *CompiledRules) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.suffixTrie.Clear()
	c.exactMap = make(map[string]string)
	c.keywordRules = make([]compiledKeywordRule, 0)
}

// Stats returns statistics about compiled rules
func (c *CompiledRules) Stats() map[string]int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return map[string]int{
		"exact_rules":   len(c.exactMap),
		"keyword_rules": len(c.keywordRules),
	}
}
