package ports

import "crstrn13/mock-backend/internal/core/models"

// RequestService defines the port for handling HTTP request operations
type RequestService interface {
	// GetRequestInfo processes an incoming request and returns its information
	GetRequestInfo(args map[string][]string, headers map[string]string, origin, url string) (*models.Request, error)
}
