package middleware

import (
	"fmt"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"

	"pedidos-online/user-service/pkg/jwt"
)

const testJWTSecret = "test-secret-key-for-testing-only"

// TestAuthMiddleware_Success tests successful authentication
func TestAuthMiddleware_Success(t *testing.T) {
	// Create test token
	userID := "test-user-123"
	email := "test@example.com"
	token, err := jwt.GenerateToken(userID, email, testJWTSecret, time.Hour)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// Create Fiber app and middleware
	app := fiber.New()
	app.Use(AuthMiddleware(testJWTSecret))
	app.Get("/test", func(c *fiber.Ctx) error {
		// Verify user info is in context
		storedUserID, err := GetUserIDFromContext(c)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		storedEmail, err := GetEmailFromContext(c)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		return c.JSON(fiber.Map{
			"userID": storedUserID,
			"email":  storedEmail,
		})
	})

	// Create request
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	// Execute request
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to execute request: %v", err)
	}

	// Verify response
	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

// TestAuthMiddleware_MissingToken tests missing Authorization header
func TestAuthMiddleware_MissingToken(t *testing.T) {
	app := fiber.New()
	app.Use(AuthMiddleware(testJWTSecret))
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	// No Authorization header

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to execute request: %v", err)
	}

	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", resp.StatusCode)
	}
}

// TestAuthMiddleware_InvalidFormat tests invalid Authorization header format
func TestAuthMiddleware_InvalidFormat(t *testing.T) {
	testCases := []struct {
		name   string
		header string
	}{
		{"No Bearer prefix", "invalid-token"},
		{"Wrong prefix", "Basic token123"},
		{"Bearer only", "Bearer "},
		{"Empty after Bearer", "Bearer    "},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			app := fiber.New()
			app.Use(AuthMiddleware(testJWTSecret))
			app.Get("/test", func(c *fiber.Ctx) error {
				return c.SendString("OK")
			})

			req := httptest.NewRequest("GET", "/test", nil)
			req.Header.Set("Authorization", tc.header)

			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("Failed to execute request: %v", err)
			}

			if resp.StatusCode != fiber.StatusUnauthorized {
				t.Errorf("Expected status 401, got %d", resp.StatusCode)
			}
		})
	}
}

// TestAuthMiddleware_InvalidToken tests invalid JWT token
func TestAuthMiddleware_InvalidToken(t *testing.T) {
	testCases := []struct {
		name  string
		token string
	}{
		{"Malformed token", "not.a.valid.token"},
		{"Random string", "random-string-123"},
		{"Wrong signature", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoidGVzdCJ9.invalid"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			app := fiber.New()
			app.Use(AuthMiddleware(testJWTSecret))
			app.Get("/test", func(c *fiber.Ctx) error {
				return c.SendString("OK")
			})

			req := httptest.NewRequest("GET", "/test", nil)
			req.Header.Set("Authorization", "Bearer "+tc.token)

			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("Failed to execute request: %v", err)
			}

			if resp.StatusCode != fiber.StatusUnauthorized {
				t.Errorf("Expected status 401, got %d", resp.StatusCode)
			}
		})
	}
}

// TestAuthMiddleware_ExpiredToken tests expired JWT token
func TestAuthMiddleware_ExpiredToken(t *testing.T) {
	// Create expired token (negative duration)
	userID := "test-user-123"
	email := "test@example.com"
	token, err := jwt.GenerateToken(userID, email, testJWTSecret, -time.Hour)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	app := fiber.New()
	app.Use(AuthMiddleware(testJWTSecret))
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to execute request: %v", err)
	}

	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", resp.StatusCode)
	}
}

// TestGetUserIDFromContext tests extracting userID from context
func TestGetUserIDFromContext(t *testing.T) {
	testCases := []struct {
		name        string
		setupFunc   func(*fiber.Ctx)
		expectError bool
		expectedID  string
	}{
		{
			name: "Valid userID",
			setupFunc: func(c *fiber.Ctx) {
				c.Locals(ContextKeyUserID, "user-123")
			},
			expectError: false,
			expectedID:  "user-123",
		},
		{
			name:        "Missing userID",
			setupFunc:   func(c *fiber.Ctx) {},
			expectError: true,
		},
		{
			name: "Invalid type",
			setupFunc: func(c *fiber.Ctx) {
				c.Locals(ContextKeyUserID, 123) // Wrong type
			},
			expectError: true,
		},
		{
			name: "Empty userID",
			setupFunc: func(c *fiber.Ctx) {
				c.Locals(ContextKeyUserID, "")
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			app := fiber.New()
			app.Get("/test", func(c *fiber.Ctx) error {
				tc.setupFunc(c)

				userID, err := GetUserIDFromContext(c)
				if tc.expectError && err == nil {
					t.Error("Expected error, got nil")
				}
				if !tc.expectError && err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
				if !tc.expectError && userID != tc.expectedID {
					t.Errorf("Expected userID %s, got %s", tc.expectedID, userID)
				}

				return c.SendString("OK")
			})

			req := httptest.NewRequest("GET", "/test", nil)
			_, err := app.Test(req)
			if err != nil {
				t.Fatalf("Failed to execute request: %v", err)
			}
		})
	}
}

// TestGetEmailFromContext tests extracting email from context
func TestGetEmailFromContext(t *testing.T) {
	testCases := []struct {
		name          string
		setupFunc     func(*fiber.Ctx)
		expectError   bool
		expectedEmail string
	}{
		{
			name: "Valid email",
			setupFunc: func(c *fiber.Ctx) {
				c.Locals(ContextKeyEmail, "test@example.com")
			},
			expectError:   false,
			expectedEmail: "test@example.com",
		},
		{
			name:        "Missing email",
			setupFunc:   func(c *fiber.Ctx) {},
			expectError: true,
		},
		{
			name: "Invalid type",
			setupFunc: func(c *fiber.Ctx) {
				c.Locals(ContextKeyEmail, 123) // Wrong type
			},
			expectError: true,
		},
		{
			name: "Empty email (valid)",
			setupFunc: func(c *fiber.Ctx) {
				c.Locals(ContextKeyEmail, "")
			},
			expectError:   false,
			expectedEmail: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			app := fiber.New()
			app.Get("/test", func(c *fiber.Ctx) error {
				tc.setupFunc(c)

				email, err := GetEmailFromContext(c)
				if tc.expectError && err == nil {
					t.Error("Expected error, got nil")
				}
				if !tc.expectError && err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
				if !tc.expectError && email != tc.expectedEmail {
					t.Errorf("Expected email %s, got %s", tc.expectedEmail, email)
				}

				return c.SendString("OK")
			})

			req := httptest.NewRequest("GET", "/test", nil)
			_, err := app.Test(req)
			if err != nil {
				t.Fatalf("Failed to execute request: %v", err)
			}
		})
	}
}

// TestOptionalAuthMiddleware tests optional authentication
func TestOptionalAuthMiddleware(t *testing.T) {
	testCases := []struct {
		name         string
		setupToken   bool
		validToken   bool
		expectUserID bool
	}{
		{"Valid token", true, true, true},
		{"No token", false, false, false},
		{"Invalid token", true, false, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			app := fiber.New()
			app.Use(OptionalAuthMiddleware(testJWTSecret))
			app.Get("/test", func(c *fiber.Ctx) error {
				userID, _ := GetUserIDFromContext(c)
				hasUserID := userID != ""

				if hasUserID != tc.expectUserID {
					t.Errorf("Expected hasUserID=%v, got %v", tc.expectUserID, hasUserID)
				}

				return c.SendString("OK")
			})

			req := httptest.NewRequest("GET", "/test", nil)

			if tc.setupToken {
				var token string
				if tc.validToken {
					token, _ = jwt.GenerateToken("user-123", "test@example.com", testJWTSecret, time.Hour)
				} else {
					token = "invalid-token"
				}
				req.Header.Set("Authorization", "Bearer "+token)
			}

			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("Failed to execute request: %v", err)
			}

			// Optional auth should always return 200 (never fails)
			if resp.StatusCode != fiber.StatusOK {
				t.Errorf("Expected status 200, got %d", resp.StatusCode)
			}
		})
	}
}

// TestRequireUserID tests the RequireUserID middleware
func TestRequireUserID(t *testing.T) {
	testCases := []struct {
		name           string
		setupFunc      func(*fiber.Ctx)
		expectedStatus int
	}{
		{
			name: "With userID",
			setupFunc: func(c *fiber.Ctx) {
				c.Locals(ContextKeyUserID, "user-123")
			},
			expectedStatus: fiber.StatusOK,
		},
		{
			name:           "Without userID",
			setupFunc:      func(c *fiber.Ctx) {},
			expectedStatus: fiber.StatusUnauthorized,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			app := fiber.New()
			app.Get("/test", func(c *fiber.Ctx) error {
				tc.setupFunc(c)
				return c.Next()
			}, RequireUserID(), func(c *fiber.Ctx) error {
				return c.SendString("OK")
			})

			req := httptest.NewRequest("GET", "/test", nil)
			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("Failed to execute request: %v", err)
			}

			if resp.StatusCode != tc.expectedStatus {
				t.Errorf("Expected status %d, got %d", tc.expectedStatus, resp.StatusCode)
			}
		})
	}
}

// TestExtractTokenFromHeader tests token extraction
func TestExtractTokenFromHeader(t *testing.T) {
	testCases := []struct {
		name        string
		header      string
		expectError bool
		expectToken string
	}{
		{"Valid header", "Bearer token123", false, "token123"},
		{"Valid with spaces", "Bearer   token123   ", false, "token123"},
		{"No Bearer prefix", "token123", true, ""},
		{"Wrong prefix", "Basic token123", true, ""},
		{"Empty token", "Bearer ", true, ""},
		{"Empty header", "", true, ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			token, err := extractTokenFromHeader(tc.header)

			if tc.expectError && err == nil {
				t.Error("Expected error, got nil")
			}
			if !tc.expectError && err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}
			if !tc.expectError && token != tc.expectToken {
				t.Errorf("Expected token %s, got %s", tc.expectToken, token)
			}
		})
	}
}

// TestBasicRateLimitMiddleware tests basic rate limiting
func TestBasicRateLimitMiddleware(t *testing.T) {
	// Create rate limit config: 3 requests per 100ms
	config := NewRateLimitConfig(3, 100*time.Millisecond)

	app := fiber.New()
	app.Use(BasicRateLimitMiddleware(config))
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	// Make 3 requests (should succeed)
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Request %d failed: %v", i+1, err)
		}
		if resp.StatusCode != fiber.StatusOK {
			t.Errorf("Request %d: expected status 200, got %d", i+1, resp.StatusCode)
		}
	}

	// 4th request should be rate limited
	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	if resp.StatusCode != fiber.StatusTooManyRequests {
		t.Errorf("Expected status 429, got %d", resp.StatusCode)
	}

	// Wait for window to expire
	time.Sleep(150 * time.Millisecond)

	// Next request should succeed
	req = httptest.NewRequest("GET", "/test", nil)
	resp, err = app.Test(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

// TestAuthMiddleware_MultipleRoutes tests middleware on multiple routes
func TestAuthMiddleware_MultipleRoutes(t *testing.T) {
	// Generate valid token
	token, err := jwt.GenerateToken("user-123", "test@example.com", testJWTSecret, time.Hour)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	app := fiber.New()

	// Public route
	app.Get("/public", func(c *fiber.Ctx) error {
		return c.SendString("public")
	})

	// Protected route group
	protected := app.Group("/protected", AuthMiddleware(testJWTSecret))
	protected.Get("/resource", func(c *fiber.Ctx) error {
		userID, _ := GetUserIDFromContext(c)
		return c.SendString(fmt.Sprintf("protected: %s", userID))
	})

	// Test public route without token
	req := httptest.NewRequest("GET", "/public", nil)
	resp, _ := app.Test(req)
	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("Public route should be accessible without token")
	}

	// Test protected route without token
	req = httptest.NewRequest("GET", "/protected/resource", nil)
	resp, _ = app.Test(req)
	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Errorf("Protected route should require token")
	}

	// Test protected route with token
	req = httptest.NewRequest("GET", "/protected/resource", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, _ = app.Test(req)
	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("Protected route should be accessible with valid token")
	}
}

// BenchmarkAuthMiddleware benchmarks the authentication middleware
func BenchmarkAuthMiddleware(b *testing.B) {
	// Generate token once
	token, _ := jwt.GenerateToken("user-123", "test@example.com", testJWTSecret, time.Hour)

	app := fiber.New()
	app.Use(AuthMiddleware(testJWTSecret))
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		app.Test(req)
	}
}
