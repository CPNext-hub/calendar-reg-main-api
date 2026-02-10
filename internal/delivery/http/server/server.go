package server

import (
	"fmt"
	"log"

	"github.com/CPNext-hub/calendar-reg-main-api/internal/config"
	"github.com/CPNext-hub/calendar-reg-main-api/internal/delivery/http/handler"
	"github.com/CPNext-hub/calendar-reg-main-api/internal/delivery/http/middleware"
	"github.com/CPNext-hub/calendar-reg-main-api/internal/delivery/http/router"
	"github.com/CPNext-hub/calendar-reg-main-api/internal/domain/usecase"
	"github.com/gofiber/fiber/v2"
)

// Start initialises dependencies and starts the HTTP server.
func Start(cfg *config.Config) {
	app := fiber.New(fiber.Config{
		AppName: cfg.AppName,
	})

	// middlewares
	middleware.SetupMiddlewares(app)

	// usecases
	healthUC := usecase.NewHealthUsecase()
	versionUC := usecase.NewVersionUsecase(cfg.AppName, cfg.AppVersion, cfg.AppEnv)

	// handlers
	healthHandler := handler.NewHealthHandler(healthUC)
	versionHandler := handler.NewVersionHandler(versionUC)

	// routes
	router.SetupRoutes(app, healthHandler, versionHandler)

	// start
	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("Server starting on %s (env=%s)", addr, cfg.AppEnv)
	if err := app.Listen(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
