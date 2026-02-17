package repository

import (
	"context"

	"github.com/CPNext-hub/calendar-reg-main-api/internal/domain/entity"
)

// CourseExternalAPI abstracts the external course data API.
type CourseExternalAPI interface {
	// FetchByCode fetches a course from the external API by its code, academic year, and semester.
	FetchByCode(ctx context.Context, code string, acadyear, semester int) (*entity.Course, error)
}
