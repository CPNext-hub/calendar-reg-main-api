package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/CPNext-hub/calendar-reg-main-api/internal/domain/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ----- Mock CronJobRepository -----

type mockCronJobRepo struct {
	mock.Mock
}

func (m *mockCronJobRepo) Create(ctx context.Context, job *entity.CronJob) error {
	args := m.Called(ctx, job)
	return args.Error(0)
}

func (m *mockCronJobRepo) GetAll(ctx context.Context) ([]*entity.CronJob, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.CronJob), args.Error(1)
}

func (m *mockCronJobRepo) GetByID(ctx context.Context, id string) (*entity.CronJob, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.CronJob), args.Error(1)
}

func (m *mockCronJobRepo) Update(ctx context.Context, job *entity.CronJob) error {
	args := m.Called(ctx, job)
	return args.Error(0)
}

func (m *mockCronJobRepo) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockCronJobRepo) GetEnabled(ctx context.Context) ([]*entity.CronJob, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.CronJob), args.Error(1)
}

// ----- Mock CronScheduler -----

type mockScheduler struct {
	mock.Mock
}

func (m *mockScheduler) AddJob(job *entity.CronJob) error {
	args := m.Called(job)
	return args.Error(0)
}

func (m *mockScheduler) RemoveJob(id string) {
	m.Called(id)
}

func (m *mockScheduler) TriggerJob(job *entity.CronJob) {
	m.Called(job)
}

// ----- Tests -----

func TestCreateCronJob_Success(t *testing.T) {
	repo := new(mockCronJobRepo)
	sched := new(mockScheduler)
	uc := NewCronJobUsecase(repo, sched)

	job := &entity.CronJob{
		Name:       "Job1",
		CronExpr:   "* * * * *",
		Enabled:    true,
		BaseEntity: entity.BaseEntity{ID: "j1"},
	}

	repo.On("Create", mock.Anything, job).Return(nil)
	sched.On("AddJob", job).Return(nil)

	err := uc.CreateCronJob(context.Background(), job)
	assert.NoError(t, err)
	repo.AssertExpectations(t)
	sched.AssertExpectations(t)
}

func TestCreateCronJob_InvalidCron(t *testing.T) {
	repo := new(mockCronJobRepo)
	sched := new(mockScheduler)
	uc := NewCronJobUsecase(repo, sched)

	job := &entity.CronJob{CronExpr: "invalid"}

	err := uc.CreateCronJob(context.Background(), job)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid cron expression")
}

func TestCreateCronJob_RepoError(t *testing.T) {
	repo := new(mockCronJobRepo)
	sched := new(mockScheduler)
	uc := NewCronJobUsecase(repo, sched)

	job := &entity.CronJob{CronExpr: "* * * * *"}
	repo.On("Create", mock.Anything, job).Return(errors.New("db error"))

	err := uc.CreateCronJob(context.Background(), job)
	assert.EqualError(t, err, "db error")
}

func TestCreateCronJob_SchedulerError(t *testing.T) {
	repo := new(mockCronJobRepo)
	sched := new(mockScheduler)
	uc := NewCronJobUsecase(repo, sched)

	job := &entity.CronJob{CronExpr: "* * * * *", Enabled: true}
	repo.On("Create", mock.Anything, job).Return(nil)
	sched.On("AddJob", job).Return(errors.New("sched error"))

	// Should not return error, just log
	err := uc.CreateCronJob(context.Background(), job)
	assert.NoError(t, err)
}

func TestGetAllCronJobs(t *testing.T) {
	repo := new(mockCronJobRepo)
	sched := new(mockScheduler)
	uc := NewCronJobUsecase(repo, sched)

	repo.On("GetAll", mock.Anything).Return([]*entity.CronJob{}, nil)

	_, err := uc.GetAllCronJobs(context.Background())
	assert.NoError(t, err)
}

func TestGetCronJobByID(t *testing.T) {
	repo := new(mockCronJobRepo)
	sched := new(mockScheduler)
	uc := NewCronJobUsecase(repo, sched)

	repo.On("GetByID", mock.Anything, "j1").Return(&entity.CronJob{}, nil)

	_, err := uc.GetCronJobByID(context.Background(), "j1")
	assert.NoError(t, err)
}

func TestUpdateCronJob_Success(t *testing.T) {
	repo := new(mockCronJobRepo)
	sched := new(mockScheduler)
	uc := NewCronJobUsecase(repo, sched)

	job := &entity.CronJob{CronExpr: "* * * * *"}
	repo.On("Update", mock.Anything, job).Return(nil)
	sched.On("AddJob", job).Return(nil)

	err := uc.UpdateCronJob(context.Background(), job)
	assert.NoError(t, err)
}

func TestUpdateCronJob_InvalidCron(t *testing.T) {
	repo := new(mockCronJobRepo)
	sched := new(mockScheduler)
	uc := NewCronJobUsecase(repo, sched)

	job := &entity.CronJob{CronExpr: "invalid"}

	err := uc.UpdateCronJob(context.Background(), job)
	assert.Error(t, err)
}

func TestUpdateCronJob_RepoError(t *testing.T) {
	repo := new(mockCronJobRepo)
	sched := new(mockScheduler)
	uc := NewCronJobUsecase(repo, sched)

	job := &entity.CronJob{CronExpr: "* * * * *"}
	repo.On("Update", mock.Anything, job).Return(errors.New("db error"))

	err := uc.UpdateCronJob(context.Background(), job)
	assert.EqualError(t, err, "db error")
}

func TestUpdateCronJob_SchedulerError(t *testing.T) {
	repo := new(mockCronJobRepo)
	sched := new(mockScheduler)
	uc := NewCronJobUsecase(repo, sched)

	job := &entity.CronJob{CronExpr: "* * * * *"}
	repo.On("Update", mock.Anything, job).Return(nil)
	sched.On("AddJob", job).Return(errors.New("sched error"))

	err := uc.UpdateCronJob(context.Background(), job)
	assert.NoError(t, err)
}

func TestDeleteCronJob(t *testing.T) {
	repo := new(mockCronJobRepo)
	sched := new(mockScheduler)
	uc := NewCronJobUsecase(repo, sched)

	repo.On("Delete", mock.Anything, "j1").Return(nil)
	sched.On("RemoveJob", "j1").Return()

	err := uc.DeleteCronJob(context.Background(), "j1")
	assert.NoError(t, err)
}

func TestDeleteCronJob_RepoError(t *testing.T) {
	repo := new(mockCronJobRepo)
	sched := new(mockScheduler)
	uc := NewCronJobUsecase(repo, sched)

	repo.On("Delete", mock.Anything, "j1").Return(errors.New("db error"))

	err := uc.DeleteCronJob(context.Background(), "j1")
	assert.EqualError(t, err, "db error")
}

func TestTriggerCronJob_Success(t *testing.T) {
	repo := new(mockCronJobRepo)
	sched := new(mockScheduler)
	uc := NewCronJobUsecase(repo, sched)

	job := &entity.CronJob{BaseEntity: entity.BaseEntity{ID: "j1"}}
	repo.On("GetByID", mock.Anything, "j1").Return(job, nil)
	sched.On("TriggerJob", job).Return()

	err := uc.TriggerCronJob(context.Background(), "j1")
	assert.NoError(t, err)
}

func TestTriggerCronJob_NotFound(t *testing.T) {
	repo := new(mockCronJobRepo)
	sched := new(mockScheduler)
	uc := NewCronJobUsecase(repo, sched)

	repo.On("GetByID", mock.Anything, "j1").Return(nil, nil)

	err := uc.TriggerCronJob(context.Background(), "j1")
	assert.EqualError(t, err, "cron job not found")
}

func TestTriggerCronJob_RepoError(t *testing.T) {
	repo := new(mockCronJobRepo)
	sched := new(mockScheduler)
	uc := NewCronJobUsecase(repo, sched)

	repo.On("GetByID", mock.Anything, "j1").Return(nil, errors.New("db error"))

	err := uc.TriggerCronJob(context.Background(), "j1")
	assert.EqualError(t, err, "db error")
}
