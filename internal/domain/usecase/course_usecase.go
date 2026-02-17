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
	GetCourseByCode(ctx context.Context, code string) (*entity.Course, error)
	DeleteCourse(ctx context.Context, code string) error
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
	existing, err := u.repo.GetByCode(ctx, course.Code)
	if err != nil {
		return err
	}
	if existing != nil {
		return errors.New("course code already exists")
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

// refreshResult holds the result from a background refresh goroutine.
type refreshResult struct {
	course *entity.Course
	err    error
}

func (u *courseUsecase) GetCourseByCode(ctx context.Context, code string) (*entity.Course, error) {
	course, err := u.repo.GetByCode(ctx, code)
	if err != nil {
		return nil, err
	}
	if course == nil {
		// Not in DB — try to fetch from external API (same pattern as stale refresh).
		if u.externalAPI == nil {
			return nil, nil
		}
		if u.refreshQueue != nil && !u.refreshQueue.Enqueue(queue.RefreshJob{Code: code}) {
			return nil, nil
		}

		ch := make(chan refreshResult, 1)
		go func() {
			fetchCtx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			defer cancel()

			fetched, err := u.externalAPI.FetchByCode(fetchCtx, code)
			if err == nil {
				fetched.CreatedAt = time.Now()
			}
			ch <- refreshResult{course: fetched, err: err}
		}()

		select {
		case res := <-ch:
			if u.refreshQueue != nil {
				u.refreshQueue.MarkDone(code)
			}
			if res.err != nil {
				log.Printf("external fetch failed for %s: %v", code, res.err)
				return nil, nil
			}
			if err := u.repo.Create(context.Background(), res.course); err != nil {
				log.Printf("failed to save fetched course %s: %v", code, err)
				return nil, err
			}
			log.Printf("course %s fetched from external API and saved", code)
			return res.course, nil

		case <-time.After(3 * time.Second):
			log.Printf("external fetch for %s taking too long, returning nil", code)

			go func() {
				if u.refreshQueue != nil {
					defer u.refreshQueue.MarkDone(code)
				}
				res := <-ch
				if res.err != nil {
					log.Printf("[background] external fetch failed for %s: %v", code, res.err)
					return
				}
				res.course.CreatedAt = time.Now()
				bgCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
				defer cancel()
				if err := u.repo.Create(bgCtx, res.course); err != nil {
					log.Printf("[background] failed to save fetched course %s: %v", code, err)
				} else {
					log.Printf("[background] fetch completed for %s", code)
				}
			}()
			return nil, nil
		}
	}

	// Check if last updated today — if not, try to refresh.
	if !isToday(course.UpdatedAt) {
		// Acquire refresh lock — skip if already in progress for this code.
		if u.externalAPI == nil {
			return course, nil
		}
		if u.refreshQueue != nil && !u.refreshQueue.Enqueue(queue.RefreshJob{Code: code}) {
			return course, nil
		}

		// Fire ONE request in a goroutine via the external API interface.
		ch := make(chan refreshResult, 1)
		go func() {
			refreshCtx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			defer cancel()

			refreshed, err := u.externalAPI.FetchByCode(refreshCtx, code)
			if err == nil {
				// Preserve identity from the existing course.
				refreshed.ID = course.ID
				refreshed.Code = course.Code
				refreshed.CreatedAt = course.CreatedAt
			}
			ch <- refreshResult{course: refreshed, err: err}
		}()

		// Wait up to 3 seconds for the response.
		select {
		case res := <-ch:
			// Got result within 3 seconds — release the lock.
			if u.refreshQueue != nil {
				u.refreshQueue.MarkDone(code)
			}
			if res.err != nil {
				log.Printf("course refresh failed for %s: %v", code, res.err)
			} else {
				if saveErr := u.repo.Update(context.Background(), res.course); saveErr != nil {
					log.Printf("failed to save refreshed course %s: %v", code, saveErr)
				}
				course = res.course
			}

		case <-time.After(3 * time.Second):
			// Timeout — return stale data, goroutine continues in background.
			log.Printf("course refresh for %s taking too long, returning stale data", code)

			go func() {
				if u.refreshQueue != nil {
					defer u.refreshQueue.MarkDone(code)
				}
				res := <-ch
				if res.err != nil {
					log.Printf("[background] refresh failed for %s: %v", code, res.err)
					return
				}
				bgCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
				defer cancel()
				if err := u.repo.Update(bgCtx, res.course); err != nil {
					log.Printf("[background] failed to save refreshed course %s: %v", code, err)
				} else {
					log.Printf("[background] refresh completed for %s", code)
				}
			}()
		}
	}

	return course, nil
}

// ProcessRefreshJob is kept for interface compatibility.
func (u *courseUsecase) ProcessRefreshJob(job queue.RefreshJob) {
	// no-op: goroutines handle the work directly
}

// isToday checks whether t falls on the current calendar day (local time).
func isToday(t time.Time) bool {
	now := time.Now()
	y1, m1, d1 := now.Date()
	y2, m2, d2 := t.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}

func (u *courseUsecase) DeleteCourse(ctx context.Context, code string) error {
	existing, err := u.repo.GetByCode(ctx, code)
	if err != nil {
		return err
	}
	if existing == nil {
		return errors.New("course not found")
	}
	return u.repo.SoftDelete(ctx, code)
}
