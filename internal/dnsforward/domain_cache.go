package dnsforward

import (
	"container/list"
	"sync"
)

// DomainCache is a thread-safe LRU cache for domain to upstream group mappings
type DomainCache struct {
	mu       sync.RWMutex
	capacity int
	cache    map[string]*list.Element
	lruList  *list.List
}

// cacheEntry represents a cache entry
type cacheEntry struct {
	domain      string
	upstreamID  string
	isNegative  bool // true if no match was found
}

// NewDomainCache creates a new domain cache with the specified capacity
func NewDomainCache(capacity int) *DomainCache {
	if capacity <= 0 {
		capacity = 1000 // default capacity
	}
	
	return &DomainCache{
		capacity: capacity,
		cache:    make(map[string]*list.Element, capacity),
		lruList:  list.New(),
	}
}

// Get retrieves a value from the cache
// Returns (upstreamID, found)
func (c *DomainCache) Get(domain string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if elem, ok := c.cache[domain]; ok {
		// Move to front (most recently used)
		c.lruList.MoveToFront(elem)
		entry := elem.Value.(*cacheEntry)
		return entry.upstreamID, true
	}

	return "", false
}

// Set adds or updates a value in the cache
func (c *DomainCache) Set(domain, upstreamID string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if already exists
	if elem, ok := c.cache[domain]; ok {
		// Update existing entry
		c.lruList.MoveToFront(elem)
		entry := elem.Value.(*cacheEntry)
		entry.upstreamID = upstreamID
		entry.isNegative = (upstreamID == "")
		return
	}

	// Add new entry
	entry := &cacheEntry{
		domain:     domain,
		upstreamID: upstreamID,
		isNegative: (upstreamID == ""),
	}
	elem := c.lruList.PushFront(entry)
	c.cache[domain] = elem

	// Evict oldest if capacity exceeded
	if c.lruList.Len() > c.capacity {
		c.evictOldest()
	}
}

// evictOldest removes the least recently used entry
// Must be called with lock held
func (c *DomainCache) evictOldest() {
	elem := c.lruList.Back()
	if elem != nil {
		c.lruList.Remove(elem)
		entry := elem.Value.(*cacheEntry)
		delete(c.cache, entry.domain)
	}
}

// Clear removes all entries from the cache
func (c *DomainCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cache = make(map[string]*list.Element, c.capacity)
	c.lruList = list.New()
}

// Len returns the current number of entries in the cache
func (c *DomainCache) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.lruList.Len()
}

// Stats returns cache statistics
type CacheStats struct {
	Size     int
	Capacity int
}

// GetStats returns current cache statistics
func (c *DomainCache) GetStats() CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return CacheStats{
		Size:     c.lruList.Len(),
		Capacity: c.capacity,
	}
}
