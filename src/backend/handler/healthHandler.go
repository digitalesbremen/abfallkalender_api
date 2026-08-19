package handler

import (
	"net/http"
)

// Health answers as soon as the process is able to serve requests. It performs
// no network calls on purpose.
//
// The upstream service is deliberately not probed here. This application is a
// caching proxy: if web.c-trace.de is unavailable it can still serve cached
// responses, and tying readiness to the upstream would take every replica out
// of the service at once during an upstream outage -- turning a partial
// degradation into a full outage.
func Health(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=UTF-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}
