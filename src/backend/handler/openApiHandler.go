package handler

import (
	"net/http"
)

// OpenApiDocumentation serves the OpenAPI specification passed in at startup.
// The spec is embedded into the binary by the main package, so serving it needs
// neither disk access per request nor a specific working directory.
func OpenApiDocumentation(spec []byte) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/yaml; charset=UTF-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(spec)
	}
}
