package dto

import "time"

// MongoTestDocument is the DTO for test CRUD operations.
type MongoTestDocument struct {
	ID        string `json:"id,omitempty"`
	Message   string `json:"message"`
	CreatedAt string `json:"created_at"`
}

// MongoPingResponse is the DTO for the MongoDB ping endpoint.
type MongoPingResponse struct {
	Status    string `json:"status"`
	Latency   string `json:"latency"`
	Timestamp string `json:"timestamp"`
}

// NewMongoPingResponse creates a ping response with timing info.
func NewMongoPingResponse(status string, latency time.Duration) MongoPingResponse {
	return MongoPingResponse{
		Status:    status,
		Latency:   latency.String(),
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
}

// MongoTestResult is the DTO for the full MongoDB test (write→read→delete).
type MongoTestResult struct {
	Ping   string `json:"ping"`
	Insert string `json:"insert"`
	Find   string `json:"find"`
	Delete string `json:"delete"`
}
