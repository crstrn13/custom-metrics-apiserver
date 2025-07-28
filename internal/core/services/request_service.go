package services

import (
	"crstrn13/mock-backend/internal/core/models"
	"crstrn13/mock-backend/internal/ports"
)

// requestService implements the RequestService port
type requestService struct{}

// NewRequestService creates a new instance of RequestService
func NewRequestService() ports.RequestService {
	return &requestService{}
}

// GetRequestInfo implements the RequestService interface
func (s *requestService) GetRequestInfo(args map[string][]string, headers map[string]string, origin, url string) (*models.Request, error) {
	return models.NewRequest(args, headers, origin, url), nil
}
