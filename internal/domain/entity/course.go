package entity

// Course represents a university course.
type Course struct {
	BaseEntity
	Code    string // e.g., "SC337861"
	Name    string // e.g., "PRINCIPLES OF REMOTE SENSING"
	Credits string // e.g., "3 (2-3-6)"
}
