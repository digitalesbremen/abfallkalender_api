package main

import (
	api "abfallkalender_api/src/backend"
	_ "embed"
	"github.com/gorilla/handlers"
	"github.com/prometheus/client_golang/prometheus"
	"log"
	"net/http"
	"os"
)

// //go:embed dist/kalender.js
var kalenderJS string

// //go:embed dist/kalender.js.map
var kalenderJSMap string

var (
	requestCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint"},
	)
	requestLatency = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request latency in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)
)

func main() {
	log.Println("Hello Bremer Abfallkalender API!!!")

	prometheus.MustRegister(requestCount)
	prometheus.MustRegister(requestLatency)

	router := api.NewRouter(kalenderJS, kalenderJSMap, requestCount, requestLatency)

	// TODO use os.lookup env
	port, portSet := os.LookupEnv("PORT")
	if !portSet {
		port = "8080"
	}

	log.Printf("Port is set to %s\n", port)

	// TODO use go routine / non blocking
	log.Fatal(http.ListenAndServe(":"+port,
		handlers.CompressHandler(
			handlers.CORS(
				handlers.AllowedOrigins([]string{"*"}))(router))))
}
