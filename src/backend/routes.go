package api

import (
	"abfallkalender_api/src/backend/handler"
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
		HandlerFunc: handler.OpenApiDocumentation,
	},
	Route{
		Name:        "Open Api documentation (yaml)",
		Method:      "GET",
		Pattern:     "/abfallkalender-api",
		HandlerFunc: handler.OpenApiDocumentation,
	},
	Route{
		Name:        "Streets",
		Method:      "GET",
		Pattern:     "/abfallkalender-api/streets",
		HandlerFunc: handler.NewController().GetStreets,
	},
	Route{
		Name:        "Street",
		Method:      "GET",
		Pattern:     "/abfallkalender-api/street/{street}",
		HandlerFunc: handler.NewController().GetStreet,
	},
	Route{
		Name:        "ICS",
		Method:      "GET",
		Pattern:     "/abfallkalender-api/street/{street}/number/{number}",
		HandlerFunc: handler.NewController().GetCalendar,
	},
	Route{
		Name:        "Next",
		Method:      "GET",
		Pattern:     "/abfallkalender-api/street/{street}/number/{number}/next",
		HandlerFunc: handler.NewController().GetNext,
	},
	Route{
		Name:        "Kalender web component",
		Method:      "GET",
		Pattern:     "/kalender.js",
		HandlerFunc: handler.GetWebComponent,
	},
	Route{
		Name:        "Kalender web component",
		Method:      "GET",
		Pattern:     "/kalender.js.map",
		HandlerFunc: handler.GetWebComponentMap,
	},
}
