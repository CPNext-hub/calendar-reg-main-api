package usecase

import (
	"context"
	"errors"
	"log"

	"github.com/CPNext-hub/calendar-reg-main-api/internal/domain/entity"
	"github.com/CPNext-hub/calendar-reg-main-api/internal/domain/repository"
	"github.com/CPNext-hub/calendar-reg-main-api/pkg/scheduler"
)

// CronScheduler defines the interface for the cron scheduler interactions.
type CronScheduler interface {
	AddJob(job *entity.CronJob) error
	RemoveJob(id string)
	TriggerJob(job *entity.CronJob)
}

// CronJobUsecase defines the business logic for cron jobs.
type CronJobUsecase interface {
	CreateCronJob(ctx context.Context, job *entity.CronJob) error
	GetAllCronJobs(ctx context.Context) ([]*entity.CronJob, error)
	GetCronJobByID(ctx context.Context, id string) (*entity.CronJob, error)
	UpdateCronJob(ctx context.Context, job *entity.CronJob) error
	DeleteCronJob(ctx context.Context, id string) error
	TriggerCronJob(ctx context.Context, id string) error
}

type cronJobUsecase struct {
	repo      repository.CronJobRepository
	scheduler CronScheduler
}

// NewCronJobUsecase creates a new instance of CronJobUsecase.
func NewCronJobUsecase(repo repository.CronJobRepository, sched CronScheduler) CronJobUsecase {
	return &cronJobUsecase{repo: repo, scheduler: sched}
}

func (u *cronJobUsecase) CreateCronJob(ctx context.Context, job *entity.CronJob) error {
	// Validate cron expression.
	if err := scheduler.ValidateCronExpr(job.CronExpr); err != nil {
		return errors.New("invalid cron expression: " + err.Error())
	}

	if err := u.repo.Create(ctx, job); err != nil {
		return err
	}

	// Register in scheduler if enabled.
	if job.Enabled {
		if err := u.scheduler.AddJob(job); err != nil {
			log.Printf("[cronjob] failed to register job %s in scheduler: %v", job.ID, err)
		}
	}

	return nil
}

func (u *cronJobUsecase) GetAllCronJobs(ctx context.Context) ([]*entity.CronJob, error) {
	return u.repo.GetAll(ctx)
}

func (u *cronJobUsecase) GetCronJobByID(ctx context.Context, id string) (*entity.CronJob, error) {
	return u.repo.GetByID(ctx, id)
}

func (u *cronJobUsecase) UpdateCronJob(ctx context.Context, job *entity.CronJob) error {
	// Validate cron expression.
	if err := scheduler.ValidateCronExpr(job.CronExpr); err != nil {
		return errors.New("invalid cron expression: " + err.Error())
	}

	if err := u.repo.Update(ctx, job); err != nil {
		return err
	}

	// Update scheduler: add (enabled) or remove (disabled).
	if err := u.scheduler.AddJob(job); err != nil {
		log.Printf("[cronjob] failed to update job %s in scheduler: %v", job.ID, err)
	}

	return nil
}

func (u *cronJobUsecase) DeleteCronJob(ctx context.Context, id string) error {
	if err := u.repo.Delete(ctx, id); err != nil {
		return err
	}

	// Remove from scheduler.
	u.scheduler.RemoveJob(id)
	return nil
}

func (u *cronJobUsecase) TriggerCronJob(ctx context.Context, id string) error {
	job, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if job == nil {
		return errors.New("cron job not found")
	}

	u.scheduler.TriggerJob(job)
	return nil
}
