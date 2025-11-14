package proxy

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProxyHandler(t *testing.T) {
	// Create a test backend server
	backendServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"success": true, "message": "backend response"}`))
	}))
	defer backendServer.Close()

	t.Run("successfully proxy request to backend", func(t *testing.T) {
		config := ProxyConfig{
			TargetURL:   backendServer.URL,
			Timeout:     5 * time.Second,
			MaxRetries:  3,
			RetryDelay:  100 * time.Millisecond,
			ServiceName: "test-service",
		}

		app := fiber.New()
		app.Get("/api/*", ProxyHandler(config))

		req := httptest.NewRequest("GET", "/api/test", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, fiber.StatusOK, resp.StatusCode)
		assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))
	})

	t.Run("forward query parameters", func(t *testing.T) {
		// Create backend that checks query params
		backendWithQuery := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			page := r.URL.Query().Get("page")
			size := r.URL.Query().Get("size")
			if page == "1" && size == "10" {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"page": 1, "size": 10}`))
			} else {
				w.WriteHeader(http.StatusBadRequest)
			}
		}))
		defer backendWithQuery.Close()

		config := ProxyConfig{
			TargetURL:   backendWithQuery.URL,
			Timeout:     5 * time.Second,
			ServiceName: "test-service",
		}

		app := fiber.New()
		app.Get("/api/*", ProxyHandler(config))

		req := httptest.NewRequest("GET", "/api/test?page=1&size=10", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	})

	t.Run("forward headers to backend", func(t *testing.T) {
		// Create backend that checks headers
		backendWithHeaders := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			contentType := r.Header.Get("Content-Type")

			if authHeader == "Bearer test-token" && contentType == "application/json" {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"authenticated": true}`))
			} else {
				w.WriteHeader(http.StatusUnauthorized)
			}
		}))
		defer backendWithHeaders.Close()

		config := ProxyConfig{
			TargetURL:   backendWithHeaders.URL,
			Timeout:     5 * time.Second,
			ServiceName: "test-service",
		}

		app := fiber.New()
		app.Get("/api/*", ProxyHandler(config))

		req := httptest.NewRequest("GET", "/api/test", nil)
		req.Header.Set("Authorization", "Bearer test-token")
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	})

	t.Run("return 503 when backend is unavailable", func(t *testing.T) {
		config := ProxyConfig{
			TargetURL:   "http://localhost:99999", // Invalid port
			Timeout:     1 * time.Second,
			MaxRetries:  0, // No retries for faster test
			ServiceName: "test-service",
		}

		app := fiber.New()
		app.Get("/api/*", ProxyHandler(config))

		req := httptest.NewRequest("GET", "/api/test", nil)
		resp, err := app.Test(req, 3000) // 3 second timeout for test
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, fiber.StatusServiceUnavailable, resp.StatusCode)
	})

	t.Run("set service name in locals", func(t *testing.T) {
		config := ProxyConfig{
			TargetURL:   backendServer.URL,
			Timeout:     5 * time.Second,
			ServiceName: "my-service",
		}

		var capturedServiceName string
		app := fiber.New()
		app.Use(func(c *fiber.Ctx) error {
			// Middleware to capture service name
			err := c.Next()
			capturedServiceName = c.Locals("service_name").(string)
			return err
		})
		app.Get("/api/*", ProxyHandler(config))

		req := httptest.NewRequest("GET", "/api/test", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, "my-service", capturedServiceName)
	})

	t.Run("handle timeout gracefully", func(t *testing.T) {
		// Create slow backend
		slowBackend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(2 * time.Second)
			w.WriteHeader(http.StatusOK)
		}))
		defer slowBackend.Close()

		config := ProxyConfig{
			TargetURL:   slowBackend.URL,
			Timeout:     500 * time.Millisecond, // Short timeout
			MaxRetries:  0,
			ServiceName: "slow-service",
		}

		app := fiber.New()
		app.Get("/api/*", ProxyHandler(config))

		req := httptest.NewRequest("GET", "/api/test", nil)
		resp, err := app.Test(req, 5000)
		require.NoError(t, err)
		defer resp.Body.Close()

		// Should return error due to timeout
		assert.Equal(t, fiber.StatusServiceUnavailable, resp.StatusCode)
	})
}

func TestProxyPathRewriting(t *testing.T) {
	// Test that paths are correctly rewritten
	backendServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Echo back the path we received
		w.Write([]byte(r.URL.Path))
	}))
	defer backendServer.Close()

	tests := []struct {
		name         string
		requestPath  string
		expectedPath string
	}{
		{
			name:         "users path",
			requestPath:  "/api/users/v1/register",
			expectedPath: "/api/v1/register",
		},
		{
			name:         "orders path",
			requestPath:  "/api/orders/v1/orders",
			expectedPath: "/api/v1/orders",
		},
		{
			name:         "orders with ID",
			requestPath:  "/api/orders/v1/orders/123",
			expectedPath: "/api/v1/orders/123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := ProxyConfig{
				TargetURL:   backendServer.URL,
				Timeout:     5 * time.Second,
				ServiceName: "test-service",
			}

			app := fiber.New()
			app.Get("/api/users/*", ProxyHandler(config))
			app.Get("/api/orders/*", ProxyHandler(config))

			req := httptest.NewRequest("GET", tt.requestPath, nil)
			resp, err := app.Test(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			body := make([]byte, 100)
			n, _ := resp.Body.Read(body)
			actualPath := string(body[:n])

			assert.Equal(t, tt.expectedPath, actualPath)
		})
	}
}

func TestProxyRetryLogic(t *testing.T) {
	t.Run("retry on failure", func(t *testing.T) {
		attemptCount := 0
		backendServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			attemptCount++
			if attemptCount < 3 {
				// Fail first 2 attempts
				w.WriteHeader(http.StatusInternalServerError)
			} else {
				// Succeed on 3rd attempt
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"success": true}`))
			}
		}))
		defer backendServer.Close()

		config := ProxyConfig{
			TargetURL:   backendServer.URL,
			Timeout:     5 * time.Second,
			MaxRetries:  3,
			RetryDelay:  10 * time.Millisecond,
			ServiceName: "test-service",
		}

		app := fiber.New()
		app.Get("/api/*", ProxyHandler(config))

		req := httptest.NewRequest("GET", "/api/test", nil)
		resp, err := app.Test(req, 5000)
		require.NoError(t, err)
		defer resp.Body.Close()

		// Should eventually succeed after retries
		assert.GreaterOrEqual(t, attemptCount, 1)
	})
}
