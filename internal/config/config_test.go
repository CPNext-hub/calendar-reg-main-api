package config

import (
	"os"
	"testing"
)

// setAllEnvVars sets every required env var to a dummy value and returns a cleanup function.
func setAllEnvVars(t *testing.T) {
	t.Helper()
	for _, key := range requiredEnvVars {
		t.Setenv(key, "test-value")
	}
}

func TestLoad_Production_AllSet(t *testing.T) {
	setAllEnvVars(t)
	t.Setenv("APP_ENV", "production")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if cfg.AppEnv != "production" {
		t.Errorf("expected AppEnv=production, got %s", cfg.AppEnv)
	}
}

func TestLoad_Production_MissingVars(t *testing.T) {
	// Clear all env vars to ensure missing keys
	for _, key := range requiredEnvVars {
		os.Unsetenv(key)
	}
	t.Setenv("APP_ENV", "production")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for missing env vars in production, got nil")
	}

	// Check that the error message mentions at least one missing key
	for _, key := range []string{"MONGO_HOST", "JWT_SECRET", "SUPER_ADMIN_USER"} {
		if !contains(err.Error(), key) {
			t.Errorf("expected error to mention %q, got: %v", key, err)
		}
	}
}

func TestLoad_Development_UsesDefaults(t *testing.T) {
	// Clear all env vars
	for _, key := range requiredEnvVars {
		os.Unsetenv(key)
	}
	t.Setenv("APP_ENV", "development")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected no error in development, got: %v", err)
	}
	if cfg.AppName != "calendar-reg-main-api" {
		t.Errorf("expected default AppName, got %s", cfg.AppName)
	}
	if cfg.Port != "8080" {
		t.Errorf("expected default Port=8080, got %s", cfg.Port)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchSubstring(s, substr)
}

func searchSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
