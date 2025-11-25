package dnsforward

import (
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPrefetchManager_Record(t *testing.T) {
	s := &Server{
		logger: slog.Default(),
	}
	pm := NewPrefetchManager(s)
	pm.threshold = 2 // Set low threshold for testing

	domain := "example.com."
	ttl := uint32(60)

	// First hit
	pm.Record(domain, ttl)
	pm.mu.RLock()
	assert.Equal(t, 1, pm.hits[domain])
	_, ok := pm.domains[domain]
	assert.False(t, ok)
	pm.mu.RUnlock()

	// Second hit - should trigger tracking
	pm.Record(domain, ttl)
	pm.mu.RLock()
	assert.Equal(t, 2, pm.hits[domain])
	expiry, ok := pm.domains[domain]
	assert.True(t, ok)
	assert.WithinDuration(t, time.Now().Add(time.Duration(ttl)*time.Second), expiry, 1*time.Second)
	pm.mu.RUnlock()
}

func TestPrefetchManager_CheckAndRefresh(t *testing.T) {
	s := &Server{
		logger: slog.Default(),
	}
	pm := NewPrefetchManager(s)

	domain := "expired.com."
	// Set expiry to past
	pm.mu.Lock()
	pm.domains[domain] = time.Now().Add(-1 * time.Second)
	pm.mu.Unlock()

	// We can't easily test the actual refresh network call in unit test without mocking,
	// but we can check if it removes the domain from the map (which happens before refresh)

	pm.checkAndRefresh()

	pm.mu.RLock()
	_, ok := pm.domains[domain]
	assert.False(t, ok, "Expired domain should be removed from map (and refreshed)")
	pm.mu.RUnlock()
}
