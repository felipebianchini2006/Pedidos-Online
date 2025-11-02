package model

import (
	"encoding/json"
	"fmt"
	"time"
)

// OrderCreatedEvent representa o evento de criação de pedido
// Publicado pelo Order Service quando um novo pedido é criado
type OrderCreatedEvent struct {
	OrderID     string      `json:"order_id"`
	UserID      string      `json:"user_id"`
	UserEmail   string      `json:"user_email"` // Email do usuário (se disponível)
	Items       []OrderItem `json:"items"`
	TotalAmount float64     `json:"total_amount"`
	Address     Address     `json:"address"`
	CreatedAt   time.Time   `json:"created_at"`
}

// OrderUpdatedEvent representa o evento de atualização de status do pedido
// Publicado pelo Order Service quando o status do pedido muda
type OrderUpdatedEvent struct {
	OrderID   string    `json:"order_id"`
	UserID    string    `json:"user_id"`
	UserEmail string    `json:"user_email"` // Email do usuário (se disponível)
	OldStatus string    `json:"old_status"`
	NewStatus string    `json:"new_status"`
	UpdatedAt time.Time `json:"updated_at"`
}

// OrderItem representa um item do pedido
type OrderItem struct {
	ProductID   string  `json:"product_id"`
	ProductName string  `json:"product_name"`
	Quantity    int     `json:"quantity"`
	Price       float64 `json:"price"`
}

// Address representa o endereço de entrega
type Address struct {
	Street     string `json:"street"`
	Number     string `json:"number"`
	City       string `json:"city"`
	State      string `json:"state"`
	ZipCode    string `json:"zip_code"`
	Complement string `json:"complement,omitempty"`
}

// ParseOrderCreatedEvent faz parse do JSON para OrderCreatedEvent
//
// Parâmetros:
//   - data: JSON em bytes
//
// Retorno:
//   - *OrderCreatedEvent: Evento parseado
//   - error: Erro de parse (se houver)
//
// Exemplo de JSON:
//
//	{
//	  "order_id": "507f1f77bcf86cd799439011",
//	  "user_id": "123e4567-e89b-12d3-a456-426614174000",
//	  "user_email": "user@example.com",
//	  "items": [
//	    {
//	      "product_id": "prod-123",
//	      "product_name": "Pizza Margherita",
//	      "quantity": 2,
//	      "price": 35.90
//	    }
//	  ],
//	  "total_amount": 71.80,
//	  "address": {
//	    "street": "Rua das Flores",
//	    "number": "123",
//	    "city": "São Paulo",
//	    "state": "SP",
//	    "zip_code": "01234-567"
//	  },
//	  "created_at": "2025-11-02T10:30:00Z"
//	}
func ParseOrderCreatedEvent(data []byte) (*OrderCreatedEvent, error) {
	var event OrderCreatedEvent
	if err := json.Unmarshal(data, &event); err != nil {
		return nil, fmt.Errorf("erro ao fazer parse de OrderCreatedEvent: %w", err)
	}

	// Validações básicas
	if event.OrderID == "" {
		return nil, fmt.Errorf("order_id é obrigatório")
	}

	if event.UserID == "" {
		return nil, fmt.Errorf("user_id é obrigatório")
	}

	if len(event.Items) == 0 {
		return nil, fmt.Errorf("pedido deve ter pelo menos 1 item")
	}

	if event.TotalAmount <= 0 {
		return nil, fmt.Errorf("total_amount deve ser maior que 0")
	}

	return &event, nil
}

// ParseOrderUpdatedEvent faz parse do JSON para OrderUpdatedEvent
//
// Parâmetros:
//   - data: JSON em bytes
//
// Retorno:
//   - *OrderUpdatedEvent: Evento parseado
//   - error: Erro de parse (se houver)
//
// Exemplo de JSON:
//
//	{
//	  "order_id": "507f1f77bcf86cd799439011",
//	  "user_id": "123e4567-e89b-12d3-a456-426614174000",
//	  "user_email": "user@example.com",
//	  "old_status": "pending",
//	  "new_status": "confirmed",
//	  "updated_at": "2025-11-02T10:35:00Z"
//	}
func ParseOrderUpdatedEvent(data []byte) (*OrderUpdatedEvent, error) {
	var event OrderUpdatedEvent
	if err := json.Unmarshal(data, &event); err != nil {
		return nil, fmt.Errorf("erro ao fazer parse de OrderUpdatedEvent: %w", err)
	}

	// Validações básicas
	if event.OrderID == "" {
		return nil, fmt.Errorf("order_id é obrigatório")
	}

	if event.UserID == "" {
		return nil, fmt.Errorf("user_id é obrigatório")
	}

	if event.OldStatus == "" {
		return nil, fmt.Errorf("old_status é obrigatório")
	}

	if event.NewStatus == "" {
		return nil, fmt.Errorf("new_status é obrigatório")
	}

	return &event, nil
}

// String retorna representação em string do OrderCreatedEvent
func (e *OrderCreatedEvent) String() string {
	return fmt.Sprintf("OrderCreatedEvent{OrderID: %s, UserID: %s, Items: %d, TotalAmount: R$ %.2f}",
		e.OrderID, e.UserID, len(e.Items), e.TotalAmount)
}

// String retorna representação em string do OrderUpdatedEvent
func (e *OrderUpdatedEvent) String() string {
	return fmt.Sprintf("OrderUpdatedEvent{OrderID: %s, UserID: %s, Status: %s -> %s}",
		e.OrderID, e.UserID, e.OldStatus, e.NewStatus)
}

// GetStatusDescription retorna descrição amigável do status em português
func GetStatusDescription(status string) string {
	descriptions := map[string]string{
		"pending":   "Pendente",
		"confirmed": "Confirmado",
		"preparing": "Em Preparação",
		"shipped":   "Enviado",
		"delivered": "Entregue",
		"cancelled": "Cancelado",
	}

	if desc, ok := descriptions[status]; ok {
		return desc
	}

	return status
}

// FormatAddress formata o endereço em string legível
func (a *Address) FormatAddress() string {
	address := fmt.Sprintf("%s, %s - %s, %s - CEP: %s",
		a.Street, a.Number, a.City, a.State, a.ZipCode)

	if a.Complement != "" {
		address += fmt.Sprintf(" (%s)", a.Complement)
	}

	return address
}

// CalculateItemTotal calcula o total de um item (price * quantity)
func (i *OrderItem) CalculateItemTotal() float64 {
	return i.Price * float64(i.Quantity)
}
