package dnsforward

import (
	"fmt"
	"sync"
	"testing"
)

func TestDomainCache_BasicOperations(t *testing.T) {
	cache := NewDomainCache(3)

	// Test Set and Get
	cache.Set("example.com", "group1")
	if val, ok := cache.Get("example.com"); !ok || val != "group1" {
		t.Errorf("Expected group1, got %s, ok=%v", val, ok)
	}

	// Test non-existent key
	if _, ok := cache.Get("notfound.com"); ok {
		t.Error("Expected key not found")
	}

	// Test negative cache (empty upstream)
	cache.Set("nogroup.com", "")
	if val, ok := cache.Get("nogroup.com"); !ok || val != "" {
		t.Errorf("Expected empty string, got %s, ok=%v", val, ok)
	}
}

func TestDomainCache_LRUEviction(t *testing.T) {
	cache := NewDomainCache(3)

	// Fill cache to capacity
	cache.Set("domain1.com", "group1")
	cache.Set("domain2.com", "group2")
	cache.Set("domain3.com", "group3")

	if cache.Len() != 3 {
		t.Errorf("Expected cache size 3, got %d", cache.Len())
	}

	// Add one more, should evict domain1.com (oldest)
	cache.Set("domain4.com", "group4")

	if cache.Len() != 3 {
		t.Errorf("Expected cache size 3 after eviction, got %d", cache.Len())
	}

	// domain1.com should be evicted
	if _, ok := cache.Get("domain1.com"); ok {
		t.Error("domain1.com should have been evicted")
	}

	// Others should still exist
	if _, ok := cache.Get("domain2.com"); !ok {
		t.Error("domain2.com should still exist")
	}
	if _, ok := cache.Get("domain3.com"); !ok {
		t.Error("domain3.com should still exist")
	}
	if _, ok := cache.Get("domain4.com"); !ok {
		t.Error("domain4.com should exist")
	}
}

func TestDomainCache_LRUOrdering(t *testing.T) {
	cache := NewDomainCache(3)

	// Fill cache
	cache.Set("domain1.com", "group1")
	cache.Set("domain2.com", "group2")
	cache.Set("domain3.com", "group3")

	// Access domain1.com to make it most recently used
	cache.Get("domain1.com")

	// Add new entry, should evict domain2.com (now oldest)
	cache.Set("domain4.com", "group4")

	// domain2.com should be evicted
	if _, ok := cache.Get("domain2.com"); ok {
		t.Error("domain2.com should have been evicted")
	}

	// domain1.com should still exist (was accessed)
	if _, ok := cache.Get("domain1.com"); !ok {
		t.Error("domain1.com should still exist")
	}
}

func TestDomainCache_Update(t *testing.T) {
	cache := NewDomainCache(3)

	// Set initial value
	cache.Set("example.com", "group1")

	// Update value
	cache.Set("example.com", "group2")

	// Should get updated value
	if val, ok := cache.Get("example.com"); !ok || val != "group2" {
		t.Errorf("Expected group2, got %s", val)
	}

	// Cache size should still be 1
	if cache.Len() != 1 {
		t.Errorf("Expected cache size 1, got %d", cache.Len())
	}
}

func TestDomainCache_Clear(t *testing.T) {
	cache := NewDomainCache(3)

	// Add entries
	cache.Set("domain1.com", "group1")
	cache.Set("domain2.com", "group2")
	cache.Set("domain3.com", "group3")

	if cache.Len() != 3 {
		t.Errorf("Expected cache size 3, got %d", cache.Len())
	}

	// Clear cache
	cache.Clear()

	if cache.Len() != 0 {
		t.Errorf("Expected cache size 0 after clear, got %d", cache.Len())
	}

	// All entries should be gone
	if _, ok := cache.Get("domain1.com"); ok {
		t.Error("Cache should be empty after clear")
	}
}

func TestDomainCache_Concurrent(t *testing.T) {
	cache := NewDomainCache(100)
	var wg sync.WaitGroup

	// Concurrent writes
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				domain := fmt.Sprintf("domain%d-%d.com", id, j)
				cache.Set(domain, fmt.Sprintf("group%d", id))
			}
		}(i)
	}

	// Concurrent reads
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				domain := fmt.Sprintf("domain%d-%d.com", id, j)
				cache.Get(domain)
			}
		}(i)
	}

	wg.Wait()

	// Cache should not exceed capacity
	if cache.Len() > 100 {
		t.Errorf("Cache size %d exceeds capacity 100", cache.Len())
	}
}

func TestDomainCache_GetStats(t *testing.T) {
	cache := NewDomainCache(10)

	// Add some entries
	cache.Set("domain1.com", "group1")
	cache.Set("domain2.com", "group2")
	cache.Set("domain3.com", "group3")

	stats := cache.GetStats()

	if stats.Size != 3 {
		t.Errorf("Expected size 3, got %d", stats.Size)
	}

	if stats.Capacity != 10 {
		t.Errorf("Expected capacity 10, got %d", stats.Capacity)
	}
}

func TestDomainCache_DefaultCapacity(t *testing.T) {
	// Test with invalid capacity
	cache := NewDomainCache(0)

	// Should use default capacity (1000)
	stats := cache.GetStats()
	if stats.Capacity != 1000 {
		t.Errorf("Expected default capacity 1000, got %d", stats.Capacity)
	}

	cache = NewDomainCache(-1)
	stats = cache.GetStats()
	if stats.Capacity != 1000 {
		t.Errorf("Expected default capacity 1000, got %d", stats.Capacity)
	}
}

// Benchmark tests
func BenchmarkDomainCache_Set(b *testing.B) {
	cache := NewDomainCache(1000)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		domain := fmt.Sprintf("domain%d.com", i%1000)
		cache.Set(domain, "group1")
	}
}

func BenchmarkDomainCache_Get(b *testing.B) {
	cache := NewDomainCache(1000)

	// Pre-populate cache
	for i := 0; i < 1000; i++ {
		domain := fmt.Sprintf("domain%d.com", i)
		cache.Set(domain, "group1")
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		domain := fmt.Sprintf("domain%d.com", i%1000)
		cache.Get(domain)
	}
}

func BenchmarkDomainCache_SetGet(b *testing.B) {
	cache := NewDomainCache(1000)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		domain := fmt.Sprintf("domain%d.com", i%1000)
		cache.Set(domain, "group1")
		cache.Get(domain)
	}
}
