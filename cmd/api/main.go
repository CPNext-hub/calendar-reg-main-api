package main

import (
	"github.com/CPNext-hub/calendar-reg-main-api/internal/config"
	"github.com/CPNext-hub/calendar-reg-main-api/internal/delivery/http/server"
)

func main() {
	cfg := config.Load()
	server.Start(cfg)
}
