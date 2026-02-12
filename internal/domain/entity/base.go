package entity

import "time"

// BaseEntity contains common fields shared by all domain entities.
type BaseEntity struct {
	ID        string
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time // nil = active, non-nil = soft-deleted
}
