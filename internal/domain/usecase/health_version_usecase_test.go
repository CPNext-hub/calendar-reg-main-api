package usecase

import (
	"strings"
	"testing"
	"time"
)

func TestHealthUsecase_GetStatus(t *testing.T) {
	uc := NewHealthUsecase()
	before := time.Now().UTC()

	status := uc.GetStatus()

	if status.Status != "ok" {
		t.Errorf("expected status='ok', got %q", status.Status)
	}

	// Verify timestamp is valid RFC3339 and recent
	ts, err := time.Parse(time.RFC3339, status.Timestamp)
	if err != nil {
		t.Fatalf("invalid timestamp format: %v", err)
	}
	if ts.Before(before.Add(-1 * time.Second)) {
		t.Error("timestamp is too far in the past")
	}
}

func TestVersionUsecase_GetVersion(t *testing.T) {
	uc := NewVersionUsecase("my-app", "1.2.3", "production")

	info := uc.GetVersion()

	if info.Name != "my-app" {
		t.Errorf("expected name='my-app', got %q", info.Name)
	}
	if info.Version != "1.2.3" {
		t.Errorf("expected version='1.2.3', got %q", info.Version)
	}
	if info.Env != "production" {
		t.Errorf("expected env='production', got %q", info.Env)
	}
}

func TestVersionUsecase_EmptyValues(t *testing.T) {
	uc := NewVersionUsecase("", "", "")

	info := uc.GetVersion()

	if info.Name != "" || info.Version != "" || info.Env != "" {
		t.Error("expected all empty strings")
	}
}

func TestHealthUsecase_TimestampFormat(t *testing.T) {
	uc := NewHealthUsecase()
	status := uc.GetStatus()

	// Should contain "T" separator (RFC3339)
	if !strings.Contains(status.Timestamp, "T") {
		t.Errorf("expected RFC3339 timestamp, got %q", status.Timestamp)
	}
}
