package api

import (
	"abfallkalender_api/src/backend/model"
	"encoding/json"
	"log"
	"net/http"
)

func Handle404() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "text/plain; charset=UTF-8")
		w.WriteHeader(http.StatusNotFound)

		log.Printf("%s\t%s\t(404 - not found)", r.Method, r.RequestURI)

		_ = json.
			NewEncoder(w).
			Encode(
				model.ProtocolError{
					Code:    http.StatusNotFound,
					Message: http.StatusText(http.StatusNotFound),
				})
	})
}
