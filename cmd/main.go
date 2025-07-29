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
	port := flag.String("port", ":8443", "HTTPS port to listen on")
	flag.Parse()

	// Initialize services
	requestService := services.NewRequestService()
	metricsProvider := metrics.NewProvider()

	// Initialize handlers
	requestHandler := handlers.NewRequestHandler(requestService, metricsProvider)

	// Register routes
	mux := http.NewServeMux()
	mux.HandleFunc("/get", requestHandler.GetHandler)
	mux.Handle("/apis/custom.metrics.k8s.io/v1beta1", metricsProvider)

	// Create HTTPS server
	server := &http.Server{
		Addr:    *port,
		Handler: mux,
	}

	// Start HTTPS server
	log.Printf("Server starting on %s (HTTPS)", *port)
	if err := server.ListenAndServeTLS(*certFile, *keyFile); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
