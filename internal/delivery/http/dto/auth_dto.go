package dto

import "github.com/CPNext-hub/calendar-reg-main-api/internal/domain/entity"

// ---------- Request DTOs ----------

// RegisterRequest represents a user registration request.
type RegisterRequest struct {
	Username string `json:"username" example:"admin1"`
	Password string `json:"password" example:"password123"`
	Role     string `json:"role" example:"admin"`
}

// LoginRequest represents a login request.
type LoginRequest struct {
	Username string `json:"username" example:"admin1"`
	Password string `json:"password" example:"password123"`
}

// ---------- Response DTOs ----------

// RegisterResponse represents the response after successful registration.
type RegisterResponse struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Role     string `json:"role"`
}

// LoginResponse represents the response after successful login.
type LoginResponse struct {
	Token string `json:"token"`
}

// ---------- Converters ----------

// ToRegisterResponse converts a User entity to a RegisterResponse.
func ToRegisterResponse(user *entity.User) *RegisterResponse {
	return &RegisterResponse{
		ID:       user.ID,
		Username: user.Username,
		Role:     string(user.Role),
	}
}
