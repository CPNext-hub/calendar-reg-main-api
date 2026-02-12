package repository

import (
	"context"

	"github.com/CPNext-hub/calendar-reg-main-api/internal/domain/entity"
)

// CourseRepository defines the interface for course data persistence.
type CourseRepository interface {
	Create(ctx context.Context, course *entity.Course) error
	GetAll(ctx context.Context) ([]*entity.Course, error)
	GetByCode(ctx context.Context, code string) (*entity.Course, error)
	Delete(ctx context.Context, code string) error
}
