package handler

import (
	"fmt"
	"net/http"
)

func GetWebComponent(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "text/javascript; charset=UTF-8")
	_, _ = fmt.Fprint(w, KalenderJS)
}

func GetWebComponentMap(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "text/javascript; charset=UTF-8")
	_, _ = fmt.Fprint(w, KalenderJSMap)
}
