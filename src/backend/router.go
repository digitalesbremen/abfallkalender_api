package api

import (
	"abfallkalender_api/src/backend/handler"
	"net/http"

	"github.com/gorilla/mux"
)

func NewRouter(kalenderJS string, kalenderJSMap string) *mux.Router {
	router := mux.NewRouter().StrictSlash(true)
	router.NotFoundHandler = handler.Handle404()

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
