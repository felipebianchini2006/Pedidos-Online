package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	// Service URLs
	GatewayURL       = "http://localhost:8000"
	UserServiceURL   = "http://localhost:8001"
	OrderServiceURL  = "http://localhost:8002"
	NotificationURL  = "http://localhost:8003"

	// Timeouts
	ServiceStartTimeout = 60 * time.Second
	HealthCheckInterval = 2 * time.Second
)

// TestUser represents a test user
type TestUser struct {
	ID       string
	Email    string
	Password string
	Name     string
	Phone    string
	Token    string
}

// TestOrder represents a test order
type TestOrder struct {
	ID          string
	UserID      string
	TotalAmount float64
	Status      string
}

// Response represents the standard API response
type Response struct {
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data,omitempty"`
	Error   string          `json:"error,omitempty"`
	Message string          `json:"message,omitempty"`
}

// WaitForService waits for a service to become healthy
func WaitForService(t *testing.T, url string, timeout time.Duration) {
	t.Helper()

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	healthURL := url + "/health"
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		resp, err := client.Get(healthURL)
		if err == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			t.Logf("Service %s is healthy", url)
			return
		}
		if resp != nil {
			resp.Body.Close()
		}

		time.Sleep(HealthCheckInterval)
	}

	t.Fatalf("Service %s did not become healthy within %v", url, timeout)
}

// WaitForAllServices waits for all services to be ready
func WaitForAllServices(t *testing.T) {
	t.Helper()

	services := []string{
		UserServiceURL,
		OrderServiceURL,
		NotificationURL,
		GatewayURL,
	}

	for _, url := range services {
		WaitForService(t, url, ServiceStartTimeout)
	}

	// Additional wait to ensure RabbitMQ connections are established
	time.Sleep(3 * time.Second)
}

// CreateTestUser creates a new test user and returns the user with token
func CreateTestUser(t *testing.T) *TestUser {
	t.Helper()

	timestamp := time.Now().Unix()
	user := &TestUser{
		Email:    fmt.Sprintf("test%d@example.com", timestamp),
		Password: "Test@123456",
		Name:     fmt.Sprintf("Test User %d", timestamp),
		Phone:    "+55119999999999",
	}

	// Register user
	registerPayload := map[string]string{
		"email":    user.Email,
		"password": user.Password,
		"name":     user.Name,
		"phone":    user.Phone,
	}

	resp := DoRequest(t, "POST", GatewayURL+"/api/v1/register", registerPayload, "")
	require.Equal(t, http.StatusCreated, resp.StatusCode, "Failed to register user")

	var registerResponse Response
	DecodeJSON(t, resp.Body, &registerResponse)
	require.True(t, registerResponse.Success, "Register response should be successful")

	var userData map[string]interface{}
	json.Unmarshal(registerResponse.Data, &userData)
	user.ID = userData["id"].(string)

	// Login to get token
	loginPayload := map[string]string{
		"email":    user.Email,
		"password": user.Password,
	}

	resp = DoRequest(t, "POST", GatewayURL+"/api/v1/login", loginPayload, "")
	require.Equal(t, http.StatusOK, resp.StatusCode, "Failed to login")

	var loginResponse Response
	DecodeJSON(t, resp.Body, &loginResponse)
	require.True(t, loginResponse.Success, "Login response should be successful")

	var loginData map[string]interface{}
	json.Unmarshal(loginResponse.Data, &loginData)
	user.Token = loginData["token"].(string)

	t.Logf("Created test user: %s with ID: %s", user.Email, user.ID)

	return user
}

// CreateTestOrder creates a test order for the given user
func CreateTestOrder(t *testing.T, token string) *TestOrder {
	t.Helper()

	orderPayload := map[string]interface{}{
		"items": []map[string]interface{}{
			{
				"product_id":   "prod-001",
				"product_name": "Test Product 1",
				"quantity":     2,
				"price":        29.99,
			},
			{
				"product_id":   "prod-002",
				"product_name": "Test Product 2",
				"quantity":     1,
				"price":        49.99,
			},
		},
		"address": map[string]string{
			"street":     "Rua Teste",
			"number":     "123",
			"city":       "São Paulo",
			"state":      "SP",
			"zip_code":   "01234-567",
			"complement": "Apto 45",
		},
	}

	resp := DoRequest(t, "POST", GatewayURL+"/api/v1/orders", orderPayload, token)
	require.Equal(t, http.StatusCreated, resp.StatusCode, "Failed to create order")

	var orderResponse Response
	DecodeJSON(t, resp.Body, &orderResponse)
	require.True(t, orderResponse.Success, "Order creation should be successful")

	var orderData map[string]interface{}
	json.Unmarshal(orderResponse.Data, &orderData)

	order := &TestOrder{
		ID:          orderData["id"].(string),
		UserID:      orderData["user_id"].(string),
		TotalAmount: orderData["total_amount"].(float64),
		Status:      orderData["status"].(string),
	}

	t.Logf("Created test order: %s for user: %s", order.ID, order.UserID)

	return order
}

// DoRequest performs an HTTP request with optional authentication
func DoRequest(t *testing.T, method, url string, body interface{}, token string) *http.Response {
	t.Helper()

	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		require.NoError(t, err, "Failed to marshal request body")
		bodyReader = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequest(method, url, bodyReader)
	require.NoError(t, err, "Failed to create request")

	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Do(req)
	require.NoError(t, err, "Request failed")

	return resp
}

// DecodeJSON decodes JSON response body into target
func DecodeJSON(t *testing.T, body io.Reader, target interface{}) {
	t.Helper()

	decoder := json.NewDecoder(body)
	err := decoder.Decode(target)
	require.NoError(t, err, "Failed to decode JSON response")
}

// AssertJSON compares expected JSON structure with actual response
func AssertJSON(t *testing.T, expected, actual interface{}) {
	t.Helper()

	expectedJSON, err := json.Marshal(expected)
	require.NoError(t, err, "Failed to marshal expected JSON")

	actualJSON, err := json.Marshal(actual)
	require.NoError(t, err, "Failed to marshal actual JSON")

	assert.JSONEq(t, string(expectedJSON), string(actualJSON))
}

// AssertResponseSuccess asserts that the response is successful
func AssertResponseSuccess(t *testing.T, resp *http.Response, expectedStatus int) *Response {
	t.Helper()

	assert.Equal(t, expectedStatus, resp.StatusCode, "Unexpected status code")

	var response Response
	DecodeJSON(t, resp.Body, &response)
	assert.True(t, response.Success, "Response should be successful")

	return &response
}

// AssertResponseError asserts that the response contains an error
func AssertResponseError(t *testing.T, resp *http.Response, expectedStatus int) *Response {
	t.Helper()

	assert.Equal(t, expectedStatus, resp.StatusCode, "Unexpected status code")

	var response Response
	DecodeJSON(t, resp.Body, &response)
	assert.False(t, response.Success, "Response should not be successful")
	assert.NotEmpty(t, response.Error, "Error message should be present")

	return &response
}

// CleanDatabase cleans all test data from databases
func CleanDatabase(t *testing.T) {
	t.Helper()

	// Note: This is a simplified version. In production, you would:
	// 1. Connect to PostgreSQL and delete test users
	// 2. Connect to MongoDB and delete test orders
	// 3. Purge RabbitMQ queues

	// For now, we'll rely on using unique emails per test run
	// and potentially running tests against a fresh database each time

	t.Log("Database cleanup - relying on test isolation via unique data")
}

// WaitForNotification waits for a notification to be processed
// This is a helper that polls or waits for async notification processing
func WaitForNotification(t *testing.T, timeout time.Duration) {
	t.Helper()

	// Give the notification service time to consume and process the message
	// In a real implementation, you might:
	// 1. Query a notifications table
	// 2. Check SMTP mock for sent emails
	// 3. Monitor RabbitMQ queue depth

	time.Sleep(timeout)
	t.Log("Waited for notification processing")
}

// GenerateTimestamp returns a unique timestamp for test data
func GenerateTimestamp() int64 {
	return time.Now().UnixNano()
}

// AssertContains checks if a string contains a substring
func AssertContains(t *testing.T, haystack, needle string) {
	t.Helper()
	assert.Contains(t, haystack, needle)
}

// AssertNotContains checks if a string does not contain a substring
func AssertNotContains(t *testing.T, haystack, needle string) {
	t.Helper()
	assert.NotContains(t, haystack, needle)
}

// ParseOrderID extracts order ID from response data
func ParseOrderID(t *testing.T, data json.RawMessage) string {
	t.Helper()

	var orderData map[string]interface{}
	err := json.Unmarshal(data, &orderData)
	require.NoError(t, err, "Failed to parse order data")

	id, ok := orderData["id"].(string)
	require.True(t, ok, "Order ID not found in response")

	return id
}

// ParseUserID extracts user ID from response data
func ParseUserID(t *testing.T, data json.RawMessage) string {
	t.Helper()

	var userData map[string]interface{}
	err := json.Unmarshal(data, &userData)
	require.NoError(t, err, "Failed to parse user data")

	id, ok := userData["id"].(string)
	require.True(t, ok, "User ID not found in response")

	return id
}

// ParseToken extracts JWT token from response data
func ParseToken(t *testing.T, data json.RawMessage) string {
	t.Helper()

	var tokenData map[string]interface{}
	err := json.Unmarshal(data, &tokenData)
	require.NoError(t, err, "Failed to parse token data")

	token, ok := tokenData["token"].(string)
	require.True(t, ok, "Token not found in response")

	return token
}
