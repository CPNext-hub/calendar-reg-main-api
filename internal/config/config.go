package config

import "os"

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
}

// Load reads configuration from environment variables with sensible defaults.
func Load() *Config {
	return &Config{
		AppName:    getEnv("APP_NAME", "calendar-reg-main-api"),
		AppVersion: getEnv("APP_VERSION", "0.1.0"),
		AppEnv:     getEnv("APP_ENV", "development"),
		Port:       getEnv("PORT", "8080"),

		MongoHost:     getEnv("MONGO_HOST", "localhost:27017"),
		MongoDBName:   getEnv("MONGO_DB_NAME", "calendar-reg"),
		MongoUser:     getEnv("MONGO_INITDB_ROOT_USERNAME", ""),
		MongoPassword: getEnv("MONGO_INITDB_ROOT_PASSWORD", ""),
	}
}

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return fallback
}
