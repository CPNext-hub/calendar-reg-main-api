package handler

import (
	"github.com/CPNext-hub/calendar-reg-main-api/internal/delivery/http/adapter"
	"github.com/CPNext-hub/calendar-reg-main-api/internal/delivery/http/dto"
	"github.com/CPNext-hub/calendar-reg-main-api/internal/domain/usecase"
	"github.com/CPNext-hub/calendar-reg-main-api/pkg/response"
	"github.com/gofiber/fiber/v2"
)

// VersionHandler handles version related HTTP requests.
type VersionHandler struct {
	usecase usecase.VersionUsecase
}

// NewVersionHandler creates a new VersionHandler instance.
func NewVersionHandler(uc usecase.VersionUsecase) *VersionHandler {
	return &VersionHandler{usecase: uc}
}

// GetVersion returns the application name, version, and environment.
func (h *VersionHandler) GetVersion(c *fiber.Ctx) error {
	info := h.usecase.GetVersion()
	res := dto.ToVersionResponse(info)
	return response.OK(adapter.NewFiberResponder(c), res)
}
