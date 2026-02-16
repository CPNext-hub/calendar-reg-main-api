package server

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/CPNext-hub/calendar-reg-main-api/internal/config"
	"github.com/CPNext-hub/calendar-reg-main-api/internal/delivery/http/handler"
	"github.com/CPNext-hub/calendar-reg-main-api/internal/delivery/http/middleware"
	"github.com/CPNext-hub/calendar-reg-main-api/internal/delivery/http/router"
	"github.com/CPNext-hub/calendar-reg-main-api/internal/domain/usecase"
	"github.com/CPNext-hub/calendar-reg-main-api/internal/infrastructure/mongodb"
	mongoRepo "github.com/CPNext-hub/calendar-reg-main-api/internal/infrastructure/repository/mongodb"
	"github.com/gofiber/fiber/v2"
)

// Start initialises dependencies and starts the HTTP server.
func Start(cfg *config.Config) {
	ctx := context.Background()

	// ---------- MongoDB ----------
	mongo, err := mongodb.Connect(ctx, cfg.MongoHost, cfg.MongoDBName, cfg.MongoUser, cfg.MongoPassword)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}

	// ---------- Fiber ----------
	app := fiber.New(fiber.Config{
		AppName: cfg.AppName,
	})

	// middlewares
	middleware.SetupMiddlewares(app)

	// usecases
	healthUC := usecase.NewHealthUsecase()
	versionUC := usecase.NewVersionUsecase(cfg.AppName, cfg.AppVersion, cfg.AppEnv)

	// repositories
	courseRepo := mongoRepo.NewCourseRepository(mongo.Database())
	courseUC := usecase.NewCourseUsecase(courseRepo)

	userRepo := mongoRepo.NewUserRepository(mongo.Database())
	authUC := usecase.NewAuthUsecase(userRepo, cfg.JWTSecret)

	// ---------- Seed superadmin ----------
	authUC.SeedSuperAdmin(ctx, cfg.SuperAdminUser, cfg.SuperAdminPass)

	// handlers
	h := &router.Handlers{
		Health:    handler.NewHealthHandler(healthUC),
		Version:   handler.NewVersionHandler(versionUC),
		MongoTest: handler.NewMongoTestHandler(mongo),
		Course:    handler.NewCourseHandler(courseUC),
		Auth:      handler.NewAuthHandler(authUC),
	}

	// routes
	router.SetupRoutes(app, h, cfg.JWTSecret)

	// ---------- Graceful Shutdown ----------
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		addr := fmt.Sprintf(":%s", cfg.Port)
		log.Printf("Server starting on %s (env=%s)", addr, cfg.AppEnv)
		if err := app.Listen(addr); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// wait for interrupt signal
	sig := <-quit
	log.Printf("Received signal %s, shutting down...", sig)

	// shutdown Fiber
	if err := app.Shutdown(); err != nil {
		log.Printf("Fiber shutdown error: %v", err)
	}

	// disconnect MongoDB
	if err := mongo.Disconnect(ctx); err != nil {
		log.Printf("MongoDB disconnect error: %v", err)
	}

	log.Println("Server stopped gracefully")

	// mongo variable is available here for future use with repositories:
	// _ = mongo.Database()
}
