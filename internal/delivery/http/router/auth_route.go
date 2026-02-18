package router

import (
	"github.com/CPNext-hub/calendar-reg-main-api/internal/delivery/http/handler"
	"github.com/CPNext-hub/calendar-reg-main-api/internal/delivery/http/middleware"
	"github.com/CPNext-hub/calendar-reg-main-api/pkg/constants"
	"github.com/gofiber/fiber/v2"
)

// RegisterAuthRoutes registers authentication routes.
func RegisterAuthRoutes(api fiber.Router, authH *handler.AuthHandler, jwtSecret string) {
	auth := api.Group("/auth")
	auth.Post("/register", authH.Register)
	auth.Post("/login", authH.Login)

	// Protected: create/list users (superadmin/admin only)
	auth.Post("/users", middleware.JWTAuth(jwtSecret), middleware.RequireRole(constants.RoleSuperAdmin, constants.RoleAdmin), authH.CreateUser)
	auth.Get("/users", middleware.JWTAuth(jwtSecret), middleware.RequireRole(constants.RoleSuperAdmin, constants.RoleAdmin), authH.GetUsers)
}
