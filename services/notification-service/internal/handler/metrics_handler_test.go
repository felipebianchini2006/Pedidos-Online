package handler

import (
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
)

// TestMetricsHandler_Metrics testa o endpoint de métricas JSON
func TestMetricsHandler_Metrics(t *testing.T) {
	handler := NewMetricsHandler()

	app := fiber.New()
	app.Get("/metrics", handler.Metrics)

	req := httptest.NewRequest("GET", "/metrics", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("Erro ao fazer requisição: %v", err)
	}
	defer resp.Body.Close()

	// Verificar status code
	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("Status code inesperado: got %d, want %d", resp.StatusCode, fiber.StatusOK)
	}

	// Verificar Content-Type
	contentType := resp.Header.Get("Content-Type")
	if !strings.Contains(contentType, "application/json") {
		t.Errorf("Content-Type inesperado: got %s, want application/json", contentType)
	}

	// Ler corpo
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Erro ao ler corpo da resposta: %v", err)
	}

	if len(body) == 0 {
		t.Error("Corpo da resposta está vazio")
	}

	// Verificar se contém campos esperados
	bodyStr := string(body)
	expectedFields := []string{
		"emails_sent_success",
		"emails_sent_failure",
		"messages_order_created",
		"messages_order_updated",
		"uptime",
	}

	for _, field := range expectedFields {
		if !strings.Contains(bodyStr, field) {
			t.Errorf("Campo %s não encontrado na resposta", field)
		}
	}

	t.Logf("Metrics JSON response: %s", string(body))
}

// TestMetricsHandler_MetricsPrometheus testa o endpoint Prometheus
func TestMetricsHandler_MetricsPrometheus(t *testing.T) {
	handler := NewMetricsHandler()

	app := fiber.New()
	app.Get("/metrics/prometheus", handler.MetricsPrometheus)

	req := httptest.NewRequest("GET", "/metrics/prometheus", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("Erro ao fazer requisição: %v", err)
	}
	defer resp.Body.Close()

	// Verificar status code
	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("Status code inesperado: got %d, want %d", resp.StatusCode, fiber.StatusOK)
	}

	// Verificar Content-Type
	contentType := resp.Header.Get("Content-Type")
	if !strings.Contains(contentType, "text/plain") {
		t.Errorf("Content-Type inesperado: got %s, want text/plain", contentType)
	}

	// Ler corpo
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Erro ao ler corpo da resposta: %v", err)
	}

	if len(body) == 0 {
		t.Error("Corpo da resposta está vazio")
	}

	// Verificar formato Prometheus
	bodyStr := string(body)
	expectedMetrics := []string{
		"notification_emails_sent_total",
		"notification_messages_processed_total",
		"notification_processing_time_avg_ms",
		"# HELP",
		"# TYPE",
	}

	for _, metric := range expectedMetrics {
		if !strings.Contains(bodyStr, metric) {
			t.Errorf("Métrica %s não encontrada na resposta Prometheus", metric)
		}
	}

	t.Logf("Metrics Prometheus response:\n%s", string(body))
}

// TestNewMetricsHandler verifica a criação do handler
func TestNewMetricsHandler(t *testing.T) {
	handler := NewMetricsHandler()

	if handler == nil {
		t.Fatal("NewMetricsHandler retornou nil")
	}

	if handler.metrics == nil {
		t.Error("Metrics não foram inicializadas")
	}
}
