package router

import (
	_ "github.com/CPNext-hub/calendar-reg-main-api/docs" // load generated docs
	"github.com/CPNext-hub/calendar-reg-main-api/internal/delivery/http/handler"
	"github.com/gofiber/fiber/v2"
	fiberSwagger "github.com/swaggo/fiber-swagger"
)

// Handlers holds all HTTP handlers for route registration.
type Handlers struct {
	Health    *handler.HealthHandler
	Version   *handler.VersionHandler
	MongoTest *handler.MongoTestHandler
}

// SetupRoutes registers all routes on the Fiber app.
func SetupRoutes(app *fiber.App, h *Handlers) {
	api := app.Group("/api/v1")

	api.Get("/status", h.Health.GetStatus)
	api.Get("/version", h.Version.GetVersion)

	// Swagger
	app.Get("/swagger/*", fiberSwagger.WrapHandler)

	// ---------- MongoDB test routes ----------
	test := api.Group("/test/mongo")
	test.Get("/ping", h.MongoTest.Ping)
	test.Post("/insert", h.MongoTest.InsertTest)
	test.Get("/find", h.MongoTest.FindAll)
	test.Delete("/delete", h.MongoTest.DeleteAll)
	test.Get("/full", h.MongoTest.FullTest)
}
