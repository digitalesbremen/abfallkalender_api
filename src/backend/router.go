package api

import (
	handler2 "abfallkalender_api/src/backend/handler"
	"net/http"

	"github.com/gorilla/mux"
)

// KalenderJS TODO global variable. Maybe pass parameter to handler?
var KalenderJS = ""
var KalenderJSMap = ""

func NewRouter(kalenderJS string, kalenderJSMap string) *mux.Router {
	router := mux.NewRouter().StrictSlash(true)
	router.NotFoundHandler = handler2.Handle404()

	KalenderJS = kalenderJS
	KalenderJSMap = kalenderJSMap

	for _, route := range routes {
		var handler http.Handler

		handler = route.HandlerFunc
		handler = handler2.Logger(handler, route.Name)

		router.
			Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			Handler(handler)
	}

	return router
}
