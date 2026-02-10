package usecase

import (
	"github.com/CPNext-hub/calendar-reg-main-api/internal/domain/entity"
)

// VersionUsecase defines the interface for version-related business logic.
type VersionUsecase interface {
	GetVersion() entity.AppInfo
}

type versionUsecase struct {
	name    string
	version string
	env     string
}

// NewVersionUsecase creates a new VersionUsecase instance.
func NewVersionUsecase(name, version, env string) VersionUsecase {
	return &versionUsecase{
		name:    name,
		version: version,
		env:     env,
	}
}

func (u *versionUsecase) GetVersion() entity.AppInfo {
	return entity.AppInfo{
		Name:    u.name,
		Version: u.version,
		Env:     u.env,
	}
}
