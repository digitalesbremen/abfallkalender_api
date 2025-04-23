package api

import (
	"abfallkalender_api/src/backend/handler"
	"abfallkalender_api/src/backend/handler/middleware"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
)

func NewRouter(kalenderJS string, kalenderJSMap string, requestCount *prometheus.CounterVec, requestLatency *prometheus.HistogramVec) *mux.Router {
	router := mux.NewRouter().StrictSlash(true)
	router.NotFoundHandler = handler.Handle404()

	router.Use(middleware.MetricsMiddleware(requestCount, requestLatency))

	// TODO signal handler

	for _, route := range routes {
		var httpHandler http.Handler

		httpHandler = route.HandlerFunc
		httpHandler = handler.Logger(httpHandler, route.Name, kalenderJS, kalenderJSMap)

		router.
			Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			Handler(httpHandler)
	}

	return router
}
