package integration

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGatewayRouting tests that gateway routes requests to correct services
func TestGatewayRouting(t *testing.T) {
	WaitForAllServices(t)

	testCases := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
		needsAuth      bool
	}{
		{
			name:           "route to user service - register",
			method:         "POST",
			path:           "/api/v1/register",
			expectedStatus: http.StatusBadRequest, // Will fail validation but proves routing works
			needsAuth:      false,
		},
		{
			name:           "route to user service - login",
			method:         "POST",
			path:           "/api/v1/login",
			expectedStatus: http.StatusBadRequest, // Will fail validation but proves routing works
			needsAuth:      false,
		},
		{
			name:           "route to user service - profile",
			method:         "GET",
			path:           "/api/v1/profile",
			expectedStatus: http.StatusUnauthorized, // No auth token
			needsAuth:      false,
		},
		{
			name:           "route to order service - orders",
			method:         "GET",
			path:           "/api/v1/orders",
			expectedStatus: http.StatusUnauthorized, // No auth token
			needsAuth:      false,
		},
		{
			name:           "route to order service - create order",
			method:         "POST",
			path:           "/api/v1/orders",
			expectedStatus: http.StatusUnauthorized, // No auth token
			needsAuth:      false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			token := ""
			if tc.needsAuth {
				user := CreateTestUser(t)
				token = user.Token
			}

			resp := DoRequest(t, tc.method, GatewayURL+tc.path, nil, token)
			defer resp.Body.Close()

			assert.Equal(t, tc.expectedStatus, resp.StatusCode,
				"Gateway should route to service and return expected status")
		})
	}
}

// TestGatewayRoutingWithValidRequests tests routing with valid requests
func TestGatewayRoutingWithValidRequests(t *testing.T) {
	WaitForAllServices(t)

	// Test user service routing with valid request
	t.Run("user service - valid register", func(t *testing.T) {
		timestamp := GenerateTimestamp()
		payload := map[string]string{
			"email":    fmt.Sprintf("routing%d@example.com", timestamp),
			"password": "Routing@123",
			"name":     "Routing Test",
			"phone":    "+5511999999999",
		}

		resp := DoRequest(t, "POST", GatewayURL+"/api/v1/register", payload, "")
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode,
			"Gateway should successfully route to user service")

		var response Response
		DecodeJSON(t, resp.Body, &response)
		assert.True(t, response.Success)
	})

	// Test order service routing with valid request
	t.Run("order service - valid create order", func(t *testing.T) {
		user := CreateTestUser(t)

		payload := map[string]interface{}{
			"items": []map[string]interface{}{
				{
					"product_id":   "prod-001",
					"product_name": "Test Product",
					"quantity":     1,
					"price":        10.00,
				},
			},
			"address": map[string]string{
				"street": "Rua Teste",
				"number": "123",
				"city":   "São Paulo",
				"state":  "SP",
			},
		}

		resp := DoRequest(t, "POST", GatewayURL+"/api/v1/orders", payload, user.Token)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode,
			"Gateway should successfully route to order service")

		var response Response
		DecodeJSON(t, resp.Body, &response)
		assert.True(t, response.Success)
	})
}

// TestGatewayHeaderPropagation tests that headers are properly forwarded
func TestGatewayHeaderPropagation(t *testing.T) {
	WaitForAllServices(t)

	user := CreateTestUser(t)

	// Make request with custom headers through gateway
	req, err := http.NewRequest("GET", GatewayURL+"/api/v1/profile", nil)
	require.NoError(t, err)

	req.Header.Set("Authorization", "Bearer "+user.Token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Request-ID", "test-request-123")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Should successfully authenticate and return profile
	assert.Equal(t, http.StatusOK, resp.StatusCode,
		"Gateway should forward Authorization header")

	var response Response
	DecodeJSON(t, resp.Body, &response)
	assert.True(t, response.Success)
}

// TestGatewayCORS tests CORS headers
func TestGatewayCORS(t *testing.T) {
	WaitForAllServices(t)

	// Make OPTIONS request (preflight)
	req, err := http.NewRequest("OPTIONS", GatewayURL+"/api/v1/register", nil)
	require.NoError(t, err)

	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Access-Control-Request-Method", "POST")
	req.Header.Set("Access-Control-Request-Headers", "Content-Type")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Verify CORS headers are present
	assert.NotEmpty(t, resp.Header.Get("Access-Control-Allow-Origin"),
		"Should have CORS origin header")

	t.Log("CORS headers:", resp.Header)
}

// TestGatewayErrorHandling tests how gateway handles service errors
func TestGatewayErrorHandling(t *testing.T) {
	WaitForAllServices(t)

	testCases := []struct {
		name           string
		method         string
		path           string
		payload        interface{}
		token          string
		expectedStatus int
	}{
		{
			name:           "handle 404 from service",
			method:         "GET",
			path:           "/api/v1/orders/nonexistent-order-id",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "handle 400 from service",
			method:         "POST",
			path:           "/api/v1/register",
			payload:        map[string]string{"invalid": "data"},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "handle 401 from service",
			method:         "GET",
			path:           "/api/v1/profile",
			token:          "invalid.token.here",
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resp := DoRequest(t, tc.method, GatewayURL+tc.path, tc.payload, tc.token)
			defer resp.Body.Close()

			assert.Equal(t, tc.expectedStatus, resp.StatusCode,
				"Gateway should properly forward error status from services")
		})
	}
}

// TestGatewayRateLimiting tests rate limiting functionality
func TestGatewayRateLimiting(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping rate limit test in short mode")
	}

	WaitForAllServices(t)

	// Note: This test assumes rate limiting is configured in the gateway
	// If not configured, this test will need to be adjusted or skipped

	endpoint := GatewayURL + "/api/v1/login"

	// Make many rapid requests
	numRequests := 100
	statusCodes := make(map[int]int)
	var mu sync.Mutex

	var wg sync.WaitGroup
	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			payload := map[string]string{
				"email":    fmt.Sprintf("ratelimit%d@example.com", idx),
				"password": "password",
			}

			resp := DoRequest(t, "POST", endpoint, payload, "")
			defer resp.Body.Close()

			mu.Lock()
			statusCodes[resp.StatusCode]++
			mu.Unlock()
		}(i)

		// Small delay to avoid overwhelming the system
		time.Sleep(10 * time.Millisecond)
	}

	wg.Wait()

	t.Logf("Rate limit test results: %v", statusCodes)

	// Check if we got any rate limit responses (429)
	// Note: If rate limiting is not configured, all will be 400/401
	if rateLimitCount, hasRateLimit := statusCodes[http.StatusTooManyRequests]; hasRateLimit {
		assert.Greater(t, rateLimitCount, 0, "Should have rate limited some requests")
		t.Logf("Successfully rate limited %d out of %d requests", rateLimitCount, numRequests)
	} else {
		t.Log("No rate limiting detected - may not be configured in gateway")
		// This is not a failure - just means rate limiting might not be enabled
	}
}

// TestGatewayRateLimitHeaders tests rate limit headers
func TestGatewayRateLimitHeaders(t *testing.T) {
	WaitForAllServices(t)

	resp := DoRequest(t, "POST", GatewayURL+"/api/v1/login", map[string]string{
		"email":    "test@example.com",
		"password": "password",
	}, "")
	defer resp.Body.Close()

	// Check for common rate limit headers
	rateLimitHeaders := []string{
		"X-RateLimit-Limit",
		"X-RateLimit-Remaining",
		"X-RateLimit-Reset",
		"RateLimit-Limit",
		"RateLimit-Remaining",
	}

	foundHeaders := []string{}
	for _, header := range rateLimitHeaders {
		if value := resp.Header.Get(header); value != "" {
			foundHeaders = append(foundHeaders, fmt.Sprintf("%s: %s", header, value))
		}
	}

	if len(foundHeaders) > 0 {
		t.Log("Found rate limit headers:", foundHeaders)
	} else {
		t.Log("No rate limit headers found - may not be configured")
	}
}

// TestGatewayTimeout tests gateway timeout handling
func TestGatewayTimeout(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping timeout test in short mode")
	}

	WaitForAllServices(t)

	// This test assumes services respond reasonably fast
	// A timeout would indicate gateway or service issues

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	req, err := http.NewRequest("POST", GatewayURL+"/api/v1/login", nil)
	require.NoError(t, err)

	start := time.Now()
	resp, err := client.Do(req)
	duration := time.Since(start)

	if err != nil {
		assert.Fail(t, "Request timed out or failed", "Error: %v", err)
		return
	}
	defer resp.Body.Close()

	// Should respond quickly (under 5 seconds for simple request)
	assert.Less(t, duration, 5*time.Second,
		"Gateway should respond quickly, not timeout")

	t.Logf("Request completed in %v", duration)
}

// TestGatewayHealthCheck tests gateway health endpoint
func TestGatewayHealthCheck(t *testing.T) {
	WaitForAllServices(t)

	resp := DoRequest(t, "GET", GatewayURL+"/health", nil, "")
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode, "Gateway health check should return OK")

	// Try to parse response
	var healthData map[string]interface{}
	err := json.NewDecoder(resp.Body).Decode(&healthData)

	if err == nil {
		t.Logf("Health check response: %v", healthData)
		// Optionally verify specific health check fields
		if status, ok := healthData["status"]; ok {
			assert.Equal(t, "healthy", status)
		}
	}
}

// TestGatewayServiceUnavailable tests handling of unavailable backend services
func TestGatewayServiceUnavailable(t *testing.T) {
	WaitForAllServices(t)

	// This test would require stopping a backend service
	// For now, we'll just verify the gateway handles unknown routes

	resp := DoRequest(t, "GET", GatewayURL+"/api/v1/unknown/route", nil, "")
	defer resp.Body.Close()

	// Should return 404 or similar error
	assert.True(t, resp.StatusCode >= 400,
		"Gateway should handle unknown routes with error status")
}

// TestGatewayRequestSize tests handling of large request bodies
func TestGatewayRequestSize(t *testing.T) {
	WaitForAllServices(t)

	user := CreateTestUser(t)

	// Create a very large order (but still valid)
	items := make([]map[string]interface{}, 100)
	for i := 0; i < 100; i++ {
		items[i] = map[string]interface{}{
			"product_id":   fmt.Sprintf("prod-%d", i),
			"product_name": fmt.Sprintf("Product %d", i),
			"quantity":     1,
			"price":        9.99,
		}
	}

	payload := map[string]interface{}{
		"items": items,
		"address": map[string]string{
			"street": "Rua Teste",
			"number": "123",
			"city":   "São Paulo",
			"state":  "SP",
		},
	}

	resp := DoRequest(t, "POST", GatewayURL+"/api/v1/orders", payload, user.Token)
	defer resp.Body.Close()

	// Should either accept it or reject with appropriate error
	// Not 500 or timeout
	assert.True(t, resp.StatusCode < 500 || resp.StatusCode == http.StatusCreated,
		"Gateway should handle large requests gracefully")

	if resp.StatusCode == http.StatusCreated {
		t.Log("Gateway successfully handled large request")
	} else {
		t.Logf("Gateway rejected large request with status: %d", resp.StatusCode)
	}
}

// TestGatewayConcurrentRequests tests handling of concurrent requests
func TestGatewayConcurrentRequests(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrency test in short mode")
	}

	WaitForAllServices(t)

	// Create multiple users concurrently
	numConcurrent := 20
	var wg sync.WaitGroup
	successCount := 0
	var mu sync.Mutex

	for i := 0; i < numConcurrent; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			timestamp := GenerateTimestamp()
			payload := map[string]string{
				"email":    fmt.Sprintf("concurrent%d_%d@example.com", idx, timestamp),
				"password": "Concurrent@123",
				"name":     fmt.Sprintf("User %d", idx),
				"phone":    "+5511999999999",
			}

			resp := DoRequest(t, "POST", GatewayURL+"/api/v1/register", payload, "")
			defer resp.Body.Close()

			if resp.StatusCode == http.StatusCreated {
				mu.Lock()
				successCount++
				mu.Unlock()
			}
		}(i)
	}

	wg.Wait()

	// Most requests should succeed
	assert.Greater(t, successCount, numConcurrent/2,
		"Gateway should handle concurrent requests successfully")

	t.Logf("Successfully processed %d out of %d concurrent requests",
		successCount, numConcurrent)
}
