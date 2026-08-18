package client

import (
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

// newCountingServer returns a server that records how often it was hit and
// always answers with the given body.
func newCountingServer(body string) (*httptest.Server, *atomic.Int32) {
	var hits atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
		hits.Add(1)
		_, _ = rw.Write([]byte(body))
	}))

	return server, &hits
}

func doGet(t *testing.T, c *Client, url string) *http.Response {
	t.Helper()

	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		t.Fatalf("could not build request: %v", err)
	}

	response, err := c.sendRequest(request, false)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	return response
}

func TestSendRequestCachesRepeatedGet(t *testing.T) {
	server, hits := newCountingServer("payload")
	defer server.Close()

	c := NewClient(server.URL)

	first := doGet(t, c, server.URL+"/some/path")
	defer func() { _ = first.Body.Close() }()

	if got := first.Header.Get("X-Cache"); got != "MISS" {
		t.Errorf("first request: expected X-Cache MISS, got %q", got)
	}
	if got := c.GetLastCacheStatus(); got != "MISS" {
		t.Errorf("first request: expected last cache status MISS, got %q", got)
	}

	firstBody, _ := io.ReadAll(first.Body)
	if string(firstBody) != "payload" {
		t.Errorf("first request: expected body %q, got %q", "payload", firstBody)
	}

	second := doGet(t, c, server.URL+"/some/path")
	defer func() { _ = second.Body.Close() }()

	if got := second.Header.Get("X-Cache"); got != "HIT" {
		t.Errorf("second request: expected X-Cache HIT, got %q", got)
	}

	secondBody, _ := io.ReadAll(second.Body)
	if string(secondBody) != "payload" {
		t.Errorf("second request: expected cached body %q, got %q", "payload", secondBody)
	}

	if got := hits.Load(); got != 1 {
		t.Errorf("expected upstream to be called once, got %d calls", got)
	}
}

func TestSendRequestDoesNotCacheForeignHosts(t *testing.T) {
	server, hits := newCountingServer("payload")
	defer server.Close()

	// BaseHost deliberately differs from the server we call, so the response
	// must not be cached.
	c := NewClient("http://some-other-host.invalid")

	for i := 0; i < 2; i++ {
		response := doGet(t, c, server.URL+"/some/path")
		if got := response.Header.Get("X-Cache"); got != "" {
			t.Errorf("request %d: expected no X-Cache header, got %q", i+1, got)
		}
		_ = response.Body.Close()
	}

	if got := c.GetLastCacheStatus(); got != "" {
		t.Errorf("expected empty last cache status for uncacheable request, got %q", got)
	}

	if got := hits.Load(); got != 2 {
		t.Errorf("expected upstream to be called twice, got %d calls", got)
	}
}

func TestSendRequestCachesPerUrl(t *testing.T) {
	server, hits := newCountingServer("payload")
	defer server.Close()

	c := NewClient(server.URL)

	first := doGet(t, c, server.URL+"/first")
	_ = first.Body.Close()
	second := doGet(t, c, server.URL+"/second")
	_ = second.Body.Close()

	if got := second.Header.Get("X-Cache"); got != "MISS" {
		t.Errorf("expected a different URL to miss the cache, got %q", got)
	}

	if got := hits.Load(); got != 2 {
		t.Errorf("expected upstream to be called twice for two URLs, got %d calls", got)
	}
}
