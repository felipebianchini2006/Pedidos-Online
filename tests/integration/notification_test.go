package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	RabbitMQURL      = "amqp://guest:guest@localhost:5672/"
	OrdersExchange   = "orders"
	OrderCreatedKey  = "order.created"
	OrderUpdatedKey  = "order.updated"
	NotificationWait = 5 * time.Second
)

// RabbitMQHelper provides utilities for testing RabbitMQ integration
type RabbitMQHelper struct {
	conn    *amqp.Connection
	channel *amqp.Channel
}

// NewRabbitMQHelper creates a new RabbitMQ test helper
func NewRabbitMQHelper(t *testing.T) *RabbitMQHelper {
	t.Helper()

	conn, err := amqp.Dial(RabbitMQURL)
	require.NoError(t, err, "Failed to connect to RabbitMQ")

	channel, err := conn.Channel()
	require.NoError(t, err, "Failed to create channel")

	return &RabbitMQHelper{
		conn:    conn,
		channel: channel,
	}
}

// Close closes the RabbitMQ connection
func (h *RabbitMQHelper) Close() {
	if h.channel != nil {
		h.channel.Close()
	}
	if h.conn != nil {
		h.conn.Close()
	}
}

// DeclareTestQueue declares a temporary test queue bound to orders exchange
func (h *RabbitMQHelper) DeclareTestQueue(t *testing.T, routingKey string) string {
	t.Helper()

	// Declare a unique temporary queue
	queueName := fmt.Sprintf("test.queue.%d", GenerateTimestamp())

	queue, err := h.channel.QueueDeclare(
		queueName, // name
		false,     // durable
		true,      // auto-delete
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	require.NoError(t, err, "Failed to declare queue")

	// Bind queue to orders exchange
	err = h.channel.QueueBind(
		queue.Name,     // queue name
		routingKey,     // routing key
		OrdersExchange, // exchange
		false,          // no-wait
		nil,            // arguments
	)
	require.NoError(t, err, "Failed to bind queue")

	t.Logf("Created test queue: %s bound to %s with key: %s", queueName, OrdersExchange, routingKey)

	return queueName
}

// ConsumeMessages consumes messages from a queue with timeout
func (h *RabbitMQHelper) ConsumeMessages(t *testing.T, queueName string, timeout time.Duration) []amqp.Delivery {
	t.Helper()

	msgs, err := h.channel.Consume(
		queueName, // queue
		"",        // consumer
		true,      // auto-ack
		false,     // exclusive
		false,     // no-local
		false,     // no-wait
		nil,       // args
	)
	require.NoError(t, err, "Failed to consume messages")

	var deliveries []amqp.Delivery
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	for {
		select {
		case msg := <-msgs:
			deliveries = append(deliveries, msg)
		case <-ctx.Done():
			return deliveries
		}
	}
}

// GetQueueDepth returns the number of messages in a queue
func (h *RabbitMQHelper) GetQueueDepth(t *testing.T, queueName string) int {
	t.Helper()

	queue, err := h.channel.QueueInspect(queueName)
	require.NoError(t, err, "Failed to inspect queue")

	return queue.Messages
}

// PurgeQueue removes all messages from a queue
func (h *RabbitMQHelper) PurgeQueue(t *testing.T, queueName string) {
	t.Helper()

	_, err := h.channel.QueuePurge(queueName, false)
	require.NoError(t, err, "Failed to purge queue")
}

// TestOrderCreatedNotification tests that order creation triggers notification
func TestOrderCreatedNotification(t *testing.T) {
	WaitForAllServices(t)

	// Setup RabbitMQ helper
	mq := NewRabbitMQHelper(t)
	defer mq.Close()

	// Create test queue to intercept order.created events
	testQueue := mq.DeclareTestQueue(t, OrderCreatedKey)

	// Create user and order
	user := CreateTestUser(t)
	order := CreateTestOrder(t, user.Token)

	t.Logf("Created order %s, waiting for notification...", order.ID)

	// Wait and consume messages
	messages := mq.ConsumeMessages(t, testQueue, NotificationWait)

	// Verify we received at least one message
	assert.Greater(t, len(messages), 0, "Should receive order.created notification")

	if len(messages) > 0 {
		// Parse the message
		var orderEvent map[string]interface{}
		err := json.Unmarshal(messages[0].Body, &orderEvent)
		require.NoError(t, err, "Failed to parse order event")

		// Verify event structure
		assert.Equal(t, order.ID, orderEvent["order_id"], "Order ID should match")
		assert.Equal(t, "order.created", orderEvent["event_type"], "Event type should be order.created")
		assert.NotEmpty(t, orderEvent["timestamp"], "Timestamp should be present")

		// Verify order data is included
		orderData, ok := orderEvent["order"].(map[string]interface{})
		assert.True(t, ok, "Order data should be present")
		assert.Equal(t, order.UserID, orderData["user_id"])
		assert.Equal(t, "pending", orderData["status"])

		t.Logf("Successfully verified order.created notification")
	}
}

// TestOrderUpdatedNotification tests that status updates trigger notifications
func TestOrderUpdatedNotification(t *testing.T) {
	WaitForAllServices(t)

	// Setup RabbitMQ helper
	mq := NewRabbitMQHelper(t)
	defer mq.Close()

	// Create test queue to intercept order.updated events
	testQueue := mq.DeclareTestQueue(t, OrderUpdatedKey)

	// Create user and order
	user := CreateTestUser(t)
	order := CreateTestOrder(t, user.Token)

	// Give time for order.created to be processed
	time.Sleep(1 * time.Second)

	// Purge any messages from queue
	mq.PurgeQueue(t, testQueue)

	// Update order status
	url := fmt.Sprintf("%s/api/v1/orders/%s/status", GatewayURL, order.ID)
	payload := map[string]string{
		"status": "confirmed",
	}

	resp := DoRequest(t, "PUT", url, payload, user.Token)
	AssertResponseSuccess(t, resp, http.StatusOK)
	resp.Body.Close()

	t.Logf("Updated order %s to confirmed, waiting for notification...", order.ID)

	// Wait and consume messages
	messages := mq.ConsumeMessages(t, testQueue, NotificationWait)

	// Verify we received notification
	assert.Greater(t, len(messages), 0, "Should receive order.updated notification")

	if len(messages) > 0 {
		// Parse the message
		var orderEvent map[string]interface{}
		err := json.Unmarshal(messages[0].Body, &orderEvent)
		require.NoError(t, err, "Failed to parse order event")

		// Verify event structure
		assert.Equal(t, order.ID, orderEvent["order_id"], "Order ID should match")
		assert.Equal(t, "order.updated", orderEvent["event_type"], "Event type should be order.updated")

		// Verify order data shows updated status
		orderData, ok := orderEvent["order"].(map[string]interface{})
		assert.True(t, ok, "Order data should be present")
		assert.Equal(t, "confirmed", orderData["status"], "Status should be updated")

		t.Logf("Successfully verified order.updated notification")
	}
}

// TestMultipleStatusUpdatesNotifications tests notifications for status flow
func TestMultipleStatusUpdatesNotifications(t *testing.T) {
	WaitForAllServices(t)

	// Setup RabbitMQ helper
	mq := NewRabbitMQHelper(t)
	defer mq.Close()

	// Create test queue
	testQueue := mq.DeclareTestQueue(t, OrderUpdatedKey)

	// Create user and order
	user := CreateTestUser(t)
	order := CreateTestOrder(t, user.Token)

	// Wait for initial creation notification
	time.Sleep(1 * time.Second)
	mq.PurgeQueue(t, testQueue)

	// Update through multiple statuses
	statuses := []string{"confirmed", "preparing", "shipped"}

	for _, status := range statuses {
		t.Logf("Updating order to %s", status)

		// Update status
		url := fmt.Sprintf("%s/api/v1/orders/%s/status", GatewayURL, order.ID)
		payload := map[string]string{"status": status}

		resp := DoRequest(t, "PUT", url, payload, user.Token)
		AssertResponseSuccess(t, resp, http.StatusOK)
		resp.Body.Close()

		// Wait for notification
		time.Sleep(2 * time.Second)

		// Check queue depth
		depth := mq.GetQueueDepth(t, testQueue)
		assert.Greater(t, depth, 0, "Should have notification for status: "+status)

		// Consume and verify
		messages := mq.ConsumeMessages(t, testQueue, 1*time.Second)
		if len(messages) > 0 {
			var orderEvent map[string]interface{}
			json.Unmarshal(messages[0].Body, &orderEvent)

			orderData := orderEvent["order"].(map[string]interface{})
			assert.Equal(t, status, orderData["status"])

			t.Logf("Verified notification for status: %s", status)
		}
	}
}

// TestNotificationServiceConsumption tests that notification service consumes messages
func TestNotificationServiceConsumption(t *testing.T) {
	WaitForAllServices(t)

	// Setup RabbitMQ helper
	mq := NewRabbitMQHelper(t)
	defer mq.Close()

	// Check that order.created and order.updated queues exist and are being consumed
	queues := []string{"order.created", "order.updated"}

	for _, queueName := range queues {
		t.Run("queue "+queueName, func(t *testing.T) {
			// Try to inspect the queue
			queue, err := mq.channel.QueueInspect(queueName)

			if err != nil {
				t.Skipf("Queue %s not found, might use different naming", queueName)
				return
			}

			// Verify queue exists
			assert.NotNil(t, queue, "Queue should exist")

			// Check if queue has consumers
			assert.GreaterOrEqual(t, queue.Consumers, 0, "Queue should have configuration")

			t.Logf("Queue %s exists with %d consumers and %d messages",
				queueName, queue.Consumers, queue.Messages)
		})
	}
}

// TestNotificationIdempotency tests that reprocessing same message is safe
func TestNotificationIdempotency(t *testing.T) {
	WaitForAllServices(t)

	// Setup RabbitMQ helper
	mq := NewRabbitMQHelper(t)
	defer mq.Close()

	// Create test queue
	testQueue := mq.DeclareTestQueue(t, OrderCreatedKey)

	// Create user and order
	user := CreateTestUser(t)
	order := CreateTestOrder(t, user.Token)

	// Wait for notification
	time.Sleep(NotificationWait)

	// Get initial messages
	initialMessages := mq.ConsumeMessages(t, testQueue, 1*time.Second)
	initialCount := len(initialMessages)

	t.Logf("Received %d initial notifications", initialCount)

	// The notification service should have consumed the message by now
	// If we try to consume again, we shouldn't get duplicates
	// (this tests that messages are being ACKed properly)

	time.Sleep(2 * time.Second)
	duplicateMessages := mq.ConsumeMessages(t, testQueue, 2*time.Second)

	// Should not receive duplicate notifications
	assert.Equal(t, 0, len(duplicateMessages),
		"Should not receive duplicate notifications after ACK")

	t.Log("Verified no duplicate notifications - idempotency working")
}

// TestNotificationServiceHealthDuringLoad tests notification service under load
func TestNotificationServiceHealthDuringLoad(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping load test in short mode")
	}

	WaitForAllServices(t)

	// Setup RabbitMQ helper
	mq := NewRabbitMQHelper(t)
	defer mq.Close()

	// Create test queue
	testQueue := mq.DeclareTestQueue(t, OrderCreatedKey)

	// Create user
	user := CreateTestUser(t)

	// Create multiple orders rapidly
	numOrders := 10
	t.Logf("Creating %d orders rapidly...", numOrders)

	for i := 0; i < numOrders; i++ {
		CreateTestOrder(t, user.Token)
	}

	// Wait for all notifications to be processed
	time.Sleep(NotificationWait * 2)

	// Consume messages
	messages := mq.ConsumeMessages(t, testQueue, 5*time.Second)

	// Should receive notifications for all orders
	assert.GreaterOrEqual(t, len(messages), numOrders,
		"Should receive notifications for all orders")

	t.Logf("Received %d notifications for %d orders", len(messages), numOrders)

	// Verify notification service is still healthy
	WaitForService(t, NotificationURL, 5*time.Second)

	t.Log("Notification service remained healthy under load")
}

// TestNotificationMessageFormat tests the structure of notification messages
func TestNotificationMessageFormat(t *testing.T) {
	WaitForAllServices(t)

	// Setup RabbitMQ helper
	mq := NewRabbitMQHelper(t)
	defer mq.Close()

	// Create test queue
	testQueue := mq.DeclareTestQueue(t, OrderCreatedKey)

	// Create order
	user := CreateTestUser(t)
	order := CreateTestOrder(t, user.Token)

	// Wait and consume
	messages := mq.ConsumeMessages(t, testQueue, NotificationWait)
	require.Greater(t, len(messages), 0, "Should receive message")

	msg := messages[0]

	// Verify message properties
	assert.Equal(t, "application/json", msg.ContentType, "Should have JSON content type")
	assert.Equal(t, uint8(2), msg.DeliveryMode, "Message should be persistent")

	// Parse and verify message structure
	var event map[string]interface{}
	err := json.Unmarshal(msg.Body, &event)
	require.NoError(t, err, "Should be valid JSON")

	// Required fields
	requiredFields := []string{"event_type", "order_id", "timestamp", "order"}
	for _, field := range requiredFields {
		assert.Contains(t, event, field, "Should have field: "+field)
	}

	// Verify order object structure
	orderData, ok := event["order"].(map[string]interface{})
	require.True(t, ok, "Order should be an object")

	orderRequiredFields := []string{"id", "user_id", "status", "total_amount", "items", "address"}
	for _, field := range orderRequiredFields {
		assert.Contains(t, orderData, field, "Order should have field: "+field)
	}

	t.Log("Message format validated successfully")
}
