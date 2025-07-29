package main

import (
	"flag"
	"log"
	"net/http"

	"crstrn13/mock-backend/internal/adapters/http/handlers"
	"crstrn13/mock-backend/internal/adapters/metrics"
	"crstrn13/mock-backend/internal/core/services"
)

func main() {
	// Define flags
	certFile := flag.String("cert", "/certs/tls.crt", "Path to TLS certificate file")
	keyFile := flag.String("key", "/certs/tls.key", "Path to TLS private key file")
	metricsPort := flag.String("metrics-port", ":8443", "HTTPS port for metrics")
	httpPort := flag.String("http-port", ":8080", "HTTP port for regular traffic")
	flag.Parse()

	// Initialize services
	requestService := services.NewRequestService()
	metricsProvider := metrics.NewProvider()

	// Initialize handlers
	requestHandler := handlers.NewRequestHandler(requestService, metricsProvider)

	// Create HTTP mux for regular traffic
	httpMux := http.NewServeMux()
	httpMux.HandleFunc("/get", requestHandler.GetHandler)

	// Create HTTPS mux for metrics
	metricsMux := http.NewServeMux()
	metricsMux.Handle("/apis/custom.metrics.k8s.io/v1beta1", metricsProvider)

	// Create HTTP server for regular traffic
	httpServer := &http.Server{
		Addr:    *httpPort,
		Handler: httpMux,
	}

	// Create HTTPS server for metrics
	metricsServer := &http.Server{
		Addr:    *metricsPort,
		Handler: metricsMux,
	}

	// Start HTTP server in a goroutine
	go func() {
		log.Printf("HTTP Server starting on %s", *httpPort)
		if err := httpServer.ListenAndServe(); err != nil {
			log.Fatalf("HTTP Server failed to start: %v", err)
		}
	}()

	// Start HTTPS server for metrics
	log.Printf("HTTPS Metrics Server starting on %s", *metricsPort)
	if err := metricsServer.ListenAndServeTLS(*certFile, *keyFile); err != nil {
		log.Fatalf("HTTPS Metrics Server failed to start: %v", err)
	}
}
