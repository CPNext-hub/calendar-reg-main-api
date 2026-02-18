package router

import (
	_ "github.com/CPNext-hub/calendar-reg-main-api/docs" // load generated docs
	"github.com/gofiber/fiber/v2"
	fiberSwagger "github.com/swaggo/fiber-swagger"
)

// SetupRoutes registers global routes (swagger, etc.) on the Fiber app.
// Domain-specific routes are registered by each module's own Register* function.
func SetupRoutes(app *fiber.App) {
	// Swagger
	app.Get("/swagger/*", fiberSwagger.WrapHandler)
}
