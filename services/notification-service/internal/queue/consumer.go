package queue

import (
	"fmt"
	"log"
	"math"
	"pedidos-online/notification-service/internal/email"
	"pedidos-online/notification-service/internal/model"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// Consumer consome mensagens do RabbitMQ e processa notificações
type Consumer struct {
	conn          *amqp.Connection
	channel       *amqp.Channel
	emailService  *email.EmailService
	rabbitMQURL   string
	reconnecting  bool
	maxRetries    int           // Máximo de tentativas de reprocessamento
	retryDelay    time.Duration // Delay inicial para retry
	maxRetryDelay time.Duration // Delay máximo para retry
}

const (
	exchangeName           = "orders"                          // Exchange principal
	exchangeType           = "topic"                           // Tipo topic para routing keys
	queueOrderCreated      = "notifications.order.created"     // Queue para pedidos criados
	queueOrderUpdated      = "notifications.order.updated"     // Queue para pedidos atualizados
	routingKeyCreated      = "order.created"                   // Routing key para criação
	routingKeyUpdated      = "order.updated"                   // Routing key para atualização
	deadLetterExchange     = "orders.dlx"                      // Dead Letter Exchange
	deadLetterQueueCreated = "notifications.order.created.dlq" // DLQ para criação
	deadLetterQueueUpdated = "notifications.order.updated.dlq" // DLQ para atualização
	maxRedeliveryAttempts  = 3                                 // Máximo de tentativas antes de ir para DLQ
)

// NewConsumer cria uma nova instância do Consumer
//
// Parâmetros:
//   - rabbitMQURL: URL de conexão do RabbitMQ (ex: amqp://guest:guest@localhost:5672/)
//   - emailService: Serviço de e-mail para envio de notificações
//
// Retorno:
//   - *Consumer: Instância do consumer
//   - error: Erro ao criar consumer (se houver)
//
// Exemplo:
//
//	consumer, err := queue.NewConsumer(cfg.RabbitMQURL, emailService)
func NewConsumer(rabbitMQURL string, emailService *email.EmailService) (*Consumer, error) {
	log.Println("🐰 Inicializando RabbitMQ Consumer...")

	consumer := &Consumer{
		emailService:  emailService,
		rabbitMQURL:   rabbitMQURL,
		maxRetries:    3,
		retryDelay:    1 * time.Second,
		maxRetryDelay: 30 * time.Second,
	}

	// Conectar ao RabbitMQ
	if err := consumer.connect(); err != nil {
		return nil, err
	}

	// Configurar topologia (exchange, queues, bindings)
	if err := consumer.setupTopology(); err != nil {
		return nil, err
	}

	log.Println("✅ RabbitMQ Consumer inicializado com sucesso")
	return consumer, nil
}

// connect estabelece conexão com o RabbitMQ
func (c *Consumer) connect() error {
	log.Printf("🔌 Conectando ao RabbitMQ: %s", c.maskURL(c.rabbitMQURL))

	conn, err := amqp.Dial(c.rabbitMQURL)
	if err != nil {
		return fmt.Errorf("erro ao conectar ao RabbitMQ: %w", err)
	}

	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return fmt.Errorf("erro ao abrir channel: %w", err)
	}

	// Configurar QoS (prefetch)
	// Processa apenas 1 mensagem por vez para garantir processamento sequencial
	if err := channel.Qos(1, 0, false); err != nil {
		channel.Close()
		conn.Close()
		return fmt.Errorf("erro ao configurar QoS: %w", err)
	}

	c.conn = conn
	c.channel = channel

	log.Println("✅ Conectado ao RabbitMQ")
	return nil
}

// setupTopology configura exchange, queues e bindings
func (c *Consumer) setupTopology() error {
	log.Println("🔧 Configurando topologia RabbitMQ...")

	// 1. Declarar Dead Letter Exchange (DLX)
	if err := c.channel.ExchangeDeclare(
		deadLetterExchange, // name
		"direct",           // type (direct para DLQ)
		true,               // durable
		false,              // auto-deleted
		false,              // internal
		false,              // no-wait
		nil,                // arguments
	); err != nil {
		return fmt.Errorf("erro ao declarar DLX: %w", err)
	}
	log.Printf("✅ Dead Letter Exchange '%s' declarado", deadLetterExchange)

	// 2. Declarar Dead Letter Queues
	if err := c.declareDLQ(deadLetterQueueCreated); err != nil {
		return err
	}
	if err := c.declareDLQ(deadLetterQueueUpdated); err != nil {
		return err
	}

	// 3. Declarar Exchange principal
	if err := c.channel.ExchangeDeclare(
		exchangeName, // name
		exchangeType, // type (topic)
		true,         // durable
		false,        // auto-deleted
		false,        // internal
		false,        // no-wait
		nil,          // arguments
	); err != nil {
		return fmt.Errorf("erro ao declarar exchange: %w", err)
	}
	log.Printf("✅ Exchange '%s' (tipo: %s) declarado", exchangeName, exchangeType)

	// 4. Declarar Queues principais com DLX
	if err := c.declareQueueWithDLX(queueOrderCreated, routingKeyCreated); err != nil {
		return err
	}
	if err := c.declareQueueWithDLX(queueOrderUpdated, routingKeyUpdated); err != nil {
		return err
	}

	log.Println("✅ Topologia RabbitMQ configurada com sucesso")
	return nil
}

// declareDLQ declara uma Dead Letter Queue
func (c *Consumer) declareDLQ(queueName string) error {
	_, err := c.channel.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		return fmt.Errorf("erro ao declarar DLQ '%s': %w", queueName, err)
	}

	// Bind DLQ ao DLX
	if err := c.channel.QueueBind(
		queueName,          // queue
		queueName,          // routing key (mesmo nome da queue)
		deadLetterExchange, // exchange
		false,              // no-wait
		nil,                // arguments
	); err != nil {
		return fmt.Errorf("erro ao bind DLQ '%s': %w", queueName, err)
	}

	log.Printf("✅ Dead Letter Queue '%s' declarada e vinculada", queueName)
	return nil
}

// declareQueueWithDLX declara uma queue com Dead Letter Exchange
func (c *Consumer) declareQueueWithDLX(queueName, routingKey string) error {
	// Arguments para configurar DLX
	args := amqp.Table{
		"x-dead-letter-exchange":    deadLetterExchange, // DLX para mensagens rejeitadas
		"x-dead-letter-routing-key": queueName + ".dlq", // Routing key para DLQ
	}

	// Declarar queue
	_, err := c.channel.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		args,      // arguments (DLX)
	)
	if err != nil {
		return fmt.Errorf("erro ao declarar queue '%s': %w", queueName, err)
	}

	// Bind queue ao exchange
	if err := c.channel.QueueBind(
		queueName,    // queue
		routingKey,   // routing key
		exchangeName, // exchange
		false,        // no-wait
		nil,          // arguments
	); err != nil {
		return fmt.Errorf("erro ao bind queue '%s': %w", queueName, err)
	}

	log.Printf("✅ Queue '%s' (routing key: %s) declarada e vinculada", queueName, routingKey)
	return nil
}

// Start inicia o consumo de mensagens das queues
//
// Este método bloqueia a goroutine atual e processa mensagens indefinidamente.
// Use em uma goroutine separada se necessário.
//
// Retorno:
//   - error: Erro ao iniciar consumo (se houver)
//
// Exemplo:
//
//	go func() {
//	    if err := consumer.Start(); err != nil {
//	        log.Fatalf("Erro no consumer: %v", err)
//	    }
//	}()
func (c *Consumer) Start() error {
	log.Println("🚀 Iniciando consumo de mensagens...")

	// Iniciar consumo da queue de pedidos criados
	deliveriesCreated, err := c.channel.Consume(
		queueOrderCreated, // queue
		"",                // consumer tag (gerado automaticamente)
		false,             // auto-ack (FALSE - ACK manual)
		false,             // exclusive
		false,             // no-local
		false,             // no-wait
		nil,               // args
	)
	if err != nil {
		return fmt.Errorf("erro ao consumir queue '%s': %w", queueOrderCreated, err)
	}
	log.Printf("✅ Consumindo mensagens de '%s'", queueOrderCreated)

	// Iniciar consumo da queue de pedidos atualizados
	deliveriesUpdated, err := c.channel.Consume(
		queueOrderUpdated, // queue
		"",                // consumer tag
		false,             // auto-ack (FALSE - ACK manual)
		false,             // exclusive
		false,             // no-local
		false,             // no-wait
		nil,               // args
	)
	if err != nil {
		return fmt.Errorf("erro ao consumir queue '%s': %w", queueOrderUpdated, err)
	}
	log.Printf("✅ Consumindo mensagens de '%s'", queueOrderUpdated)

	// Processar mensagens em goroutines separadas
	go c.processDeliveries(deliveriesCreated, "order.created")
	go c.processDeliveries(deliveriesUpdated, "order.updated")

	log.Println("✅ Consumer rodando e aguardando mensagens...")

	// Bloquear para manter o consumer ativo
	select {}
}

// processDeliveries processa mensagens de um canal de deliveries
func (c *Consumer) processDeliveries(deliveries <-chan amqp.Delivery, eventType string) {
	for delivery := range deliveries {
		log.Printf("📨 Mensagem recebida: %s (DeliveryTag: %d)", eventType, delivery.DeliveryTag)

		// Verificar número de tentativas (x-death header)
		attempts := c.getDeliveryAttempts(delivery)
		log.Printf("   Tentativa: %d/%d", attempts, maxRedeliveryAttempts)

		// Processar com retry logic
		err := c.processDeliveryWithRetry(delivery, eventType, attempts)

		if err != nil {
			log.Printf("❌ Erro ao processar mensagem (tentativa %d/%d): %v", attempts, maxRedeliveryAttempts, err)

			// Se excedeu o número de tentativas, rejeitar sem requeue (vai para DLQ)
			if attempts >= maxRedeliveryAttempts {
				log.Printf("⚠️  Máximo de tentativas atingido. Enviando para DLQ...")
				if nackErr := delivery.Nack(false, false); nackErr != nil {
					log.Printf("❌ Erro ao NACK mensagem: %v", nackErr)
				} else {
					log.Printf("✅ Mensagem enviada para DLQ")
				}
			} else {
				// Rejeitar com requeue (será reprocessada)
				log.Printf("🔄 Reenfileirando mensagem para nova tentativa...")
				if nackErr := delivery.Nack(false, true); nackErr != nil {
					log.Printf("❌ Erro ao NACK mensagem: %v", nackErr)
				}
			}
		} else {
			// Sucesso - ACK manual
			log.Printf("✅ Mensagem processada com sucesso (DeliveryTag: %d)", delivery.DeliveryTag)
			if ackErr := delivery.Ack(false); ackErr != nil {
				log.Printf("❌ Erro ao ACK mensagem: %v", ackErr)
			}
		}
	}

	log.Printf("⚠️  Canal de deliveries '%s' fechado", eventType)
}

// processDeliveryWithRetry processa uma mensagem com retry logic e backoff exponencial
func (c *Consumer) processDeliveryWithRetry(delivery amqp.Delivery, eventType string, currentAttempt int) error {
	var err error

	// Calcular delay com backoff exponencial
	delay := c.calculateBackoffDelay(currentAttempt)

	// Aguardar delay antes de processar (exceto na primeira tentativa)
	if currentAttempt > 1 {
		log.Printf("⏳ Aguardando %v antes de reprocessar...", delay)
		time.Sleep(delay)
	}

	// Processar mensagem de acordo com o tipo
	switch eventType {
	case "order.created":
		err = c.handleOrderCreated(delivery)
	case "order.updated":
		err = c.handleOrderUpdated(delivery)
	default:
		err = fmt.Errorf("tipo de evento desconhecido: %s", eventType)
	}

	return err
}

// handleOrderCreated processa evento de pedido criado
func (c *Consumer) handleOrderCreated(delivery amqp.Delivery) error {
	log.Printf("🆕 Processando OrderCreatedEvent...")

	// Parse JSON para OrderCreatedEvent
	event, err := model.ParseOrderCreatedEvent(delivery.Body)
	if err != nil {
		return fmt.Errorf("erro ao fazer parse de OrderCreatedEvent: %w", err)
	}

	log.Printf("   %s", event.String())

	// Enviar e-mail de confirmação
	if err := c.emailService.SendOrderConfirmation(event); err != nil {
		return fmt.Errorf("erro ao enviar e-mail de confirmação: %w", err)
	}

	log.Printf("✅ OrderCreatedEvent processado: %s", event.OrderID)
	return nil
}

// handleOrderUpdated processa evento de pedido atualizado
func (c *Consumer) handleOrderUpdated(delivery amqp.Delivery) error {
	log.Printf("🔄 Processando OrderUpdatedEvent...")

	// Parse JSON para OrderUpdatedEvent
	event, err := model.ParseOrderUpdatedEvent(delivery.Body)
	if err != nil {
		return fmt.Errorf("erro ao fazer parse de OrderUpdatedEvent: %w", err)
	}

	log.Printf("   %s", event.String())

	// Enviar e-mail de atualização de status
	if err := c.emailService.SendOrderStatusUpdate(event); err != nil {
		return fmt.Errorf("erro ao enviar e-mail de atualização: %w", err)
	}

	log.Printf("✅ OrderUpdatedEvent processado: %s", event.OrderID)
	return nil
}

// getDeliveryAttempts retorna o número de tentativas de entrega
func (c *Consumer) getDeliveryAttempts(delivery amqp.Delivery) int {
	// Verificar header x-death para contar redeliveries
	if xDeath, ok := delivery.Headers["x-death"].([]interface{}); ok && len(xDeath) > 0 {
		if death, ok := xDeath[0].(amqp.Table); ok {
			if count, ok := death["count"].(int64); ok {
				return int(count) + 1
			}
		}
	}

	// Primeira tentativa
	return 1
}

// calculateBackoffDelay calcula o delay de retry com backoff exponencial
func (c *Consumer) calculateBackoffDelay(attempt int) time.Duration {
	// Backoff exponencial: delay * 2^(attempt-1)
	// Exemplo: 1s, 2s, 4s, 8s, 16s, 30s (max)
	delay := time.Duration(math.Pow(2, float64(attempt-1))) * c.retryDelay

	// Limitar ao delay máximo
	if delay > c.maxRetryDelay {
		delay = c.maxRetryDelay
	}

	return delay
}

// IsConnected verifica se o consumer está conectado ao RabbitMQ
func (c *Consumer) IsConnected() bool {
	return c.conn != nil && !c.conn.IsClosed()
}

// Close fecha a conexão com o RabbitMQ
func (c *Consumer) Close() error {
	log.Println("🔌 Fechando conexão com RabbitMQ...")

	if c.channel != nil {
		if err := c.channel.Close(); err != nil {
			log.Printf("⚠️  Erro ao fechar channel: %v", err)
		}
	}

	if c.conn != nil {
		if err := c.conn.Close(); err != nil {
			log.Printf("⚠️  Erro ao fechar conexão: %v", err)
		}
	}

	log.Println("✅ Conexão com RabbitMQ fechada")
	return nil
}

// maskURL mascara a senha na URL do RabbitMQ para logs
func (c *Consumer) maskURL(url string) string {
	if len(url) > 30 {
		return url[:15] + "***" + url[len(url)-10:]
	}
	return "***"
}
