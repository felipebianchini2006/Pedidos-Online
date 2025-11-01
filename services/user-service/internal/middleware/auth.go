package middleware

import (
	"log"
	"strings"

	"github.com/gofiber/fiber/v2"

	"pedidos-online/user-service/internal/service"
)

// AuthMiddleware creates a middleware for JWT authentication
func AuthMiddleware(userService service.UserService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get Authorization header
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			log.Printf("Missing Authorization header")
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"error":   "Missing authorization header",
			})
		}

		// Extract token from "Bearer <token>" format
		token := extractToken(authHeader)
		if token == "" {
			log.Printf("Invalid Authorization header format")
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"error":   "Invalid authorization header format",
			})
		}

		// Validate token and get user ID
		userID, err := userService.ValidateToken(token)
		if err != nil {
			log.Printf("Invalid token: %v", err)
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"error":   "Invalid or expired token",
			})
		}

		// Store user ID in context for handlers to use
		c.Locals("userID", userID)

		// Continue to next handler
		return c.Next()
	}
}

// extractToken extracts the token from Authorization header
// Expected format: "Bearer <token>"
func extractToken(authHeader string) string {
	const bearerPrefix = "Bearer "

	if !strings.HasPrefix(authHeader, bearerPrefix) {
		return ""
	}

	token := strings.TrimSpace(authHeader[len(bearerPrefix):])
	return token
}

// OptionalAuthMiddleware is similar to AuthMiddleware but doesn't fail if token is missing
// Useful for endpoints that work differently for authenticated users
func OptionalAuthMiddleware(userService service.UserService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get Authorization header
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			// No token, but that's okay for optional auth
			return c.Next()
		}

		// Extract token
		token := extractToken(authHeader)
		if token == "" {
			// Invalid format, but continue without auth
			return c.Next()
		}

		// Try to validate token
		userID, err := userService.ValidateToken(token)
		if err != nil {
			// Invalid token, but continue without auth
			return c.Next()
		}

		// Store user ID in context if valid
		c.Locals("userID", userID)

		return c.Next()
	}
}
