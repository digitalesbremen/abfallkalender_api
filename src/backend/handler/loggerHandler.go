package handler

import (
	"log"
	"net/http"
	"time"
)

// KalenderJS TODO move parameter to handler
var KalenderJS = ""

// KalenderJSMap TODO move parameter to handler
var KalenderJSMap = ""

func Logger(inner http.Handler, name string, kalenderJS string, kalenderJSMap string) http.Handler {
	KalenderJS = kalenderJS
	KalenderJSMap = kalenderJSMap

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		log.Printf("%s\t%s\t%s\t(in progress)", r.Method, r.RequestURI, name)

		inner.ServeHTTP(w, r)

		log.Printf("%s\t%s\t%s\ttakes %s", r.Method, r.RequestURI, name, time.Since(start))
	})
}
