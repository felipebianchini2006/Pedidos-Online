package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"testing"

	"pedidos-online/order-service/internal/model"
	"pedidos-online/order-service/internal/service"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MockOrderService is a mock implementation of OrderService
type MockOrderService struct {
	mock.Mock
}

func (m *MockOrderService) CreateOrder(ctx context.Context, userID string, items []model.OrderItem, address model.Address) (*model.Order, error) {
	args := m.Called(ctx, userID, items, address)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Order), args.Error(1)
}

func (m *MockOrderService) GetOrder(ctx context.Context, orderID, userID string) (*model.Order, error) {
	args := m.Called(ctx, orderID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Order), args.Error(1)
}

func (m *MockOrderService) ListOrders(ctx context.Context, userID string, page, pageSize int) ([]*model.Order, int64, error) {
	args := m.Called(ctx, userID, page, pageSize)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*model.Order), args.Get(1).(int64), args.Error(2)
}

func (m *MockOrderService) ListAllOrders(ctx context.Context, page, pageSize int) ([]*model.Order, int64, error) {
	args := m.Called(ctx, page, pageSize)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*model.Order), args.Get(1).(int64), args.Error(2)
}

func (m *MockOrderService) UpdateOrderStatus(ctx context.Context, orderID, newStatus string) (*model.Order, error) {
	args := m.Called(ctx, orderID, newStatus)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Order), args.Error(1)
}

func (m *MockOrderService) CancelOrder(ctx context.Context, orderID, userID string) error {
	args := m.Called(ctx, orderID, userID)
	return args.Error(0)
}

func setupTestApp() *fiber.App {
	return fiber.New()
}

func TestCreateOrder(t *testing.T) {
	t.Run("successfully create order", func(t *testing.T) {
		mockService := new(MockOrderService)
		handler := NewOrderHandler(mockService)
		app := setupTestApp()

		expectedOrder := &model.Order{
			ID:          primitive.NewObjectID(),
			UserID:      "user-123",
			Status:      model.OrderStatusPending,
			TotalAmount: 59.98,
		}

		mockService.On("CreateOrder", mock.Anything, "user-123", mock.Anything, mock.Anything).
			Return(expectedOrder, nil)

		requestBody := map[string]interface{}{
			"items": []map[string]interface{}{
				{
					"product_id":   "prod-001",
					"product_name": "Product 1",
					"quantity":     2,
					"price":        29.99,
				},
			},
			"address": map[string]string{
				"street":   "Rua Teste",
				"number":   "123",
				"city":     "São Paulo",
				"state":    "SP",
				"zip_code": "01234567",
			},
		}
		body, _ := json.Marshal(requestBody)

		req := httptest.NewRequest("POST", "/api/v1/orders", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		app.Post("/api/v1/orders", func(c *fiber.Ctx) error {
			c.Locals("userID", "user-123")
			return handler.CreateOrder(c)
		})

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, fiber.StatusCreated, resp.StatusCode)

		var response SuccessResponse
		json.NewDecoder(resp.Body).Decode(&response)

		assert.True(t, response.Success)
		assert.NotNil(t, response.Data)
		mockService.AssertExpectations(t)
	})

	t.Run("error without authentication", func(t *testing.T) {
		mockService := new(MockOrderService)
		handler := NewOrderHandler(mockService)
		app := setupTestApp()

		app.Post("/api/v1/orders", handler.CreateOrder)

		requestBody := map[string]interface{}{
			"items": []map[string]interface{}{
				{"product_id": "prod-001", "product_name": "Product 1", "quantity": 1, "price": 10.0},
			},
			"address": map[string]string{"street": "Rua", "number": "1", "city": "City", "state": "SP", "zip_code": "12345678"},
		}
		body, _ := json.Marshal(requestBody)

		req := httptest.NewRequest("POST", "/api/v1/orders", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("error with invalid request body", func(t *testing.T) {
		mockService := new(MockOrderService)
		handler := NewOrderHandler(mockService)
		app := setupTestApp()

		app.Post("/api/v1/orders", func(c *fiber.Ctx) error {
			c.Locals("userID", "user-123")
			return handler.CreateOrder(c)
		})

		req := httptest.NewRequest("POST", "/api/v1/orders", bytes.NewReader([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	})
}

func TestGetOrder(t *testing.T) {
	t.Run("successfully get order", func(t *testing.T) {
		mockService := new(MockOrderService)
		handler := NewOrderHandler(mockService)
		app := setupTestApp()

		orderID := primitive.NewObjectID().Hex()
		expectedOrder := &model.Order{
			ID:     primitive.NewObjectID(),
			UserID: "user-123",
			Status: model.OrderStatusPending,
		}

		mockService.On("GetOrder", mock.Anything, orderID, "user-123").Return(expectedOrder, nil)

		app.Get("/api/v1/orders/:id", func(c *fiber.Ctx) error {
			c.Locals("userID", "user-123")
			return handler.GetOrder(c)
		})

		req := httptest.NewRequest("GET", "/api/v1/orders/"+orderID, nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		var response SuccessResponse
		json.NewDecoder(resp.Body).Decode(&response)

		assert.True(t, response.Success)
		mockService.AssertExpectations(t)
	})

	t.Run("error when order not found", func(t *testing.T) {
		mockService := new(MockOrderService)
		handler := NewOrderHandler(mockService)
		app := setupTestApp()

		orderID := primitive.NewObjectID().Hex()

		mockService.On("GetOrder", mock.Anything, orderID, "user-123").
			Return(nil, errors.New("pedido não encontrado"))

		app.Get("/api/v1/orders/:id", func(c *fiber.Ctx) error {
			c.Locals("userID", "user-123")
			return handler.GetOrder(c)
		})

		req := httptest.NewRequest("GET", "/api/v1/orders/"+orderID, nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)
		mockService.AssertExpectations(t)
	})

	t.Run("error when accessing another user's order", func(t *testing.T) {
		mockService := new(MockOrderService)
		handler := NewOrderHandler(mockService)
		app := setupTestApp()

		orderID := primitive.NewObjectID().Hex()

		mockService.On("GetOrder", mock.Anything, orderID, "user-123").
			Return(nil, errors.New("acesso negado: este pedido não pertence ao usuário"))

		app.Get("/api/v1/orders/:id", func(c *fiber.Ctx) error {
			c.Locals("userID", "user-123")
			return handler.GetOrder(c)
		})

		req := httptest.NewRequest("GET", "/api/v1/orders/"+orderID, nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
		mockService.AssertExpectations(t)
	})
}

func TestListOrders(t *testing.T) {
	t.Run("successfully list orders", func(t *testing.T) {
		mockService := new(MockOrderService)
		handler := NewOrderHandler(mockService)
		app := setupTestApp()

		expectedOrders := []*model.Order{
			{ID: primitive.NewObjectID(), UserID: "user-123", Status: model.OrderStatusPending},
		}

		mockService.On("ListOrders", mock.Anything, "user-123", 1, 10).
			Return(expectedOrders, int64(1), nil)

		app.Get("/api/v1/orders", func(c *fiber.Ctx) error {
			c.Locals("userID", "user-123")
			return handler.ListOrders(c)
		})

		req := httptest.NewRequest("GET", "/api/v1/orders?page=1&page_size=10", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		var response OrderListResponse
		json.NewDecoder(resp.Body).Decode(&response)

		assert.True(t, response.Success)
		assert.Len(t, response.Data, 1)
		assert.Equal(t, int64(1), response.Pagination.Total)
		mockService.AssertExpectations(t)
	})
}

func TestUpdateStatus(t *testing.T) {
	t.Run("successfully update status", func(t *testing.T) {
		mockService := new(MockOrderService)
		handler := NewOrderHandler(mockService)
		app := setupTestApp()

		orderID := primitive.NewObjectID().Hex()
		updatedOrder := &model.Order{
			ID:     primitive.NewObjectID(),
			Status: model.OrderStatusConfirmed,
		}

		mockService.On("UpdateOrderStatus", mock.Anything, orderID, model.OrderStatusConfirmed).
			Return(updatedOrder, nil)

		app.Put("/api/v1/orders/:id/status", handler.UpdateStatus)

		requestBody := map[string]string{"status": "confirmed"}
		body, _ := json.Marshal(requestBody)

		req := httptest.NewRequest("PUT", "/api/v1/orders/"+orderID+"/status", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		var response SuccessResponse
		json.NewDecoder(resp.Body).Decode(&response)

		assert.True(t, response.Success)
		mockService.AssertExpectations(t)
	})

	t.Run("error with invalid status transition", func(t *testing.T) {
		mockService := new(MockOrderService)
		handler := NewOrderHandler(mockService)
		app := setupTestApp()

		orderID := primitive.NewObjectID().Hex()

		mockService.On("UpdateOrderStatus", mock.Anything, orderID, model.OrderStatusPending).
			Return(nil, errors.New("transição de status inválida"))

		app.Put("/api/v1/orders/:id/status", handler.UpdateStatus)

		requestBody := map[string]string{"status": "pending"}
		body, _ := json.Marshal(requestBody)

		req := httptest.NewRequest("PUT", "/api/v1/orders/"+orderID+"/status", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
		mockService.AssertExpectations(t)
	})
}

func TestCancelOrder(t *testing.T) {
	t.Run("successfully cancel order", func(t *testing.T) {
		mockService := new(MockOrderService)
		handler := NewOrderHandler(mockService)
		app := setupTestApp()

		orderID := primitive.NewObjectID().Hex()

		mockService.On("CancelOrder", mock.Anything, orderID, "user-123").Return(nil)

		app.Delete("/api/v1/orders/:id", func(c *fiber.Ctx) error {
			c.Locals("userID", "user-123")
			return handler.CancelOrder(c)
		})

		req := httptest.NewRequest("DELETE", "/api/v1/orders/"+orderID, nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, fiber.StatusNoContent, resp.StatusCode)
		mockService.AssertExpectations(t)
	})

	t.Run("error when cannot cancel shipped order", func(t *testing.T) {
		mockService := new(MockOrderService)
		handler := NewOrderHandler(mockService)
		app := setupTestApp()

		orderID := primitive.NewObjectID().Hex()

		mockService.On("CancelOrder", mock.Anything, orderID, "user-123").
			Return(errors.New("pedido não pode ser cancelado no status 'shipped'"))

		app.Delete("/api/v1/orders/:id", func(c *fiber.Ctx) error {
			c.Locals("userID", "user-123")
			return handler.CancelOrder(c)
		})

		req := httptest.NewRequest("DELETE", "/api/v1/orders/"+orderID, nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, fiber.StatusConflict, resp.StatusCode)
		mockService.AssertExpectations(t)
	})
}
