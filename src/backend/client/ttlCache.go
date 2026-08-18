package client

import (
	"sync"
	"time"
)

// ttlCache is a minimal concurrency-safe cache with a fixed time-to-live per
// entry. It replaces github.com/patrickmn/go-cache, whose last release predates
// Go modules and which is no longer maintained.
//
// Only the two operations this project actually needs are provided: Get and
// SetDefault. Expired entries are dropped lazily on read and, in addition,
// purged periodically so the map cannot grow without bound.
type ttlCache[V any] struct {
	mu         sync.RWMutex
	items      map[string]ttlItem[V]
	defaultTTL time.Duration
}

type ttlItem[V any] struct {
	value     V
	expiresAt time.Time
}

// newTTLCache creates a cache whose entries expire after defaultTTL. When
// cleanupInterval is greater than zero a janitor goroutine purges expired
// entries at that interval. The janitor deliberately has no stop channel: the
// cache lives for the lifetime of the process, which is what time.Tick is
// documented for.
func newTTLCache[V any](defaultTTL time.Duration, cleanupInterval time.Duration) *ttlCache[V] {
	c := &ttlCache[V]{
		items:      make(map[string]ttlItem[V]),
		defaultTTL: defaultTTL,
	}

	if cleanupInterval > 0 {
		go func() {
			for range time.Tick(cleanupInterval) {
				c.deleteExpired()
			}
		}()
	}

	return c
}

// Get returns the value stored under key and reports whether it was present and
// has not expired yet.
func (c *ttlCache[V]) Get(key string) (V, bool) {
	c.mu.RLock()
	item, found := c.items[key]
	c.mu.RUnlock()

	if !found || time.Now().After(item.expiresAt) {
		var zero V
		return zero, false
	}

	return item.value, true
}

// SetDefault stores value under key using the cache's default TTL.
func (c *ttlCache[V]) SetDefault(key string, value V) {
	c.mu.Lock()
	c.items[key] = ttlItem[V]{value: value, expiresAt: time.Now().Add(c.defaultTTL)}
	c.mu.Unlock()
}

func (c *ttlCache[V]) deleteExpired() {
	now := time.Now()

	c.mu.Lock()
	for key, item := range c.items {
		if now.After(item.expiresAt) {
			delete(c.items, key)
		}
	}
	c.mu.Unlock()
}
