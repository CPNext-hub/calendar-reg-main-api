package main

import (
	"log"

	"github.com/CPNext-hub/calendar-reg-main-api/internal/config"
	"github.com/CPNext-hub/calendar-reg-main-api/internal/delivery/http/server"
)

// @title Calendar Reg Main API
// @version 1.0
// @description This is the main API server for the Calendar Reg application.

// @contact.name API Support
// @contact.email [EMAIL_ADDRESS]

// @host localhost:8080
// @BasePath /api/v1
func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}
	server.Start(cfg)
}
