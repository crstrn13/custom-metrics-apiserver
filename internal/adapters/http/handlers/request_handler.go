package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"crstrn13/mock-backend/internal/adapters/metrics"
	"crstrn13/mock-backend/internal/ports"
)

// RequestHandler handles HTTP requests for the /get endpoint
type RequestHandler struct {
	requestService ports.RequestService
	metricsHandler *metrics.Provider
}

// NewRequestHandler creates a new RequestHandler instance
func NewRequestHandler(requestService ports.RequestService, metricsHandler *metrics.Provider) *RequestHandler {
	return &RequestHandler{
		requestService: requestService,
		metricsHandler: metricsHandler,
	}
}

// GetHandler handles GET requests
func (h *RequestHandler) GetHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Increment request counter
	h.metricsHandler.IncrementRequestCount()

	// Extract headers
	headers := make(map[string]string)
	for name, values := range r.Header {
		headers[strings.ToLower(name)] = values[0]
	}

	// Get request info from service
	requestInfo, err := h.requestService.GetRequestInfo(
		r.URL.Query(),
		headers,
		r.RemoteAddr,
		r.URL.String(),
	)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Encode and send response
	json.NewEncoder(w).Encode(requestInfo)
}
