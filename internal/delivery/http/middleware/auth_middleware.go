package middleware

import (
	"strings"

	"github.com/dermasense/backend/internal/service"
	"github.com/gofiber/fiber/v2"
)

// AuthMiddleware creates a JWT authentication middleware.
func AuthMiddleware(authService *service.AuthService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "missing authorization header",
			})
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid authorization format, use: Bearer <token>",
			})
		}

		userID, err := authService.ValidateJWT(parts[1])
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid or expired token",
			})
		}

		c.Locals("userID", userID)
		return c.Next()
	}
}
