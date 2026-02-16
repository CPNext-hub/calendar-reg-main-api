package constants

// Role constants for the system.
const (
	RoleSuperAdmin = "superadmin"
	RoleAdmin      = "admin"
	RoleStudent    = "student"
)

// ValidRoles contains all valid roles for validation.
var ValidRoles = map[string]bool{
	RoleSuperAdmin: true,
	RoleAdmin:      true,
	RoleStudent:    true,
}

// PrivilegedRoles are roles that can manage users (e.g. create admin accounts).
var PrivilegedRoles = map[string]bool{
	RoleSuperAdmin: true,
	RoleAdmin:      true,
}
