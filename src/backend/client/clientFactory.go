package client

import (
	"net/http"
	"net/url"
	"time"

	cache "github.com/patrickmn/go-cache"
)

type Client struct {
	BaseURL    string
	BaseHost   string
	HTTPClient *http.Client
	Cache      *cache.Cache
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
