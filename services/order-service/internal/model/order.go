package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Constantes de status de pedido
const (
	OrderStatusPending   = "pending"   // Pedido criado, aguardando confirmação
	OrderStatusConfirmed = "confirmed" // Pedido confirmado pelo sistema
	OrderStatusPreparing = "preparing" // Pedido em preparação
	OrderStatusShipped   = "shipped"   // Pedido enviado para entrega
	OrderStatusDelivered = "delivered" // Pedido entregue ao cliente
	OrderStatusCancelled = "cancelled" // Pedido cancelado
)

// Order representa um pedido no sistema
type Order struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID      string             `bson:"user_id" json:"user_id" validate:"required"`
	Items       []OrderItem        `bson:"items" json:"items" validate:"required,min=1,dive"`
	TotalAmount float64            `bson:"total_amount" json:"total_amount"`
	Status      string             `bson:"status" json:"status" validate:"required,oneof=pending confirmed preparing shipped delivered cancelled"`
	Address     Address            `bson:"address" json:"address" validate:"required"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updated_at"`
}

// OrderItem representa um item dentro de um pedido
type OrderItem struct {
	ProductID   string  `bson:"product_id" json:"product_id" validate:"required"`
	ProductName string  `bson:"product_name" json:"product_name" validate:"required,min=2"`
	Quantity    int     `bson:"quantity" json:"quantity" validate:"required,min=1"`
	Price       float64 `bson:"price" json:"price" validate:"required,gt=0"`
}

// Address representa o endereço de entrega do pedido
type Address struct {
	Street     string `bson:"street" json:"street" validate:"required,min=3"`
	Number     string `bson:"number" json:"number" validate:"required"`
	City       string `bson:"city" json:"city" validate:"required,min=2"`
	State      string `bson:"state" json:"state" validate:"required,len=2"`
	ZipCode    string `bson:"zip_code" json:"zip_code" validate:"required,len=8"`
	Complement string `bson:"complement" json:"complement"`
}

// CalculateTotal calcula o valor total do pedido baseado nos itens
func (o *Order) CalculateTotal() {
	total := 0.0
	for _, item := range o.Items {
		total += item.Price * float64(item.Quantity)
	}
	o.TotalAmount = total
}

// IsValidStatus verifica se o status da ordem é válido
func (o *Order) IsValidStatus() bool {
	return IsValidStatus(o.Status)
}

// IsValidStatus verifica se o status fornecido é um dos valores permitidos
func IsValidStatus(status string) bool {
	validStatuses := []string{
		OrderStatusPending,
		OrderStatusConfirmed,
		OrderStatusPreparing,
		OrderStatusShipped,
		OrderStatusDelivered,
		OrderStatusCancelled,
	}

	for _, validStatus := range validStatuses {
		if status == validStatus {
			return true
		}
	}

	return false
}

// CanTransitionTo verifica se é possível transitar do status atual para o novo status
func (o *Order) CanTransitionTo(newStatus string) bool {
	// Mapa de transições válidas
	validTransitions := map[string][]string{
		OrderStatusPending: {
			OrderStatusConfirmed,
			OrderStatusCancelled,
		},
		OrderStatusConfirmed: {
			OrderStatusPreparing,
			OrderStatusCancelled,
		},
		OrderStatusPreparing: {
			OrderStatusShipped,
			OrderStatusCancelled,
		},
		OrderStatusShipped: {
			OrderStatusDelivered,
		},
		OrderStatusDelivered: {}, // Estado final, sem transições
		OrderStatusCancelled: {}, // Estado final, sem transições
	}

	allowedTransitions, exists := validTransitions[o.Status]
	if !exists {
		return false
	}

	for _, allowed := range allowedTransitions {
		if allowed == newStatus {
			return true
		}
	}

	return false
}

// CreateOrderRequest representa o payload para criar um novo pedido
type CreateOrderRequest struct {
	Items   []OrderItem `json:"items" validate:"required,min=1,dive"`
	Address Address     `json:"address" validate:"required"`
}

// UpdateOrderStatusRequest representa o payload para atualizar o status do pedido
type UpdateOrderStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=pending confirmed preparing shipped delivered cancelled"`
}

// OrderEvent representa um evento de pedido para publicação no RabbitMQ
type OrderEvent struct {
	EventType string    `json:"event_type"` // "order.created" ou "order.updated"
	OrderID   string    `json:"order_id"`
	UserID    string    `json:"user_id"`
	Status    string    `json:"status"`
	Total     float64   `json:"total"`
	Timestamp time.Time `json:"timestamp"`
}
