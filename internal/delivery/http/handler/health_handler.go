package handler

import (
	"github.com/CPNext-hub/calendar-reg-main-api/internal/delivery/http/adapter"
	"github.com/CPNext-hub/calendar-reg-main-api/internal/delivery/http/dto"
	"github.com/CPNext-hub/calendar-reg-main-api/internal/domain/usecase"
	"github.com/CPNext-hub/calendar-reg-main-api/pkg/response"
	"github.com/gofiber/fiber/v2"
)

// HealthHandler handles health/status related HTTP requests.
type HealthHandler struct {
	usecase usecase.HealthUsecase
}

// NewHealthHandler creates a new HealthHandler instance.
func NewHealthHandler(uc usecase.HealthUsecase) *HealthHandler {
	return &HealthHandler{usecase: uc}
}

// GetStatus returns the current health status of the service.
// @Summary Check service health
// @Description Get the current health status of the service
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} dto.HealthResponse
// @Router /status [get]
func (h *HealthHandler) GetStatus(c *fiber.Ctx) error {
	status := h.usecase.GetStatus()
	res := dto.ToHealthResponse(status)
	return response.OK(adapter.NewFiberResponder(c), res)
}
