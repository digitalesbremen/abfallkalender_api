package main

import (
	api "abfallkalender_api/src/backend"
	"context"
	_ "embed"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/handlers"
	"github.com/prometheus/client_golang/prometheus"
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

	port, portSet := os.LookupEnv("PORT")
	if !portSet {
		port = "8080"
	}

	log.Printf("Port is set to %s\n", port)

	// Respect reverse proxy headers for scheme/host (X-Forwarded-*)
	var wrapped http.Handler = handlers.ProxyHeaders(router)
	// Enable compression and permissive CORS (as before)
	wrapped = handlers.CompressHandler(wrapped)
	wrapped = handlers.CORS(handlers.AllowedOrigins([]string{"*"}))(wrapped)

	// --- Graceful shutdown setup ---
	// We run the HTTP server in a goroutine and listen for OS signals.
	// On SIGINT/SIGTERM we initiate a graceful shutdown with a timeout,
	// giving in-flight requests time to complete.

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: wrapped,
	}

	// Channel for server errors
	serverErr := make(chan error, 1)

	go func() {
		log.Printf("HTTP server listening on %s\n", srv.Addr)
		// http.ErrServerClosed is expected on graceful shutdown
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErr <- err
		}
		close(serverErr)
	}()

	// Signal handling
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-stop:
		log.Printf("Received signal %s. Shutting down gracefully...\n", sig.String())
		// Allow up to 10 seconds for graceful shutdown
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			log.Printf("Graceful shutdown failed: %v\n", err)
		} else {
			log.Println("Server shut down gracefully")
		}
	case err := <-serverErr:
		// Server crashed while running
		if err != nil {
			log.Fatalf("HTTP server error: %v\n", err)
		}
	}
}
