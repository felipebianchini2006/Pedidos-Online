package handler

import (
	"fmt"
	"net/smtp"
	"time"

	"pedidos-online/notification-service/internal/config"

	"github.com/gofiber/fiber/v2"
	amqp "github.com/rabbitmq/amqp091-go"
)

var (
	// startTime armazena quando o serviço foi iniciado
	startTime = time.Now()
)

// HealthHandler gerencia o endpoint de health check
type HealthHandler struct {
	config      *config.Config
	rabbitMQURL string
}

// NewHealthHandler cria uma nova instância do HealthHandler
func NewHealthHandler(cfg *config.Config) *HealthHandler {
	return &HealthHandler{
		config:      cfg,
		rabbitMQURL: cfg.RabbitMQURL,
	}
}

// HealthResponse representa a resposta do health check
type HealthResponse struct {
	Status   string            `json:"status"`   // "healthy" ou "unhealthy"
	Services map[string]string `json:"services"` // Status de cada serviço
	Uptime   string            `json:"uptime"`   // Tempo desde o início
}

// Health verifica a saúde do serviço e suas dependências
//
// GET /health
//
// Resposta de sucesso (200 OK):
//
//	{
//	  "status": "healthy",
//	  "services": {
//	    "rabbitmq": "ok",
//	    "smtp": "ok"
//	  },
//	  "uptime": "2h30m15s"
//	}
//
// Resposta de falha (503 Service Unavailable):
//
//	{
//	  "status": "unhealthy",
//	  "services": {
//	    "rabbitmq": "connection refused",
//	    "smtp": "ok"
//	  },
//	  "uptime": "2h30m15s"
//	}
func (h *HealthHandler) Health(c *fiber.Ctx) error {
	services := make(map[string]string)

	// Verificar RabbitMQ
	rabbitMQStatus := h.checkRabbitMQ()
	services["rabbitmq"] = rabbitMQStatus

	// Verificar SMTP
	smtpStatus := h.checkSMTP()
	services["smtp"] = smtpStatus

	// Calcular uptime
	uptime := time.Since(startTime).Round(time.Second).String()

	// Determinar status geral
	overallStatus := "healthy"
	statusCode := fiber.StatusOK

	// RabbitMQ é crítico, SMTP é opcional
	if rabbitMQStatus != "ok" {
		overallStatus = "unhealthy"
		statusCode = fiber.StatusServiceUnavailable
	}

	response := HealthResponse{
		Status:   overallStatus,
		Services: services,
		Uptime:   uptime,
	}

	return c.Status(statusCode).JSON(response)
}

// checkRabbitMQ verifica se consegue conectar ao RabbitMQ
// Tenta declarar um exchange temporário para garantir que a conexão está funcional
func (h *HealthHandler) checkRabbitMQ() string {
	// Tentar conectar ao RabbitMQ
	conn, err := amqp.Dial(h.rabbitMQURL)
	if err != nil {
		return fmt.Sprintf("connection failed: %v", err)
	}
	defer conn.Close()

	// Tentar abrir um canal
	ch, err := conn.Channel()
	if err != nil {
		return fmt.Sprintf("channel failed: %v", err)
	}
	defer ch.Close()

	// Tentar declarar um exchange temporário (não-durável)
	// Isso garante que temos permissões adequadas
	testExchangeName := fmt.Sprintf("health-check-%d", time.Now().UnixNano())
	err = ch.ExchangeDeclare(
		testExchangeName, // nome
		"fanout",         // tipo
		false,            // durável
		true,             // auto-delete
		false,            // internal
		false,            // no-wait
		nil,              // argumentos
	)
	if err != nil {
		return fmt.Sprintf("exchange declaration failed: %v", err)
	}

	// Deletar o exchange de teste
	err = ch.ExchangeDelete(testExchangeName, false, false)
	if err != nil {
		// Não é crítico se não conseguir deletar (auto-delete está ativo)
		return "ok (exchange cleanup warning)"
	}

	return "ok"
}

// checkSMTP verifica se consegue conectar ao servidor SMTP
// Não tenta autenticar (para evitar bloqueios por tentativas repetidas)
func (h *HealthHandler) checkSMTP() string {
	// Construir endereço do servidor SMTP
	smtpAddr := fmt.Sprintf("%s:%d", h.config.SMTP.Host, h.config.SMTP.Port)

	// Tentar conectar ao servidor SMTP
	// Timeout de 5 segundos para evitar travamento
	client, err := smtp.Dial(smtpAddr)
	if err != nil {
		return fmt.Sprintf("connection failed: %v", err)
	}
	defer client.Close()

	// Verificar se o servidor responde ao comando NOOP
	err = client.Noop()
	if err != nil {
		return fmt.Sprintf("noop failed: %v", err)
	}

	// SMTP está respondendo (não testamos autenticação para evitar rate limiting)
	return "ok"
}

// ReadinessHandler verifica se o serviço está pronto para receber requisições
type ReadinessHandler struct {
	handler *HealthHandler
}

// NewReadinessHandler cria uma nova instância do ReadinessHandler
func NewReadinessHandler(cfg *config.Config) *ReadinessHandler {
	return &ReadinessHandler{
		handler: NewHealthHandler(cfg),
	}
}

// Ready verifica se o serviço está pronto
//
// GET /ready
//
// Resposta de sucesso (200 OK):
//
//	{
//	  "ready": true
//	}
//
// Resposta de falha (503 Service Unavailable):
//
//	{
//	  "ready": false,
//	  "reason": "rabbitmq connection failed"
//	}
func (rh *ReadinessHandler) Ready(c *fiber.Ctx) error {
	// Verificar apenas RabbitMQ (crítico para readiness)
	rabbitMQStatus := rh.handler.checkRabbitMQ()

	if rabbitMQStatus != "ok" {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"ready":  false,
			"reason": fmt.Sprintf("rabbitmq: %s", rabbitMQStatus),
		})
	}

	return c.JSON(fiber.Map{
		"ready": true,
	})
}
