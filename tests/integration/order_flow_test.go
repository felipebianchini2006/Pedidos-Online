package integration

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCreateOrder tests creating a new order
func TestCreateOrder(t *testing.T) {
	WaitForAllServices(t)

	// Create authenticated user
	user := CreateTestUser(t)

	// Create order
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

	resp := DoRequest(t, "POST", GatewayURL+"/api/v1/orders", orderPayload, user.Token)
	defer resp.Body.Close()

	// Verify status
	apiResp := AssertResponseSuccess(t, resp, http.StatusCreated)

	// Parse order data
	var orderData map[string]interface{}
	err := json.Unmarshal(apiResp.Data, &orderData)
	require.NoError(t, err)

	// Verify order structure
	assert.NotEmpty(t, orderData["id"], "Order ID should be present")
	assert.Equal(t, user.ID, orderData["user_id"], "User ID should match")
	assert.Equal(t, "pending", orderData["status"], "Initial status should be pending")
	assert.Equal(t, 109.97, orderData["total_amount"], "Total should be calculated correctly")

	// Verify items
	items, ok := orderData["items"].([]interface{})
	assert.True(t, ok, "Items should be an array")
	assert.Len(t, items, 2, "Should have 2 items")

	// Verify address
	address, ok := orderData["address"].(map[string]interface{})
	assert.True(t, ok, "Address should be an object")
	assert.Equal(t, "Rua Teste", address["street"])
	assert.Equal(t, "São Paulo", address["city"])

	// Verify timestamps
	assert.NotEmpty(t, orderData["created_at"])
	assert.NotEmpty(t, orderData["updated_at"])
}

// TestCreateOrderWithoutAuth tests creating order without authentication
func TestCreateOrderWithoutAuth(t *testing.T) {
	WaitForAllServices(t)

	orderPayload := map[string]interface{}{
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

	resp := DoRequest(t, "POST", GatewayURL+"/api/v1/orders", orderPayload, "")
	defer resp.Body.Close()

	// Should return unauthorized
	AssertResponseError(t, resp, http.StatusUnauthorized)
}

// TestCreateOrderInvalidData tests validation of order data
func TestCreateOrderInvalidData(t *testing.T) {
	WaitForAllServices(t)

	user := CreateTestUser(t)

	testCases := []struct {
		name    string
		payload map[string]interface{}
	}{
		{
			name: "empty items",
			payload: map[string]interface{}{
				"items": []map[string]interface{}{},
				"address": map[string]string{
					"street": "Rua Teste",
					"number": "123",
					"city":   "São Paulo",
					"state":  "SP",
				},
			},
		},
		{
			name: "missing items",
			payload: map[string]interface{}{
				"address": map[string]string{
					"street": "Rua Teste",
					"number": "123",
					"city":   "São Paulo",
					"state":  "SP",
				},
			},
		},
		{
			name: "missing address",
			payload: map[string]interface{}{
				"items": []map[string]interface{}{
					{
						"product_id":   "prod-001",
						"product_name": "Test",
						"quantity":     1,
						"price":        10.00,
					},
				},
			},
		},
		{
			name: "invalid quantity",
			payload: map[string]interface{}{
				"items": []map[string]interface{}{
					{
						"product_id":   "prod-001",
						"product_name": "Test",
						"quantity":     0,
						"price":        10.00,
					},
				},
				"address": map[string]string{
					"street": "Rua Teste",
					"number": "123",
					"city":   "São Paulo",
					"state":  "SP",
				},
			},
		},
		{
			name: "invalid price",
			payload: map[string]interface{}{
				"items": []map[string]interface{}{
					{
						"product_id":   "prod-001",
						"product_name": "Test",
						"quantity":     1,
						"price":        -10.00,
					},
				},
				"address": map[string]string{
					"street": "Rua Teste",
					"number": "123",
					"city":   "São Paulo",
					"state":  "SP",
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resp := DoRequest(t, "POST", GatewayURL+"/api/v1/orders", tc.payload, user.Token)
			defer resp.Body.Close()

			AssertResponseError(t, resp, http.StatusBadRequest)
		})
	}
}

// TestListOrders tests listing user orders
func TestListOrders(t *testing.T) {
	WaitForAllServices(t)

	// Create user
	user := CreateTestUser(t)

	// Create multiple orders
	order1 := CreateTestOrder(t, user.Token)
	time.Sleep(100 * time.Millisecond)
	order2 := CreateTestOrder(t, user.Token)

	// List orders
	resp := DoRequest(t, "GET", GatewayURL+"/api/v1/orders", nil, user.Token)
	defer resp.Body.Close()

	// Verify success
	apiResp := AssertResponseSuccess(t, resp, http.StatusOK)

	// Parse orders
	var ordersData []map[string]interface{}
	err := json.Unmarshal(apiResp.Data, &ordersData)
	require.NoError(t, err)

	// Verify we have at least our 2 orders
	assert.GreaterOrEqual(t, len(ordersData), 2, "Should have at least 2 orders")

	// Verify order IDs are present
	orderIDs := make(map[string]bool)
	for _, order := range ordersData {
		orderIDs[order["id"].(string)] = true
	}

	assert.True(t, orderIDs[order1.ID], "Order 1 should be in list")
	assert.True(t, orderIDs[order2.ID], "Order 2 should be in list")
}

// TestListOrdersWithoutAuth tests listing orders without authentication
func TestListOrdersWithoutAuth(t *testing.T) {
	WaitForAllServices(t)

	resp := DoRequest(t, "GET", GatewayURL+"/api/v1/orders", nil, "")
	defer resp.Body.Close()

	// Should return unauthorized
	AssertResponseError(t, resp, http.StatusUnauthorized)
}

// TestListOrdersIsolation tests that users can only see their own orders
func TestListOrdersIsolation(t *testing.T) {
	WaitForAllServices(t)

	// Create two users
	user1 := CreateTestUser(t)
	user2 := CreateTestUser(t)

	// Create order for user1
	order1 := CreateTestOrder(t, user1.Token)

	// Create order for user2
	order2 := CreateTestOrder(t, user2.Token)

	// User1 should only see their order
	resp := DoRequest(t, "GET", GatewayURL+"/api/v1/orders", nil, user1.Token)
	defer resp.Body.Close()

	apiResp := AssertResponseSuccess(t, resp, http.StatusOK)

	var user1Orders []map[string]interface{}
	json.Unmarshal(apiResp.Data, &user1Orders)

	// Verify user1 sees their order but not user2's
	hasOrder1 := false
	hasOrder2 := false
	for _, order := range user1Orders {
		if order["id"].(string) == order1.ID {
			hasOrder1 = true
		}
		if order["id"].(string) == order2.ID {
			hasOrder2 = true
		}
	}

	assert.True(t, hasOrder1, "User1 should see their own order")
	assert.False(t, hasOrder2, "User1 should not see user2's order")
}

// TestGetOrderDetails tests fetching order details
func TestGetOrderDetails(t *testing.T) {
	WaitForAllServices(t)

	// Create user and order
	user := CreateTestUser(t)
	order := CreateTestOrder(t, user.Token)

	// Get order details
	url := fmt.Sprintf("%s/api/v1/orders/%s", GatewayURL, order.ID)
	resp := DoRequest(t, "GET", url, nil, user.Token)
	defer resp.Body.Close()

	// Verify success
	apiResp := AssertResponseSuccess(t, resp, http.StatusOK)

	// Parse order data
	var orderData map[string]interface{}
	err := json.Unmarshal(apiResp.Data, &orderData)
	require.NoError(t, err)

	// Verify order details
	assert.Equal(t, order.ID, orderData["id"])
	assert.Equal(t, order.UserID, orderData["user_id"])
	assert.Equal(t, order.Status, orderData["status"])
	assert.Equal(t, order.TotalAmount, orderData["total_amount"])

	// Verify items are present
	items, ok := orderData["items"].([]interface{})
	assert.True(t, ok, "Items should be present")
	assert.Greater(t, len(items), 0, "Should have items")

	// Verify address is present
	address, ok := orderData["address"].(map[string]interface{})
	assert.True(t, ok, "Address should be present")
	assert.NotEmpty(t, address["street"])
}

// TestGetOrderDetailsNotFound tests fetching non-existent order
func TestGetOrderDetailsNotFound(t *testing.T) {
	WaitForAllServices(t)

	user := CreateTestUser(t)

	// Try to get non-existent order
	url := fmt.Sprintf("%s/api/v1/orders/%s", GatewayURL, "507f1f77bcf86cd799439011")
	resp := DoRequest(t, "GET", url, nil, user.Token)
	defer resp.Body.Close()

	// Should return not found
	AssertResponseError(t, resp, http.StatusNotFound)
}

// TestGetOrderDetailsUnauthorized tests access control
func TestGetOrderDetailsUnauthorized(t *testing.T) {
	WaitForAllServices(t)

	// Create two users
	user1 := CreateTestUser(t)
	user2 := CreateTestUser(t)

	// User1 creates an order
	order := CreateTestOrder(t, user1.Token)

	// User2 tries to access user1's order
	url := fmt.Sprintf("%s/api/v1/orders/%s", GatewayURL, order.ID)
	resp := DoRequest(t, "GET", url, nil, user2.Token)
	defer resp.Body.Close()

	// Should return forbidden or not found
	assert.True(t, resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusNotFound,
		"Should deny access to other user's order")
}

// TestUpdateOrderStatus tests updating order status
func TestUpdateOrderStatus(t *testing.T) {
	WaitForAllServices(t)

	// Create user and order
	user := CreateTestUser(t)
	order := CreateTestOrder(t, user.Token)

	// Valid status transitions
	statusFlow := []string{"confirmed", "preparing", "shipped", "delivered"}

	for _, newStatus := range statusFlow {
		t.Run("update to "+newStatus, func(t *testing.T) {
			// Update status
			url := fmt.Sprintf("%s/api/v1/orders/%s/status", GatewayURL, order.ID)
			payload := map[string]string{
				"status": newStatus,
			}

			resp := DoRequest(t, "PUT", url, payload, user.Token)
			defer resp.Body.Close()

			// Verify success
			apiResp := AssertResponseSuccess(t, resp, http.StatusOK)

			// Parse updated order
			var orderData map[string]interface{}
			json.Unmarshal(apiResp.Data, &orderData)

			// Verify status was updated
			assert.Equal(t, newStatus, orderData["status"])

			// Small delay to allow notification processing
			time.Sleep(100 * time.Millisecond)
		})
	}
}

// TestUpdateOrderStatusInvalid tests invalid status updates
func TestUpdateOrderStatusInvalid(t *testing.T) {
	WaitForAllServices(t)

	user := CreateTestUser(t)
	order := CreateTestOrder(t, user.Token)

	invalidStatuses := []string{"invalid", "random", ""}

	for _, invalidStatus := range invalidStatuses {
		t.Run("invalid status "+invalidStatus, func(t *testing.T) {
			url := fmt.Sprintf("%s/api/v1/orders/%s/status", GatewayURL, order.ID)
			payload := map[string]string{
				"status": invalidStatus,
			}

			resp := DoRequest(t, "PUT", url, payload, user.Token)
			defer resp.Body.Close()

			AssertResponseError(t, resp, http.StatusBadRequest)
		})
	}
}

// TestCancelOrder tests canceling an order
func TestCancelOrder(t *testing.T) {
	WaitForAllServices(t)

	// Create user and order
	user := CreateTestUser(t)
	order := CreateTestOrder(t, user.Token)

	// Cancel order
	url := fmt.Sprintf("%s/api/v1/orders/%s/status", GatewayURL, order.ID)
	payload := map[string]string{
		"status": "cancelled",
	}

	resp := DoRequest(t, "PUT", url, payload, user.Token)
	defer resp.Body.Close()

	// Verify success
	apiResp := AssertResponseSuccess(t, resp, http.StatusOK)

	var orderData map[string]interface{}
	json.Unmarshal(apiResp.Data, &orderData)
	assert.Equal(t, "cancelled", orderData["status"])
}

// TestCompleteOrderFlow tests the complete order lifecycle
func TestCompleteOrderFlow(t *testing.T) {
	WaitForAllServices(t)

	// Step 1: Create authenticated user
	t.Log("Step 1: Creating authenticated user")
	user := CreateTestUser(t)

	// Step 2: Create order
	t.Log("Step 2: Creating new order")
	orderPayload := map[string]interface{}{
		"items": []map[string]interface{}{
			{
				"product_id":   "prod-test-001",
				"product_name": "Complete Flow Product",
				"quantity":     3,
				"price":        15.50,
			},
		},
		"address": map[string]string{
			"street":     "Av. Paulista",
			"number":     "1000",
			"city":       "São Paulo",
			"state":      "SP",
			"zip_code":   "01310-100",
			"complement": "Conj. 101",
		},
	}

	resp := DoRequest(t, "POST", GatewayURL+"/api/v1/orders", orderPayload, user.Token)
	apiResp := AssertResponseSuccess(t, resp, http.StatusCreated)
	orderID := ParseOrderID(t, apiResp.Data)
	resp.Body.Close()

	t.Logf("Created order ID: %s", orderID)

	// Step 3: List orders and verify it's there
	t.Log("Step 3: Listing orders")
	resp = DoRequest(t, "GET", GatewayURL+"/api/v1/orders", nil, user.Token)
	apiResp = AssertResponseSuccess(t, resp, http.StatusOK)

	var orders []map[string]interface{}
	json.Unmarshal(apiResp.Data, &orders)

	found := false
	for _, order := range orders {
		if order["id"].(string) == orderID {
			found = true
			assert.Equal(t, "pending", order["status"])
			break
		}
	}
	assert.True(t, found, "Order should be in list")
	resp.Body.Close()

	// Step 4: Get order details
	t.Log("Step 4: Getting order details")
	url := fmt.Sprintf("%s/api/v1/orders/%s", GatewayURL, orderID)
	resp = DoRequest(t, "GET", url, nil, user.Token)
	apiResp = AssertResponseSuccess(t, resp, http.StatusOK)

	var orderDetails map[string]interface{}
	json.Unmarshal(apiResp.Data, &orderDetails)
	assert.Equal(t, orderID, orderDetails["id"])
	assert.Equal(t, 46.50, orderDetails["total_amount"]) // 3 * 15.50
	resp.Body.Close()

	// Step 5: Update status through flow
	t.Log("Step 5: Progressing order through status flow")
	statuses := []string{"confirmed", "preparing", "shipped", "delivered"}

	for _, status := range statuses {
		t.Logf("Updating to status: %s", status)
		url = fmt.Sprintf("%s/api/v1/orders/%s/status", GatewayURL, orderID)
		payload := map[string]string{"status": status}

		resp = DoRequest(t, "PUT", url, payload, user.Token)
		apiResp = AssertResponseSuccess(t, resp, http.StatusOK)

		var updatedOrder map[string]interface{}
		json.Unmarshal(apiResp.Data, &updatedOrder)
		assert.Equal(t, status, updatedOrder["status"])
		resp.Body.Close()

		// Give time for notifications to be sent
		time.Sleep(200 * time.Millisecond)
	}

	// Step 6: Verify final state
	t.Log("Step 6: Verifying final order state")
	url = fmt.Sprintf("%s/api/v1/orders/%s", GatewayURL, orderID)
	resp = DoRequest(t, "GET", url, nil, user.Token)
	apiResp = AssertResponseSuccess(t, resp, http.StatusOK)

	json.Unmarshal(apiResp.Data, &orderDetails)
	assert.Equal(t, "delivered", orderDetails["status"])
	resp.Body.Close()

	t.Log("Complete order flow test passed!")
}
