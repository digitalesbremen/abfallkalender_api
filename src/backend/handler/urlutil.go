package handler

import (
	"net/http"
	"strings"
)

// requestScheme determines the effective scheme (http/https) considering
// standard reverse proxy headers. Fallback is based on the presence of TLS.
func requestScheme(r *http.Request) string {
	// RFC 7239 Forwarded: proto=..., host=...
	if fwd := r.Header.Get("Forwarded"); fwd != "" {
		parts := strings.Split(strings.ToLower(fwd), ";")
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if strings.HasPrefix(p, "proto=") {
				return strings.Trim(p[len("proto="):], "\"")
			}
		}
	}
	// De-facto headers used by many proxies
	if xfproto := r.Header.Get("X-Forwarded-Proto"); xfproto != "" {
		// Only the first value if multiple are present
		return strings.Split(xfproto, ",")[0]
	}
	if r.TLS != nil {
		return "https"
	}
	// Default to https to generate secure self links when no proxy hints are present
	// (e.g. in tests or when running behind TLS-terminating proxies without headers).
	return "https"
}

// requestHost determines the effective host considering reverse proxy headers.
func requestHost(r *http.Request) string {
	if fwd := r.Header.Get("Forwarded"); fwd != "" {
		parts := strings.Split(fwd, ";")
		for _, p := range parts {
			low := strings.TrimSpace(strings.ToLower(p))
			if strings.HasPrefix(low, "host=") {
				// Use original substring to keep case, but trim quotes
				return strings.Trim(p[len("host="):], "\"")
			}
		}
	}
	if xfh := r.Header.Get("X-Forwarded-Host"); xfh != "" {
		return strings.Split(xfh, ",")[0]
	}
	return r.Host
}

// requestPrefix returns an optional path prefix supplied by a reverse proxy.
func requestPrefix(r *http.Request) string {
	if xfp := r.Header.Get("X-Forwarded-Prefix"); xfp != "" {
		// First value, trimmed of trailing slashes
		return strings.TrimRight(strings.Split(xfp, ",")[0], "/")
	}
	return ""
}

// absoluteURL builds an absolute URL combining scheme, host, an optional
// forwarded prefix, and the supplied absolute path (starting with '/').
func absoluteURL(r *http.Request, path string) string {
	scheme := requestScheme(r)
	host := requestHost(r)
	prefix := requestPrefix(r)
	return scheme + "://" + host + prefix + path
}
