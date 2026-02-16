package router

import (
	_ "github.com/CPNext-hub/calendar-reg-main-api/docs" // load generated docs
	"github.com/CPNext-hub/calendar-reg-main-api/internal/delivery/http/handler"
	"github.com/CPNext-hub/calendar-reg-main-api/internal/delivery/http/middleware"
	"github.com/CPNext-hub/calendar-reg-main-api/pkg/constants"
	"github.com/gofiber/fiber/v2"
	fiberSwagger "github.com/swaggo/fiber-swagger"
)

// Handlers holds all HTTP handlers for route registration.
type Handlers struct {
	Health    *handler.HealthHandler
	Version   *handler.VersionHandler
	MongoTest *handler.MongoTestHandler
	Course    *handler.CourseHandler
	Auth      *handler.AuthHandler
}

// SetupRoutes registers all routes on the Fiber app.
func SetupRoutes(app *fiber.App, h *Handlers, jwtSecret string) {
	api := app.Group("/api/v1")

	api.Get("/status", h.Health.GetStatus)
	api.Get("/version", h.Version.GetVersion)

	// ---------- Auth routes (public) ----------
	auth := api.Group("/auth")
	auth.Post("/register", h.Auth.Register)
	auth.Post("/login", h.Auth.Login)
	// Protected: create user with any role (superadmin/admin only)
	auth.Post("/users", middleware.JWTAuth(jwtSecret), middleware.RequireRole(constants.RoleSuperAdmin, constants.RoleAdmin), h.Auth.CreateUser)

	// ---------- Course routes ----------
	courses := api.Group("/courses")
	// Public: read-only
	courses.Get("/", h.Course.GetCourses)
	courses.Get("/:code", h.Course.GetCourse)
	// Protected: superadmin and admin
	adminCourses := courses.Group("", middleware.JWTAuth(jwtSecret), middleware.RequireRole(constants.RoleSuperAdmin, constants.RoleAdmin))
	adminCourses.Post("/", h.Course.CreateCourse)
	adminCourses.Delete("/:code", h.Course.DeleteCourse)

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
