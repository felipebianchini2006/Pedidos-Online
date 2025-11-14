package middleware

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDefaultCORSMiddleware(t *testing.T) {
	t.Run("set CORS headers with wildcard origin", func(t *testing.T) {
		app := fiber.New()
		app.Use(NewDefaultCORSMiddleware([]string{"*"}))
		app.Get("/test", func(c *fiber.Ctx) error {
			return c.SendString("OK")
		})

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Origin", "http://localhost:3000")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, fiber.StatusOK, resp.StatusCode)
		assert.NotEmpty(t, resp.Header.Get("Access-Control-Allow-Origin"))
	})

	t.Run("handle preflight OPTIONS request", func(t *testing.T) {
		app := fiber.New()
		app.Use(NewDefaultCORSMiddleware([]string{"http://localhost:3000"}))
		app.Post("/test", func(c *fiber.Ctx) error {
			return c.SendString("OK")
		})

		req := httptest.NewRequest("OPTIONS", "/test", nil)
		req.Header.Set("Origin", "http://localhost:3000")
		req.Header.Set("Access-Control-Request-Method", "POST")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, fiber.StatusNoContent, resp.StatusCode)
		assert.NotEmpty(t, resp.Header.Get("Access-Control-Allow-Methods"))
		assert.NotEmpty(t, resp.Header.Get("Access-Control-Allow-Headers"))
	})

	t.Run("allow specific origins", func(t *testing.T) {
		allowedOrigins := []string{"http://localhost:3000", "http://localhost:5173"}
		app := fiber.New()
		app.Use(NewDefaultCORSMiddleware(allowedOrigins))
		app.Get("/test", func(c *fiber.Ctx) error {
			return c.SendString("OK")
		})

		for _, origin := range allowedOrigins {
			req := httptest.NewRequest("GET", "/test", nil)
			req.Header.Set("Origin", origin)

			resp, err := app.Test(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, fiber.StatusOK, resp.StatusCode)
		}
	})
}

func TestNewCORSMiddleware(t *testing.T) {
	t.Run("use custom configuration", func(t *testing.T) {
		config := CORSConfig{
			AllowedOrigins:   []string{"https://example.com"},
			AllowedMethods:   []string{"GET", "POST"},
			AllowedHeaders:   []string{"Authorization", "Content-Type"},
			AllowCredentials: true,
			MaxAge:           7200,
		}

		app := fiber.New()
		app.Use(NewCORSMiddleware(config))
		app.Get("/test", func(c *fiber.Ctx) error {
			return c.SendString("OK")
		})

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Origin", "https://example.com")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	})

	t.Run("use default values when not specified", func(t *testing.T) {
		config := CORSConfig{
			AllowedOrigins: []string{"*"},
		}

		app := fiber.New()
		app.Use(NewCORSMiddleware(config))
		app.Get("/test", func(c *fiber.Ctx) error {
			return c.SendString("OK")
		})

		req := httptest.NewRequest("GET", "/test", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	})

	t.Run("set max age header", func(t *testing.T) {
		config := CORSConfig{
			AllowedOrigins: []string{"*"},
			MaxAge:         3600,
		}

		app := fiber.New()
		app.Use(NewCORSMiddleware(config))
		app.Get("/test", func(c *fiber.Ctx) error {
			return c.SendString("OK")
		})

		req := httptest.NewRequest("OPTIONS", "/test", nil)
		req.Header.Set("Origin", "http://localhost:3000")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		maxAge := resp.Header.Get("Access-Control-Max-Age")
		assert.NotEmpty(t, maxAge)
	})
}
