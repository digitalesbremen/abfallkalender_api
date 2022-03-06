package api

import (
	"net/http"
)

type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

type Routes []Route

var routes = Routes{
	Route{
		Name:        "Open Api documentation (yaml)",
		Method:      "GET",
		Pattern:     "/",
		HandlerFunc: OpenApiDocumentation,
	},
	Route{
		Name:        "Open Api documentation (yaml)",
		Method:      "GET",
		Pattern:     "/api",
		HandlerFunc: OpenApiDocumentation,
	},
	Route{
		Name:        "Streets",
		Method:      "GET",
		Pattern:     "/api/streets",
		HandlerFunc: GetStreets,
	},
	Route{
		Name:        "Kalender web component",
		Method:      "GET",
		Pattern:     "/kalender.js",
		HandlerFunc: GetWebComponent,
	},
	Route{
		Name:        "Kalender web component",
		Method:      "GET",
		Pattern:     "/kalender.js.map",
		HandlerFunc: GetWebComponentMap,
	},
}
