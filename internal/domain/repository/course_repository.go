package repository

import (
	"context"

	"github.com/CPNext-hub/calendar-reg-main-api/internal/domain/entity"
)

// CourseRepository defines the interface for course data persistence.
type CourseRepository interface {
	Create(ctx context.Context, course *entity.Course) error
	GetAll(ctx context.Context) ([]*entity.Course, error)
	GetPaginated(ctx context.Context, page, limit int, includeSections bool) ([]*entity.Course, int64, error)
	GetByKey(ctx context.Context, code string, year, semester int) (*entity.Course, error)
	Update(ctx context.Context, course *entity.Course) error
	SoftDelete(ctx context.Context, code string, year, semester int) error
}
