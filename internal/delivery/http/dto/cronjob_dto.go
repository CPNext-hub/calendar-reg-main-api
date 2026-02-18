package dto

import (
	"time"

	"github.com/CPNext-hub/calendar-reg-main-api/internal/domain/entity"
)

// --- CronJob Request DTOs ---

// CreateCronJobRequest represents the request body for creating a cron job.
type CreateCronJobRequest struct {
	Name        string   `json:"name"`
	CourseCodes []string `json:"course_codes"`
	Acadyear    int      `json:"acadyear"`
	Semester    int      `json:"semester"`
	CronExpr    string   `json:"cron_expr"` // e.g. "0 */6 * * *"
	Enabled     bool     `json:"enabled"`
}

// ToEntity converts a CreateCronJobRequest to a domain entity.
func (r *CreateCronJobRequest) ToEntity() *entity.CronJob {
	return &entity.CronJob{
		Name:        r.Name,
		CourseCodes: r.CourseCodes,
		Acadyear:    r.Acadyear,
		Semester:    r.Semester,
		CronExpr:    r.CronExpr,
		Enabled:     r.Enabled,
	}
}

// UpdateCronJobRequest represents the request body for updating a cron job.
type UpdateCronJobRequest struct {
	Name        string   `json:"name"`
	CourseCodes []string `json:"course_codes"`
	Acadyear    int      `json:"acadyear"`
	Semester    int      `json:"semester"`
	CronExpr    string   `json:"cron_expr"`
	Enabled     bool     `json:"enabled"`
}

// ToEntity converts an UpdateCronJobRequest to a domain entity with the given ID.
func (r *UpdateCronJobRequest) ToEntity(id string) *entity.CronJob {
	return &entity.CronJob{
		BaseEntity:  entity.BaseEntity{ID: id},
		Name:        r.Name,
		CourseCodes: r.CourseCodes,
		Acadyear:    r.Acadyear,
		Semester:    r.Semester,
		CronExpr:    r.CronExpr,
		Enabled:     r.Enabled,
	}
}

// --- CronJob Response DTOs ---

// CronJobResponse represents the response body for a cron job.
type CronJobResponse struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	CourseCodes []string `json:"course_codes"`
	Acadyear    int      `json:"acadyear"`
	Semester    int      `json:"semester"`
	CronExpr    string   `json:"cron_expr"`
	Enabled     bool     `json:"enabled"`
	CreatedAt   string   `json:"created_at"`
	UpdatedAt   string   `json:"updated_at"`
}

// ToCronJobResponse converts a CronJob entity to a CronJobResponse DTO.
func ToCronJobResponse(j *entity.CronJob) *CronJobResponse {
	if j == nil {
		return nil
	}
	return &CronJobResponse{
		ID:          j.ID,
		Name:        j.Name,
		CourseCodes: j.CourseCodes,
		Acadyear:    j.Acadyear,
		Semester:    j.Semester,
		CronExpr:    j.CronExpr,
		Enabled:     j.Enabled,
		CreatedAt:   j.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   j.UpdatedAt.Format(time.RFC3339),
	}
}

// ToCronJobResponses converts a slice of CronJob entities to CronJobResponse DTOs.
func ToCronJobResponses(jobs []*entity.CronJob) []*CronJobResponse {
	responses := make([]*CronJobResponse, len(jobs))
	for i, j := range jobs {
		responses[i] = ToCronJobResponse(j)
	}
	return responses
}
