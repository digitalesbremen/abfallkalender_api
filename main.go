package main

import (
	api "abfallkalender_api/src/backend"
	"context"
	_ "embed"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/handlers"
	"github.com/prometheus/client_golang/prometheus"
)

// shutdownTimeout bounds how long in-flight requests may finish after SIGTERM.
// Keep this below the Kubernetes terminationGracePeriodSeconds (30s by
// default), otherwise the pod is SIGKILLed mid-shutdown.
const shutdownTimeout = 15 * time.Second

// openApiSpec is embedded at build time. The Docker build substitutes
// ${VERSION} in the file before compiling, so the served spec carries the
// release version.
//
//go:embed open-api-3.yaml
var openApiSpec []byte

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

	router := api.NewRouter(openApiSpec, requestCount, requestLatency)

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

	server := &http.Server{
		Addr:    ":" + port,
		Handler: wrapped,
		// Bound the time spent reading request headers so a stalled client
		// cannot hold a connection open indefinitely.
		ReadHeaderTimeout: 10 * time.Second,
	}

	shutdownDone := make(chan struct{})
	go func() {
		defer close(shutdownDone)

		signals := make(chan os.Signal, 1)
		signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
		receivedSignal := <-signals

		// Kubernetes sends SIGTERM before removing a pod. Without this the
		// server would keep serving until SIGKILL, cutting off in-flight
		// requests on every rolling update.
		log.Printf("Received %s, shutting down within %s", receivedSignal, shutdownTimeout)

		ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			log.Printf("Graceful shutdown failed, closing connections: %v", err)
			_ = server.Close()
		}
	}()

	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("Server stopped unexpectedly: %v", err)
	}

	// ListenAndServe returns as soon as Shutdown is called, so wait for the
	// shutdown to actually complete before leaving main.
	<-shutdownDone
	log.Println("Shutdown complete")
}
