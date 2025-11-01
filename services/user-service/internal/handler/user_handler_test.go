package handler

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
)

// Test Response struct
func TestResponse_Struct(t *testing.T) {
	resp := Response{
		Success: true,
		Data:    "test data",
		Error:   "",
		Message: "test message",
	}

	if !resp.Success {
		t.Error("Success should be true")
	}
	if resp.Data != "test data" {
		t.Error("Data mismatch")
	}
}

// Test RegisterRequest struct
func TestRegisterRequest_Struct(t *testing.T) {
	req := RegisterRequest{
		Email:    "test@example.com",
		Password: "password123",
		Name:     "Test User",
		Phone:    "11999999999",
	}

	if req.Email != "test@example.com" {
		t.Error("Email mismatch")
	}
}

// Test LoginRequest struct
func TestLoginRequest_Struct(t *testing.T) {
	req := LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	}

	if req.Email != "test@example.com" {
		t.Error("Email mismatch")
	}
}

// Test UpdateProfileRequest struct
func TestUpdateProfileRequest_Struct(t *testing.T) {
	req := UpdateProfileRequest{
		Name:  "New Name",
		Phone: "11988888888",
	}

	if req.Name != "New Name" {
		t.Error("Name mismatch")
	}
}

// Test AuthResponse struct
func TestAuthResponse_Struct(t *testing.T) {
	resp := AuthResponse{
		Token: "jwt-token",
		User:  map[string]string{"id": "123"},
	}

	if resp.Token != "jwt-token" {
		t.Error("Token mismatch")
	}
}

// Test NewUserHandler
func TestNewUserHandler(t *testing.T) {
	handler := NewUserHandler(nil)

	if handler == nil {
		t.Error("Handler should not be nil")
	}

	if handler.validator == nil {
		t.Error("Validator should be initialized")
	}
}

// Test formatValidationError
func TestFormatValidationError(t *testing.T) {
	// Create a simple error to test
	err := fiber.NewError(fiber.StatusBadRequest, "test error")
	
	result := formatValidationError(err)
	
	if result == "" {
		t.Error("Result should not be empty")
	}
}

// Test Register endpoint structure (without actual service call)
func TestRegister_InvalidJSON(t *testing.T) {
	app := fiber.New()
	handler := NewUserHandler(nil)
	
	app.Post("/register", handler.Register)

	// Create request with invalid JSON
	req := httptest.NewRequest("POST", "/register", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to test request: %v", err)
	}

	if resp.StatusCode != fiber.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", fiber.StatusBadRequest, resp.StatusCode)
	}
}

// Test Login endpoint structure (without actual service call)
func TestLogin_InvalidJSON(t *testing.T) {
	app := fiber.New()
	handler := NewUserHandler(nil)
	
	app.Post("/login", handler.Login)

	// Create request with invalid JSON
	req := httptest.NewRequest("POST", "/login", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to test request: %v", err)
	}

	if resp.StatusCode != fiber.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", fiber.StatusBadRequest, resp.StatusCode)
	}
}

// Test GetProfile without authentication
func TestGetProfile_Unauthorized(t *testing.T) {
	app := fiber.New()
	handler := NewUserHandler(nil)
	
	app.Get("/profile", handler.GetProfile)

	// Create request without userID in context
	req := httptest.NewRequest("GET", "/profile", nil)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to test request: %v", err)
	}

	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", fiber.StatusUnauthorized, resp.StatusCode)
	}
}

// Test UpdateProfile without authentication
func TestUpdateProfile_Unauthorized(t *testing.T) {
	app := fiber.New()
	handler := NewUserHandler(nil)
	
	app.Put("/profile", handler.UpdateProfile)

	// Create request without userID in context
	reqBody, _ := json.Marshal(UpdateProfileRequest{
		Name:  "New Name",
		Phone: "11999999999",
	})
	req := httptest.NewRequest("PUT", "/profile", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to test request: %v", err)
	}

	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", fiber.StatusUnauthorized, resp.StatusCode)
	}
}

// Test UpdateProfile with empty fields
func TestUpdateProfile_EmptyFields(t *testing.T) {
	app := fiber.New()
	handler := NewUserHandler(nil)
	
	// Create a middleware to inject userID
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("userID", "test-user-id")
		return c.Next()
	})
	
	app.Put("/profile", handler.UpdateProfile)

	// Create request with empty fields
	reqBody, _ := json.Marshal(UpdateProfileRequest{
		Name:  "",
		Phone: "",
	})
	req := httptest.NewRequest("PUT", "/profile", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to test request: %v", err)
	}

	if resp.StatusCode != fiber.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", fiber.StatusBadRequest, resp.StatusCode)
	}
}

// Test RegisterRoutes function
func TestRegisterRoutes(t *testing.T) {
	app := fiber.New()
	handler := NewUserHandler(nil)
	
	// Create a dummy auth middleware
	authMiddleware := func(c *fiber.Ctx) error {
		return c.Next()
	}

	// This should not panic
	RegisterRoutes(app, handler, authMiddleware)

	// Verify routes are registered (by testing a request)
	req := httptest.NewRequest("POST", "/api/v1/register", bytes.NewBuffer([]byte("{}")))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to test request: %v", err)
	}

	// Should get a response (even if it's an error)
	if resp == nil {
		t.Error("Expected a response")
	}
}

// Benchmark tests
func BenchmarkFormatValidationError(b *testing.B) {
	err := fiber.NewError(fiber.StatusBadRequest, "test error")
	for i := 0; i < b.N; i++ {
		formatValidationError(err)
	}
}

func BenchmarkNewUserHandler(b *testing.B) {
	for i := 0; i < b.N; i++ {
		NewUserHandler(nil)
	}
}

// Note: Full integration tests would require:
// 1. Mock service implementation
// 2. Test database setup
// 3. Complete request/response flow testing
// 4. Authentication middleware testing
//
// Consider using:
// - github.com/stretchr/testify/mock for mocking
// - github.com/stretchr/testify/assert for assertions
// - Complete service mocks for isolated handler testing
