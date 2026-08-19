package api

import (
	"abfallkalender_api/src/backend/handler"
	"abfallkalender_api/src/backend/handler/middleware"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
)

func NewRouter(openApiSpec []byte, requestCount *prometheus.CounterVec, requestLatency *prometheus.HistogramVec) *mux.Router {
	router := mux.NewRouter().StrictSlash(true)
	router.NotFoundHandler = handler.Handle404()

	router.Use(middleware.MetricsMiddleware(requestCount, requestLatency))

	// TODO signal handler

	for _, route := range newRoutes(openApiSpec) {
		var httpHandler http.Handler

		httpHandler = route.HandlerFunc
		if !route.SkipLog {
			httpHandler = handler.Logger(httpHandler, route.Name)
		}

		router.
			Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			Handler(httpHandler)
	}

	return router
}
