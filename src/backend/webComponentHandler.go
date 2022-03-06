package api

import (
	"fmt"
	"net/http"
)

func GetWebComponent(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "text/javascript; charset=UTF-8")
	//w.Header().Set("Access-Control-Allow-Origin", "*")
	_, _ = fmt.Fprint(w, WebComponentJs)
}
