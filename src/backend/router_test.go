package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
)

func newTestRouter() http.Handler {
	// Fresh collectors per router: the global ones in main are registered with
	// the default registry, which would panic on duplicate registration here.
	requestCount := prometheus.NewCounterVec(
		prometheus.CounterOpts{Name: "test_http_requests_total", Help: "test"},
		[]string{"method", "endpoint"},
	)
	requestLatency := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{Name: "test_http_request_duration_seconds", Help: "test"},
		[]string{"method", "endpoint"},
	)

	return NewRouter([]byte("openapi: 3.0.3"), requestCount, requestLatency)
}

// TestProbesAreWiredUp guards the Kubernetes probe endpoints: a handler that
// works but is not routed would fail readiness in the cluster only.
func TestProbesAreWiredUp(t *testing.T) {
	router := newTestRouter()

	for _, path := range []string{"/livez", "/readyz"} {
		t.Run(path, func(t *testing.T) {
			response := httptest.NewRecorder()

			router.ServeHTTP(response, httptest.NewRequest("GET", path, nil))

			if response.Code != http.StatusOK {
				t.Errorf("expected 200 for %s, got %d", path, response.Code)
			}
			if response.Body.String() != "ok" {
				t.Errorf("expected body %q for %s, got %q", "ok", path, response.Body.String())
			}
		})
	}
}

func TestOpenApiSpecIsServed(t *testing.T) {
	router := newTestRouter()

	for _, path := range []string{"/", "/abfallkalender-api"} {
		t.Run(path, func(t *testing.T) {
			response := httptest.NewRecorder()

			router.ServeHTTP(response, httptest.NewRequest("GET", path, nil))

			if response.Code != http.StatusOK {
				t.Errorf("expected 200 for %s, got %d", path, response.Code)
			}
			if response.Body.String() != "openapi: 3.0.3" {
				t.Errorf("expected the embedded spec for %s, got %q", path, response.Body.String())
			}
		})
	}
}

// The web component routes were removed along with the frontend.
func TestRemovedFrontendRoutesReturn404(t *testing.T) {
	router := newTestRouter()

	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest("GET", "/kalender.js", nil))

	if response.Code != http.StatusNotFound {
		t.Errorf("expected 404 for /kalender.js, got %d", response.Code)
	}
}
