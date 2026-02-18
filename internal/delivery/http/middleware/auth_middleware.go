package middleware

import (
	"strings"

	"github.com/CPNext-hub/calendar-reg-main-api/internal/delivery/http/adapter"
	"github.com/CPNext-hub/calendar-reg-main-api/pkg/response"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

// JWTAuth returns a Fiber middleware that validates JWT tokens from the
// Authorization header (Bearer scheme). On success it stores the parsed
// claims in c.Locals("user") for downstream handlers/middleware.
func JWTAuth(secret string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return response.Unauthorized(adapter.NewFiberResponder(c), "Missing authorization header")
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			return response.Unauthorized(adapter.NewFiberResponder(c), "Invalid authorization format. Use: Bearer <token>")
		}

		tokenStr := parts[1]

		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			// Ensure signing method is HMAC
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(secret), nil
		})
		if err != nil || !token.Valid {
			return response.Unauthorized(adapter.NewFiberResponder(c), "Invalid or expired token")
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return response.Unauthorized(adapter.NewFiberResponder(c), "Invalid token claims")
		}

		// Store claims for downstream use
		c.Locals("user", claims)
		return c.Next()
	}
}
