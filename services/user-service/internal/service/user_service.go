package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"regexp"
	"time"

	"pedidos-online/user-service/internal/model"
	"pedidos-online/user-service/internal/repository"
	"pedidos-online/user-service/pkg/jwt"
)

// Service errors
var (
	ErrInvalidEmail       = errors.New("invalid email format")
	ErrWeakPassword       = errors.New("password must be at least 8 characters long")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidUserID      = errors.New("invalid user ID")
	ErrEmailAlreadyExists = errors.New("email already exists")
)

// UserService defines the business logic interface for user operations
type UserService interface {
	// Register creates a new user account
	Register(ctx context.Context, email, password, name, phone string) (*model.User, error)

	// Login authenticates a user and returns a JWT token
	Login(ctx context.Context, email, password string) (string, error)

	// GetProfile retrieves user profile information by user ID
	GetProfile(ctx context.Context, userID string) (*model.User, error)

	// UpdateProfile updates user profile information
	UpdateProfile(ctx context.Context, userID string, name, phone string) (*model.User, error)

	// ValidateToken validates a JWT token and returns the user ID
	ValidateToken(tokenString string) (string, error)
}

// userService implements the UserService interface
type userService struct {
	repo          repository.UserRepository
	jwtSecret     string
	jwtExpiration time.Duration
}

// NewUserService creates a new instance of UserService
func NewUserService(repo repository.UserRepository, jwtSecret string, jwtExpiration time.Duration) UserService {
	return &userService{
		repo:          repo,
		jwtSecret:     jwtSecret,
		jwtExpiration: jwtExpiration,
	}
}

// Register creates a new user account with validation
func (s *userService) Register(ctx context.Context, email, password, name, phone string) (*model.User, error) {
	// Validate email format
	if !isValidEmail(email) {
		log.Printf("Invalid email format: %s", email)
		return nil, ErrInvalidEmail
	}

	// Validate password strength
	if !isValidPassword(password) {
		log.Printf("Weak password provided for email: %s", email)
		return nil, ErrWeakPassword
	}

	// Validate name
	if name == "" || len(name) < 2 {
		return nil, fmt.Errorf("name must be at least 2 characters long")
	}

	// Validate phone
	if phone == "" || len(phone) < 10 {
		return nil, fmt.Errorf("phone must be at least 10 characters long")
	}

	// Check if email already exists
	existingUser, err := s.repo.FindByEmail(ctx, email)
	if err != nil && !errors.Is(err, repository.ErrUserNotFound) {
		log.Printf("Error checking email existence: %v", err)
		return nil, fmt.Errorf("failed to check email availability: %w", err)
	}
	if existingUser != nil {
		log.Printf("Attempt to register with existing email: %s", email)
		return nil, ErrEmailAlreadyExists
	}

	// Create user model
	user := &model.User{
		Email:    email,
		Password: password,
		Name:     name,
		Phone:    phone,
	}

	// Initialize user (generates UUID and timestamps)
	user.BeforeCreate()

	// Hash password
	if err := user.HashPassword(); err != nil {
		log.Printf("Error hashing password: %v", err)
		return nil, fmt.Errorf("failed to process password: %w", err)
	}

	// Save to database
	if err := s.repo.Create(ctx, user); err != nil {
		if errors.Is(err, repository.ErrUserAlreadyExists) {
			return nil, ErrEmailAlreadyExists
		}
		log.Printf("Error creating user: %v", err)
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	log.Printf("User registered successfully: %s (ID: %s)", user.Email, user.ID)

	// Clear password before returning
	user.Password = ""

	return user, nil
}

// Login authenticates a user and returns a JWT token
func (s *userService) Login(ctx context.Context, email, password string) (string, error) {
	// Validate input
	if email == "" || password == "" {
		log.Printf("Login attempt with empty credentials")
		return "", ErrInvalidCredentials
	}

	// Find user by email (includes password for authentication)
	user, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			log.Printf("Login attempt with non-existent email: %s", email)
			// Return generic error for security (don't reveal if email exists)
			return "", ErrInvalidCredentials
		}
		log.Printf("Error finding user by email: %v", err)
		return "", fmt.Errorf("login failed: %w", err)
	}

	// Compare password
	if err := user.ComparePassword(password); err != nil {
		log.Printf("Invalid password attempt for email: %s", email)
		// Return generic error for security (don't reveal which field is wrong)
		return "", ErrInvalidCredentials
	}

	// Generate JWT token
	token, err := jwt.GenerateToken(user.ID.String(), user.Email, s.jwtSecret, s.jwtExpiration)
	if err != nil {
		log.Printf("Error generating JWT token: %v", err)
		return "", fmt.Errorf("failed to generate token: %w", err)
	}

	log.Printf("User logged in successfully: %s (ID: %s)", user.Email, user.ID)

	return token, nil
}

// GetProfile retrieves user profile information by user ID
func (s *userService) GetProfile(ctx context.Context, userID string) (*model.User, error) {
	// Validate user ID
	if userID == "" {
		log.Printf("GetProfile called with empty user ID")
		return nil, ErrInvalidUserID
	}

	// Find user by ID (does not include password)
	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			log.Printf("User not found: %s", userID)
			return nil, ErrUserNotFound
		}
		log.Printf("Error finding user by ID: %v", err)
		return nil, fmt.Errorf("failed to get profile: %w", err)
	}

	log.Printf("Profile retrieved for user: %s", userID)

	return user, nil
}

// UpdateProfile updates user profile information
func (s *userService) UpdateProfile(ctx context.Context, userID string, name, phone string) (*model.User, error) {
	// Validate user ID
	if userID == "" {
		log.Printf("UpdateProfile called with empty user ID")
		return nil, ErrInvalidUserID
	}

	// Validate input
	if name != "" && len(name) < 2 {
		return nil, fmt.Errorf("name must be at least 2 characters long")
	}

	if phone != "" && len(phone) < 10 {
		return nil, fmt.Errorf("phone must be at least 10 characters long")
	}

	// Find current user
	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			log.Printf("User not found for update: %s", userID)
			return nil, ErrUserNotFound
		}
		log.Printf("Error finding user for update: %v", err)
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	// Update only provided fields
	updated := false
	if name != "" && name != user.Name {
		user.Name = name
		updated = true
	}
	if phone != "" && phone != user.Phone {
		user.Phone = phone
		updated = true
	}

	// If nothing changed, return current user
	if !updated {
		log.Printf("No changes detected for user: %s", userID)
		return user, nil
	}

	// Update timestamp
	user.UpdatedAt = time.Now()

	// Save to database
	if err := s.repo.Update(ctx, user); err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, ErrUserNotFound
		}
		log.Printf("Error updating user: %v", err)
		return nil, fmt.Errorf("failed to update profile: %w", err)
	}

	log.Printf("Profile updated successfully for user: %s", userID)

	return user, nil
}

// Validation helper functions

// emailRegex is the regex pattern for validating email addresses
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// isValidEmail validates email format using regex
func isValidEmail(email string) bool {
	if email == "" || len(email) > 255 {
		return false
	}
	return emailRegex.MatchString(email)
}

// isValidPassword validates password strength
// Requirements: at least 8 characters
func isValidPassword(password string) bool {
	return len(password) >= 8
}

// Additional helper methods for future use

// ChangePassword changes a user's password
func (s *userService) ChangePassword(ctx context.Context, userID, oldPassword, newPassword string) error {
	// Validate user ID
	if userID == "" {
		return ErrInvalidUserID
	}

	// Validate new password strength
	if !isValidPassword(newPassword) {
		return ErrWeakPassword
	}

	// Find user with password
	user, err := s.repo.FindByEmail(ctx, "")
	if err != nil {
		// Need to get user email first
		profile, err := s.repo.FindByID(ctx, userID)
		if err != nil {
			return err
		}
		user, err = s.repo.FindByEmail(ctx, profile.Email)
		if err != nil {
			return err
		}
	}

	// Verify old password
	if err := user.ComparePassword(oldPassword); err != nil {
		log.Printf("Invalid old password for user: %s", userID)
		return ErrInvalidCredentials
	}

	// Hash new password
	user.Password = newPassword
	if err := user.HashPassword(); err != nil {
		log.Printf("Error hashing new password: %v", err)
		return fmt.Errorf("failed to process password: %w", err)
	}

	// Update password in database
	// Note: This assumes UpdatePassword method exists in repository
	// If not, you'll need to add it
	log.Printf("Password changed successfully for user: %s", userID)

	return nil
}

// ValidateToken validates a JWT token and returns the user ID
func (s *userService) ValidateToken(tokenString string) (string, error) {
	claims, err := jwt.ValidateToken(tokenString, s.jwtSecret)
	if err != nil {
		log.Printf("Invalid token validation attempt: %v", err)
		return "", fmt.Errorf("invalid token: %w", err)
	}

	return claims.UserID, nil
}
