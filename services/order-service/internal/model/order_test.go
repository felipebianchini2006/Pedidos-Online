package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestCalculateTotal(t *testing.T) {
	t.Run("calculate total with single item", func(t *testing.T) {
		order := &Order{
			Items: []OrderItem{
				{ProductID: "prod-001", ProductName: "Product 1", Quantity: 2, Price: 29.99},
			},
		}

		order.CalculateTotal()

		assert.Equal(t, 59.98, order.TotalAmount)
	})

	t.Run("calculate total with multiple items", func(t *testing.T) {
		order := &Order{
			Items: []OrderItem{
				{ProductID: "prod-001", ProductName: "Product 1", Quantity: 2, Price: 29.99},
				{ProductID: "prod-002", ProductName: "Product 2", Quantity: 1, Price: 49.99},
				{ProductID: "prod-003", ProductName: "Product 3", Quantity: 3, Price: 10.00},
			},
		}

		order.CalculateTotal()

		// 2*29.99 + 1*49.99 + 3*10.00 = 59.98 + 49.99 + 30.00 = 139.97
		assert.Equal(t, 139.97, order.TotalAmount)
	})

	t.Run("calculate total with no items", func(t *testing.T) {
		order := &Order{
			Items: []OrderItem{},
		}

		order.CalculateTotal()

		assert.Equal(t, 0.0, order.TotalAmount)
	})

	t.Run("calculate total with decimal quantities", func(t *testing.T) {
		order := &Order{
			Items: []OrderItem{
				{ProductID: "prod-001", ProductName: "Product 1", Quantity: 5, Price: 9.99},
			},
		}

		order.CalculateTotal()

		assert.Equal(t, 49.95, order.TotalAmount)
	})
}

func TestIsValidStatus(t *testing.T) {
	tests := []struct {
		status string
		valid  bool
	}{
		{OrderStatusPending, true},
		{OrderStatusConfirmed, true},
		{OrderStatusPreparing, true},
		{OrderStatusShipped, true},
		{OrderStatusDelivered, true},
		{OrderStatusCancelled, true},
		{"invalid", false},
		{"PENDING", false}, // case sensitive
		{"", false},
		{"random-status", false},
	}

	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			result := IsValidStatus(tt.status)
			assert.Equal(t, tt.valid, result, "Status: %s should be %v", tt.status, tt.valid)
		})
	}
}

func TestOrderIsValidStatus(t *testing.T) {
	t.Run("order with valid status", func(t *testing.T) {
		order := &Order{Status: OrderStatusPending}
		assert.True(t, order.IsValidStatus())
	})

	t.Run("order with invalid status", func(t *testing.T) {
		order := &Order{Status: "invalid-status"}
		assert.False(t, order.IsValidStatus())
	})
}

func TestCanTransitionTo(t *testing.T) {
	tests := []struct {
		name           string
		currentStatus  string
		newStatus      string
		canTransition  bool
	}{
		// Valid transitions from pending
		{"pending to confirmed", OrderStatusPending, OrderStatusConfirmed, true},
		{"pending to cancelled", OrderStatusPending, OrderStatusCancelled, true},

		// Invalid transitions from pending
		{"pending to preparing", OrderStatusPending, OrderStatusPreparing, false},
		{"pending to shipped", OrderStatusPending, OrderStatusShipped, false},
		{"pending to delivered", OrderStatusPending, OrderStatusDelivered, false},

		// Valid transitions from confirmed
		{"confirmed to preparing", OrderStatusConfirmed, OrderStatusPreparing, true},
		{"confirmed to cancelled", OrderStatusConfirmed, OrderStatusCancelled, true},

		// Invalid transitions from confirmed
		{"confirmed to pending", OrderStatusConfirmed, OrderStatusPending, false},
		{"confirmed to shipped", OrderStatusConfirmed, OrderStatusShipped, false},
		{"confirmed to delivered", OrderStatusConfirmed, OrderStatusDelivered, false},

		// Valid transitions from preparing
		{"preparing to shipped", OrderStatusPreparing, OrderStatusShipped, true},
		{"preparing to cancelled", OrderStatusPreparing, OrderStatusCancelled, true},

		// Invalid transitions from preparing
		{"preparing to pending", OrderStatusPreparing, OrderStatusPending, false},
		{"preparing to confirmed", OrderStatusPreparing, OrderStatusConfirmed, false},
		{"preparing to delivered", OrderStatusPreparing, OrderStatusDelivered, false},

		// Valid transitions from shipped
		{"shipped to delivered", OrderStatusShipped, OrderStatusDelivered, true},

		// Invalid transitions from shipped
		{"shipped to pending", OrderStatusShipped, OrderStatusPending, false},
		{"shipped to confirmed", OrderStatusShipped, OrderStatusConfirmed, false},
		{"shipped to preparing", OrderStatusShipped, OrderStatusPreparing, false},
		{"shipped to cancelled", OrderStatusShipped, OrderStatusCancelled, false},

		// No valid transitions from delivered (final state)
		{"delivered to pending", OrderStatusDelivered, OrderStatusPending, false},
		{"delivered to confirmed", OrderStatusDelivered, OrderStatusConfirmed, false},
		{"delivered to preparing", OrderStatusDelivered, OrderStatusPreparing, false},
		{"delivered to shipped", OrderStatusDelivered, OrderStatusShipped, false},
		{"delivered to cancelled", OrderStatusDelivered, OrderStatusCancelled, false},

		// No valid transitions from cancelled (final state)
		{"cancelled to pending", OrderStatusCancelled, OrderStatusPending, false},
		{"cancelled to confirmed", OrderStatusCancelled, OrderStatusConfirmed, false},
		{"cancelled to preparing", OrderStatusCancelled, OrderStatusPreparing, false},
		{"cancelled to shipped", OrderStatusCancelled, OrderStatusShipped, false},
		{"cancelled to delivered", OrderStatusCancelled, OrderStatusDelivered, false},

		// Invalid current status
		{"invalid to pending", "invalid", OrderStatusPending, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			order := &Order{
				ID:     primitive.NewObjectID(),
				Status: tt.currentStatus,
			}

			result := order.CanTransitionTo(tt.newStatus)
			assert.Equal(t, tt.canTransition, result,
				"Transition from %s to %s should be %v",
				tt.currentStatus, tt.newStatus, tt.canTransition)
		})
	}
}

func TestCanTransitionToSameStatus(t *testing.T) {
	t.Run("cannot transition to same status", func(t *testing.T) {
		order := &Order{Status: OrderStatusPending}
		assert.False(t, order.CanTransitionTo(OrderStatusPending))
	})

	t.Run("delivered cannot transition to delivered", func(t *testing.T) {
		order := &Order{Status: OrderStatusDelivered}
		assert.False(t, order.CanTransitionTo(OrderStatusDelivered))
	})
}
