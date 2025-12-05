package dnsroutingfiles

import (
	"os"
	"sync"
	"time"

	"github.com/AdguardTeam/AdGuardHome/internal/dnsrouting"
)

// ruleFileCache caches parsed rule files with modification time tracking.
// This avoids re-parsing files that haven't changed.
type ruleFileCache struct {
	mu    sync.RWMutex
	cache map[int64]*cachedRuleFile
}

// cachedRuleFile represents a cached rule file with its metadata.
type cachedRuleFile struct {
	rules      []dnsrouting.ParsedRule
	modTime    time.Time
	rulesCount int
}

// newRuleFileCache creates a new rule file cache.
func newRuleFileCache() *ruleFileCache {
	return &ruleFileCache{
		cache: make(map[int64]*cachedRuleFile),
	}
}

// get retrieves cached rules if the file hasn't been modified.
// Returns nil if cache miss or file has been modified.
func (c *ruleFileCache) get(ruleID int64, filePath string) []dnsrouting.ParsedRule {
	c.mu.RLock()
	cached, exists := c.cache[ruleID]
	c.mu.RUnlock()

	if !exists {
		return nil
	}

	// Check if file has been modified
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		// File doesn't exist or error, invalidate cache
		c.invalidate(ruleID)
		return nil
	}

	// Compare modification times
	if !fileInfo.ModTime().Equal(cached.modTime) {
		// File has been modified, invalidate cache
		c.invalidate(ruleID)
		return nil
	}

	// Cache hit
	return cached.rules
}

// set stores parsed rules in the cache with the file's modification time.
func (c *ruleFileCache) set(ruleID int64, filePath string, rules []dnsrouting.ParsedRule) {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		// Can't get file info, don't cache
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.cache[ruleID] = &cachedRuleFile{
		rules:      rules,
		modTime:    fileInfo.ModTime(),
		rulesCount: len(rules),
	}
}

// invalidate removes a rule from the cache.
func (c *ruleFileCache) invalidate(ruleID int64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.cache, ruleID)
}

// clear removes all entries from the cache.
func (c *ruleFileCache) clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cache = make(map[int64]*cachedRuleFile)
}

// stats returns cache statistics.
func (c *ruleFileCache) stats() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	totalRules := 0
	for _, cached := range c.cache {
		totalRules += cached.rulesCount
	}

	return map[string]interface{}{
		"cached_files":  len(c.cache),
		"total_rules":   totalRules,
	}
}
