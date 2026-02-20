package usecase

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/CPNext-hub/calendar-reg-main-api/internal/domain/entity"
	"github.com/CPNext-hub/calendar-reg-main-api/pkg/pagination"
	"github.com/CPNext-hub/calendar-reg-main-api/pkg/queue"
)

// ----- mock CourseRepository -----

type mockCourseRepo struct {
	courses    map[string]*entity.Course // key = "code:year:semester"
	createErr  error
	getByErr   error
	getAllErr  error
	pagErr     error
	updateErr  error
	deleteErr  error
	allCourses []*entity.Course
}

// ----- mock CourseExternalAPI -----

type mockExternalAPI struct {
	fetchByCodeFunc func(ctx context.Context, code string, acadyear, semester int) (*entity.Course, error)
}

func (m *mockExternalAPI) FetchByCode(ctx context.Context, code string, acadyear, semester int) (*entity.Course, error) {
	if m.fetchByCodeFunc != nil {
		return m.fetchByCodeFunc(ctx, code, acadyear, semester)
	}
	return nil, errors.New("mock external api not implemented")
}

func newMockCourseRepo() *mockCourseRepo {
	return &mockCourseRepo{courses: make(map[string]*entity.Course)}
}

func mockKey(code string, year, semester int) string {
	return fmt.Sprintf("%s:%d:%d", code, year, semester)
}

func (m *mockCourseRepo) Create(_ context.Context, c *entity.Course) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.courses[c.Key()] = c
	return nil
}

func (m *mockCourseRepo) GetAll(_ context.Context) ([]*entity.Course, error) {
	if m.getAllErr != nil {
		return nil, m.getAllErr
	}
	return m.allCourses, nil
}

func (m *mockCourseRepo) GetPaginated(_ context.Context, page, limit int, includeSections bool) ([]*entity.Course, int64, error) {
	if m.pagErr != nil {
		return nil, 0, m.pagErr
	}
	return m.allCourses, int64(len(m.allCourses)), nil
}

func (m *mockCourseRepo) GetByKey(_ context.Context, code string, year, semester int) (*entity.Course, error) {
	if m.getByErr != nil {
		return nil, m.getByErr
	}
	c, ok := m.courses[mockKey(code, year, semester)]
	if !ok {
		return nil, nil
	}
	return c, nil
}

func (m *mockCourseRepo) SoftDelete(_ context.Context, code string, year, semester int) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	delete(m.courses, mockKey(code, year, semester))
	return nil
}

func (m *mockCourseRepo) Update(_ context.Context, c *entity.Course) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	m.courses[c.Key()] = c
	return nil
}

// ----- CreateCourse tests -----

func TestCreateCourse_Success(t *testing.T) {
	repo := newMockCourseRepo()
	uc := NewCourseUsecase(repo, nil, nil)

	course := &entity.Course{Code: "CS101", Year: 2568, Semester: 1}
	err := uc.CreateCourse(context.Background(), course)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if _, ok := repo.courses[course.Key()]; !ok {
		t.Error("expected course to be stored in repo")
	}
}

func TestCreateCourse_AlreadyExists(t *testing.T) {
	repo := newMockCourseRepo()
	c := &entity.Course{Code: "CS101", Year: 2568, Semester: 1}
	repo.courses[c.Key()] = c
	uc := NewCourseUsecase(repo, nil, nil)

	err := uc.CreateCourse(context.Background(), &entity.Course{Code: "CS101", Year: 2568, Semester: 1})
	if err == nil {
		t.Fatal("expected error for duplicate course")
	}
}

func TestCreateCourse_RepoGetByKeyError(t *testing.T) {
	repo := newMockCourseRepo()
	repo.getByErr = errors.New("db error")
	uc := NewCourseUsecase(repo, nil, nil)

	err := uc.CreateCourse(context.Background(), &entity.Course{Code: "CS101", Year: 2568, Semester: 1})
	if err == nil || err.Error() != "db error" {
		t.Errorf("expected 'db error', got %v", err)
	}
}

func TestCreateCourse_RepoCreateError(t *testing.T) {
	repo := newMockCourseRepo()
	repo.createErr = errors.New("insert failed")
	uc := NewCourseUsecase(repo, nil, nil)

	err := uc.CreateCourse(context.Background(), &entity.Course{Code: "CS101", Year: 2568, Semester: 1})
	if err == nil || err.Error() != "insert failed" {
		t.Errorf("expected 'insert failed', got %v", err)
	}
}

// ----- GetAllCourses tests -----

func TestGetAllCourses_Success(t *testing.T) {
	repo := newMockCourseRepo()
	repo.allCourses = []*entity.Course{{Code: "CS101"}, {Code: "CS102"}}
	uc := NewCourseUsecase(repo, nil, nil)

	courses, err := uc.GetAllCourses(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(courses) != 2 {
		t.Errorf("expected 2 courses, got %d", len(courses))
	}
}

func TestGetAllCourses_Error(t *testing.T) {
	repo := newMockCourseRepo()
	repo.getAllErr = errors.New("find failed")
	uc := NewCourseUsecase(repo, nil, nil)

	_, err := uc.GetAllCourses(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
}

// ----- GetCoursesPaginated tests -----

func TestGetCoursesPaginated_Success(t *testing.T) {
	repo := newMockCourseRepo()
	repo.allCourses = []*entity.Course{{Code: "CS101"}, {Code: "CS102"}, {Code: "CS103"}}
	uc := NewCourseUsecase(repo, nil, nil)

	pq := pagination.PaginationQuery{Page: 1, Limit: 10}
	result, err := uc.GetCoursesPaginated(context.Background(), pq)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Total != 3 {
		t.Errorf("expected total=3, got %d", result.Total)
	}
	if len(result.Items) != 3 {
		t.Errorf("expected 3 items, got %d", len(result.Items))
	}
	if result.Page != 1 {
		t.Errorf("expected page=1, got %d", result.Page)
	}
}

func TestGetCoursesPaginated_LimitZero(t *testing.T) {
	repo := newMockCourseRepo()
	repo.allCourses = []*entity.Course{{Code: "CS101"}}
	uc := NewCourseUsecase(repo, nil, nil)

	pq := pagination.PaginationQuery{Page: 1, Limit: 0}
	result, err := uc.GetCoursesPaginated(context.Background(), pq)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.TotalPages != 1 {
		t.Errorf("expected totalPages=1 for limit=0, got %d", result.TotalPages)
	}
}

func TestGetCoursesPaginated_Error(t *testing.T) {
	repo := newMockCourseRepo()
	repo.pagErr = errors.New("paginate failed")
	uc := NewCourseUsecase(repo, nil, nil)

	pq := pagination.PaginationQuery{Page: 1, Limit: 10}
	_, err := uc.GetCoursesPaginated(context.Background(), pq)
	if err == nil {
		t.Fatal("expected error")
	}
}

// ----- GetCourseByCode tests -----

func TestGetCourseByCode_Found(t *testing.T) {
	repo := newMockCourseRepo()
	c := &entity.Course{Code: "CS101", Year: 2568, Semester: 1, NameEN: "Intro CS", BaseEntity: entity.BaseEntity{UpdatedAt: time.Now()}}
	repo.courses[c.Key()] = c
	uc := NewCourseUsecase(repo, nil, nil)

	course, err := uc.GetCourseByCode(context.Background(), "CS101", 2568, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if course == nil || course.NameEN != "Intro CS" {
		t.Error("expected to find course with correct name")
	}
}

func TestGetCourseByCode_NotFound_NoExternal(t *testing.T) {
	repo := newMockCourseRepo()
	uc := NewCourseUsecase(repo, nil, nil)

	course, err := uc.GetCourseByCode(context.Background(), "NOPE", 2568, 1)
	if !errors.Is(err, ErrCourseNotFound) {
		t.Fatalf("expected ErrCourseNotFound, got: %v", err)
	}
	if course != nil {
		t.Error("expected nil for non-existent course")
	}
}

func TestGetCourseByCode_Error(t *testing.T) {
	repo := newMockCourseRepo()
	repo.getByErr = errors.New("db error")
	uc := NewCourseUsecase(repo, nil, nil)

	_, err := uc.GetCourseByCode(context.Background(), "CS101", 2568, 1)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestGetCourseByCode_FetchNew_Success(t *testing.T) {
	repo := newMockCourseRepo()
	extAPI := &mockExternalAPI{
		fetchByCodeFunc: func(ctx context.Context, code string, acadyear, semester int) (*entity.Course, error) {
			return &entity.Course{Code: code, Year: acadyear, Semester: semester}, nil
		},
	}
	q := queue.New(10, 1)
	uc := NewCourseUsecase(repo, extAPI, q)

	// Start a worker that simulates success
	q.Start(func(job queue.RefreshJob) {
		job.Result <- queue.JobResult{Data: &entity.Course{Code: job.Code}}
	})
	defer q.Stop()

	course, err := uc.GetCourseByCode(context.Background(), "NEW101", 2568, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if course == nil || course.Code != "NEW101" {
		t.Errorf("expected fetched course, got %v", course)
	}
}

func TestGetCourseByCode_FetchNew_AlreadyInFlight(t *testing.T) {
	repo := newMockCourseRepo()
	extAPI := &mockExternalAPI{}
	q := queue.New(10, 1)
	uc := NewCourseUsecase(repo, extAPI, q)

	// Manually enqueue to block the key
	q.Enqueue(queue.RefreshJob{Code: "BUSY", Acadyear: 2568, Semester: 1})

	// Try fetching same key
	_, err := uc.GetCourseByCode(context.Background(), "BUSY", 2568, 1)
	if !errors.Is(err, ErrCourseFetchPending) {
		t.Errorf("expected ErrCourseFetchPending, got %v", err)
	}
}

func TestGetCourseByCode_FetchNew_Error(t *testing.T) {
	repo := newMockCourseRepo()
	extAPI := &mockExternalAPI{}
	q := queue.New(10, 1)
	uc := NewCourseUsecase(repo, extAPI, q)

	q.Start(func(job queue.RefreshJob) {
		job.Result <- queue.JobResult{Err: errors.New("fetch failed")}
	})
	defer q.Stop()

	_, err := uc.GetCourseByCode(context.Background(), "NEW101", 2568, 1)
	if !errors.Is(err, ErrCourseNotFound) {
		t.Errorf("expected ErrCourseNotFound, got %v", err)
	}
}

func TestGetCourseByCode_FetchNew_InvalidData(t *testing.T) {
	repo := newMockCourseRepo()
	extAPI := &mockExternalAPI{}
	q := queue.New(10, 1)
	uc := NewCourseUsecase(repo, extAPI, q)

	q.Start(func(job queue.RefreshJob) {
		// Return unexpected type
		job.Result <- queue.JobResult{Data: "not a course"}
	})
	defer q.Stop()

	_, err := uc.GetCourseByCode(context.Background(), "NEW101", 2568, 1)
	if !errors.Is(err, ErrCourseNotFound) {
		t.Errorf("expected ErrCourseNotFound, got %v", err)
	}
}

func TestGetCourseByCode_FetchNew_Timeout(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping timeout test in short mode")
	}
	repo := newMockCourseRepo()
	extAPI := &mockExternalAPI{}
	q := queue.New(10, 1)
	uc := NewCourseUsecase(repo, extAPI, q)

	// Worker sleeps longer than 3s
	q.Start(func(job queue.RefreshJob) {
		time.Sleep(3100 * time.Millisecond)
		job.Result <- queue.JobResult{Data: &entity.Course{Code: job.Code}}
	})
	defer q.Stop()

	// Only wait slightly longer than the usecase timeout (3s) to avoid test hanging too long
	// But usecase handles timeout internally.
	_, err := uc.GetCourseByCode(context.Background(), "SLOW101", 2568, 1)
	if !errors.Is(err, ErrCourseFetchPending) {
		t.Errorf("expected ErrCourseFetchPending, got %v", err)
	}
}

func TestGetCourseByCode_RefreshesStale(t *testing.T) {
	repo := newMockCourseRepo()
	// Stale course (yesterday)
	staleTime := time.Now().AddDate(0, 0, -1)
	c := &entity.Course{Code: "STALE", Year: 2568, Semester: 1, BaseEntity: entity.BaseEntity{UpdatedAt: staleTime}}
	repo.courses[c.Key()] = c

	extAPI := &mockExternalAPI{}
	q := queue.New(10, 1)
	uc := NewCourseUsecase(repo, extAPI, q)

	// Use a channel to detect if refresh was enqueued
	refreshed := make(chan bool, 1)
	q.Start(func(job queue.RefreshJob) {
		if job.Code == "STALE" && !job.IsNew {
			refreshed <- true
		}
	})
	defer q.Stop()

	course, err := uc.GetCourseByCode(context.Background(), "STALE", 2568, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if course == nil {
		t.Fatal("expected course")
	}

	select {
	case <-refreshed:
		// success
	case <-time.After(1 * time.Second):
		t.Error("expected background refresh to be enqueued")
	}
}

func TestGetCourseByCode_Stale_NoRefreshConfigured(t *testing.T) {
	repo := newMockCourseRepo()
	// Stale course
	staleTime := time.Now().AddDate(0, 0, -1)
	c := &entity.Course{Code: "STALE", Year: 2568, Semester: 1, BaseEntity: entity.BaseEntity{UpdatedAt: staleTime}}
	repo.courses[c.Key()] = c

	// No external API or Queue
	uc := NewCourseUsecase(repo, nil, nil)

	course, err := uc.GetCourseByCode(context.Background(), "STALE", 2568, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should return course without error
	if course == nil {
		t.Fatal("expected course")
	}
}

// ----- ProcessRefreshJob Tests (Worker Logic) -----

func TestProcessRefreshJob_New_Success(t *testing.T) {
	repo := newMockCourseRepo()
	extAPI := &mockExternalAPI{
		fetchByCodeFunc: func(ctx context.Context, code string, acadyear, semester int) (*entity.Course, error) {
			return &entity.Course{Code: code, Year: acadyear, Semester: semester}, nil
		},
	}
	q := queue.New(10, 1)
	uc := NewCourseUsecase(repo, extAPI, q)

	resultCh := make(chan queue.JobResult, 1)
	job := queue.RefreshJob{Code: "NEW", Acadyear: 2568, Semester: 1, IsNew: true, Result: resultCh}

	// We call ProcessRefreshJob directly as if we are the worker
	q.Enqueue(job) // Just to set inflight
	uc.ProcessRefreshJob(job)

	res := <-resultCh
	if res.Err != nil {
		t.Fatalf("expected success, got error: %v", res.Err)
	}
	c := res.Data.(*entity.Course)
	if c.Code != "NEW" {
		t.Errorf("expected NEW code, got %s", c.Code)
	}

	// Verify in repo
	if _, err := repo.GetByKey(context.Background(), "NEW", 2568, 1); err != nil {
		t.Error("expected course to be saved in repo")
	}
}

func TestProcessRefreshJob_FetchError(t *testing.T) {
	repo := newMockCourseRepo()
	extAPI := &mockExternalAPI{
		fetchByCodeFunc: func(ctx context.Context, code string, acadyear, semester int) (*entity.Course, error) {
			return nil, errors.New("external fail")
		},
	}
	q := queue.New(10, 1)
	uc := NewCourseUsecase(repo, extAPI, q)

	resultCh := make(chan queue.JobResult, 1)
	job := queue.RefreshJob{Code: "ERR", Acadyear: 2568, Semester: 1, IsNew: true, Result: resultCh}

	q.Enqueue(job)
	uc.ProcessRefreshJob(job)

	res := <-resultCh
	if res.Err == nil {
		t.Error("expected error result")
	}
}

func TestProcessRefreshJob_SaveError(t *testing.T) {
	repo := newMockCourseRepo()
	repo.createErr = errors.New("save fail")
	extAPI := &mockExternalAPI{
		fetchByCodeFunc: func(ctx context.Context, code string, acadyear, semester int) (*entity.Course, error) {
			return &entity.Course{Code: code}, nil
		},
	}
	q := queue.New(10, 1)
	uc := NewCourseUsecase(repo, extAPI, q)

	resultCh := make(chan queue.JobResult, 1)
	job := queue.RefreshJob{Code: "SAVE_ERR", Acadyear: 2568, Semester: 1, IsNew: true, Result: resultCh}

	q.Enqueue(job)
	uc.ProcessRefreshJob(job)

	res := <-resultCh
	if res.Err == nil {
		t.Error("expected error result")
	}
}

func TestProcessRefreshJob_Update_Success(t *testing.T) {
	repo := newMockCourseRepo()
	existing := &entity.Course{Code: "EXIST", Year: 2568, Semester: 1, NameEN: "Old Name"}
	repo.courses[existing.Key()] = existing

	extAPI := &mockExternalAPI{
		fetchByCodeFunc: func(ctx context.Context, code string, acadyear, semester int) (*entity.Course, error) {
			return &entity.Course{Code: code, Year: acadyear, Semester: semester, NameEN: "New Name"}, nil
		},
	}
	q := queue.New(10, 1)
	uc := NewCourseUsecase(repo, extAPI, q)

	job := queue.RefreshJob{Code: "EXIST", Acadyear: 2568, Semester: 1, IsNew: false}
	q.Enqueue(job)
	uc.ProcessRefreshJob(job)

	updated, _ := repo.GetByKey(context.Background(), "EXIST", 2568, 1)
	if updated.NameEN != "New Name" {
		t.Errorf("expected name updated to 'New Name', got '%s'", updated.NameEN)
	}
}

func TestProcessRefreshJob_Update_NotFound(t *testing.T) {
	repo := newMockCourseRepo()
	// Course NOT in repo

	extAPI := &mockExternalAPI{
		fetchByCodeFunc: func(ctx context.Context, code string, acadyear, semester int) (*entity.Course, error) {
			return &entity.Course{Code: code}, nil
		},
	}
	q := queue.New(10, 1)
	uc := NewCourseUsecase(repo, extAPI, q)

	job := queue.RefreshJob{Code: "MISSING", Acadyear: 2568, Semester: 1, IsNew: false}
	q.Enqueue(job)
	// Should log error and return without result (no crash)
	uc.ProcessRefreshJob(job)
}

func TestProcessRefreshJob_Update_SaveError(t *testing.T) {
	repo := newMockCourseRepo()
	existing := &entity.Course{Code: "EXIST", Year: 2568, Semester: 1}
	repo.courses[existing.Key()] = existing
	repo.updateErr = errors.New("update fail")

	extAPI := &mockExternalAPI{
		fetchByCodeFunc: func(ctx context.Context, code string, acadyear, semester int) (*entity.Course, error) {
			return &entity.Course{Code: code, Year: acadyear, Semester: semester, NameEN: "New Name"}, nil
		},
	}
	q := queue.New(10, 1)
	uc := NewCourseUsecase(repo, extAPI, q)

	job := queue.RefreshJob{Code: "EXIST", Acadyear: 2568, Semester: 1, IsNew: false}
	q.Enqueue(job)
	uc.ProcessRefreshJob(job)
	// Expect log "failed to update course"
}

// ----- DeleteCourse tests -----

func TestDeleteCourse_Success(t *testing.T) {
	repo := newMockCourseRepo()
	c := &entity.Course{Code: "CS101", Year: 2568, Semester: 1}
	repo.courses[c.Key()] = c
	uc := NewCourseUsecase(repo, nil, nil)

	err := uc.DeleteCourse(context.Background(), "CS101", 2568, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := repo.courses[c.Key()]; ok {
		t.Error("expected course to be deleted from repo")
	}
}

func TestDeleteCourse_NotFound(t *testing.T) {
	repo := newMockCourseRepo()
	uc := NewCourseUsecase(repo, nil, nil)

	err := uc.DeleteCourse(context.Background(), "NOPE", 2568, 1)
	if err == nil {
		t.Fatal("expected error for non-existent course")
	}
	if err.Error() != "course not found" {
		t.Errorf("expected 'course not found', got %q", err.Error())
	}
}

func TestDeleteCourse_RepoError(t *testing.T) {
	repo := newMockCourseRepo()
	c := &entity.Course{Code: "CS101", Year: 2568, Semester: 1}
	repo.courses[c.Key()] = c
	repo.deleteErr = errors.New("delete failed")
	uc := NewCourseUsecase(repo, nil, nil)

	err := uc.DeleteCourse(context.Background(), "CS101", 2568, 1)
	if err == nil || err.Error() != "delete failed" {
		t.Errorf("expected 'delete failed', got %v", err)
	}
}

func TestDeleteCourse_GetError(t *testing.T) {
	repo := newMockCourseRepo()
	repo.getByErr = errors.New("db error")
	uc := NewCourseUsecase(repo, nil, nil)

	err := uc.DeleteCourse(context.Background(), "CS101", 2568, 1)
	if err == nil || err.Error() != "db error" {
		t.Errorf("expected 'db error', got %v", err)
	}
}
