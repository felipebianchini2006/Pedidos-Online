package repository

import (
	"context"
	"testing"
	"time"

	"pedidos-online/order-service/internal/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
)

func TestCreate(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("successfully create order", func(mt *mtest.T) {
		repo := NewOrderRepository(mt.DB)

		order := &model.Order{
			UserID: "user-123",
			Items: []model.OrderItem{
				{ProductID: "prod-001", ProductName: "Product 1", Quantity: 2, Price: 29.99},
			},
			TotalAmount: 59.98,
			Status:      model.OrderStatusPending,
			Address: model.Address{
				Street:  "Rua Teste",
				Number:  "123",
				City:    "São Paulo",
				State:   "SP",
				ZipCode: "01234567",
			},
		}

		// Mock insert response
		mt.AddMockResponses(mtest.CreateSuccessResponse())

		err := repo.Create(context.Background(), order)

		require.NoError(t, err)
		assert.NotEqual(t, primitive.NilObjectID, order.ID)
		assert.False(t, order.CreatedAt.IsZero())
		assert.False(t, order.UpdatedAt.IsZero())
	})

	mt.Run("error when database insert fails", func(mt *mtest.T) {
		repo := NewOrderRepository(mt.DB)

		order := &model.Order{
			UserID: "user-123",
			Items: []model.OrderItem{
				{ProductID: "prod-001", ProductName: "Product 1", Quantity: 1, Price: 10.00},
			},
			Status: model.OrderStatusPending,
		}

		// Mock insert error
		mt.AddMockResponses(mtest.CreateWriteErrorsResponse(mtest.WriteError{
			Index:   0,
			Code:    11000,
			Message: "duplicate key error",
		}))

		err := repo.Create(context.Background(), order)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "erro ao criar pedido")
	})
}

func TestFindByID(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("successfully find order by ID", func(mt *mtest.T) {
		repo := NewOrderRepository(mt.DB)

		orderID := primitive.NewObjectID()
		expectedOrder := model.Order{
			ID:          orderID,
			UserID:      "user-123",
			Items:       []model.OrderItem{{ProductID: "prod-001", ProductName: "Product 1", Quantity: 1, Price: 29.99}},
			TotalAmount: 29.99,
			Status:      model.OrderStatusPending,
			Address: model.Address{
				Street:  "Rua Teste",
				Number:  "123",
				City:    "São Paulo",
				State:   "SP",
				ZipCode: "01234567",
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		// Mock find response
		mt.AddMockResponses(mtest.CreateCursorResponse(1, "orders_db.orders", mtest.FirstBatch, bson.D{
			{Key: "_id", Value: expectedOrder.ID},
			{Key: "user_id", Value: expectedOrder.UserID},
			{Key: "total_amount", Value: expectedOrder.TotalAmount},
			{Key: "status", Value: expectedOrder.Status},
		}))

		order, err := repo.FindByID(context.Background(), orderID.Hex())

		require.NoError(t, err)
		require.NotNil(t, order)
		assert.Equal(t, expectedOrder.ID, order.ID)
		assert.Equal(t, expectedOrder.UserID, order.UserID)
		assert.Equal(t, expectedOrder.Status, order.Status)
	})

	mt.Run("error with invalid ObjectID", func(mt *mtest.T) {
		repo := NewOrderRepository(mt.DB)

		order, err := repo.FindByID(context.Background(), "invalid-id")

		require.Error(t, err)
		assert.Nil(t, order)
		assert.Contains(t, err.Error(), "ID inválido")
	})

	mt.Run("error when order not found", func(mt *mtest.T) {
		repo := NewOrderRepository(mt.DB)

		orderID := primitive.NewObjectID()

		// Mock no documents response
		mt.AddMockResponses(mtest.CreateCursorResponse(0, "orders_db.orders", mtest.FirstBatch))

		order, err := repo.FindByID(context.Background(), orderID.Hex())

		require.Error(t, err)
		assert.Nil(t, order)
		assert.Contains(t, err.Error(), "pedido não encontrado")
	})
}

func TestFindByUserID(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("successfully find orders by user ID", func(mt *mtest.T) {
		repo := NewOrderRepository(mt.DB)

		userID := "user-123"
		order1ID := primitive.NewObjectID()
		order2ID := primitive.NewObjectID()

		// Mock find response with multiple orders
		first := mtest.CreateCursorResponse(1, "orders_db.orders", mtest.FirstBatch, bson.D{
			{Key: "_id", Value: order1ID},
			{Key: "user_id", Value: userID},
			{Key: "status", Value: model.OrderStatusPending},
			{Key: "total_amount", Value: 100.0},
		})
		second := mtest.CreateCursorResponse(1, "orders_db.orders", mtest.NextBatch, bson.D{
			{Key: "_id", Value: order2ID},
			{Key: "user_id", Value: userID},
			{Key: "status", Value: model.OrderStatusConfirmed},
			{Key: "total_amount", Value: 200.0},
		})
		killCursors := mtest.CreateCursorResponse(0, "orders_db.orders", mtest.NextBatch)

		mt.AddMockResponses(first, second, killCursors)

		orders, err := repo.FindByUserID(context.Background(), userID, 10, 0)

		require.NoError(t, err)
		require.NotNil(t, orders)
		assert.Len(t, orders, 2)
		assert.Equal(t, order1ID, orders[0].ID)
		assert.Equal(t, order2ID, orders[1].ID)
	})

	mt.Run("return empty list when no orders found", func(mt *mtest.T) {
		repo := NewOrderRepository(mt.DB)

		// Mock empty response
		mt.AddMockResponses(mtest.CreateCursorResponse(0, "orders_db.orders", mtest.FirstBatch))

		orders, err := repo.FindByUserID(context.Background(), "user-999", 10, 0)

		require.NoError(t, err)
		require.NotNil(t, orders)
		assert.Len(t, orders, 0)
	})

	mt.Run("use default limit when zero", func(mt *mtest.T) {
		repo := NewOrderRepository(mt.DB)

		mt.AddMockResponses(mtest.CreateCursorResponse(0, "orders_db.orders", mtest.FirstBatch))

		orders, err := repo.FindByUserID(context.Background(), "user-123", 0, 0)

		require.NoError(t, err)
		assert.NotNil(t, orders)
	})

	mt.Run("cap limit at maximum", func(mt *mtest.T) {
		repo := NewOrderRepository(mt.DB)

		mt.AddMockResponses(mtest.CreateCursorResponse(0, "orders_db.orders", mtest.FirstBatch))

		orders, err := repo.FindByUserID(context.Background(), "user-123", 200, 0)

		require.NoError(t, err)
		assert.NotNil(t, orders)
		// Limit should be capped at 100
	})
}

func TestUpdateStatus(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("successfully update order status", func(mt *mtest.T) {
		repo := NewOrderRepository(mt.DB)

		orderID := primitive.NewObjectID()

		// Mock update response
		mt.AddMockResponses(bson.D{
			{Key: "ok", Value: 1},
			{Key: "value", Value: bson.D{
				{Key: "_id", Value: orderID},
				{Key: "status", Value: model.OrderStatusConfirmed},
			}},
		})

		err := repo.UpdateStatus(context.Background(), orderID.Hex(), model.OrderStatusConfirmed)

		require.NoError(t, err)
	})

	mt.Run("error with invalid status", func(mt *mtest.T) {
		repo := NewOrderRepository(mt.DB)

		orderID := primitive.NewObjectID()

		err := repo.UpdateStatus(context.Background(), orderID.Hex(), "invalid-status")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "status inválido")
	})

	mt.Run("error with invalid ObjectID", func(mt *mtest.T) {
		repo := NewOrderRepository(mt.DB)

		err := repo.UpdateStatus(context.Background(), "invalid-id", model.OrderStatusConfirmed)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "ID inválido")
	})

	mt.Run("error when order not found", func(mt *mtest.T) {
		repo := NewOrderRepository(mt.DB)

		orderID := primitive.NewObjectID()

		// Mock no match response
		mt.AddMockResponses(bson.D{
			{Key: "ok", Value: 1},
			{Key: "value", Value: nil},
		})

		err := repo.UpdateStatus(context.Background(), orderID.Hex(), model.OrderStatusConfirmed)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "pedido não encontrado")
	})
}

func TestUpdate(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("successfully update order", func(mt *mtest.T) {
		repo := NewOrderRepository(mt.DB)

		order := &model.Order{
			ID:          primitive.NewObjectID(),
			UserID:      "user-123",
			Items:       []model.OrderItem{{ProductID: "prod-001", ProductName: "Product 1", Quantity: 2, Price: 29.99}},
			TotalAmount: 59.98,
			Status:      model.OrderStatusConfirmed,
			Address: model.Address{
				Street:  "Rua Atualizada",
				Number:  "456",
				City:    "Rio de Janeiro",
				State:   "RJ",
				ZipCode: "20000000",
			},
		}

		// Mock update response
		mt.AddMockResponses(bson.D{
			{Key: "ok", Value: 1},
			{Key: "value", Value: bson.D{
				{Key: "_id", Value: order.ID},
			}},
		})

		err := repo.Update(context.Background(), order)

		require.NoError(t, err)
		assert.False(t, order.UpdatedAt.IsZero())
	})

	mt.Run("error when order not found", func(mt *mtest.T) {
		repo := NewOrderRepository(mt.DB)

		order := &model.Order{
			ID:     primitive.NewObjectID(),
			UserID: "user-123",
			Status: model.OrderStatusPending,
		}

		// Mock no match response
		mt.AddMockResponses(bson.D{
			{Key: "ok", Value: 1},
			{Key: "value", Value: nil},
		})

		err := repo.Update(context.Background(), order)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "pedido não encontrado")
	})
}

func TestDelete(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("successfully delete order", func(mt *mtest.T) {
		repo := NewOrderRepository(mt.DB)

		orderID := primitive.NewObjectID()

		// Mock delete response
		mt.AddMockResponses(bson.D{
			{Key: "ok", Value: 1},
			{Key: "acknowledged", Value: true},
			{Key: "n", Value: 1},
		})

		err := repo.Delete(context.Background(), orderID.Hex())

		require.NoError(t, err)
	})

	mt.Run("error with invalid ObjectID", func(mt *mtest.T) {
		repo := NewOrderRepository(mt.DB)

		err := repo.Delete(context.Background(), "invalid-id")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "ID inválido")
	})

	mt.Run("error when order not found", func(mt *mtest.T) {
		repo := NewOrderRepository(mt.DB)

		orderID := primitive.NewObjectID()

		// Mock delete with no matches
		mt.AddMockResponses(bson.D{
			{Key: "ok", Value: 1},
			{Key: "acknowledged", Value: true},
			{Key: "n", Value: 0},
		})

		err := repo.Delete(context.Background(), orderID.Hex())

		require.Error(t, err)
		assert.Contains(t, err.Error(), "pedido não encontrado")
	})
}

func TestCount(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("successfully count user orders", func(mt *mtest.T) {
		repo := NewOrderRepository(mt.DB)

		expectedCount := int64(5)

		// Mock count response
		mt.AddMockResponses(bson.D{
			{Key: "ok", Value: 1},
			{Key: "n", Value: expectedCount},
		})

		count, err := repo.Count(context.Background(), "user-123")

		require.NoError(t, err)
		assert.Equal(t, expectedCount, count)
	})

	mt.Run("return zero when no orders found", func(mt *mtest.T) {
		repo := NewOrderRepository(mt.DB)

		// Mock count response with zero
		mt.AddMockResponses(bson.D{
			{Key: "ok", Value: 1},
			{Key: "n", Value: 0},
		})

		count, err := repo.Count(context.Background(), "user-999")

		require.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})
}

func TestFindAll(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("successfully find all orders", func(mt *mtest.T) {
		repo := NewOrderRepository(mt.DB)

		order1ID := primitive.NewObjectID()
		order2ID := primitive.NewObjectID()

		// Mock find all response
		first := mtest.CreateCursorResponse(1, "orders_db.orders", mtest.FirstBatch, bson.D{
			{Key: "_id", Value: order1ID},
			{Key: "user_id", Value: "user-123"},
			{Key: "status", Value: model.OrderStatusPending},
		})
		second := mtest.CreateCursorResponse(1, "orders_db.orders", mtest.NextBatch, bson.D{
			{Key: "_id", Value: order2ID},
			{Key: "user_id", Value: "user-456"},
			{Key: "status", Value: model.OrderStatusDelivered},
		})
		killCursors := mtest.CreateCursorResponse(0, "orders_db.orders", mtest.NextBatch)

		mt.AddMockResponses(first, second, killCursors)

		orders, err := repo.FindAll(context.Background(), 10, 0)

		require.NoError(t, err)
		require.NotNil(t, orders)
		assert.Len(t, orders, 2)
	})

	mt.Run("return empty list when no orders", func(mt *mtest.T) {
		repo := NewOrderRepository(mt.DB)

		mt.AddMockResponses(mtest.CreateCursorResponse(0, "orders_db.orders", mtest.FirstBatch))

		orders, err := repo.FindAll(context.Background(), 10, 0)

		require.NoError(t, err)
		require.NotNil(t, orders)
		assert.Len(t, orders, 0)
	})
}

func TestCountAll(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("successfully count all orders", func(mt *mtest.T) {
		repo := NewOrderRepository(mt.DB)

		expectedCount := int64(42)

		// Mock count all response
		mt.AddMockResponses(bson.D{
			{Key: "ok", Value: 1},
			{Key: "n", Value: expectedCount},
		})

		count, err := repo.CountAll(context.Background())

		require.NoError(t, err)
		assert.Equal(t, expectedCount, count)
	})

	mt.Run("return zero when no orders", func(mt *mtest.T) {
		repo := NewOrderRepository(mt.DB)

		// Mock count response with zero
		mt.AddMockResponses(bson.D{
			{Key: "ok", Value: 1},
			{Key: "n", Value: 0},
		})

		count, err := repo.CountAll(context.Background())

		require.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})
}

func TestCreateIndexes(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("successfully create indexes", func(mt *mtest.T) {
		repo := NewOrderRepository(mt.DB)

		// Mock create indexes response
		mt.AddMockResponses(bson.D{
			{Key: "ok", Value: 1},
			{Key: "createdCollectionAutomatically", Value: false},
			{Key: "numIndexesBefore", Value: 1},
			{Key: "numIndexesAfter", Value: 5},
		})

		err := repo.CreateIndexes(context.Background())

		require.NoError(t, err)
	})
}
