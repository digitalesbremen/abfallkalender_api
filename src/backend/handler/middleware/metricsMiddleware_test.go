package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

func newInstrumentedRouter() (*mux.Router, *prometheus.Registry) {
	registry := prometheus.NewRegistry()

	requestCount := prometheus.NewCounterVec(
		prometheus.CounterOpts{Name: "http_requests_total", Help: "test"},
		[]string{"method", "endpoint"},
	)
	requestLatency := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{Name: "http_request_duration_seconds", Help: "test"},
		[]string{"method", "endpoint"},
	)
	registry.MustRegister(requestCount, requestLatency)

	router := mux.NewRouter()
	router.Use(MetricsMiddleware(requestCount, requestLatency))
	router.
		Methods("GET").
		Path("/abfallkalender-api/street/{street}").
		Name("Street").
		HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })

	return router, registry
}

func metricFamily(t *testing.T, registry *prometheus.Registry, name string) *dto.MetricFamily {
	t.Helper()

	families, err := registry.Gather()
	if err != nil {
		t.Fatalf("could not gather metrics: %v", err)
	}

	for _, family := range families {
		if family.GetName() == name {
			return family
		}
	}

	t.Fatalf("metric family %q not found", name)
	return nil
}

// TestRequestCountIsLabelledByRouteNotPath is the regression guard for a
// cardinality bug: the counter used to be labelled with r.URL.Path, so every
// street and house number produced its own time series.
func TestRequestCountIsLabelledByRouteNotPath(t *testing.T) {
	router, registry := newInstrumentedRouter()

	for _, street := range []string{"Aachener%20Stra%C3%9Fe", "Bahnhofstrasse", "Zuericher%20Strasse"} {
		router.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/abfallkalender-api/street/"+street, nil))
	}

	family := metricFamily(t, registry, "http_requests_total")

	if len(family.GetMetric()) != 1 {
		t.Fatalf("expected 3 requests to 3 different streets to share 1 time series, got %d", len(family.GetMetric()))
	}

	metric := family.GetMetric()[0]

	for _, label := range metric.GetLabel() {
		if label.GetName() != "endpoint" {
			continue
		}
		if label.GetValue() != "Street" {
			t.Errorf("expected endpoint label %q, got %q", "Street", label.GetValue())
		}
	}

	if got := metric.GetCounter().GetValue(); got != 3 {
		t.Errorf("expected the shared series to have counted 3 requests, got %v", got)
	}
}

// The latency histogram was already labelled correctly; keep it that way.
func TestRequestLatencyIsLabelledByRoute(t *testing.T) {
	router, registry := newInstrumentedRouter()

	router.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/abfallkalender-api/street/Aachener", nil))
	router.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/abfallkalender-api/street/Bahnhofstrasse", nil))

	family := metricFamily(t, registry, "http_request_duration_seconds")

	if len(family.GetMetric()) != 1 {
		t.Fatalf("expected 1 histogram series across both streets, got %d", len(family.GetMetric()))
	}

	if got := family.GetMetric()[0].GetHistogram().GetSampleCount(); got != 2 {
		t.Errorf("expected 2 observations, got %d", got)
	}
}
