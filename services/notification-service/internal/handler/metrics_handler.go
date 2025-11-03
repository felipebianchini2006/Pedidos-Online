package handler

import (
	"pedidos-online/notification-service/internal/metrics"

	"github.com/gofiber/fiber/v2"
)

// MetricsHandler gerencia o endpoint de métricas
type MetricsHandler struct {
	metrics *metrics.Metrics
}

// NewMetricsHandler cria uma nova instância do MetricsHandler
func NewMetricsHandler() *MetricsHandler {
	return &MetricsHandler{
		metrics: metrics.GetMetrics(),
	}
}

// Metrics retorna as métricas do serviço em formato JSON
//
// GET /metrics
//
// Resposta (200 OK):
//
//	{
//	  "emails_sent_success": 150,
//	  "emails_sent_failure": 3,
//	  "messages_order_created": 75,
//	  "messages_order_updated": 78,
//	  "messages_process_failed": 2,
//	  "average_processing_time_ms": 250,
//	  "total_messages_processed": 153,
//	  "uptime": "2h30m15s"
//	}
func (h *MetricsHandler) Metrics(c *fiber.Ctx) error {
	snapshot := h.metrics.GetSnapshot()
	return c.JSON(snapshot)
}

// MetricsPrometheus retorna as métricas em formato Prometheus
//
// GET /metrics/prometheus
//
// Resposta (200 OK - text/plain):
//
//	# HELP notification_emails_sent_total Total de e-mails enviados
//	# TYPE notification_emails_sent_total counter
//	notification_emails_sent_total{status="success"} 150
//	notification_emails_sent_total{status="failure"} 3
//	...
func (h *MetricsHandler) MetricsPrometheus(c *fiber.Ctx) error {
	prometheusFormat := h.metrics.FormatPrometheus()
	c.Set("Content-Type", "text/plain; version=0.0.4")
	return c.SendString(prometheusFormat)
}
