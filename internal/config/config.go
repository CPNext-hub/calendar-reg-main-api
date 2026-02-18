package config

import (
	"fmt"
	"os"
	"strings"
)

// Config holds the application configuration.
type Config struct {
	AppName    string
	AppVersion string
	AppEnv     string
	Port       string

	// MongoDB — host/db จาก ConfigMap, credentials จาก Secret
	MongoHost     string
	MongoDBName   string
	MongoUser     string
	MongoPassword string

	// JWT
	JWTSecret string

	// Superadmin seed
	SuperAdminUser string
	SuperAdminPass string

	// External Course gRPC
	CourseGRPCAddr string
}

// requiredEnvVars lists every environment variable that must be set in production.
var requiredEnvVars = []string{
	"APP_NAME", "APP_VERSION", "APP_ENV", "PORT",
	"MONGO_HOST", "MONGO_DB_NAME",
	"MONGO_INITDB_ROOT_USERNAME", "MONGO_INITDB_ROOT_PASSWORD",
	"JWT_SECRET",
	"SUPER_ADMIN_USER", "SUPER_ADMIN_PASS",
}

// Load reads configuration from environment variables.
// In production all variables must be explicitly set; missing ones cause an error.
func Load() (*Config, error) {
	appEnv := getEnv("APP_ENV", "development")

	if appEnv == "production" {
		if err := validateProduction(); err != nil {
			return nil, err
		}
	}

	return &Config{
		AppName:    getEnv("APP_NAME", "calendar-reg-main-api"),
		AppVersion: getEnv("APP_VERSION", "0.1.0"),
		AppEnv:     appEnv,
		Port:       getEnv("PORT", "8080"),

		MongoHost:     getEnv("MONGO_HOST", "localhost:27017"),
		MongoDBName:   getEnv("MONGO_DB_NAME", "calendar-reg"),
		MongoUser:     getEnv("MONGO_INITDB_ROOT_USERNAME", ""),
		MongoPassword: getEnv("MONGO_INITDB_ROOT_PASSWORD", ""),

		JWTSecret: getEnv("JWT_SECRET", "change-me"),

		SuperAdminUser: getEnv("SUPER_ADMIN_USER", "superadmin"),
		SuperAdminPass: getEnv("SUPER_ADMIN_PASS", "superadmin123"),

		CourseGRPCAddr: getEnv("COURSE_GRPC_ADDR", "localhost:50051"),
	}, nil
}

// validateProduction checks that every required environment variable is set.
func validateProduction() error {
	var missing []string
	for _, key := range requiredEnvVars {
		if _, ok := os.LookupEnv(key); !ok {
			missing = append(missing, key)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("production mode: missing required environment variables: %s", strings.Join(missing, ", "))
	}
	return nil
}

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return fallback
}
