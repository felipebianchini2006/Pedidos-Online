package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// Publisher gerencia a publicação de eventos no RabbitMQ
type Publisher struct {
	conn      *amqp.Connection
	channel   *amqp.Channel
	exchange  string
	connected bool
}

// NewPublisher cria uma nova instância do Publisher
func NewPublisher(url, exchange string) (*Publisher, error) {
	// Conectar ao RabbitMQ
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("erro ao conectar ao RabbitMQ: %w", err)
	}

	// Criar canal
	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("erro ao criar canal RabbitMQ: %w", err)
	}

	// Declarar exchange (topic)
	err = channel.ExchangeDeclare(
		exchange, // nome
		"topic",  // tipo
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
		conn:      conn,
		channel:   channel,
		exchange:  exchange,
		connected: true,
	}, nil
}

// Publish publica uma mensagem no exchange com a routing key especificada
func (p *Publisher) Publish(routingKey string, message interface{}) error {
	if !p.connected {
		return fmt.Errorf("publisher não está conectado")
	}

	// Serializar mensagem para JSON
	body, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("erro ao serializar mensagem: %w", err)
	}

	// Criar contexto com timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Publicar mensagem
	err = p.channel.PublishWithContext(
		ctx,
		p.exchange, // exchange
		routingKey, // routing key
		false,      // mandatory
		false,      // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent, // mensagem persistente
			Timestamp:    time.Now(),
		},
	)

	if err != nil {
		return fmt.Errorf("erro ao publicar mensagem: %w", err)
	}

	log.Printf("📤 Mensagem publicada: exchange=%s, routing_key=%s", p.exchange, routingKey)
	return nil
}

// Close fecha a conexão com o RabbitMQ
func (p *Publisher) Close() error {
	p.connected = false

	if p.channel != nil {
		if err := p.channel.Close(); err != nil {
			log.Printf("⚠️  Erro ao fechar canal RabbitMQ: %v", err)
		}
	}

	if p.conn != nil {
		if err := p.conn.Close(); err != nil {
			log.Printf("⚠️  Erro ao fechar conexão RabbitMQ: %v", err)
			return err
		}
	}

	log.Println("✅ RabbitMQ Publisher desconectado")
	return nil
}

// IsConnected verifica se o publisher está conectado
func (p *Publisher) IsConnected() bool {
	return p.connected && p.conn != nil && !p.conn.IsClosed()
}
