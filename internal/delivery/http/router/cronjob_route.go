package router

import (
	"github.com/CPNext-hub/calendar-reg-main-api/internal/delivery/http/handler"
	"github.com/CPNext-hub/calendar-reg-main-api/internal/delivery/http/middleware"
	"github.com/CPNext-hub/calendar-reg-main-api/pkg/constants"
	"github.com/gofiber/fiber/v2"
)

// RegisterCronJobRoutes registers cron job routes (admin only).
func RegisterCronJobRoutes(api fiber.Router, cronJobH *handler.CronJobHandler, jwtSecret string) {
	cronjobs := api.Group("/cronjobs", middleware.JWTAuth(jwtSecret), middleware.RequireRole(constants.RoleSuperAdmin, constants.RoleAdmin))
	cronjobs.Post("/", cronJobH.CreateCronJob)
	cronjobs.Get("/", cronJobH.GetCronJobs)
	cronjobs.Get("/:id", cronJobH.GetCronJob)
	cronjobs.Put("/:id", cronJobH.UpdateCronJob)
	cronjobs.Delete("/:id", cronJobH.DeleteCronJob)
	cronjobs.Post("/:id/trigger", cronJobH.TriggerCronJob)
}
