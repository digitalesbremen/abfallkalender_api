package api

import (
	"net/http"

	"github.com/gorilla/mux"
)

// WebComponentJs TODO global variable. Maybe pass parameter to handler?
var WebComponentJs = ""

func NewRouter(webComponentJs string) *mux.Router {
	router := mux.NewRouter().StrictSlash(true)
	router.NotFoundHandler = Handle404()

	WebComponentJs = webComponentJs

	for _, route := range routes {
		var handler http.Handler

		handler = route.HandlerFunc
		handler = Logger(handler, route.Name)

		router.
			Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			Handler(handler)
	}

	return router
}
