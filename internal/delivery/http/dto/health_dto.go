package dto

import "github.com/CPNext-hub/calendar-reg-main-api/internal/domain/entity"

// HealthResponse is the DTO for the health/status endpoint.
type HealthResponse struct {
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
}

// ToHealthResponse converts a domain Health entity to a HealthResponse DTO.
func ToHealthResponse(h entity.Health) HealthResponse {
	return HealthResponse{
		Status:    h.Status,
		Timestamp: h.Timestamp,
	}
}
