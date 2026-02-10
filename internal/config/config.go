package config

import "os"

// Config holds the application configuration.
type Config struct {
	AppName    string
	AppVersion string
	AppEnv     string
	Port       string
}

// Load reads configuration from environment variables with sensible defaults.
func Load() *Config {
	return &Config{
		AppName:    getEnv("APP_NAME", "calendar-reg-main-api"),
		AppVersion: getEnv("APP_VERSION", "0.1.0"),
		AppEnv:     getEnv("APP_ENV", "development"),
		Port:       getEnv("PORT", "8080"),
	}
}

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return fallback
}
