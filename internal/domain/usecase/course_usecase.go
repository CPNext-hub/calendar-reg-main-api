package usecase

import (
	"context"
	"errors"

	"github.com/CPNext-hub/calendar-reg-main-api/internal/domain/entity"
	"github.com/CPNext-hub/calendar-reg-main-api/internal/domain/repository"
)

// CourseUsecase defines the business logic for courses.
type CourseUsecase interface {
	CreateCourse(ctx context.Context, code, name, credits string) error
	GetAllCourses(ctx context.Context) ([]*entity.Course, error)
	GetCourseByCode(ctx context.Context, code string) (*entity.Course, error)
	DeleteCourse(ctx context.Context, code string) error
}

type courseUsecase struct {
	repo repository.CourseRepository
}

// NewCourseUsecase creates a new instance of CourseUsecase.
func NewCourseUsecase(repo repository.CourseRepository) CourseUsecase {
	return &courseUsecase{repo: repo}
}

func (u *courseUsecase) CreateCourse(ctx context.Context, code, name, credits string) error {
	// Check if exists
	existing, err := u.repo.GetByCode(ctx, code)
	if err != nil {
		return err
	}
	if existing != nil {
		return errors.New("course code already exists")
	}

	course := &entity.Course{
		Code:    code,
		Name:    name,
		Credits: credits,
	}

	return u.repo.Create(ctx, course)
}

func (u *courseUsecase) GetAllCourses(ctx context.Context) ([]*entity.Course, error) {
	return u.repo.GetAll(ctx)
}

func (u *courseUsecase) GetCourseByCode(ctx context.Context, code string) (*entity.Course, error) {
	return u.repo.GetByCode(ctx, code)
}

func (u *courseUsecase) DeleteCourse(ctx context.Context, code string) error {
	// Verify course exists before soft-deleting
	existing, err := u.repo.GetByCode(ctx, code)
	if err != nil {
		return err
	}
	if existing == nil {
		return errors.New("course not found")
	}
	return u.repo.SoftDelete(ctx, code)
}
