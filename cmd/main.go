package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"log"
	"math/big"
	"net/http"
	"time"

	"crstrn13/mock-backend/internal/adapters/http/handlers"
	"crstrn13/mock-backend/internal/adapters/metrics"
	"crstrn13/mock-backend/internal/core/services"
)

func generateTLSConfig() (*tls.Config, error) {
	// Generate private key
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}

	// Create certificate template
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"Mock Backend"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(time.Hour * 24 * 365),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	// Create certificate
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		return nil, err
	}

	// Create PEM blocks
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})

	// Parse certificate
	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return nil, err
	}

	return &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
	}, nil
}

func main() {
	// Initialize services
	requestService := services.NewRequestService()
	metricsProvider := metrics.NewProvider()

	// Initialize handlers
	requestHandler := handlers.NewRequestHandler(requestService, metricsProvider)

	// Register routes
	mux := http.NewServeMux()
	mux.HandleFunc("/get", requestHandler.GetHandler)
	mux.Handle("/apis/custom.metrics.k8s.io/v1beta1", metricsProvider)

	// Generate TLS config
	tlsConfig, err := generateTLSConfig()
	if err != nil {
		log.Fatalf("Failed to generate TLS config: %v", err)
	}

	// Create HTTPS server
	server := &http.Server{
		Addr:      ":8080",
		Handler:   mux,
		TLSConfig: tlsConfig,
	}

	// Start HTTPS server
	log.Printf("Server starting on :8080 (HTTPS)")
	if err := server.ListenAndServeTLS("", ""); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
