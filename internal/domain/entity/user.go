package entity

// User represents a registered user.
type User struct {
	BaseEntity
	Username string
	Password string // bcrypt-hashed
	Role     string
}
