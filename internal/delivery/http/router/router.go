package router

import (
	"github.com/CPNext-hub/calendar-reg-main-api/internal/delivery/http/handler"
	"github.com/gofiber/fiber/v2"
)

// SetupRoutes registers all routes on the Fiber app.
func SetupRoutes(app *fiber.App, healthHandler *handler.HealthHandler, versionHandler *handler.VersionHandler) {
	api := app.Group("/api/v1")

	api.Get("/status", healthHandler.GetStatus)
	api.Get("/version", versionHandler.GetVersion)
}
