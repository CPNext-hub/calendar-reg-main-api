package middleware

import (
	"github.com/CPNext-hub/calendar-reg-main-api/internal/delivery/http/adapter"
	"github.com/CPNext-hub/calendar-reg-main-api/pkg/response"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

// RequireRole returns a Fiber middleware that checks whether the
// authenticated user's role is one of the allowed roles.
// It expects c.Locals("user") to contain jwt.MapClaims (set by JWTAuth).
func RequireRole(roles ...string) fiber.Handler {
	allowed := make(map[string]bool, len(roles))
	for _, r := range roles {
		allowed[r] = true
	}

	return func(c *fiber.Ctx) error {
		claims, ok := c.Locals("user").(jwt.MapClaims)
		if !ok {
			return response.Unauthorized(adapter.NewFiberResponder(c), "Authentication required")
		}

		role, ok := claims["role"].(string)
		if !ok || !allowed[role] {
			return response.Forbidden(adapter.NewFiberResponder(c), "Insufficient permissions")
		}

		return c.Next()
	}
}
