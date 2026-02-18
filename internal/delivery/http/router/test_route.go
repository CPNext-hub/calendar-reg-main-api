package router

import (
	"github.com/CPNext-hub/calendar-reg-main-api/internal/delivery/http/handler"
	"github.com/gofiber/fiber/v2"
)

// RegisterTestRoutes registers MongoDB test routes.
func RegisterTestRoutes(api fiber.Router, mongoTestH *handler.MongoTestHandler) {
	test := api.Group("/test/mongo")
	test.Get("/ping", mongoTestH.Ping)
	test.Post("/insert", mongoTestH.InsertTest)
	test.Get("/find", mongoTestH.FindAll)
	test.Delete("/delete", mongoTestH.DeleteAll)
	test.Get("/full", mongoTestH.FullTest)
}
