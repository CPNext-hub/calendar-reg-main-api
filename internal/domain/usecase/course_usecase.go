package usecase

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/CPNext-hub/calendar-reg-main-api/internal/domain/entity"
	"github.com/CPNext-hub/calendar-reg-main-api/internal/domain/repository"
	"github.com/CPNext-hub/calendar-reg-main-api/pkg/pagination"
	"github.com/CPNext-hub/calendar-reg-main-api/pkg/queue"
)

// CourseUsecase defines the business logic for courses.
type CourseUsecase interface {
	CreateCourse(ctx context.Context, course *entity.Course) error
	GetAllCourses(ctx context.Context) ([]*entity.Course, error)
	GetCoursesPaginated(ctx context.Context, pq pagination.PaginationQuery) (*pagination.PaginatedResult[*entity.Course], error)
	GetCourseByCode(ctx context.Context, code string, acadyear, semester int) (*entity.Course, error)
	DeleteCourse(ctx context.Context, code string, year, semester int) error
	ProcessRefreshJob(job queue.RefreshJob)
}

type courseUsecase struct {
	repo         repository.CourseRepository
	externalAPI  repository.CourseExternalAPI
	refreshQueue *queue.RefreshQueue
}

// NewCourseUsecase creates a new instance of CourseUsecase.
func NewCourseUsecase(repo repository.CourseRepository, extAPI repository.CourseExternalAPI, q *queue.RefreshQueue) CourseUsecase {
	return &courseUsecase{repo: repo, externalAPI: extAPI, refreshQueue: q}
}

func (u *courseUsecase) CreateCourse(ctx context.Context, course *entity.Course) error {
	existing, err := u.repo.GetByKey(ctx, course.Code, course.Year, course.Semester)
	if err != nil {
		return err
	}
	if existing != nil {
		return errors.New("course already exists for this code/year/semester")
	}
	return u.repo.Create(ctx, course)
}

func (u *courseUsecase) GetAllCourses(ctx context.Context) ([]*entity.Course, error) {
	return u.repo.GetAll(ctx)
}

func (u *courseUsecase) GetCoursesPaginated(ctx context.Context, pq pagination.PaginationQuery) (*pagination.PaginatedResult[*entity.Course], error) {
	items, total, err := u.repo.GetPaginated(ctx, pq.Page, pq.Limit)
	if err != nil {
		return nil, err
	}
	result := pagination.NewResult(items, pq.Page, pq.Limit, total)
	return &result, nil
}

func (u *courseUsecase) GetCourseByCode(ctx context.Context, code string, acadyear, semester int) (*entity.Course, error) {
	course, err := u.repo.GetByKey(ctx, code, acadyear, semester)
	if err != nil {
		return nil, err
	}

	if course == nil {
		// Not in DB — enqueue a first-fetch job and wait up to 3 seconds.
		if u.externalAPI == nil || u.refreshQueue == nil {
			return nil, nil
		}

		resultCh := make(chan queue.JobResult, 1)
		if !u.refreshQueue.Enqueue(queue.RefreshJob{Code: code, Acadyear: acadyear, Semester: semester, IsNew: true, Result: resultCh}) {
			return nil, nil
		}

		select {
		case res := <-resultCh:
			if res.Err != nil {
				log.Printf("external fetch failed for %s: %v", code, res.Err)
				return nil, nil
			}
			if c, ok := res.Data.(*entity.Course); ok {
				return c, nil
			}
			return nil, nil

		case <-time.After(3 * time.Second):
			log.Printf("external fetch for %s taking too long, returning nil", code)
			return nil, nil
		}
	}

	// Check if last updated today — if not, enqueue a background refresh.
	if !isToday(course.UpdatedAt) {
		if u.externalAPI != nil && u.refreshQueue != nil {
			u.refreshQueue.Enqueue(queue.RefreshJob{Code: code, Acadyear: course.Year, Semester: course.Semester, IsNew: false})
		}
	}

	return course, nil
}

// ProcessRefreshJob is called by worker pool goroutines to fetch and save course data.
func (u *courseUsecase) ProcessRefreshJob(job queue.RefreshJob) {
	defer u.refreshQueue.MarkDone(job.Key())

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	fetched, err := u.externalAPI.FetchByCode(ctx, job.Code, job.Acadyear, job.Semester)
	if err != nil {
		log.Printf("[worker] fetch failed for %s: %v", job.Key(), err)
		if job.Result != nil {
			job.Result <- queue.JobResult{Err: err}
		}
		return
	}

	if job.IsNew {
		// First fetch — create new record.
		fetched.CreatedAt = time.Now()
		if saveErr := u.repo.Create(context.Background(), fetched); saveErr != nil {
			log.Printf("[worker] failed to save new course %s: %v", job.Key(), saveErr)
			if job.Result != nil {
				job.Result <- queue.JobResult{Err: saveErr}
			}
			return
		}
		log.Printf("[worker] new course %s fetched and saved", job.Key())
	} else {
		// Stale refresh — update existing record, preserving identity.
		existing, getErr := u.repo.GetByKey(ctx, job.Code, job.Acadyear, job.Semester)
		if getErr != nil || existing == nil {
			log.Printf("[worker] could not find existing course %s for refresh: %v", job.Key(), getErr)
			return
		}
		fetched.ID = existing.ID
		fetched.Code = existing.Code
		fetched.CreatedAt = existing.CreatedAt

		if saveErr := u.repo.Update(context.Background(), fetched); saveErr != nil {
			log.Printf("[worker] failed to update course %s: %v", job.Key(), saveErr)
			return
		}
		log.Printf("[worker] course %s refreshed and saved", job.Key())
	}

	// Send result back to caller if they're waiting.
	if job.Result != nil {
		job.Result <- queue.JobResult{Data: fetched}
	}
}

// isToday checks whether t falls on the current calendar day (local time).
func isToday(t time.Time) bool {
	now := time.Now()
	y1, m1, d1 := now.Date()
	y2, m2, d2 := t.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}

func (u *courseUsecase) DeleteCourse(ctx context.Context, code string, year, semester int) error {
	existing, err := u.repo.GetByKey(ctx, code, year, semester)
	if err != nil {
		return err
	}
	if existing == nil {
		return errors.New("course not found")
	}
	return u.repo.SoftDelete(ctx, code, year, semester)
}
