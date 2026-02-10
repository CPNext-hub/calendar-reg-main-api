package dto

import "github.com/CPNext-hub/calendar-reg-main-api/internal/domain/entity"

// VersionResponse is the DTO for the version endpoint.
type VersionResponse struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Env     string `json:"env"`
}

// ToVersionResponse converts a domain AppInfo entity to a VersionResponse DTO.
func ToVersionResponse(info entity.AppInfo) VersionResponse {
	return VersionResponse{
		Name:    info.Name,
		Version: info.Version,
		Env:     info.Env,
	}
}
