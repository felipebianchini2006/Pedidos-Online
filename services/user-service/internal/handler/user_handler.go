package handler

import (
	"errors"
	"log"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"

	"pedidos-online/user-service/internal/service"
)

// Response is the standardized response format for all endpoints
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Message string      `json:"message,omitempty"`
}

// RegisterRequest represents the request body for user registration
type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
	Name     string `json:"name" validate:"required,min=2,max=100"`
	Phone    string `json:"phone" validate:"required,min=10,max=15"`
}

// LoginRequest represents the request body for user login
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// UpdateProfileRequest represents the request body for updating user profile
type UpdateProfileRequest struct {
	Name  string `json:"name" validate:"omitempty,min=2,max=100"`
	Phone string `json:"phone" validate:"omitempty,min=10,max=15"`
}

// AuthResponse represents the response body for login
type AuthResponse struct {
	Token string      `json:"token"`
	User  interface{} `json:"user"`
}

// UserHandler handles HTTP requests for user operations
type UserHandler struct {
	service   service.UserService
	validator *validator.Validate
}

// NewUserHandler creates a new instance of UserHandler
func NewUserHandler(service service.UserService) *UserHandler {
	return &UserHandler{
		service:   service,
		validator: validator.New(),
	}
}

// Register handles user registration
// POST /api/v1/register
func (h *UserHandler) Register(c *fiber.Ctx) error {
	var req RegisterRequest

	// Parse request body
	if err := c.BodyParser(&req); err != nil {
		log.Printf("Error parsing register request: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(Response{
			Success: false,
			Error:   "Invalid request body",
		})
	}

	// Validate request
	if err := h.validator.Struct(req); err != nil {
		log.Printf("Validation error in register: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(Response{
			Success: false,
			Error:   formatValidationError(err),
		})
	}

	// Call service to register user
	user, err := h.service.Register(
		c.Context(),
		req.Email,
		req.Password,
		req.Name,
		req.Phone,
	)

	if err != nil {
		// Handle specific errors
		switch {
		case errors.Is(err, service.ErrInvalidEmail):
			return c.Status(fiber.StatusBadRequest).JSON(Response{
				Success: false,
				Error:   "Invalid email format",
			})
		case errors.Is(err, service.ErrWeakPassword):
			return c.Status(fiber.StatusBadRequest).JSON(Response{
				Success: false,
				Error:   "Password must be at least 8 characters long",
			})
		case errors.Is(err, service.ErrEmailAlreadyExists):
			return c.Status(fiber.StatusConflict).JSON(Response{
				Success: false,
				Error:   "Email already exists",
			})
		default:
			log.Printf("Error registering user: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(Response{
				Success: false,
				Error:   "Failed to register user",
			})
		}
	}

	log.Printf("User registered successfully: %s", user.Email)

	// Return success response
	return c.Status(fiber.StatusCreated).JSON(Response{
		Success: true,
		Data:    user.ToResponse(),
		Message: "User registered successfully",
	})
}

// Login handles user authentication
// POST /api/v1/login
func (h *UserHandler) Login(c *fiber.Ctx) error {
	var req LoginRequest

	// Parse request body
	if err := c.BodyParser(&req); err != nil {
		log.Printf("Error parsing login request: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(Response{
			Success: false,
			Error:   "Invalid request body",
		})
	}

	// Validate request
	if err := h.validator.Struct(req); err != nil {
		log.Printf("Validation error in login: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(Response{
			Success: false,
			Error:   formatValidationError(err),
		})
	}

	// Call service to login
	token, err := h.service.Login(
		c.Context(),
		req.Email,
		req.Password,
	)

	if err != nil {
		// Handle specific errors
		if errors.Is(err, service.ErrInvalidCredentials) {
			return c.Status(fiber.StatusUnauthorized).JSON(Response{
				Success: false,
				Error:   "Invalid email or password",
			})
		}

		log.Printf("Error logging in user: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(Response{
			Success: false,
			Error:   "Failed to login",
		})
	}

	// Get user profile for response
	userID, err := h.service.ValidateToken(token)
	if err != nil {
		log.Printf("Error validating token after login: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(Response{
			Success: false,
			Error:   "Failed to generate session",
		})
	}

	user, err := h.service.GetProfile(c.Context(), userID)
	if err != nil {
		log.Printf("Error getting profile after login: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(Response{
			Success: false,
			Error:   "Failed to get user profile",
		})
	}

	log.Printf("User logged in successfully: %s", req.Email)

	// Return success response
	return c.Status(fiber.StatusOK).JSON(Response{
		Success: true,
		Data: AuthResponse{
			Token: token,
			User:  user.ToResponse(),
		},
		Message: "Login successful",
	})
}

// GetProfile retrieves the authenticated user's profile
// GET /api/v1/profile
func (h *UserHandler) GetProfile(c *fiber.Ctx) error {
	// Extract user ID from context (set by auth middleware)
	userID := c.Locals("userID")
	if userID == nil {
		log.Printf("User ID not found in context")
		return c.Status(fiber.StatusUnauthorized).JSON(Response{
			Success: false,
			Error:   "Unauthorized",
		})
	}

	userIDStr, ok := userID.(string)
	if !ok {
		log.Printf("Invalid user ID type in context")
		return c.Status(fiber.StatusUnauthorized).JSON(Response{
			Success: false,
			Error:   "Invalid authentication",
		})
	}

	// Call service to get profile
	user, err := h.service.GetProfile(c.Context(), userIDStr)
	if err != nil {
		// Handle specific errors
		switch {
		case errors.Is(err, service.ErrUserNotFound):
			return c.Status(fiber.StatusNotFound).JSON(Response{
				Success: false,
				Error:   "User not found",
			})
		case errors.Is(err, service.ErrInvalidUserID):
			return c.Status(fiber.StatusBadRequest).JSON(Response{
				Success: false,
				Error:   "Invalid user ID",
			})
		default:
			log.Printf("Error getting profile: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(Response{
				Success: false,
				Error:   "Failed to get profile",
			})
		}
	}

	log.Printf("Profile retrieved for user: %s", userIDStr)

	// Return success response
	return c.Status(fiber.StatusOK).JSON(Response{
		Success: true,
		Data:    user.ToResponse(),
	})
}

// UpdateProfile updates the authenticated user's profile
// PUT /api/v1/profile
func (h *UserHandler) UpdateProfile(c *fiber.Ctx) error {
	// Extract user ID from context (set by auth middleware)
	userID := c.Locals("userID")
	if userID == nil {
		log.Printf("User ID not found in context")
		return c.Status(fiber.StatusUnauthorized).JSON(Response{
			Success: false,
			Error:   "Unauthorized",
		})
	}

	userIDStr, ok := userID.(string)
	if !ok {
		log.Printf("Invalid user ID type in context")
		return c.Status(fiber.StatusUnauthorized).JSON(Response{
			Success: false,
			Error:   "Invalid authentication",
		})
	}

	var req UpdateProfileRequest

	// Parse request body
	if err := c.BodyParser(&req); err != nil {
		log.Printf("Error parsing update profile request: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(Response{
			Success: false,
			Error:   "Invalid request body",
		})
	}

	// Validate request
	if err := h.validator.Struct(req); err != nil {
		log.Printf("Validation error in update profile: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(Response{
			Success: false,
			Error:   formatValidationError(err),
		})
	}

	// Check if at least one field is provided
	if req.Name == "" && req.Phone == "" {
		return c.Status(fiber.StatusBadRequest).JSON(Response{
			Success: false,
			Error:   "At least one field (name or phone) must be provided",
		})
	}

	// Call service to update profile
	user, err := h.service.UpdateProfile(
		c.Context(),
		userIDStr,
		req.Name,
		req.Phone,
	)

	if err != nil {
		// Handle specific errors
		switch {
		case errors.Is(err, service.ErrUserNotFound):
			return c.Status(fiber.StatusNotFound).JSON(Response{
				Success: false,
				Error:   "User not found",
			})
		case errors.Is(err, service.ErrInvalidUserID):
			return c.Status(fiber.StatusBadRequest).JSON(Response{
				Success: false,
				Error:   "Invalid user ID",
			})
		default:
			log.Printf("Error updating profile: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(Response{
				Success: false,
				Error:   "Failed to update profile",
			})
		}
	}

	log.Printf("Profile updated for user: %s", userIDStr)

	// Return success response
	return c.Status(fiber.StatusOK).JSON(Response{
		Success: true,
		Data:    user.ToResponse(),
		Message: "Profile updated successfully",
	})
}

// RegisterRoutes registers all user routes
func RegisterRoutes(app *fiber.App, handler *UserHandler, authMiddleware fiber.Handler) {
	// API v1 group
	api := app.Group("/api/v1")

	// Public routes (no authentication required)
	api.Post("/register", handler.Register)
	api.Post("/login", handler.Login)

	// Protected routes (authentication required)
	api.Get("/profile", authMiddleware, handler.GetProfile)
	api.Put("/profile", authMiddleware, handler.UpdateProfile)
}

// Helper functions

// formatValidationError formats validation errors into a user-friendly message
func formatValidationError(err error) string {
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		if len(validationErrors) > 0 {
			field := validationErrors[0].Field()
			tag := validationErrors[0].Tag()

			switch tag {
			case "required":
				return field + " is required"
			case "email":
				return "Invalid email format"
			case "min":
				return field + " is too short"
			case "max":
				return field + " is too long"
			default:
				return "Validation error: " + field
			}
		}
	}
	return "Invalid input"
}
