package handler

import (
	"context"
	"time"

	"github.com/CPNext-hub/calendar-reg-main-api/internal/delivery/http/adapter"
	"github.com/CPNext-hub/calendar-reg-main-api/internal/delivery/http/dto"
	"github.com/CPNext-hub/calendar-reg-main-api/internal/domain/usecase"
	"github.com/CPNext-hub/calendar-reg-main-api/pkg/constants"
	"github.com/CPNext-hub/calendar-reg-main-api/pkg/response"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

// AuthHandler handles authentication-related HTTP requests.
type AuthHandler struct {
	usecase usecase.AuthUsecase
}

// NewAuthHandler creates a new AuthHandler instance.
func NewAuthHandler(uc usecase.AuthUsecase) *AuthHandler {
	return &AuthHandler{usecase: uc}
}

// Register creates a new student user (public).
// @Summary Register a new student
// @Description Create a new user with student role (public endpoint)
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.RegisterRequest true "Register Request"
// @Success 201 {object} dto.RegisterResponse
// @Failure 400 {object} interface{}
// @Failure 409 {object} interface{}
// @Failure 500 {object} interface{}
// @Router /auth/register [post]
func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req dto.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(adapter.NewFiberResponder(c), "Invalid request body")
	}

	if req.Username == "" || req.Password == "" {
		return response.BadRequest(adapter.NewFiberResponder(c), "Username and password are required")
	}

	if len(req.Password) < 6 {
		return response.BadRequest(adapter.NewFiberResponder(c), "Password must be at least 6 characters")
	}

	// Public register is always student role
	role := constants.RoleStudent

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	user, err := h.usecase.Register(ctx, req.Username, req.Password, role, nil)
	if err != nil {
		if err.Error() == "username already exists" {
			return response.Conflict(adapter.NewFiberResponder(c), err.Error())
		}
		return response.InternalError(adapter.NewFiberResponder(c), err.Error())
	}

	return response.Created(adapter.NewFiberResponder(c), dto.ToRegisterResponse(user))
}

// CreateUser creates a user with any role (admin-only).
// @Summary Create a user (admin)
// @Description Create a new user with any role. Requires superadmin or admin JWT.
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.RegisterRequest true "Register Request"
// @Security BearerAuth
// @Success 201 {object} dto.RegisterResponse
// @Failure 400 {object} interface{}
// @Failure 403 {object} interface{}
// @Failure 409 {object} interface{}
// @Failure 500 {object} interface{}
// @Router /auth/users [post]
func (h *AuthHandler) CreateUser(c *fiber.Ctx) error {
	var req dto.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(adapter.NewFiberResponder(c), "Invalid request body")
	}

	if req.Username == "" || req.Password == "" {
		return response.BadRequest(adapter.NewFiberResponder(c), "Username and password are required")
	}

	if len(req.Password) < 6 {
		return response.BadRequest(adapter.NewFiberResponder(c), "Password must be at least 6 characters")
	}

	role := req.Role
	if req.Role == "" {
		role = constants.RoleStudent
	} else if !constants.ValidRoles[role] {
		return response.BadRequest(adapter.NewFiberResponder(c), "Invalid role. Allowed: superadmin, admin, student")
	}

	// Extract caller role from JWT claims
	claims, ok := c.Locals("user").(jwt.MapClaims)
	if !ok {
		return response.Unauthorized(adapter.NewFiberResponder(c), "Authentication required")
	}
	callerRole, _ := claims["role"].(string)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	user, err := h.usecase.Register(ctx, req.Username, req.Password, role, &callerRole)
	if err != nil {
		switch err.Error() {
		case "username already exists":
			return response.Conflict(adapter.NewFiberResponder(c), err.Error())
		case "only superadmin or admin can create admin users",
			"superadmin cannot be created via registration":
			return response.Forbidden(adapter.NewFiberResponder(c), err.Error())
		}
		return response.InternalError(adapter.NewFiberResponder(c), err.Error())
	}

	return response.Created(adapter.NewFiberResponder(c), dto.ToRegisterResponse(user))
}

// Login authenticates a user and returns a JWT token.
// @Summary Login
// @Description Authenticate with username and password to receive a JWT token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.LoginRequest true "Login Request"
// @Success 200 {object} dto.LoginResponse
// @Failure 400 {object} interface{}
// @Failure 401 {object} interface{}
// @Failure 500 {object} interface{}
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req dto.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(adapter.NewFiberResponder(c), "Invalid request body")
	}

	if req.Username == "" || req.Password == "" {
		return response.BadRequest(adapter.NewFiberResponder(c), "Username and password are required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	token, err := h.usecase.Login(ctx, req.Username, req.Password)
	if err != nil {
		if err.Error() == "invalid credentials" {
			return response.Unauthorized(adapter.NewFiberResponder(c), "Invalid username or password")
		}
		return response.InternalError(adapter.NewFiberResponder(c), err.Error())
	}

	return response.OK(adapter.NewFiberResponder(c), dto.LoginResponse{Token: token})
}
