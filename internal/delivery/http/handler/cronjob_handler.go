package handler

import (
	"context"
	"time"

	"github.com/CPNext-hub/calendar-reg-main-api/internal/delivery/http/adapter"
	"github.com/CPNext-hub/calendar-reg-main-api/internal/delivery/http/dto"
	"github.com/CPNext-hub/calendar-reg-main-api/internal/domain/usecase"
	"github.com/CPNext-hub/calendar-reg-main-api/pkg/response"
	"github.com/gofiber/fiber/v2"
)

// CronJobHandler handles HTTP requests for cron jobs.
type CronJobHandler struct {
	usecase usecase.CronJobUsecase
}

// NewCronJobHandler creates a new CronJobHandler instance.
func NewCronJobHandler(uc usecase.CronJobUsecase) *CronJobHandler {
	return &CronJobHandler{usecase: uc}
}

// CreateCronJob creates a new cron job.
// @Summary Create a new cron job
// @Description Create a new cron job for scheduled course data refresh
// @Tags cronjobs
// @Accept json
// @Produce json
// @Param request body dto.CreateCronJobRequest true "CronJob Request"
// @Success 201 {object} dto.CronJobResponse
// @Failure 400 {object} interface{}
// @Failure 500 {object} interface{}
// @Router /cronjobs [post]
func (h *CronJobHandler) CreateCronJob(c *fiber.Ctx) error {
	var req dto.CreateCronJobRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(adapter.NewFiberResponder(c), "Invalid request body")
	}

	if req.Name == "" || len(req.CourseCodes) == 0 || req.CronExpr == "" {
		return response.BadRequest(adapter.NewFiberResponder(c), "Missing required fields: name, course_codes, cron_expr")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	job := req.ToEntity()
	if err := h.usecase.CreateCronJob(ctx, job); err != nil {
		return response.InternalError(adapter.NewFiberResponder(c), err.Error())
	}

	return response.Created(adapter.NewFiberResponder(c), dto.ToCronJobResponse(job))
}

// GetCronJobs retrieves all cron jobs.
// @Summary Get all cron jobs
// @Description Retrieve all cron jobs
// @Tags cronjobs
// @Produce json
// @Success 200 {array} dto.CronJobResponse
// @Failure 500 {object} interface{}
// @Router /cronjobs [get]
func (h *CronJobHandler) GetCronJobs(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	jobs, err := h.usecase.GetAllCronJobs(ctx)
	if err != nil {
		return response.InternalError(adapter.NewFiberResponder(c), err.Error())
	}

	return response.OK(adapter.NewFiberResponder(c), dto.ToCronJobResponses(jobs))
}

// GetCronJob retrieves a cron job by ID.
// @Summary Get cron job by ID
// @Description Retrieve a specific cron job by its ID
// @Tags cronjobs
// @Produce json
// @Param id path string true "CronJob ID"
// @Success 200 {object} dto.CronJobResponse
// @Failure 404 {object} interface{}
// @Failure 500 {object} interface{}
// @Router /cronjobs/{id} [get]
func (h *CronJobHandler) GetCronJob(c *fiber.Ctx) error {
	id := c.Params("id")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	job, err := h.usecase.GetCronJobByID(ctx, id)
	if err != nil {
		return response.InternalError(adapter.NewFiberResponder(c), err.Error())
	}
	if job == nil {
		return response.NotFound(adapter.NewFiberResponder(c), "Cron job not found")
	}

	return response.OK(adapter.NewFiberResponder(c), dto.ToCronJobResponse(job))
}

// UpdateCronJob updates an existing cron job.
// @Summary Update cron job
// @Description Update an existing cron job by ID
// @Tags cronjobs
// @Accept json
// @Produce json
// @Param id path string true "CronJob ID"
// @Param request body dto.UpdateCronJobRequest true "CronJob Update Request"
// @Success 200 {object} dto.CronJobResponse
// @Failure 400 {object} interface{}
// @Failure 500 {object} interface{}
// @Router /cronjobs/{id} [put]
func (h *CronJobHandler) UpdateCronJob(c *fiber.Ctx) error {
	id := c.Params("id")

	var req dto.UpdateCronJobRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(adapter.NewFiberResponder(c), "Invalid request body")
	}

	if req.Name == "" || len(req.CourseCodes) == 0 || req.CronExpr == "" {
		return response.BadRequest(adapter.NewFiberResponder(c), "Missing required fields: name, course_codes, cron_expr")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	job := req.ToEntity(id)
	if err := h.usecase.UpdateCronJob(ctx, job); err != nil {
		return response.InternalError(adapter.NewFiberResponder(c), err.Error())
	}

	return response.OK(adapter.NewFiberResponder(c), dto.ToCronJobResponse(job))
}

// DeleteCronJob deletes a cron job.
// @Summary Delete cron job
// @Description Soft delete a cron job by ID
// @Tags cronjobs
// @Param id path string true "CronJob ID"
// @Success 200 {object} interface{}
// @Failure 500 {object} interface{}
// @Router /cronjobs/{id} [delete]
func (h *CronJobHandler) DeleteCronJob(c *fiber.Ctx) error {
	id := c.Params("id")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := h.usecase.DeleteCronJob(ctx, id); err != nil {
		return response.InternalError(adapter.NewFiberResponder(c), err.Error())
	}

	return response.OK(adapter.NewFiberResponder(c), map[string]string{"message": "Cron job deleted"})
}

// TriggerCronJob manually triggers a cron job immediately.
// @Summary Trigger cron job
// @Description Manually trigger a cron job to run immediately
// @Tags cronjobs
// @Param id path string true "CronJob ID"
// @Success 200 {object} interface{}
// @Failure 404 {object} interface{}
// @Failure 500 {object} interface{}
// @Router /cronjobs/{id}/trigger [post]
func (h *CronJobHandler) TriggerCronJob(c *fiber.Ctx) error {
	id := c.Params("id")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := h.usecase.TriggerCronJob(ctx, id); err != nil {
		return response.InternalError(adapter.NewFiberResponder(c), err.Error())
	}

	return response.OK(adapter.NewFiberResponder(c), map[string]string{"message": "Cron job triggered"})
}
