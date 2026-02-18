package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/CPNext-hub/calendar-reg-main-api/internal/domain/entity"
	"github.com/CPNext-hub/calendar-reg-main-api/pkg/queue"
	"github.com/CPNext-hub/calendar-reg-main-api/pkg/scheduler"
)

// ----- mock CronJobRepository -----

type mockCronJobRepo struct {
	jobs      map[string]*entity.CronJob
	createErr error
	getAllErr error
	getByErr  error
	updateErr error
	deleteErr error
	idCounter int
}

func newMockCronJobRepo() *mockCronJobRepo {
	return &mockCronJobRepo{jobs: make(map[string]*entity.CronJob)}
}

func (m *mockCronJobRepo) Create(_ context.Context, job *entity.CronJob) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.idCounter++
	job.ID = "mock-id-" + string(rune('0'+m.idCounter))
	m.jobs[job.ID] = job
	return nil
}

func (m *mockCronJobRepo) GetAll(_ context.Context) ([]*entity.CronJob, error) {
	if m.getAllErr != nil {
		return nil, m.getAllErr
	}
	result := make([]*entity.CronJob, 0, len(m.jobs))
	for _, j := range m.jobs {
		result = append(result, j)
	}
	return result, nil
}

func (m *mockCronJobRepo) GetByID(_ context.Context, id string) (*entity.CronJob, error) {
	if m.getByErr != nil {
		return nil, m.getByErr
	}
	j, ok := m.jobs[id]
	if !ok {
		return nil, nil
	}
	return j, nil
}

func (m *mockCronJobRepo) Update(_ context.Context, job *entity.CronJob) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	if _, ok := m.jobs[job.ID]; !ok {
		return errors.New("cron job not found")
	}
	m.jobs[job.ID] = job
	return nil
}

func (m *mockCronJobRepo) Delete(_ context.Context, id string) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	if _, ok := m.jobs[id]; !ok {
		return errors.New("cron job not found")
	}
	delete(m.jobs, id)
	return nil
}

func (m *mockCronJobRepo) GetEnabled(_ context.Context) ([]*entity.CronJob, error) {
	result := make([]*entity.CronJob, 0)
	for _, j := range m.jobs {
		if j.Enabled {
			result = append(result, j)
		}
	}
	return result, nil
}

// ----- helper: create a real scheduler backed by a queue -----

func newTestScheduler() *scheduler.Scheduler {
	q := queue.New(10, 1)
	q.Start(func(job queue.RefreshJob) {
		q.MarkDone(job.Key())
	})
	return scheduler.New(q)
}

// ----- CreateCronJob tests -----

func TestCreateCronJob_Success(t *testing.T) {
	repo := newMockCronJobRepo()
	sched := newTestScheduler()
	defer sched.Stop()
	uc := NewCronJobUsecase(repo, sched)

	job := &entity.CronJob{
		Name:        "Test Job",
		CourseCodes: []string{"CP353004"},
		Acadyear:    2568,
		Semester:    2,
		CronExpr:    "0 */6 * * *",
		Enabled:     true,
	}
	err := uc.CreateCronJob(context.Background(), job)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if job.ID == "" {
		t.Error("expected job ID to be set")
	}
	if _, ok := repo.jobs[job.ID]; !ok {
		t.Error("expected job to be stored in repo")
	}
}

func TestCreateCronJob_InvalidCronExpr(t *testing.T) {
	repo := newMockCronJobRepo()
	sched := newTestScheduler()
	defer sched.Stop()
	uc := NewCronJobUsecase(repo, sched)

	job := &entity.CronJob{
		Name:        "Bad Expr",
		CourseCodes: []string{"CP353004"},
		Acadyear:    2568,
		Semester:    2,
		CronExpr:    "not-a-cron-expression",
		Enabled:     true,
	}
	err := uc.CreateCronJob(context.Background(), job)
	if err == nil {
		t.Fatal("expected error for invalid cron expression")
	}
}

func TestCreateCronJob_RepoError(t *testing.T) {
	repo := newMockCronJobRepo()
	repo.createErr = errors.New("insert failed")
	sched := newTestScheduler()
	defer sched.Stop()
	uc := NewCronJobUsecase(repo, sched)

	job := &entity.CronJob{
		Name:        "Fail Job",
		CourseCodes: []string{"CP353004"},
		Acadyear:    2568,
		Semester:    2,
		CronExpr:    "0 */6 * * *",
		Enabled:     true,
	}
	err := uc.CreateCronJob(context.Background(), job)
	if err == nil || err.Error() != "insert failed" {
		t.Errorf("expected 'insert failed', got %v", err)
	}
}

// ----- GetAllCronJobs tests -----

func TestGetAllCronJobs_Success(t *testing.T) {
	repo := newMockCronJobRepo()
	repo.jobs["1"] = &entity.CronJob{BaseEntity: entity.BaseEntity{ID: "1"}, Name: "Job1"}
	repo.jobs["2"] = &entity.CronJob{BaseEntity: entity.BaseEntity{ID: "2"}, Name: "Job2"}
	sched := newTestScheduler()
	defer sched.Stop()
	uc := NewCronJobUsecase(repo, sched)

	jobs, err := uc.GetAllCronJobs(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(jobs) != 2 {
		t.Errorf("expected 2 jobs, got %d", len(jobs))
	}
}

func TestGetAllCronJobs_Error(t *testing.T) {
	repo := newMockCronJobRepo()
	repo.getAllErr = errors.New("find failed")
	sched := newTestScheduler()
	defer sched.Stop()
	uc := NewCronJobUsecase(repo, sched)

	_, err := uc.GetAllCronJobs(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
}

// ----- GetCronJobByID tests -----

func TestGetCronJobByID_Found(t *testing.T) {
	repo := newMockCronJobRepo()
	repo.jobs["abc"] = &entity.CronJob{BaseEntity: entity.BaseEntity{ID: "abc"}, Name: "Test"}
	sched := newTestScheduler()
	defer sched.Stop()
	uc := NewCronJobUsecase(repo, sched)

	job, err := uc.GetCronJobByID(context.Background(), "abc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if job == nil || job.Name != "Test" {
		t.Error("expected to find job with correct name")
	}
}

func TestGetCronJobByID_NotFound(t *testing.T) {
	repo := newMockCronJobRepo()
	sched := newTestScheduler()
	defer sched.Stop()
	uc := NewCronJobUsecase(repo, sched)

	job, err := uc.GetCronJobByID(context.Background(), "nonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if job != nil {
		t.Error("expected nil for non-existent job")
	}
}

// ----- UpdateCronJob tests -----

func TestUpdateCronJob_Success(t *testing.T) {
	repo := newMockCronJobRepo()
	repo.jobs["abc"] = &entity.CronJob{
		BaseEntity:  entity.BaseEntity{ID: "abc"},
		Name:        "Old Name",
		CourseCodes: []string{"CP353004"},
		CronExpr:    "0 */6 * * *",
		Enabled:     true,
	}
	sched := newTestScheduler()
	defer sched.Stop()
	uc := NewCronJobUsecase(repo, sched)

	updated := &entity.CronJob{
		BaseEntity:  entity.BaseEntity{ID: "abc"},
		Name:        "New Name",
		CourseCodes: []string{"CP353004", "SC313002"},
		CronExpr:    "0 0 * * *",
		Enabled:     true,
	}
	err := uc.UpdateCronJob(context.Background(), updated)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.jobs["abc"].Name != "New Name" {
		t.Error("expected name to be updated")
	}
}

func TestUpdateCronJob_InvalidCronExpr(t *testing.T) {
	repo := newMockCronJobRepo()
	repo.jobs["abc"] = &entity.CronJob{BaseEntity: entity.BaseEntity{ID: "abc"}, CronExpr: "0 */6 * * *"}
	sched := newTestScheduler()
	defer sched.Stop()
	uc := NewCronJobUsecase(repo, sched)

	updated := &entity.CronJob{
		BaseEntity: entity.BaseEntity{ID: "abc"},
		CronExpr:   "invalid",
	}
	err := uc.UpdateCronJob(context.Background(), updated)
	if err == nil {
		t.Fatal("expected error for invalid cron expression")
	}
}

func TestUpdateCronJob_NotFound(t *testing.T) {
	repo := newMockCronJobRepo()
	sched := newTestScheduler()
	defer sched.Stop()
	uc := NewCronJobUsecase(repo, sched)

	job := &entity.CronJob{
		BaseEntity: entity.BaseEntity{ID: "nope"},
		CronExpr:   "0 */6 * * *",
	}
	err := uc.UpdateCronJob(context.Background(), job)
	if err == nil {
		t.Fatal("expected error for non-existent job")
	}
}

// ----- DeleteCronJob tests -----

func TestDeleteCronJob_Success(t *testing.T) {
	repo := newMockCronJobRepo()
	repo.jobs["abc"] = &entity.CronJob{BaseEntity: entity.BaseEntity{ID: "abc"}, Name: "Test"}
	sched := newTestScheduler()
	defer sched.Stop()
	uc := NewCronJobUsecase(repo, sched)

	err := uc.DeleteCronJob(context.Background(), "abc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := repo.jobs["abc"]; ok {
		t.Error("expected job to be deleted from repo")
	}
}

func TestDeleteCronJob_NotFound(t *testing.T) {
	repo := newMockCronJobRepo()
	sched := newTestScheduler()
	defer sched.Stop()
	uc := NewCronJobUsecase(repo, sched)

	err := uc.DeleteCronJob(context.Background(), "nope")
	if err == nil {
		t.Fatal("expected error for non-existent job")
	}
}

// ----- TriggerCronJob tests -----

func TestTriggerCronJob_Success(t *testing.T) {
	repo := newMockCronJobRepo()
	repo.jobs["abc"] = &entity.CronJob{
		BaseEntity:  entity.BaseEntity{ID: "abc"},
		Name:        "Test",
		CourseCodes: []string{"CP353004"},
		Acadyear:    2568,
		Semester:    2,
		CronExpr:    "0 */6 * * *",
		Enabled:     true,
	}
	sched := newTestScheduler()
	defer sched.Stop()
	uc := NewCronJobUsecase(repo, sched)

	err := uc.TriggerCronJob(context.Background(), "abc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTriggerCronJob_NotFound(t *testing.T) {
	repo := newMockCronJobRepo()
	sched := newTestScheduler()
	defer sched.Stop()
	uc := NewCronJobUsecase(repo, sched)

	err := uc.TriggerCronJob(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for non-existent job")
	}
}
