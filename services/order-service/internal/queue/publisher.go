package queue

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"pedidos-online/order-service/internal/model"

	amqp "github.com/rabbitmq/amqp091-go"
)

// Publisher gerencia a publicação de eventos no RabbitMQ
type Publisher struct {
	conn     *amqp.Connection
	channel  *amqp.Channel
	exchange string
}

// OrderCreatedEvent representa o evento de pedido criado
type OrderCreatedEvent struct {
	OrderID     string            `json:"order_id"`
	UserID      string            `json:"user_id"`
	Items       []model.OrderItem `json:"items"`
	TotalAmount float64           `json:"total_amount"`
	Status      string            `json:"status"`
	Address     model.Address     `json:"address"`
	CreatedAt   time.Time         `json:"created_at"`
}

// OrderUpdatedEvent representa o evento de pedido atualizado
type OrderUpdatedEvent struct {
	OrderID   string    `json:"order_id"`
	OldStatus string    `json:"old_status"`
	NewStatus string    `json:"new_status"`
	UpdatedAt time.Time `json:"updated_at"`
}

// NewPublisher cria um novo publisher com retry logic
func NewPublisher(rabbitMQURL string) (*Publisher, error) {
	const maxRetries = 3
	const exchange = "orders"

	// Conectar com retry
	conn, err := connectWithRetry(rabbitMQURL, maxRetries)
	if err != nil {
		return nil, fmt.Errorf("falha ao conectar ao RabbitMQ após %d tentativas: %w", maxRetries, err)
	}

	// Criar channel
	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("erro ao criar channel: %w", err)
	}

	// Declarar exchange topic "orders" (durable)
	err = channel.ExchangeDeclare(
		exchange, // name
		"topic",  // type
		true,     // durable
		false,    // auto-deleted
		false,    // internal
		false,    // no-wait
		nil,      // arguments
	)
	if err != nil {
		channel.Close()
		conn.Close()
		return nil, fmt.Errorf("erro ao declarar exchange: %w", err)
	}

	log.Printf("✅ RabbitMQ Publisher conectado ao exchange '%s'", exchange)

	return &Publisher{
		conn:     conn,
		channel:  channel,
		exchange: exchange,
	}, nil
}

// PublishOrderCreated publica evento de pedido criado
func (p *Publisher) PublishOrderCreated(order *model.Order) error {
	// Criar evento
	event := OrderCreatedEvent{
		OrderID:     order.ID.Hex(),
		UserID:      order.UserID,
		Items:       order.Items,
		TotalAmount: order.TotalAmount,
		Status:      order.Status,
		Address:     order.Address,
		CreatedAt:   order.CreatedAt,
	}

	// Serializar para JSON
	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("erro ao serializar evento: %w", err)
	}

	// Preparar headers
	headers := amqp.Table{
		"event_type": "order.created",
		"timestamp":  time.Now().Unix(),
		"order_id":   order.ID.Hex(),
		"user_id":    order.UserID,
	}

	// Configurar mensagem
	msg := amqp.Publishing{
		DeliveryMode: amqp.Persistent, // mensagem persistente
		ContentType:  "application/json",
		Body:         body,
		Headers:      headers,
		Timestamp:    time.Now(),
	}

	// Publicar na exchange com routing key "order.created"
	err = p.channel.Publish(
		p.exchange,      // exchange
		"order.created", // routing key
		false,           // mandatory
		false,           // immediate
		msg,             // message
	)

	if err != nil {
		return fmt.Errorf("erro ao publicar evento order.created: %w", err)
	}

	log.Printf("📨 Evento 'order.created' publicado para pedido %s", order.ID.Hex())
	return nil
}

// PublishOrderUpdated publica evento de pedido atualizado
func (p *Publisher) PublishOrderUpdated(orderID, oldStatus, newStatus string) error {
	// Criar evento
	event := OrderUpdatedEvent{
		OrderID:   orderID,
		OldStatus: oldStatus,
		NewStatus: newStatus,
		UpdatedAt: time.Now(),
	}

	// Serializar para JSON
	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("erro ao serializar evento: %w", err)
	}

	// Preparar headers
	headers := amqp.Table{
		"event_type": "order.updated",
		"timestamp":  time.Now().Unix(),
		"order_id":   orderID,
		"old_status": oldStatus,
		"new_status": newStatus,
	}

	// Configurar mensagem
	msg := amqp.Publishing{
		DeliveryMode: amqp.Persistent, // mensagem persistente
		ContentType:  "application/json",
		Body:         body,
		Headers:      headers,
		Timestamp:    time.Now(),
	}

	// Publicar na exchange com routing key "order.updated"
	err = p.channel.Publish(
		p.exchange,      // exchange
		"order.updated", // routing key
		false,           // mandatory
		false,           // immediate
		msg,             // message
	)

	if err != nil {
		return fmt.Errorf("erro ao publicar evento order.updated: %w", err)
	}

	log.Printf("📨 Evento 'order.updated' publicado para pedido %s (%s → %s)", orderID, oldStatus, newStatus)
	return nil
}

// Close fecha o channel e a conexão
func (p *Publisher) Close() error {
	var errs []error

	// Fechar channel
	if p.channel != nil {
		if err := p.channel.Close(); err != nil {
			errs = append(errs, fmt.Errorf("erro ao fechar channel: %w", err))
		} else {
			log.Println("✅ RabbitMQ Channel fechado")
		}
	}

	// Fechar conexão
	if p.conn != nil {
		if err := p.conn.Close(); err != nil {
			errs = append(errs, fmt.Errorf("erro ao fechar conexão: %w", err))
		} else {
			log.Println("✅ RabbitMQ Connection fechada")
		}
	}

	// Retornar erros se houver
	if len(errs) > 0 {
		return fmt.Errorf("erros ao fechar publisher: %v", errs)
	}

	return nil
}

// IsConnected verifica se a conexão está ativa
func (p *Publisher) IsConnected() bool {
	return p.conn != nil && !p.conn.IsClosed()
}

// connectWithRetry tenta conectar ao RabbitMQ com retry e backoff exponencial
func connectWithRetry(url string, maxRetries int) (*amqp.Connection, error) {
	var conn *amqp.Connection
	var err error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		log.Printf("🔄 Tentativa %d/%d de conexão ao RabbitMQ...", attempt, maxRetries)

		conn, err = amqp.Dial(url)
		if err == nil {
			log.Printf("✅ Conectado ao RabbitMQ na tentativa %d", attempt)
			return conn, nil
		}

		log.Printf("❌ Falha na tentativa %d: %v", attempt, err)

		// Se não for a última tentativa, aguardar com backoff exponencial
		if attempt < maxRetries {
			backoff := time.Duration(attempt*attempt) * time.Second
			log.Printf("⏳ Aguardando %v antes da próxima tentativa...", backoff)
			time.Sleep(backoff)
		}
	}

	return nil, fmt.Errorf("todas as %d tentativas falharam: %w", maxRetries, err)
}

// Reconnect tenta reconectar ao RabbitMQ
func (p *Publisher) Reconnect(rabbitMQURL string) error {
	log.Println("🔄 Tentando reconectar ao RabbitMQ...")

	// Fechar conexões antigas
	if err := p.Close(); err != nil {
		log.Printf("⚠️  Erro ao fechar conexões antigas: %v", err)
	}

	// Conectar novamente
	conn, err := connectWithRetry(rabbitMQURL, 3)
	if err != nil {
		return fmt.Errorf("erro ao reconectar: %w", err)
	}

	// Criar novo channel
	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return fmt.Errorf("erro ao criar channel: %w", err)
	}

	// Declarar exchange novamente
	err = channel.ExchangeDeclare(
		p.exchange, // name
		"topic",    // type
		true,       // durable
		false,      // auto-deleted
		false,      // internal
		false,      // no-wait
		nil,        // arguments
	)
	if err != nil {
		channel.Close()
		conn.Close()
		return fmt.Errorf("erro ao declarar exchange: %w", err)
	}

	// Atualizar conexão e channel
	p.conn = conn
	p.channel = channel

	log.Println("✅ Reconectado ao RabbitMQ com sucesso")
	return nil
}

// PublishWithRetry publica uma mensagem com retry automático em caso de falha
func (p *Publisher) PublishWithRetry(routingKey string, body []byte, headers amqp.Table) error {
	const maxRetries = 3

	msg := amqp.Publishing{
		DeliveryMode: amqp.Persistent,
		ContentType:  "application/json",
		Body:         body,
		Headers:      headers,
		Timestamp:    time.Now(),
	}

	for attempt := 1; attempt <= maxRetries; attempt++ {
		err := p.channel.Publish(
			p.exchange,
			routingKey,
			false,
			false,
			msg,
		)

		if err == nil {
			return nil
		}

		log.Printf("❌ Tentativa %d/%d falhou: %v", attempt, maxRetries, err)

		// Se a conexão foi perdida, tentar reconectar
		if !p.IsConnected() {
			log.Println("🔄 Conexão perdida, tentando reconectar...")
			// Nota: Você precisaria passar a URL aqui ou armazená-la no struct
		}

		if attempt < maxRetries {
			backoff := time.Duration(attempt) * time.Second
			time.Sleep(backoff)
		}
	}

	return fmt.Errorf("falha ao publicar após %d tentativas", maxRetries)
}
