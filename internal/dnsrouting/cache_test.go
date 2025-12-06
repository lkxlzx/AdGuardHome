package dnsrouting

import (
	"testing"
	"time"
)

func TestLRUCache_Basic(t *testing.T) {
	cache := NewLRUCache(3, 0) // No TTL

	// Test Put and Get
	cache.Put("example.com", "group1", true)
	entry, found := cache.Get("example.com")
	if !found {
		t.Fatal("Expected to find example.com in cache")
	}
	if entry.UpstreamGroup != "group1" {
		t.Errorf("Expected group1, got %s", entry.UpstreamGroup)
	}

	// Test cache miss
	_, found = cache.Get("notfound.com")
	if found {
		t.Error("Expected cache miss for notfound.com")
	}
}

func TestLRUCache_Eviction(t *testing.T) {
	cache := NewLRUCache(2, 0) // Capacity of 2

	cache.Put("domain1.com", "group1", true)
	cache.Put("domain2.com", "group2", true)
	cache.Put("domain3.com", "group3", true) // Should evict domain1

	// domain1 should be evicted
	_, found := cache.Get("domain1.com")
	if found {
		t.Error("domain1.com should have been evicted")
	}

	// domain2 and domain3 should still be there
	_, found = cache.Get("domain2.com")
	if !found {
		t.Error("domain2.com should be in cache")
	}

	_, found = cache.Get("domain3.com")
	if !found {
		t.Error("domain3.com should be in cache")
	}
}

func TestLRUCache_LRUOrder(t *testing.T) {
	cache := NewLRUCache(2, 0)

	cache.Put("domain1.com", "group1", true)
	cache.Put("domain2.com", "group2", true)

	// Access domain1 to make it most recently used
	cache.Get("domain1.com")

	// Add domain3, should evict domain2 (least recently used)
	cache.Put("domain3.com", "group3", true)

	// domain2 should be evicted
	_, found := cache.Get("domain2.com")
	if found {
		t.Error("domain2.com should have been evicted")
	}

	// domain1 and domain3 should still be there
	_, found = cache.Get("domain1.com")
	if !found {
		t.Error("domain1.com should be in cache")
	}

	_, found = cache.Get("domain3.com")
	if !found {
		t.Error("domain3.com should be in cache")
	}
}

func TestLRUCache_TTL(t *testing.T) {
	cache := NewLRUCache(10, 100*time.Millisecond)

	cache.Put("example.com", "group1", true)

	// Should be found immediately
	_, found := cache.Get("example.com")
	if !found {
		t.Error("Entry should be found immediately")
	}

	// Wait for TTL to expire
	time.Sleep(150 * time.Millisecond)

	// Should not be found after TTL
	_, found = cache.Get("example.com")
	if found {
		t.Error("Entry should have expired")
	}
}

func TestLRUCache_Update(t *testing.T) {
	cache := NewLRUCache(10, 0)

	cache.Put("example.com", "group1", true)
	cache.Put("example.com", "group2", true) // Update

	entry, found := cache.Get("example.com")
	if !found {
		t.Fatal("Entry should be found")
	}

	if entry.UpstreamGroup != "group2" {
		t.Errorf("Expected group2, got %s", entry.UpstreamGroup)
	}
}

func TestLRUCache_Clear(t *testing.T) {
	cache := NewLRUCache(10, 0)

	cache.Put("domain1.com", "group1", true)
	cache.Put("domain2.com", "group2", true)

	if cache.Len() != 2 {
		t.Errorf("Expected cache size 2, got %d", cache.Len())
	}

	cache.Clear()

	if cache.Len() != 0 {
		t.Errorf("Expected cache size 0 after clear, got %d", cache.Len())
	}

	_, found := cache.Get("domain1.com")
	if found {
		t.Error("Cache should be empty after clear")
	}
}

func TestCacheMetrics(t *testing.T) {
	metrics := &cacheMetrics{}

	// Record some hits and misses
	metrics.RecordHit()
	metrics.RecordHit()
	metrics.RecordMiss()
	metrics.RecordMiss()
	metrics.RecordMiss()

	hitRate := metrics.GetHitRate()
	expected := 2.0 / 5.0 // 2 hits out of 5 total

	if hitRate != expected {
		t.Errorf("Expected hit rate %.2f, got %.2f", expected, hitRate)
	}

	// Test reset
	metrics.Reset()
	hitRate = metrics.GetHitRate()
	if hitRate != 0.0 {
		t.Errorf("Expected hit rate 0.0 after reset, got %.2f", hitRate)
	}
}

func BenchmarkLRUCache_Get(b *testing.B) {
	cache := NewLRUCache(1000, 0)

	// Pre-populate cache
	for i := 0; i < 1000; i++ {
		domain := "domain" + string(rune(i)) + ".com"
		cache.Put(domain, "group1", true)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Get("domain500.com")
	}
}

func BenchmarkLRUCache_Put(b *testing.B) {
	cache := NewLRUCache(1000, 0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		domain := "domain" + string(rune(i%1000)) + ".com"
		cache.Put(domain, "group1", true)
	}
}

func BenchmarkLRUCache_Concurrent(b *testing.B) {
	cache := NewLRUCache(1000, 0)

	// Pre-populate cache
	for i := 0; i < 1000; i++ {
		domain := "domain" + string(rune(i)) + ".com"
		cache.Put(domain, "group1", true)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			domain := "domain" + string(rune(i%1000)) + ".com"
			cache.Get(domain)
			i++
		}
	})
}
