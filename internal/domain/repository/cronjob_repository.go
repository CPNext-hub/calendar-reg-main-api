package repository

import (
	"context"

	"github.com/CPNext-hub/calendar-reg-main-api/internal/domain/entity"
)

// CronJobRepository defines the interface for cron job data persistence.
type CronJobRepository interface {
	Create(ctx context.Context, job *entity.CronJob) error
	GetAll(ctx context.Context) ([]*entity.CronJob, error)
	GetByID(ctx context.Context, id string) (*entity.CronJob, error)
	Update(ctx context.Context, job *entity.CronJob) error
	Delete(ctx context.Context, id string) error
	GetEnabled(ctx context.Context) ([]*entity.CronJob, error)
}
