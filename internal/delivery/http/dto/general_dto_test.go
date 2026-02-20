package dto

import (
	"testing"
	"time"

	"github.com/CPNext-hub/calendar-reg-main-api/internal/domain/entity"
	"github.com/stretchr/testify/assert"
)

// --- CronJob DTO Tests ---

func TestCreateCronJobRequest_ToEntity(t *testing.T) {
	req := CreateCronJobRequest{
		Name:        "Test Job",
		CourseCodes: []string{"A", "B"},
		Acadyear:    2024,
		Semester:    1,
		CronExpr:    "* * * * *",
		Enabled:     true,
	}
	e := req.ToEntity()
	assert.Equal(t, "Test Job", e.Name)
	assert.Equal(t, []string{"A", "B"}, e.CourseCodes)
	assert.Equal(t, 2024, e.Acadyear)
	assert.Equal(t, 1, e.Semester)
	assert.Equal(t, "* * * * *", e.CronExpr)
	assert.True(t, e.Enabled)
}

func TestUpdateCronJobRequest_ToEntity(t *testing.T) {
	req := UpdateCronJobRequest{
		Name:        "Updated Job",
		CourseCodes: []string{"C"},
		Acadyear:    2025,
		Semester:    2,
		CronExpr:    "0 * * * *",
		Enabled:     false,
	}
	e := req.ToEntity("job-123")
	assert.Equal(t, "job-123", e.ID)
	assert.Equal(t, "Updated Job", e.Name)
	assert.Equal(t, []string{"C"}, e.CourseCodes)
	assert.Equal(t, 2025, e.Acadyear)
	assert.Equal(t, 2, e.Semester)
	assert.Equal(t, "0 * * * *", e.CronExpr)
	assert.False(t, e.Enabled)
}

func TestToCronJobResponse(t *testing.T) {
	now := time.Now()
	e := &entity.CronJob{
		BaseEntity:  entity.BaseEntity{ID: "job-1", CreatedAt: now, UpdatedAt: now},
		Name:        "Job 1",
		CourseCodes: []string{"A"},
		Acadyear:    2024,
		Semester:    1,
		CronExpr:    "* * * * *",
		Enabled:     true,
	}
	res := ToCronJobResponse(e)
	assert.Equal(t, "job-1", res.ID)
	assert.Equal(t, "Job 1", res.Name)
	assert.Equal(t, now.Format(time.RFC3339), res.CreatedAt)
	assert.Equal(t, now.Format(time.RFC3339), res.UpdatedAt)
}

func TestToCronJobResponse_Nil(t *testing.T) {
	assert.Nil(t, ToCronJobResponse(nil))
}

func TestToCronJobResponses(t *testing.T) {
	jobs := []*entity.CronJob{
		{Name: "J1"},
		{Name: "J2"},
	}
	res := ToCronJobResponses(jobs)
	assert.Len(t, res, 2)
	assert.Equal(t, "J1", res[0].Name)
	assert.Equal(t, "J2", res[1].Name)
}

// --- Health DTO Tests ---

func TestToHealthResponse(t *testing.T) {
	h := entity.Health{Status: "OK", Timestamp: "2024-01-01"}
	res := ToHealthResponse(h)
	assert.Equal(t, "OK", res.Status)
	assert.Equal(t, "2024-01-01", res.Timestamp)
}

// --- Version DTO Tests ---

func TestToVersionResponse(t *testing.T) {
	v := entity.AppInfo{Name: "App", Version: "1.0", Env: "dev"}
	res := ToVersionResponse(v)
	assert.Equal(t, "App", res.Name)
	assert.Equal(t, "1.0", res.Version)
	assert.Equal(t, "dev", res.Env)
}

// --- Mongo Test DTO Tests ---

func TestNewMongoPingResponse(t *testing.T) {
	latency := 100 * time.Millisecond
	res := NewMongoPingResponse("Connected", latency)
	assert.Equal(t, "Connected", res.Status)
	assert.Equal(t, "100ms", res.Latency)
	assert.NotEmpty(t, res.Timestamp)
}

// --- Auth DTO Tests ---

func TestToRegisterResponse(t *testing.T) {
	u := &entity.User{
		BaseEntity: entity.BaseEntity{ID: "u1"},
		Username:   "admin",
		Role:       "admin",
	}
	res := ToRegisterResponse(u)
	assert.Equal(t, "u1", res.ID)
	assert.Equal(t, "admin", res.Username)
	assert.Equal(t, "admin", res.Role)
}

func TestToUserResponse(t *testing.T) {
	u := &entity.User{
		BaseEntity: entity.BaseEntity{ID: "u1"},
		Username:   "user1",
		Role:       "user",
	}
	res := ToUserResponse(u)
	assert.Equal(t, "u1", res.ID)
	assert.Equal(t, "user1", res.Username)
	assert.Equal(t, "user", res.Role)
}

func TestToUserResponses(t *testing.T) {
	users := []*entity.User{
		{Username: "u1"},
		{Username: "u2"},
	}
	res := ToUserResponses(users)
	assert.Len(t, res, 2)
	assert.Equal(t, "u1", res[0].Username)
	assert.Equal(t, "u2", res[1].Username)
}
