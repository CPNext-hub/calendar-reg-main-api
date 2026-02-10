package usecase

import (
	"time"

	"github.com/CPNext-hub/calendar-reg-main-api/internal/domain/entity"
)

// HealthUsecase defines the interface for health-related business logic.
type HealthUsecase interface {
	GetStatus() entity.Health
}

type healthUsecase struct{}

// NewHealthUsecase creates a new HealthUsecase instance.
func NewHealthUsecase() HealthUsecase {
	return &healthUsecase{}
}

func (u *healthUsecase) GetStatus() entity.Health {
	return entity.Health{
		Status:    "ok",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
}
