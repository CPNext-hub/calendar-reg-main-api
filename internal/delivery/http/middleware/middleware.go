package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	fiberlogger "github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

// SetupMiddlewares registers global middlewares on the Fiber app.
func SetupMiddlewares(app *fiber.App) {
	app.Use(recover.New())
	app.Use(fiberlogger.New())
	app.Use(cors.New())
}
