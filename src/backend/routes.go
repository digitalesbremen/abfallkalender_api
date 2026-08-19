package api

import (
	"abfallkalender_api/src/backend/handler"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
)

type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
	// SkipLog suppresses the per-request log line. Used for the Kubernetes
	// probes, which would otherwise dominate the log at roughly one line every
	// few seconds per replica.
	SkipLog bool
}

type Routes []Route

// newRoutes builds the route table. All routes share a single controller, and
// therefore a single upstream HTTP client with one cache — previously every
// route constructed its own, which meant four independent caches querying the
// same upstream.
func newRoutes(openApiSpec []byte) Routes {
	controller := handler.NewController()
	openApiDocumentation := handler.OpenApiDocumentation(openApiSpec)

	return Routes{
		Route{
			Name:        "Metrics",
			Method:      "GET",
			Pattern:     "/metrics",
			HandlerFunc: promhttp.Handler().ServeHTTP,
		},
		Route{
			Name:        "Liveness probe",
			Method:      "GET",
			Pattern:     "/livez",
			HandlerFunc: handler.Health,
			SkipLog:     true,
		},
		Route{
			Name:        "Readiness probe",
			Method:      "GET",
			Pattern:     "/readyz",
			HandlerFunc: handler.Health,
			SkipLog:     true,
		},
		Route{
			Name:        "Open Api documentation (yaml)",
			Method:      "GET",
			Pattern:     "/",
			HandlerFunc: openApiDocumentation,
		},
		Route{
			Name:        "Open Api documentation (yaml)",
			Method:      "GET",
			Pattern:     "/abfallkalender-api",
			HandlerFunc: openApiDocumentation,
		},
		Route{
			Name:        "Streets",
			Method:      "GET",
			Pattern:     "/abfallkalender-api/streets",
			HandlerFunc: controller.GetStreets,
		},
		Route{
			Name:        "Street",
			Method:      "GET",
			Pattern:     "/abfallkalender-api/street/{street}",
			HandlerFunc: controller.GetStreet,
		},
		Route{
			Name:        "ICS",
			Method:      "GET",
			Pattern:     "/abfallkalender-api/street/{street}/number/{number}",
			HandlerFunc: controller.GetCalendar,
		},
		Route{
			Name:        "Next",
			Method:      "GET",
			Pattern:     "/abfallkalender-api/street/{street}/number/{number}/next",
			HandlerFunc: controller.GetNext,
		},
	}
}
