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
	"github.com/CPNext-hub/calendar-reg-main-api/internal/infrastructure/externalapi"
	"github.com/CPNext-hub/calendar-reg-main-api/internal/infrastructure/mongodb"
	mongoRepo "github.com/CPNext-hub/calendar-reg-main-api/internal/infrastructure/repository/mongodb"
	"github.com/CPNext-hub/calendar-reg-main-api/pkg/queue"
	"github.com/CPNext-hub/calendar-reg-main-api/pkg/scheduler"
	"github.com/gofiber/fiber/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Start initialises dependencies and starts the HTTP server.
func Start(cfg *config.Config) {
	ctx := context.Background()

	// ========== Infrastructure ==========

	// ---------- MongoDB ----------
	mongo, err := mongodb.Connect(ctx, cfg.MongoHost, cfg.MongoDBName, cfg.MongoUser, cfg.MongoPassword)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}

	// ---------- gRPC ----------
	grpcConn, err := grpc.NewClient(cfg.CourseGRPCAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("Failed to connect to course gRPC service at %s: %v", cfg.CourseGRPCAddr, err)
	}
	log.Printf("gRPC client connected to %s", cfg.CourseGRPCAddr)

	// ---------- Background Queue ----------
	refreshQueue := queue.New(100, 5)

	// ========== Fiber ==========

	app := fiber.New(fiber.Config{
		AppName: cfg.AppName,
	})
	middleware.SetupMiddlewares(app)
	api := app.Group("/api/v1")

	// ========== Module: Health & Version ==========

	healthUC := usecase.NewHealthUsecase()
	versionUC := usecase.NewVersionUsecase(cfg.AppName, cfg.AppVersion, cfg.AppEnv)
	healthH := handler.NewHealthHandler(healthUC)
	versionH := handler.NewVersionHandler(versionUC)
	router.RegisterHealthRoutes(api, healthH, versionH)

	// ========== Module: Auth ==========

	userRepo := mongoRepo.NewUserRepository(mongo.Database())
	authUC := usecase.NewAuthUsecase(userRepo, cfg.JWTSecret)
	authH := handler.NewAuthHandler(authUC)
	router.RegisterAuthRoutes(api, authH, cfg.JWTSecret)

	authUC.SeedSuperAdmin(ctx, cfg.SuperAdminUser, cfg.SuperAdminPass)

	// ========== Module: Course ==========

	courseRepo := mongoRepo.NewCourseRepository(mongo.Database())
	courseExtAPI := externalapi.NewCourseExternalAPI(grpcConn)
	courseUC := usecase.NewCourseUsecase(courseRepo, courseExtAPI, refreshQueue)
	courseH := handler.NewCourseHandler(courseUC)
	queueH := handler.NewQueueHandler(refreshQueue)
	router.RegisterCourseRoutes(api, courseH, queueH, cfg.JWTSecret)

	refreshQueue.Start(courseUC.ProcessRefreshJob)

	// ========== Module: CronJob ==========

	cronJobRepo := mongoRepo.NewCronJobRepository(mongo.Database())
	cronScheduler := scheduler.New(refreshQueue)
	cronJobUC := usecase.NewCronJobUsecase(cronJobRepo, cronScheduler)
	cronJobH := handler.NewCronJobHandler(cronJobUC)
	router.RegisterCronJobRoutes(api, cronJobH, cfg.JWTSecret)

	enabledJobs, err := cronJobRepo.GetEnabled(ctx)
	if err != nil {
		log.Printf("Failed to load cron jobs: %v", err)
	} else {
		cronScheduler.LoadJobs(enabledJobs)
	}
	cronScheduler.Start()

	// ========== Module: Test (MongoDB) ==========

	mongoTestH := handler.NewMongoTestHandler(mongo)
	router.RegisterTestRoutes(api, mongoTestH)

	// ========== Global Routes (Swagger) ==========

	router.SetupRoutes(app)

	// ========== Graceful Shutdown ==========

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

	// stop cron scheduler
	cronScheduler.Stop()

	// stop background worker (drain remaining jobs)
	refreshQueue.Stop()

	// shutdown Fiber
	if err := app.Shutdown(); err != nil {
		log.Printf("Fiber shutdown error: %v", err)
	}

	// close gRPC connection
	if err := grpcConn.Close(); err != nil {
		log.Printf("gRPC connection close error: %v", err)
	}

	// disconnect MongoDB
	if err := mongo.Disconnect(ctx); err != nil {
		log.Printf("MongoDB disconnect error: %v", err)
	}

	log.Println("Server stopped gracefully")
}
