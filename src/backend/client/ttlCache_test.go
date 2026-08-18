package client

import (
	"sync"
	"testing"
	"time"
)

func TestTTLCacheReturnsStoredValue(t *testing.T) {
	c := newTTLCache[string](time.Minute, 0)

	c.SetDefault("key", "value")

	got, found := c.Get("key")
	if !found {
		t.Fatal("expected key to be found")
	}
	if got != "value" {
		t.Errorf("expected \"value\", got %q", got)
	}
}

func TestTTLCacheReportsMissForUnknownKey(t *testing.T) {
	c := newTTLCache[string](time.Minute, 0)

	got, found := c.Get("missing")
	if found {
		t.Fatal("expected missing key to report a miss")
	}
	if got != "" {
		t.Errorf("expected zero value on miss, got %q", got)
	}
}

func TestTTLCacheExpiresEntries(t *testing.T) {
	c := newTTLCache[string](time.Millisecond, 0)

	c.SetDefault("key", "value")
	time.Sleep(10 * time.Millisecond)

	if _, found := c.Get("key"); found {
		t.Error("expected entry to be expired")
	}
}

func TestTTLCacheDeleteExpiredRemovesOnlyExpiredEntries(t *testing.T) {
	c := newTTLCache[string](time.Minute, 0)

	c.SetDefault("fresh", "value")
	// Insert an already-expired entry directly; SetDefault always applies the
	// default TTL, so there is no other way to construct one.
	c.mu.Lock()
	c.items["stale"] = ttlItem[string]{value: "value", expiresAt: time.Now().Add(-time.Minute)}
	c.mu.Unlock()

	c.deleteExpired()

	c.mu.RLock()
	_, staleExists := c.items["stale"]
	_, freshExists := c.items["fresh"]
	c.mu.RUnlock()

	if staleExists {
		t.Error("expected expired entry to be removed")
	}
	if !freshExists {
		t.Error("expected unexpired entry to be kept")
	}
}

// TestTTLCacheConcurrentAccess is meaningful under -race, which CI runs.
func TestTTLCacheConcurrentAccess(t *testing.T) {
	c := newTTLCache[int](time.Minute, 0)

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			c.SetDefault("key", i)
		}()
		go func() {
			defer wg.Done()
			c.Get("key")
		}()
	}
	wg.Wait()
}
