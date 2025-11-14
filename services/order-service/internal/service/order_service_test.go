package service

import (
	"context"
	"errors"
	"testing"

	"pedidos-online/order-service/internal/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MockOrderRepository is a mock implementation of OrderRepository
type MockOrderRepository struct {
	mock.Mock
}

func (m *MockOrderRepository) Create(ctx context.Context, order *model.Order) error {
	args := m.Called(ctx, order)
	if args.Get(0) == nil {
		// Simulate database setting ID
		order.ID = primitive.NewObjectID()
		return nil
	}
	return args.Error(0)
}

func (m *MockOrderRepository) FindByID(ctx context.Context, id string) (*model.Order, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Order), args.Error(1)
}

func (m *MockOrderRepository) FindByUserID(ctx context.Context, userID string, limit, skip int) ([]*model.Order, error) {
	args := m.Called(ctx, userID, limit, skip)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Order), args.Error(1)
}

func (m *MockOrderRepository) FindAll(ctx context.Context, limit, skip int) ([]*model.Order, error) {
	args := m.Called(ctx, limit, skip)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Order), args.Error(1)
}

func (m *MockOrderRepository) Update(ctx context.Context, order *model.Order) error {
	args := m.Called(ctx, order)
	return args.Error(0)
}

func (m *MockOrderRepository) UpdateStatus(ctx context.Context, id, status string) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

func (m *MockOrderRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockOrderRepository) Count(ctx context.Context, userID string) (int64, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockOrderRepository) CountAll(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockOrderRepository) CreateIndexes(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// MockEventPublisher is a mock implementation of EventPublisher
type MockEventPublisher struct {
	mock.Mock
}

func (m *MockEventPublisher) PublishOrderCreated(order *model.Order) error {
	args := m.Called(order)
	return args.Error(0)
}

func (m *MockEventPublisher) PublishOrderUpdated(orderID, oldStatus, newStatus string) error {
	args := m.Called(orderID, oldStatus, newStatus)
	return args.Error(0)
}

func (m *MockEventPublisher) Close() error {
	args := m.Called()
	return args.Error(0)
}

// Test CreateOrder
func TestCreateOrder(t *testing.T) {
	t.Run("successfully create order with valid data", func(t *testing.T) {
		mockRepo := new(MockOrderRepository)
		mockPublisher := new(MockEventPublisher)
		service := NewOrderService(mockRepo, mockPublisher)

		ctx := context.Background()
		userID := "user-123"
		items := []model.OrderItem{
			{ProductID: "prod-001", ProductName: "Product 1", Quantity: 2, Price: 29.99},
		}
		address := model.Address{
			Street:  "Rua Teste",
			Number:  "123",
			City:    "São Paulo",
			State:   "SP",
			ZipCode: "01234567",
		}

		mockRepo.On("Create", ctx, mock.MatchedBy(func(order *model.Order) bool {
			return order.UserID == userID && len(order.Items) == 1 && order.TotalAmount == 59.98
		})).Return(nil)

		mockPublisher.On("PublishOrderCreated", mock.Anything).Return(nil)

		order, err := service.CreateOrder(ctx, userID, items, address)

		require.NoError(t, err)
		require.NotNil(t, order)
		assert.Equal(t, userID, order.UserID)
		assert.Equal(t, model.OrderStatusPending, order.Status)
		assert.Equal(t, 59.98, order.TotalAmount)
		assert.NotEqual(t, primitive.NilObjectID, order.ID)
		mockRepo.AssertExpectations(t)
		mockPublisher.AssertExpectations(t)
	})

	t.Run("error with empty userID", func(t *testing.T) {
		mockRepo := new(MockOrderRepository)
		mockPublisher := new(MockEventPublisher)
		service := NewOrderService(mockRepo, mockPublisher)

		ctx := context.Background()
		items := []model.OrderItem{
			{ProductID: "prod-001", ProductName: "Product 1", Quantity: 1, Price: 10.00},
		}
		address := model.Address{Street: "Rua", Number: "1", City: "City", State: "SP", ZipCode: "12345678"}

		order, err := service.CreateOrder(ctx, "", items, address)

		require.Error(t, err)
		assert.Nil(t, order)
		assert.Contains(t, err.Error(), "user_id é obrigatório")
		mockRepo.AssertExpectations(t)
	})

	t.Run("error with empty items", func(t *testing.T) {
		mockRepo := new(MockOrderRepository)
		mockPublisher := new(MockEventPublisher)
		service := NewOrderService(mockRepo, mockPublisher)

		ctx := context.Background()
		items := []model.OrderItem{}
		address := model.Address{Street: "Rua", Number: "1", City: "City", State: "SP", ZipCode: "12345678"}

		order, err := service.CreateOrder(ctx, "user-123", items, address)

		require.Error(t, err)
		assert.Nil(t, order)
		assert.Contains(t, err.Error(), "pelo menos um item")
	})

	t.Run("error with too many items", func(t *testing.T) {
		mockRepo := new(MockOrderRepository)
		mockPublisher := new(MockEventPublisher)
		service := NewOrderService(mockRepo, mockPublisher)

		ctx := context.Background()
		items := make([]model.OrderItem, 51) // Exceeds limit of 50
		for i := range items {
			items[i] = model.OrderItem{ProductID: "prod", ProductName: "Product", Quantity: 1, Price: 10.0}
		}
		address := model.Address{Street: "Rua", Number: "1", City: "City", State: "SP", ZipCode: "12345678"}

		order, err := service.CreateOrder(ctx, "user-123", items, address)

		require.Error(t, err)
		assert.Nil(t, order)
		assert.Contains(t, err.Error(), "mais de 50 itens")
	})

	t.Run("error with invalid item quantity", func(t *testing.T) {
		mockRepo := new(MockOrderRepository)
		mockPublisher := new(MockEventPublisher)
		service := NewOrderService(mockRepo, mockPublisher)

		ctx := context.Background()
		items := []model.OrderItem{
			{ProductID: "prod-001", ProductName: "Product 1", Quantity: 0, Price: 10.00},
		}
		address := model.Address{Street: "Rua", Number: "1", City: "City", State: "SP", ZipCode: "12345678"}

		order, err := service.CreateOrder(ctx, "user-123", items, address)

		require.Error(t, err)
		assert.Nil(t, order)
		assert.Contains(t, err.Error(), "quantidade deve ser maior que zero")
	})

	t.Run("error with invalid item price", func(t *testing.T) {
		mockRepo := new(MockOrderRepository)
		mockPublisher := new(MockEventPublisher)
		service := NewOrderService(mockRepo, mockPublisher)

		ctx := context.Background()
		items := []model.OrderItem{
			{ProductID: "prod-001", ProductName: "Product 1", Quantity: 1, Price: 0.00},
		}
		address := model.Address{Street: "Rua", Number: "1", City: "City", State: "SP", ZipCode: "12345678"}

		order, err := service.CreateOrder(ctx, "user-123", items, address)

		require.Error(t, err)
		assert.Nil(t, order)
		assert.Contains(t, err.Error(), "preço deve ser maior que zero")
	})

	t.Run("error with invalid address", func(t *testing.T) {
		mockRepo := new(MockOrderRepository)
		mockPublisher := new(MockEventPublisher)
		service := NewOrderService(mockRepo, mockPublisher)

		ctx := context.Background()
		items := []model.OrderItem{
			{ProductID: "prod-001", ProductName: "Product 1", Quantity: 1, Price: 10.00},
		}
		address := model.Address{Street: "", Number: "1", City: "City", State: "SP", ZipCode: "12345678"}

		order, err := service.CreateOrder(ctx, "user-123", items, address)

		require.Error(t, err)
		assert.Nil(t, order)
		assert.Contains(t, err.Error(), "rua é obrigatória")
	})

	t.Run("success even when event publishing fails", func(t *testing.T) {
		mockRepo := new(MockOrderRepository)
		mockPublisher := new(MockEventPublisher)
		service := NewOrderService(mockRepo, mockPublisher)

		ctx := context.Background()
		userID := "user-123"
		items := []model.OrderItem{
			{ProductID: "prod-001", ProductName: "Product 1", Quantity: 1, Price: 10.00},
		}
		address := model.Address{Street: "Rua", Number: "1", City: "City", State: "SP", ZipCode: "12345678"}

		mockRepo.On("Create", ctx, mock.Anything).Return(nil)
		mockPublisher.On("PublishOrderCreated", mock.Anything).Return(errors.New("rabbitmq error"))

		order, err := service.CreateOrder(ctx, userID, items, address)

		// Order should still be created successfully
		require.NoError(t, err)
		require.NotNil(t, order)
		mockRepo.AssertExpectations(t)
		mockPublisher.AssertExpectations(t)
	})
}

// Test GetOrder
func TestGetOrder(t *testing.T) {
	t.Run("successfully get order by ID", func(t *testing.T) {
		mockRepo := new(MockOrderRepository)
		mockPublisher := new(MockEventPublisher)
		service := NewOrderService(mockRepo, mockPublisher)

		ctx := context.Background()
		orderID := primitive.NewObjectID().Hex()
		userID := "user-123"

		expectedOrder := &model.Order{
			ID:     primitive.NewObjectID(),
			UserID: userID,
			Status: model.OrderStatusPending,
		}

		mockRepo.On("FindByID", ctx, orderID).Return(expectedOrder, nil)

		order, err := service.GetOrder(ctx, orderID, userID)

		require.NoError(t, err)
		require.NotNil(t, order)
		assert.Equal(t, expectedOrder.UserID, order.UserID)
		mockRepo.AssertExpectations(t)
	})

	t.Run("error when order not found", func(t *testing.T) {
		mockRepo := new(MockOrderRepository)
		mockPublisher := new(MockEventPublisher)
		service := NewOrderService(mockRepo, mockPublisher)

		ctx := context.Background()
		orderID := primitive.NewObjectID().Hex()

		mockRepo.On("FindByID", ctx, orderID).Return(nil, errors.New("not found"))

		order, err := service.GetOrder(ctx, orderID, "user-123")

		require.Error(t, err)
		assert.Nil(t, order)
		assert.Contains(t, err.Error(), "pedido não encontrado")
		mockRepo.AssertExpectations(t)
	})

	t.Run("error when user tries to access another user's order", func(t *testing.T) {
		mockRepo := new(MockOrderRepository)
		mockPublisher := new(MockEventPublisher)
		service := NewOrderService(mockRepo, mockPublisher)

		ctx := context.Background()
		orderID := primitive.NewObjectID().Hex()

		expectedOrder := &model.Order{
			ID:     primitive.NewObjectID(),
			UserID: "user-123",
			Status: model.OrderStatusPending,
		}

		mockRepo.On("FindByID", ctx, orderID).Return(expectedOrder, nil)

		order, err := service.GetOrder(ctx, orderID, "user-456") // Different user

		require.Error(t, err)
		assert.Nil(t, order)
		assert.Contains(t, err.Error(), "acesso negado")
		mockRepo.AssertExpectations(t)
	})

	t.Run("error with empty orderID", func(t *testing.T) {
		mockRepo := new(MockOrderRepository)
		mockPublisher := new(MockEventPublisher)
		service := NewOrderService(mockRepo, mockPublisher)

		ctx := context.Background()

		order, err := service.GetOrder(ctx, "", "user-123")

		require.Error(t, err)
		assert.Nil(t, order)
		assert.Contains(t, err.Error(), "order_id é obrigatório")
	})

	t.Run("error with empty userID", func(t *testing.T) {
		mockRepo := new(MockOrderRepository)
		mockPublisher := new(MockEventPublisher)
		service := NewOrderService(mockRepo, mockPublisher)

		ctx := context.Background()
		orderID := primitive.NewObjectID().Hex()

		order, err := service.GetOrder(ctx, orderID, "")

		require.Error(t, err)
		assert.Nil(t, order)
		assert.Contains(t, err.Error(), "user_id é obrigatório")
	})
}

// Test ListOrders
func TestListOrders(t *testing.T) {
	t.Run("successfully list orders with pagination", func(t *testing.T) {
		mockRepo := new(MockOrderRepository)
		mockPublisher := new(MockEventPublisher)
		service := NewOrderService(mockRepo, mockPublisher)

		ctx := context.Background()
		userID := "user-123"

		expectedOrders := []*model.Order{
			{ID: primitive.NewObjectID(), UserID: userID, Status: model.OrderStatusPending},
			{ID: primitive.NewObjectID(), UserID: userID, Status: model.OrderStatusDelivered},
		}

		mockRepo.On("FindByUserID", ctx, userID, 10, 0).Return(expectedOrders, nil)
		mockRepo.On("Count", ctx, userID).Return(int64(2), nil)

		orders, total, err := service.ListOrders(ctx, userID, 1, 10)

		require.NoError(t, err)
		assert.Len(t, orders, 2)
		assert.Equal(t, int64(2), total)
		mockRepo.AssertExpectations(t)
	})

	t.Run("adjust invalid page and pageSize", func(t *testing.T) {
		mockRepo := new(MockOrderRepository)
		mockPublisher := new(MockEventPublisher)
		service := NewOrderService(mockRepo, mockPublisher)

		ctx := context.Background()
		userID := "user-123"

		mockRepo.On("FindByUserID", ctx, userID, 10, 0).Return([]*model.Order{}, nil)
		mockRepo.On("Count", ctx, userID).Return(int64(0), nil)

		// Invalid page and pageSize should be adjusted
		orders, total, err := service.ListOrders(ctx, userID, 0, 0)

		require.NoError(t, err)
		assert.NotNil(t, orders)
		assert.Equal(t, int64(0), total)
		mockRepo.AssertExpectations(t)
	})

	t.Run("cap pageSize at maximum", func(t *testing.T) {
		mockRepo := new(MockOrderRepository)
		mockPublisher := new(MockEventPublisher)
		service := NewOrderService(mockRepo, mockPublisher)

		ctx := context.Background()
		userID := "user-123"

		// PageSize capped at 100
		mockRepo.On("FindByUserID", ctx, userID, 100, 0).Return([]*model.Order{}, nil)
		mockRepo.On("Count", ctx, userID).Return(int64(0), nil)

		orders, total, err := service.ListOrders(ctx, userID, 1, 200)

		require.NoError(t, err)
		assert.NotNil(t, orders)
		mockRepo.AssertExpectations(t)
	})

	t.Run("error with empty userID", func(t *testing.T) {
		mockRepo := new(MockOrderRepository)
		mockPublisher := new(MockEventPublisher)
		service := NewOrderService(mockRepo, mockPublisher)

		ctx := context.Background()

		orders, total, err := service.ListOrders(ctx, "", 1, 10)

		require.Error(t, err)
		assert.Nil(t, orders)
		assert.Equal(t, int64(0), total)
		assert.Contains(t, err.Error(), "user_id é obrigatório")
	})
}

// Test UpdateOrderStatus
func TestUpdateOrderStatus(t *testing.T) {
	t.Run("successfully update order status", func(t *testing.T) {
		mockRepo := new(MockOrderRepository)
		mockPublisher := new(MockEventPublisher)
		service := NewOrderService(mockRepo, mockPublisher)

		ctx := context.Background()
		orderID := primitive.NewObjectID().Hex()

		existingOrder := &model.Order{
			ID:     primitive.NewObjectID(),
			UserID: "user-123",
			Status: model.OrderStatusPending,
		}

		mockRepo.On("FindByID", ctx, orderID).Return(existingOrder, nil)
		mockRepo.On("UpdateStatus", ctx, orderID, model.OrderStatusConfirmed).Return(nil)
		mockPublisher.On("PublishOrderUpdated", orderID, model.OrderStatusPending, model.OrderStatusConfirmed).Return(nil)

		order, err := service.UpdateOrderStatus(ctx, orderID, model.OrderStatusConfirmed)

		require.NoError(t, err)
		require.NotNil(t, order)
		assert.Equal(t, model.OrderStatusConfirmed, order.Status)
		mockRepo.AssertExpectations(t)
		mockPublisher.AssertExpectations(t)
	})

	t.Run("error with invalid status transition", func(t *testing.T) {
		mockRepo := new(MockOrderRepository)
		mockPublisher := new(MockEventPublisher)
		service := NewOrderService(mockRepo, mockPublisher)

		ctx := context.Background()
		orderID := primitive.NewObjectID().Hex()

		existingOrder := &model.Order{
			ID:     primitive.NewObjectID(),
			Status: model.OrderStatusDelivered, // Final state
		}

		mockRepo.On("FindByID", ctx, orderID).Return(existingOrder, nil)

		order, err := service.UpdateOrderStatus(ctx, orderID, model.OrderStatusPending)

		require.Error(t, err)
		assert.Nil(t, order)
		assert.Contains(t, err.Error(), "transição de status inválida")
		mockRepo.AssertExpectations(t)
	})

	t.Run("error with invalid status value", func(t *testing.T) {
		mockRepo := new(MockOrderRepository)
		mockPublisher := new(MockEventPublisher)
		service := NewOrderService(mockRepo, mockPublisher)

		ctx := context.Background()
		orderID := primitive.NewObjectID().Hex()

		order, err := service.UpdateOrderStatus(ctx, orderID, "invalid-status")

		require.Error(t, err)
		assert.Nil(t, order)
		assert.Contains(t, err.Error(), "status inválido")
	})
}

// Test CancelOrder
func TestCancelOrder(t *testing.T) {
	t.Run("successfully cancel order in pending status", func(t *testing.T) {
		mockRepo := new(MockOrderRepository)
		mockPublisher := new(MockEventPublisher)
		service := NewOrderService(mockRepo, mockPublisher)

		ctx := context.Background()
		orderID := primitive.NewObjectID().Hex()
		userID := "user-123"

		existingOrder := &model.Order{
			ID:     primitive.NewObjectID(),
			UserID: userID,
			Status: model.OrderStatusPending,
		}

		mockRepo.On("FindByID", ctx, orderID).Return(existingOrder, nil)
		mockRepo.On("UpdateStatus", ctx, orderID, model.OrderStatusCancelled).Return(nil)
		mockPublisher.On("PublishOrderUpdated", orderID, model.OrderStatusPending, model.OrderStatusCancelled).Return(nil)

		err := service.CancelOrder(ctx, orderID, userID)

		require.NoError(t, err)
		mockRepo.AssertExpectations(t)
		mockPublisher.AssertExpectations(t)
	})

	t.Run("error when cancelling shipped order", func(t *testing.T) {
		mockRepo := new(MockOrderRepository)
		mockPublisher := new(MockEventPublisher)
		service := NewOrderService(mockRepo, mockPublisher)

		ctx := context.Background()
		orderID := primitive.NewObjectID().Hex()
		userID := "user-123"

		existingOrder := &model.Order{
			ID:     primitive.NewObjectID(),
			UserID: userID,
			Status: model.OrderStatusShipped, // Cannot cancel
		}

		mockRepo.On("FindByID", ctx, orderID).Return(existingOrder, nil)

		err := service.CancelOrder(ctx, orderID, userID)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "não pode ser cancelado")
		mockRepo.AssertExpectations(t)
	})

	t.Run("error when user tries to cancel another user's order", func(t *testing.T) {
		mockRepo := new(MockOrderRepository)
		mockPublisher := new(MockEventPublisher)
		service := NewOrderService(mockRepo, mockPublisher)

		ctx := context.Background()
		orderID := primitive.NewObjectID().Hex()

		existingOrder := &model.Order{
			ID:     primitive.NewObjectID(),
			UserID: "user-123",
			Status: model.OrderStatusPending,
		}

		mockRepo.On("FindByID", ctx, orderID).Return(existingOrder, nil)

		err := service.CancelOrder(ctx, orderID, "user-456") // Different user

		require.Error(t, err)
		assert.Contains(t, err.Error(), "acesso negado")
		mockRepo.AssertExpectations(t)
	})
}

// Test ListAllOrders (admin function)
func TestListAllOrders(t *testing.T) {
	t.Run("successfully list all orders", func(t *testing.T) {
		mockRepo := new(MockOrderRepository)
		mockPublisher := new(MockEventPublisher)
		service := NewOrderService(mockRepo, mockPublisher)

		ctx := context.Background()

		expectedOrders := []*model.Order{
			{ID: primitive.NewObjectID(), UserID: "user-123", Status: model.OrderStatusPending},
			{ID: primitive.NewObjectID(), UserID: "user-456", Status: model.OrderStatusDelivered},
		}

		mockRepo.On("FindAll", ctx, 10, 0).Return(expectedOrders, nil)
		mockRepo.On("CountAll", ctx).Return(int64(2), nil)

		orders, total, err := service.ListAllOrders(ctx, 1, 10)

		require.NoError(t, err)
		assert.Len(t, orders, 2)
		assert.Equal(t, int64(2), total)
		mockRepo.AssertExpectations(t)
	})
}
