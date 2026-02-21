package handler

import (
	"github.com/dermasense/backend/internal/service"
	"github.com/gofiber/fiber/v2"
)

// AuthHandler handles authentication endpoints.
type AuthHandler struct {
	authService *service.AuthService
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// GoogleLoginRequest is the request body for Google login.
type GoogleLoginRequest struct {
	IDToken string `json:"id_token"`
}

// GoogleLogin verifies a Google OAuth token and returns a JWT.
func (h *AuthHandler) GoogleLogin(c *fiber.Ctx) error {
	var req GoogleLoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	if req.IDToken == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "id_token is required",
		})
	}

	resp, err := h.authService.GoogleLogin(c.Context(), req.IDToken)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    resp,
	})
}
