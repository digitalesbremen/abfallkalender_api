package api

import (
	handler2 "abfallkalender_api/src/backend/handler"
	"net/http"

	"github.com/gorilla/mux"
)

func NewRouter(kalenderJS string, kalenderJSMap string) *mux.Router {
	router := mux.NewRouter().StrictSlash(true)
	router.NotFoundHandler = handler2.Handle404()

	for _, route := range routes {
		var handler http.Handler

		handler = route.HandlerFunc
		handler = handler2.Logger(handler, route.Name, kalenderJS, kalenderJSMap)

		router.
			Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			Handler(handler)
	}

	return router
}
