package handler

import (
	"io"
	"net/http/httptest"
	"testing"

	"pedidos-online/notification-service/internal/config"

	"github.com/gofiber/fiber/v2"
)

// TestHealthHandler_Health testa o endpoint de health check
func TestHealthHandler_Health(t *testing.T) {
	// Configuração de teste
	cfg := &config.Config{
		RabbitMQURL: "amqp://guest:guest@localhost:5672/",
		SMTP: config.SMTPConfig{
			Host:     "smtp.gmail.com",
			Port:     587,
			User:     "test@example.com",
			Password: "testpassword",
			From:     "test@example.com",
		},
	}

	handler := NewHealthHandler(cfg)

	// Criar aplicação Fiber para teste
	app := fiber.New()
	app.Get("/health", handler.Health)

	// Teste: Fazer requisição ao endpoint
	req := httptest.NewRequest("GET", "/health", nil)
	resp, err := app.Test(req, -1) // -1 = sem timeout
	if err != nil {
		t.Fatalf("Erro ao fazer requisição: %v", err)
	}
	defer resp.Body.Close()

	// Verificar status code (pode ser 200 ou 503 dependendo do RabbitMQ estar rodando)
	if resp.StatusCode != fiber.StatusOK && resp.StatusCode != fiber.StatusServiceUnavailable {
		t.Errorf("Status code inesperado: got %d, want %d ou %d", resp.StatusCode, fiber.StatusOK, fiber.StatusServiceUnavailable)
	}

	// Ler corpo da resposta
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Erro ao ler corpo da resposta: %v", err)
	}

	// Verificar se o corpo não está vazio
	if len(body) == 0 {
		t.Error("Corpo da resposta está vazio")
	}

	t.Logf("Health check response: %s", string(body))
}

// TestReadinessHandler_Ready testa o endpoint de readiness
func TestReadinessHandler_Ready(t *testing.T) {
	cfg := &config.Config{
		RabbitMQURL: "amqp://guest:guest@localhost:5672/",
		SMTP: config.SMTPConfig{
			Host:     "smtp.gmail.com",
			Port:     587,
			User:     "test@example.com",
			Password: "testpassword",
			From:     "test@example.com",
		},
	}

	handler := NewReadinessHandler(cfg)

	app := fiber.New()
	app.Get("/ready", handler.Ready)

	req := httptest.NewRequest("GET", "/ready", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("Erro ao fazer requisição: %v", err)
	}
	defer resp.Body.Close()

	// Verificar status code (pode ser 200 ou 503 dependendo do RabbitMQ)
	if resp.StatusCode != fiber.StatusOK && resp.StatusCode != fiber.StatusServiceUnavailable {
		t.Errorf("Status code inesperado: got %d, want %d ou %d", resp.StatusCode, fiber.StatusOK, fiber.StatusServiceUnavailable)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Erro ao ler corpo da resposta: %v", err)
	}

	if len(body) == 0 {
		t.Error("Corpo da resposta está vazio")
	}

	t.Logf("Readiness check response: %s", string(body))
}

// TestNewHealthHandler verifica a criação do handler
func TestNewHealthHandler(t *testing.T) {
	cfg := &config.Config{
		RabbitMQURL: "amqp://guest:guest@localhost:5672/",
		SMTP: config.SMTPConfig{
			Host: "smtp.gmail.com",
			Port: 587,
		},
	}

	handler := NewHealthHandler(cfg)

	if handler == nil {
		t.Fatal("NewHealthHandler retornou nil")
	}

	if handler.config != cfg {
		t.Error("Config não foi setada corretamente")
	}

	if handler.rabbitMQURL != cfg.RabbitMQURL {
		t.Errorf("RabbitMQURL incorreta: got %s, want %s", handler.rabbitMQURL, cfg.RabbitMQURL)
	}
}

// TestNewReadinessHandler verifica a criação do handler
func TestNewReadinessHandler(t *testing.T) {
	cfg := &config.Config{
		RabbitMQURL: "amqp://guest:guest@localhost:5672/",
	}

	handler := NewReadinessHandler(cfg)

	if handler == nil {
		t.Fatal("NewReadinessHandler retornou nil")
	}

	if handler.handler == nil {
		t.Error("Handler interno não foi criado")
	}
}
