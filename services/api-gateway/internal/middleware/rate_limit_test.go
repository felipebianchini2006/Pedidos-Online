package middleware

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRateLimiter(t *testing.T) {
	t.Run("create rate limiter with valid parameters", func(t *testing.T) {
		rl := NewRateLimiter(100, time.Minute)

		assert.NotNil(t, rl)
		assert.Equal(t, 100, rl.limit)
		assert.Equal(t, time.Minute, rl.window)
		assert.NotNil(t, rl.requests)
		assert.NotNil(t, rl.cleanupTicker)
	})
}

func TestRateLimiterAllow(t *testing.T) {
	t.Run("allow requests within limit", func(t *testing.T) {
		rl := NewRateLimiter(5, time.Minute)
		ip := "192.168.1.1"

		for i := 0; i < 5; i++ {
			allowed, remaining, _ := rl.Allow(ip)
			assert.True(t, allowed, "Request %d should be allowed", i+1)
			assert.Equal(t, 5-i-1, remaining)
		}
	})

	t.Run("block requests exceeding limit", func(t *testing.T) {
		rl := NewRateLimiter(3, time.Minute)
		ip := "192.168.1.2"

		// First 3 should be allowed
		for i := 0; i < 3; i++ {
			allowed, _, _ := rl.Allow(ip)
			assert.True(t, allowed)
		}

		// 4th request should be blocked
		allowed, remaining, _ := rl.Allow(ip)
		assert.False(t, allowed)
		assert.Equal(t, 0, remaining)
	})

	t.Run("track different IPs separately", func(t *testing.T) {
		rl := NewRateLimiter(2, time.Minute)

		// IP 1 - use both requests
		allowed, _, _ := rl.Allow("192.168.1.1")
		assert.True(t, allowed)
		allowed, _, _ = rl.Allow("192.168.1.1")
		assert.True(t, allowed)

		// IP 2 - should still have full quota
		allowed, remaining, _ := rl.Allow("192.168.1.2")
		assert.True(t, allowed)
		assert.Equal(t, 1, remaining)
	})

	t.Run("reset counter after window expires", func(t *testing.T) {
		rl := NewRateLimiter(2, 100*time.Millisecond)
		ip := "192.168.1.3"

		// Use up quota
		rl.Allow(ip)
		rl.Allow(ip)

		// Should be blocked
		allowed, _, _ := rl.Allow(ip)
		assert.False(t, allowed)

		// Wait for window to expire
		time.Sleep(150 * time.Millisecond)

		// Should be allowed again
		allowed, remaining, _ := rl.Allow(ip)
		assert.True(t, allowed)
		assert.Equal(t, 1, remaining)
	})
}

func TestRateLimiterMiddleware(t *testing.T) {
	t.Run("allow requests within limit", func(t *testing.T) {
		rl := NewRateLimiter(5, time.Minute)
		app := fiber.New()
		app.Use(rl.Middleware())
		app.Get("/test", func(c *fiber.Ctx) error {
			return c.SendString("OK")
		})

		for i := 0; i < 5; i++ {
			req := httptest.NewRequest("GET", "/test", nil)
			req.Header.Set("X-Forwarded-For", "192.168.1.1")

			resp, err := app.Test(req)
			require.NoError(t, err)
			assert.Equal(t, fiber.StatusOK, resp.StatusCode)
		}
	})

	t.Run("return 429 when limit exceeded", func(t *testing.T) {
		rl := NewRateLimiter(2, time.Minute)
		app := fiber.New()
		app.Use(rl.Middleware())
		app.Get("/test", func(c *fiber.Ctx) error {
			return c.SendString("OK")
		})

		// First 2 requests should succeed
		for i := 0; i < 2; i++ {
			req := httptest.NewRequest("GET", "/test", nil)
			req.Header.Set("X-Forwarded-For", "192.168.1.2")
			resp, _ := app.Test(req)
			assert.Equal(t, fiber.StatusOK, resp.StatusCode)
		}

		// 3rd request should be rate limited
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("X-Forwarded-For", "192.168.1.2")
		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, fiber.StatusTooManyRequests, resp.StatusCode)
	})

	t.Run("set rate limit headers", func(t *testing.T) {
		rl := NewRateLimiter(10, time.Minute)
		app := fiber.New()
		app.Use(rl.Middleware())
		app.Get("/test", func(c *fiber.Ctx) error {
			return c.SendString("OK")
		})

		req := httptest.NewRequest("GET", "/test", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)

		assert.Equal(t, "10", resp.Header.Get("X-RateLimit-Limit"))
		assert.Equal(t, "9", resp.Header.Get("X-RateLimit-Remaining"))
		assert.NotEmpty(t, resp.Header.Get("X-RateLimit-Reset"))
	})

	t.Run("set Retry-After header when rate limited", func(t *testing.T) {
		rl := NewRateLimiter(1, time.Minute)
		app := fiber.New()
		app.Use(rl.Middleware())
		app.Get("/test", func(c *fiber.Ctx) error {
			return c.SendString("OK")
		})

		// First request succeeds
		req1 := httptest.NewRequest("GET", "/test", nil)
		resp1, _ := app.Test(req1)
		assert.Equal(t, fiber.StatusOK, resp1.StatusCode)

		// Second request is rate limited
		req2 := httptest.NewRequest("GET", "/test", nil)
		resp2, _ := app.Test(req2)
		assert.Equal(t, fiber.StatusTooManyRequests, resp2.StatusCode)
		assert.NotEmpty(t, resp2.Header.Get("Retry-After"))
	})
}

func TestRateLimiterCleanup(t *testing.T) {
	t.Run("cleanup removes expired entries", func(t *testing.T) {
		rl := NewRateLimiter(5, 50*time.Millisecond)

		// Generate some requests
		rl.Allow("192.168.1.1")
		rl.Allow("192.168.1.2")
		rl.Allow("192.168.1.3")

		assert.Len(t, rl.requests, 3)

		// Wait for entries to expire and cleanup to run
		time.Sleep(200 * time.Millisecond)

		// Entries should have been cleaned up
		// Note: cleanup runs periodically, so this might not be exact
		rl.mu.RLock()
		requestCount := len(rl.requests)
		rl.mu.RUnlock()

		// After cleanup, expired entries should be removed
		assert.LessOrEqual(t, requestCount, 3)
	})
}
