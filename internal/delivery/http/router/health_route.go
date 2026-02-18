package router

import (
	"github.com/CPNext-hub/calendar-reg-main-api/internal/delivery/http/handler"
	"github.com/gofiber/fiber/v2"
)

// RegisterHealthRoutes registers health and version routes.
func RegisterHealthRoutes(api fiber.Router, healthH *handler.HealthHandler, versionH *handler.VersionHandler) {
	api.Get("/status", healthH.GetStatus)
	api.Get("/version", versionH.GetVersion)
}
