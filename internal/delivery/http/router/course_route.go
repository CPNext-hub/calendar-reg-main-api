package router

import (
	"github.com/CPNext-hub/calendar-reg-main-api/internal/delivery/http/handler"
	"github.com/CPNext-hub/calendar-reg-main-api/internal/delivery/http/middleware"
	"github.com/CPNext-hub/calendar-reg-main-api/pkg/constants"
	"github.com/gofiber/fiber/v2"
)

// RegisterCourseRoutes registers course routes.
func RegisterCourseRoutes(api fiber.Router, courseH *handler.CourseHandler, queueH *handler.QueueHandler, jwtSecret string) {
	courses := api.Group("/courses")

	// Public: read-only
	courses.Get("/", courseH.GetCourses)
	courses.Get("/:code", courseH.GetCourse)

	// Protected: superadmin and admin
	adminCourses := courses.Group("", middleware.JWTAuth(jwtSecret), middleware.RequireRole(constants.RoleSuperAdmin, constants.RoleAdmin))
	adminCourses.Post("/", courseH.CreateCourse)
	adminCourses.Delete("/:code", courseH.DeleteCourse)

	// Queue status
	api.Get("/queue/status", queueH.GetStatus)
}
