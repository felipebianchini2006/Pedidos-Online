package middleware

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	"pedidos-online/user-service/pkg/jwt"
)

// Context keys for storing user information
const (
	ContextKeyUserID = "userID"
	ContextKeyEmail  = "email"
)

// Error messages
var (
	ErrTokenNotProvided = errors.New("Token de autenticação não fornecido")
	ErrTokenInvalid     = errors.New("Token de autenticação inválido")
	ErrTokenExpired     = errors.New("Token de autenticação expirado")
	ErrInvalidFormat    = errors.New("Formato de token inválido")
	ErrUserIDNotFound   = errors.New("userID not found in context")
	ErrEmailNotFound    = errors.New("email not found in context")
)

// AuthMiddleware creates a middleware for JWT authentication
// It validates the JWT token and stores user information in the context
func AuthMiddleware(jwtSecret string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get Authorization header
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			logAccessDenied(c, "Token ausente")
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"error":   ErrTokenNotProvided.Error(),
			})
		}

		// Extract token from "Bearer <token>" format
		token, err := extractTokenFromHeader(authHeader)
		if err != nil {
			logAccessDenied(c, fmt.Sprintf("Formato inválido: %v", err))
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"error":   ErrInvalidFormat.Error(),
			})
		}

		// Validate token and extract claims
		claims, err := jwt.ValidateToken(token, jwtSecret)
		if err != nil {
			// Check if it's an expiration error
			if strings.Contains(err.Error(), "expired") {
				logAccessDenied(c, fmt.Sprintf("Token expirado: %v", err))
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"success": false,
					"error":   ErrTokenExpired.Error(),
				})
			}

			logAccessDenied(c, fmt.Sprintf("Token inválido: %v", err))
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"error":   ErrTokenInvalid.Error(),
			})
		}

		// Validate claims
		if claims.UserID == "" {
			logAccessDenied(c, "UserID ausente no token")
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"error":   ErrTokenInvalid.Error(),
			})
		}

		// Store user information in context
		c.Locals(ContextKeyUserID, claims.UserID)
		c.Locals(ContextKeyEmail, claims.Email)

		// Log successful authentication
		log.Printf("[AUTH] User authenticated - UserID: %s, Email: %s, IP: %s, Path: %s",
			claims.UserID, claims.Email, c.IP(), c.Path())

		// Continue to next handler
		return c.Next()
	}
}

// extractTokenFromHeader extracts the token from Authorization header
// Expected format: "Bearer <token>"
func extractTokenFromHeader(authHeader string) (string, error) {
	const bearerPrefix = "Bearer "

	// Check if header has the correct prefix
	if !strings.HasPrefix(authHeader, bearerPrefix) {
		return "", fmt.Errorf("header must start with 'Bearer '")
	}

	// Extract token after "Bearer "
	token := strings.TrimSpace(authHeader[len(bearerPrefix):])
	if token == "" {
		return "", fmt.Errorf("token is empty")
	}

	return token, nil
}

// logAccessDenied logs authentication failures with IP and timestamp
func logAccessDenied(c *fiber.Ctx, reason string) {
	log.Printf("[AUTH DENIED] IP: %s, Path: %s, Method: %s, Reason: %s, Timestamp: %s, UserAgent: %s",
		c.IP(),
		c.Path(),
		c.Method(),
		reason,
		time.Now().Format(time.RFC3339),
		c.Get("User-Agent"),
	)
}

// GetUserIDFromContext retrieves the user ID from the Fiber context
// Returns error if user ID is not found or has invalid type
func GetUserIDFromContext(c *fiber.Ctx) (string, error) {
	userID := c.Locals(ContextKeyUserID)
	if userID == nil {
		return "", ErrUserIDNotFound
	}

	userIDStr, ok := userID.(string)
	if !ok {
		return "", fmt.Errorf("userID has invalid type: expected string, got %T", userID)
	}

	if userIDStr == "" {
		return "", fmt.Errorf("userID is empty")
	}

	return userIDStr, nil
}

// GetEmailFromContext retrieves the email from the Fiber context
// Returns error if email is not found or has invalid type
func GetEmailFromContext(c *fiber.Ctx) (string, error) {
	email := c.Locals(ContextKeyEmail)
	if email == nil {
		return "", ErrEmailNotFound
	}

	emailStr, ok := email.(string)
	if !ok {
		return "", fmt.Errorf("email has invalid type: expected string, got %T", email)
	}

	// Email can be empty (not all tokens include it)
	return emailStr, nil
}

// OptionalAuthMiddleware is similar to AuthMiddleware but doesn't fail if token is missing
// Useful for endpoints that work differently for authenticated users
// If a valid token is provided, user information is stored in context
func OptionalAuthMiddleware(jwtSecret string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get Authorization header
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			// No token provided, continue without authentication
			return c.Next()
		}

		// Try to extract token
		token, err := extractTokenFromHeader(authHeader)
		if err != nil {
			// Invalid format, but continue without authentication
			log.Printf("[OPTIONAL AUTH] Invalid token format from IP %s: %v", c.IP(), err)
			return c.Next()
		}

		// Try to validate token
		claims, err := jwt.ValidateToken(token, jwtSecret)
		if err != nil {
			// Invalid or expired token, continue without authentication
			log.Printf("[OPTIONAL AUTH] Invalid token from IP %s: %v", c.IP(), err)
			return c.Next()
		}

		// Store user information in context if valid
		if claims.UserID != "" {
			c.Locals(ContextKeyUserID, claims.UserID)
			c.Locals(ContextKeyEmail, claims.Email)
			log.Printf("[OPTIONAL AUTH] User authenticated - UserID: %s, IP: %s", claims.UserID, c.IP())
		}

		return c.Next()
	}
}

// RequireUserID is a helper middleware that can be chained after AuthMiddleware
// to ensure the userID was properly extracted
func RequireUserID() fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, err := GetUserIDFromContext(c)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"error":   "Autenticação necessária",
			})
		}

		if userID == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"error":   "ID do usuário não encontrado",
			})
		}

		return c.Next()
	}
}

// RateLimitConfig holds configuration for basic rate limiting
type RateLimitConfig struct {
	// Max requests per window
	MaxRequests int
	// Window duration
	Window time.Duration
	// Store for tracking requests (simple in-memory map for now)
	// In production, use Redis or similar
	requests map[string][]time.Time
}

// NewRateLimitConfig creates a new rate limit configuration
func NewRateLimitConfig(maxRequests int, window time.Duration) *RateLimitConfig {
	return &RateLimitConfig{
		MaxRequests: maxRequests,
		Window:      window,
		requests:    make(map[string][]time.Time),
	}
}

// BasicRateLimitMiddleware implements a simple in-memory rate limiter
// Note: This is a basic implementation. For production, use Redis-based rate limiting
// or implement this at the API Gateway level
func BasicRateLimitMiddleware(config *RateLimitConfig) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Use IP address as identifier
		identifier := c.IP()

		// Get current time
		now := time.Now()

		// Get request history for this identifier
		requests, exists := config.requests[identifier]
		if !exists {
			requests = []time.Time{}
		}

		// Filter out old requests outside the window
		validRequests := []time.Time{}
		for _, reqTime := range requests {
			if now.Sub(reqTime) < config.Window {
				validRequests = append(validRequests, reqTime)
			}
		}

		// Check if limit exceeded
		if len(validRequests) >= config.MaxRequests {
			log.Printf("[RATE LIMIT] Limit exceeded for IP: %s, Requests: %d, Window: %s",
				identifier, len(validRequests), config.Window)

			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"success": false,
				"error":   "Limite de requisições excedido. Tente novamente mais tarde.",
			})
		}

		// Add current request
		validRequests = append(validRequests, now)
		config.requests[identifier] = validRequests

		return c.Next()
	}
}
