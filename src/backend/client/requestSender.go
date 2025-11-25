package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type errorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type cachedResponse struct {
	StatusCode int
	Header     http.Header
	Body       []byte
}

func (c *Client) sendRequest(originalRequest *http.Request, autoCloseBody bool) (*http.Response, error) {
	request := originalRequest
	request.Header.Set("Content-Type", "application/json; charset=utf-8")
	request.Header.Set("Accept", "application/json; charset=utf-8")

	// Decide if this request is cacheable: only GET/HEAD and only for BaseHost
	cacheable := false
	if c.Cache != nil && (request.Method == http.MethodGet || request.Method == http.MethodHead) {
		if u, err := url.Parse(request.URL.String()); err == nil {
			// Match by host to ensure only calls to web.c-trace.de are cached
			if u.Host == c.BaseHost {
				cacheable = true
			}
		}
	}

	cacheKey := request.Method + " " + request.URL.String()

	if cacheable {
		if v, found := c.Cache.Get(cacheKey); found {
			if cr, ok := v.(cachedResponse); ok {
				// Build a fresh http.Response from cached data
				resp := &http.Response{
					StatusCode: cr.StatusCode,
					Status:     fmt.Sprintf("%d %s", cr.StatusCode, http.StatusText(cr.StatusCode)),
					Header:     cr.Header.Clone(),
					Body:       io.NopCloser(bytes.NewReader(cr.Body)),
					// Minimal fields required by callers
				}
				resp.Header.Set("X-Cache", "HIT")
				if autoCloseBody {
					defer func(Body io.ReadCloser) { _ = Body.Close() }(resp.Body)
				}
				return resp, nil
			}
		}
	}

	res, err := c.HTTPClient.Do(request)
	if err != nil {
		return nil, err
	}

	if autoCloseBody {
		defer func(Body io.ReadCloser) {
			_ = Body.Close()
		}(res.Body)
	}

	if res.StatusCode < http.StatusOK || res.StatusCode >= http.StatusBadRequest {
		var errRes errorResponse
		if err = json.NewDecoder(res.Body).Decode(&errRes); err == nil {
			return nil, errors.New(errRes.Message)
		}

		return nil, fmt.Errorf("unknown error, status code: %d", res.StatusCode)
	}

	// On success: possibly store in cache
	if cacheable {
		// Read body fully to cache it, but we also need to return a readable body to the caller.
		// So we copy it into bytes and create a new ReadCloser for the response we return.
		bodyBytes, readErr := io.ReadAll(res.Body)
		if !autoCloseBody {
			// Ensure original body is closed regardless
			defer func(Body io.ReadCloser) { _ = Body.Close() }(res.Body)
		} else {
			_ = res.Body.Close()
		}
		if readErr != nil {
			return nil, readErr
		}

		cr := cachedResponse{
			StatusCode: res.StatusCode,
			Header:     res.Header.Clone(),
			Body:       bodyBytes,
		}
		// Store with default expiration (24h configured in NewClient)
		c.Cache.SetDefault(cacheKey, cr)

		// Return a new response object with the cached body
		resp := &http.Response{
			StatusCode: cr.StatusCode,
			Status:     fmt.Sprintf("%d %s", cr.StatusCode, http.StatusText(cr.StatusCode)),
			Header:     cr.Header.Clone(),
			Body:       io.NopCloser(bytes.NewReader(cr.Body)),
		}
		resp.Header.Set("X-Cache", "MISS")
		if autoCloseBody {
			defer func(Body io.ReadCloser) { _ = Body.Close() }(resp.Body)
		}
		return resp, nil
	}

	// Non-cacheable success: return as-is
	return res, nil
}
