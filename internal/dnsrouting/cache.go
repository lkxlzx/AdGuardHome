package dnsrouting

import (
	"container/list"
	"sync"
	"time"
)

// CacheEntry represents a cached routing result
type CacheEntry struct {
	Domain        string
	UpstreamGroup string
	Matched       bool
	Timestamp     time.Time
}

// LRUCache implements a thread-safe LRU cache for DNS routing results
type LRUCache struct {
	mu       sync.RWMutex
	capacity int
	ttl      time.Duration
	items    map[string]*list.Element
	lruList  *list.List
}

// cacheItem is the internal structure stored in the LRU list
type cacheItem struct {
	key   string
	entry CacheEntry
}

// NewLRUCache creates a new LRU cache with the specified capacity and TTL
func NewLRUCache(capacity int, ttl time.Duration) *LRUCache {
	return &LRUCache{
		capacity: capacity,
		ttl:      ttl,
		items:    make(map[string]*list.Element),
		lruList:  list.New(),
	}
}

// Get retrieves a value from the cache
// Returns the cached entry and true if found and not expired, otherwise nil and false
func (c *LRUCache) Get(domain string) (*CacheEntry, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	element, exists := c.items[domain]
	if !exists {
		return nil, false
	}

	item := element.Value.(*cacheItem)

	// Check if entry has expired
	if c.ttl > 0 && time.Since(item.entry.Timestamp) > c.ttl {
		// Remove expired entry
		c.lruList.Remove(element)
		delete(c.items, domain)
		return nil, false
	}

	// Move to front (most recently used)
	c.lruList.MoveToFront(element)

	return &item.entry, true
}

// Put adds or updates a value in the cache
func (c *LRUCache) Put(domain string, upstreamGroup string, matched bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry := CacheEntry{
		Domain:        domain,
		UpstreamGroup: upstreamGroup,
		Matched:       matched,
		Timestamp:     time.Now(),
	}

	// Check if key already exists
	if element, exists := c.items[domain]; exists {
		// Update existing entry
		c.lruList.MoveToFront(element)
		element.Value.(*cacheItem).entry = entry
		return
	}

	// Add new entry
	item := &cacheItem{
		key:   domain,
		entry: entry,
	}
	element := c.lruList.PushFront(item)
	c.items[domain] = element

	// Evict least recently used if over capacity
	if c.lruList.Len() > c.capacity {
		c.evictOldest()
	}
}

// evictOldest removes the least recently used item
func (c *LRUCache) evictOldest() {
	element := c.lruList.Back()
	if element != nil {
		c.lruList.Remove(element)
		item := element.Value.(*cacheItem)
		delete(c.items, item.key)
	}
}

// Clear removes all entries from the cache
func (c *LRUCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = make(map[string]*list.Element)
	c.lruList = list.New()
}

// Len returns the current number of items in the cache
func (c *LRUCache) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.lruList.Len()
}

// Stats returns cache statistics
type CacheStats struct {
	Size     int
	Capacity int
	HitRate  float64
}

// cacheMetrics tracks cache performance
type cacheMetrics struct {
	mu    sync.RWMutex
	hits  uint64
	total uint64
}

// RecordHit records a cache hit
func (m *cacheMetrics) RecordHit() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.hits++
	m.total++
}

// RecordMiss records a cache miss
func (m *cacheMetrics) RecordMiss() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.total++
}

// GetHitRate returns the cache hit rate
func (m *cacheMetrics) GetHitRate() float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.total == 0 {
		return 0.0
	}
	return float64(m.hits) / float64(m.total)
}

// Reset resets the metrics
func (m *cacheMetrics) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.hits = 0
	m.total = 0
}
