package middleware

import (
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"net/http"
)

func MetricsMiddleware(requestCount *prometheus.CounterVec, requestLatency *prometheus.HistogramVec) mux.MiddlewareFunc {
	return func(inner http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Route-Name aus mux extrahieren
			route := mux.CurrentRoute(r)
			routeName := "unknown"
			if route != nil {
				if name := route.GetName(); name != "" {
					routeName = name
				}
			}

			timer := prometheus.NewTimer(requestLatency.WithLabelValues(r.Method, routeName))
			defer timer.ObserveDuration()

			// Label by route name, never by r.URL.Path: the path contains street
			// names and house numbers, so using it would create one time series
			// per queried address and blow up Prometheus cardinality.
			requestCount.WithLabelValues(r.Method, routeName).Inc()

			inner.ServeHTTP(w, r)
		})
	}
}
