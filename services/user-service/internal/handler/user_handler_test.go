package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"pedidos-online/user-service/internal/model"
	"pedidos-online/user-service/internal/service"
)

// MockUserService is a mock implementation of UserService
type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) Register(ctx context.Context, email, password, name, phone string) (*model.User, error) {
	args := m.Called(ctx, email, password, name, phone)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserService) Login(ctx context.Context, email, password string) (string, error) {
	args := m.Called(ctx, email, password)
	return args.String(0), args.Error(1)
}

func (m *MockUserService) GetProfile(ctx context.Context, userID string) (*model.User, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserService) UpdateProfile(ctx context.Context, userID, name, phone string) (*model.User, error) {
	args := m.Called(ctx, userID, name, phone)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserService) ValidateToken(tokenString string) (string, error) {
	args := m.Called(tokenString)
	return args.String(0), args.Error(1)
}

func setupTestApp() *fiber.App {
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		},
	})
	return app
}

func TestRegisterHandler(t *testing.T) {
	t.Run("successfully register with valid data", func(t *testing.T) {
		mockService := new(MockUserService)
		handler := NewUserHandler(mockService)
		app := setupTestApp()

		user := &model.User{
			ID:    uuid.New(),
			Email: "test@example.com",
			Name:  "Test User",
			Phone: "1234567890",
		}

		mockService.On("Register", mock.Anything, "test@example.com", "password123", "Test User", "1234567890").
			Return(user, nil)

		requestBody := map[string]string{
			"email":    "test@example.com",
			"password": "password123",
			"name":     "Test User",
			"phone":    "1234567890",
		}
		body, _ := json.Marshal(requestBody)

		req := httptest.NewRequest("POST", "/register", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		// Perform actual registration
		app.Post("/register", handler.Register)
		resp, err = app.Test(req)
		require.NoError(t, err)

		assert.Equal(t, fiber.StatusCreated, resp.StatusCode)

		var response Response
		json.NewDecoder(resp.Body).Decode(&response)

		assert.True(t, response.Success)
		assert.NotNil(t, response.Data)
		mockService.AssertExpectations(t)
	})

	t.Run("error with invalid request body", func(t *testing.T) {
		mockService := new(MockUserService)
		handler := NewUserHandler(mockService)
		app := setupTestApp()
		app.Post("/register", handler.Register)

		req := httptest.NewRequest("POST", "/register", bytes.NewReader([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

		var response Response
		json.NewDecoder(resp.Body).Decode(&response)

		assert.False(t, response.Success)
		assert.NotEmpty(t, response.Error)
		mockService.AssertExpectations(t)
	})

	t.Run("error with invalid email", func(t *testing.T) {
		mockService := new(MockUserService)
		handler := NewUserHandler(mockService)
		app := setupTestApp()
		app.Post("/register", handler.Register)

		requestBody := map[string]string{
			"email":    "invalid-email",
			"password": "password123",
			"name":     "Test User",
			"phone":    "1234567890",
		}
		body, _ := json.Marshal(requestBody)

		req := httptest.NewRequest("POST", "/register", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

		var response Response
		json.NewDecoder(resp.Body).Decode(&response)

		assert.False(t, response.Success)
		mockService.AssertExpectations(t)
	})

	t.Run("error when email already exists", func(t *testing.T) {
		mockService := new(MockUserService)
		handler := NewUserHandler(mockService)
		app := setupTestApp()
		app.Post("/register", handler.Register)

		mockService.On("Register", mock.Anything, "existing@example.com", "password123", "Test User", "1234567890").
			Return(nil, service.ErrEmailAlreadyExists)

		requestBody := map[string]string{
			"email":    "existing@example.com",
			"password": "password123",
			"name":     "Test User",
			"phone":    "1234567890",
		}
		body, _ := json.Marshal(requestBody)

		req := httptest.NewRequest("POST", "/register", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, fiber.StatusConflict, resp.StatusCode)

		var response Response
		json.NewDecoder(resp.Body).Decode(&response)

		assert.False(t, response.Success)
		assert.Contains(t, response.Error, "already exists")
		mockService.AssertExpectations(t)
	})

	t.Run("error with weak password", func(t *testing.T) {
		mockService := new(MockUserService)
		handler := NewUserHandler(mockService)
		app := setupTestApp()
		app.Post("/register", handler.Register)

		// Password "weak" has only 4 characters, validator requires min=8
		// So the service will never be called, validation fails first
		requestBody := map[string]string{
			"email":    "test@example.com",
			"password": "weak",
			"name":     "Test User",
			"phone":    "1234567890",
		}
		body, _ := json.Marshal(requestBody)

		req := httptest.NewRequest("POST", "/register", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

		var response Response
		json.NewDecoder(resp.Body).Decode(&response)

		assert.False(t, response.Success)
		assert.Contains(t, response.Error, "Password")
		mockService.AssertExpectations(t)
	})
}

func TestLoginHandler(t *testing.T) {
	t.Run("successfully login with valid credentials", func(t *testing.T) {
		mockService := new(MockUserService)
		handler := NewUserHandler(mockService)
		app := setupTestApp()
		app.Post("/login", handler.Login)

		userID := uuid.New().String()
		token := "valid-jwt-token"
		user := &model.User{
			ID:    uuid.MustParse(userID),
			Email: "test@example.com",
			Name:  "Test User",
			Phone: "1234567890",
		}

		mockService.On("Login", mock.Anything, "test@example.com", "password123").
			Return(token, nil)
		mockService.On("ValidateToken", token).Return(userID, nil)
		mockService.On("GetProfile", mock.Anything, userID).Return(user, nil)

		requestBody := map[string]string{
			"email":    "test@example.com",
			"password": "password123",
		}
		body, _ := json.Marshal(requestBody)

		req := httptest.NewRequest("POST", "/login", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		var response Response
		json.NewDecoder(resp.Body).Decode(&response)

		assert.True(t, response.Success)
		assert.NotNil(t, response.Data)
		mockService.AssertExpectations(t)
	})

	t.Run("error with invalid credentials", func(t *testing.T) {
		mockService := new(MockUserService)
		handler := NewUserHandler(mockService)
		app := setupTestApp()
		app.Post("/login", handler.Login)

		mockService.On("Login", mock.Anything, "test@example.com", "wrongpassword").
			Return("", service.ErrInvalidCredentials)

		requestBody := map[string]string{
			"email":    "test@example.com",
			"password": "wrongpassword",
		}
		body, _ := json.Marshal(requestBody)

		req := httptest.NewRequest("POST", "/login", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)

		var response Response
		json.NewDecoder(resp.Body).Decode(&response)

		assert.False(t, response.Success)
		mockService.AssertExpectations(t)
	})

	t.Run("error with invalid request body", func(t *testing.T) {
		mockService := new(MockUserService)
		handler := NewUserHandler(mockService)
		app := setupTestApp()
		app.Post("/login", handler.Login)

		req := httptest.NewRequest("POST", "/login", bytes.NewReader([]byte("invalid")))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	})
}

func TestGetProfileHandler(t *testing.T) {
	t.Run("successfully get profile with valid token", func(t *testing.T) {
		mockService := new(MockUserService)
		handler := NewUserHandler(mockService)
		app := setupTestApp()

		userID := uuid.New().String()
		user := &model.User{
			ID:    uuid.MustParse(userID),
			Email: "test@example.com",
			Name:  "Test User",
			Phone: "1234567890",
		}

		mockService.On("GetProfile", mock.Anything, userID).Return(user, nil)

		app.Get("/profile", func(c *fiber.Ctx) error {
			c.Locals("userID", userID)
			return handler.GetProfile(c)
		})

		req := httptest.NewRequest("GET", "/profile", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		var response Response
		json.NewDecoder(resp.Body).Decode(&response)

		assert.True(t, response.Success)
		assert.NotNil(t, response.Data)
		mockService.AssertExpectations(t)
	})

	t.Run("error when user ID not in context", func(t *testing.T) {
		mockService := new(MockUserService)
		handler := NewUserHandler(mockService)
		app := setupTestApp()
		app.Get("/profile", handler.GetProfile)

		req := httptest.NewRequest("GET", "/profile", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("error when user not found", func(t *testing.T) {
		mockService := new(MockUserService)
		handler := NewUserHandler(mockService)
		app := setupTestApp()

		userID := uuid.New().String()
		mockService.On("GetProfile", mock.Anything, userID).
			Return(nil, service.ErrUserNotFound)

		app.Get("/profile", func(c *fiber.Ctx) error {
			c.Locals("userID", userID)
			return handler.GetProfile(c)
		})

		req := httptest.NewRequest("GET", "/profile", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)
		mockService.AssertExpectations(t)
	})
}

func TestUpdateProfileHandler(t *testing.T) {
	t.Run("successfully update profile", func(t *testing.T) {
		mockService := new(MockUserService)
		handler := NewUserHandler(mockService)
		app := setupTestApp()

		userID := uuid.New().String()
		updatedUser := &model.User{
			ID:    uuid.MustParse(userID),
			Email: "test@example.com",
			Name:  "Updated Name",
			Phone: "9876543210",
		}

		mockService.On("UpdateProfile", mock.Anything, userID, "Updated Name", "9876543210").
			Return(updatedUser, nil)

		app.Put("/profile", func(c *fiber.Ctx) error {
			c.Locals("userID", userID)
			return handler.UpdateProfile(c)
		})

		requestBody := map[string]string{
			"name":  "Updated Name",
			"phone": "9876543210",
		}
		body, _ := json.Marshal(requestBody)

		req := httptest.NewRequest("PUT", "/profile", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		var response Response
		json.NewDecoder(resp.Body).Decode(&response)

		assert.True(t, response.Success)
		mockService.AssertExpectations(t)
	})

	t.Run("error when user ID not in context", func(t *testing.T) {
		mockService := new(MockUserService)
		handler := NewUserHandler(mockService)
		app := setupTestApp()
		app.Put("/profile", handler.UpdateProfile)

		requestBody := map[string]string{
			"name": "Updated Name",
		}
		body, _ := json.Marshal(requestBody)

		req := httptest.NewRequest("PUT", "/profile", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("error with invalid request body", func(t *testing.T) {
		mockService := new(MockUserService)
		handler := NewUserHandler(mockService)
		app := setupTestApp()

		userID := uuid.New().String()

		app.Put("/profile", func(c *fiber.Ctx) error {
			c.Locals("userID", userID)
			return handler.UpdateProfile(c)
		})

		req := httptest.NewRequest("PUT", "/profile", bytes.NewReader([]byte("invalid")))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	})

	t.Run("error when no fields provided", func(t *testing.T) {
		mockService := new(MockUserService)
		handler := NewUserHandler(mockService)
		app := setupTestApp()

		userID := uuid.New().String()

		app.Put("/profile", func(c *fiber.Ctx) error {
			c.Locals("userID", userID)
			return handler.UpdateProfile(c)
		})

		requestBody := map[string]string{}
		body, _ := json.Marshal(requestBody)

		req := httptest.NewRequest("PUT", "/profile", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	})

	t.Run("error when user not found", func(t *testing.T) {
		mockService := new(MockUserService)
		handler := NewUserHandler(mockService)
		app := setupTestApp()

		userID := uuid.New().String()

		mockService.On("UpdateProfile", mock.Anything, userID, "New Name", "").
			Return(nil, service.ErrUserNotFound)

		app.Put("/profile", func(c *fiber.Ctx) error {
			c.Locals("userID", userID)
			return handler.UpdateProfile(c)
		})

		requestBody := map[string]string{
			"name": "New Name",
		}
		body, _ := json.Marshal(requestBody)

		req := httptest.NewRequest("PUT", "/profile", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)
		mockService.AssertExpectations(t)
	})

	t.Run("error for internal server error", func(t *testing.T) {
		mockService := new(MockUserService)
		handler := NewUserHandler(mockService)
		app := setupTestApp()

		userID := uuid.New().String()

		mockService.On("UpdateProfile", mock.Anything, userID, "New Name", "").
			Return(nil, errors.New("database error"))

		app.Put("/profile", func(c *fiber.Ctx) error {
			c.Locals("userID", userID)
			return handler.UpdateProfile(c)
		})

		requestBody := map[string]string{
			"name": "New Name",
		}
		body, _ := json.Marshal(requestBody)

		req := httptest.NewRequest("PUT", "/profile", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
		mockService.AssertExpectations(t)
	})
}
