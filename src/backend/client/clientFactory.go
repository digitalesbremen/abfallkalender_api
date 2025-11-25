package client

import (
	"net/http"
	"net/url"
	"sync"
	"time"

	cache "github.com/patrickmn/go-cache"
)

type Client struct {
	BaseURL    string
	BaseHost   string
	HTTPClient *http.Client
	Cache      *cache.Cache
	// LastCacheStatus reflects the X-Cache value (HIT/MISS) of the last
	// successful outbound request performed via this client instance.
	// It is guarded by cacheStatusMu to be safe under concurrent access.
	lastCacheStatus string
	cacheStatusMu   sync.RWMutex
}

func NewClient(baseURL string) *Client {
	// Parse host for later cache scoping
	u, _ := url.Parse(baseURL)

	client := Client{
		BaseURL: baseURL,
		BaseHost: func() string {
			if u != nil {
				return u.Host
			}
			return ""
		}(),
		HTTPClient: &http.Client{
			Timeout: time.Minute,
			// do not follow redirects
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
		// Default TTL 24h, cleanup every 30m
		Cache: cache.New(24*time.Hour, 30*time.Minute),
	}
	return &client
}

// setLastCacheStatus stores the provided status in a concurrency-safe manner.
// Expected values are "HIT", "MISS" or empty string for not-applicable.
func (c *Client) setLastCacheStatus(status string) {
	c.cacheStatusMu.Lock()
	c.lastCacheStatus = status
	c.cacheStatusMu.Unlock()
}

// GetLastCacheStatus returns the most recently observed cache status.
// Returns empty string if not applicable/unknown.
func (c *Client) GetLastCacheStatus() string {
	c.cacheStatusMu.RLock()
	defer c.cacheStatusMu.RUnlock()
	return c.lastCacheStatus
}
