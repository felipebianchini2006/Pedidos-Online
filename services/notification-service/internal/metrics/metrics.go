package metrics

import (
	"fmt"
	"sync"
	"time"
)

// Metrics armazena métricas do Notification Service
type Metrics struct {
	mu sync.RWMutex

	// Contadores de e-mails
	EmailsSentSuccess int64 `json:"emails_sent_success"` // E-mails enviados com sucesso
	EmailsSentFailure int64 `json:"emails_sent_failure"` // E-mails que falharam

	// Contadores de mensagens processadas por tipo
	MessagesOrderCreated  int64 `json:"messages_order_created"`  // Eventos order.created processados
	MessagesOrderUpdated  int64 `json:"messages_order_updated"`  // Eventos order.updated processados
	MessagesProcessFailed int64 `json:"messages_process_failed"` // Mensagens que falharam no processamento

	// Tempo de processamento
	TotalProcessingTime time.Duration `json:"total_processing_time_ms"` // Tempo total de processamento
	ProcessedMessages   int64         `json:"processed_messages"`       // Total de mensagens processadas com sucesso

	// Timestamp de início
	StartTime time.Time `json:"start_time"`
}

var (
	// globalMetrics é a instância global de métricas
	globalMetrics *Metrics
	once          sync.Once
)

// GetMetrics retorna a instância global de métricas
func GetMetrics() *Metrics {
	once.Do(func() {
		globalMetrics = &Metrics{
			StartTime: time.Now(),
		}
	})
	return globalMetrics
}

// IncrementEmailSuccess incrementa o contador de e-mails enviados com sucesso
func (m *Metrics) IncrementEmailSuccess() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.EmailsSentSuccess++
}

// IncrementEmailFailure incrementa o contador de e-mails que falharam
func (m *Metrics) IncrementEmailFailure() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.EmailsSentFailure++
}

// IncrementOrderCreated incrementa o contador de eventos order.created
func (m *Metrics) IncrementOrderCreated() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.MessagesOrderCreated++
}

// IncrementOrderUpdated incrementa o contador de eventos order.updated
func (m *Metrics) IncrementOrderUpdated() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.MessagesOrderUpdated++
}

// IncrementProcessFailed incrementa o contador de mensagens que falharam
func (m *Metrics) IncrementProcessFailed() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.MessagesProcessFailed++
}

// RecordProcessingTime registra o tempo de processamento de uma mensagem
func (m *Metrics) RecordProcessingTime(duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.TotalProcessingTime += duration
	m.ProcessedMessages++
}

// GetAverageProcessingTime retorna o tempo médio de processamento
func (m *Metrics) GetAverageProcessingTime() time.Duration {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.ProcessedMessages == 0 {
		return 0
	}

	return m.TotalProcessingTime / time.Duration(m.ProcessedMessages)
}

// GetSnapshot retorna um snapshot thread-safe das métricas
func (m *Metrics) GetSnapshot() MetricsSnapshot {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return MetricsSnapshot{
		EmailsSentSuccess:     m.EmailsSentSuccess,
		EmailsSentFailure:     m.EmailsSentFailure,
		MessagesOrderCreated:  m.MessagesOrderCreated,
		MessagesOrderUpdated:  m.MessagesOrderUpdated,
		MessagesProcessFailed: m.MessagesProcessFailed,
		AverageProcessingTime: m.GetAverageProcessingTime(),
		TotalMessages:         m.MessagesOrderCreated + m.MessagesOrderUpdated,
		Uptime:                time.Since(m.StartTime).Round(time.Second).String(),
	}
}

// MetricsSnapshot representa um snapshot das métricas em um momento específico
type MetricsSnapshot struct {
	EmailsSentSuccess     int64         `json:"emails_sent_success"`
	EmailsSentFailure     int64         `json:"emails_sent_failure"`
	MessagesOrderCreated  int64         `json:"messages_order_created"`
	MessagesOrderUpdated  int64         `json:"messages_order_updated"`
	MessagesProcessFailed int64         `json:"messages_process_failed"`
	AverageProcessingTime time.Duration `json:"average_processing_time_ms"`
	TotalMessages         int64         `json:"total_messages_processed"`
	Uptime                string        `json:"uptime"`
}

// FormatPrometheus formata as métricas no formato Prometheus
func (m *Metrics) FormatPrometheus() string {
	snapshot := m.GetSnapshot()

	output := ""

	// HELP e TYPE para cada métrica
	output += "# HELP notification_emails_sent_total Total de e-mails enviados\n"
	output += "# TYPE notification_emails_sent_total counter\n"
	output += fmt.Sprintf("notification_emails_sent_total{status=\"success\"} %d\n", snapshot.EmailsSentSuccess)
	output += fmt.Sprintf("notification_emails_sent_total{status=\"failure\"} %d\n", snapshot.EmailsSentFailure)
	output += "\n"

	output += "# HELP notification_messages_processed_total Total de mensagens processadas por tipo\n"
	output += "# TYPE notification_messages_processed_total counter\n"
	output += fmt.Sprintf("notification_messages_processed_total{event_type=\"order.created\"} %d\n", snapshot.MessagesOrderCreated)
	output += fmt.Sprintf("notification_messages_processed_total{event_type=\"order.updated\"} %d\n", snapshot.MessagesOrderUpdated)
	output += "\n"

	output += "# HELP notification_messages_failed_total Total de mensagens que falharam no processamento\n"
	output += "# TYPE notification_messages_failed_total counter\n"
	output += fmt.Sprintf("notification_messages_failed_total %d\n", snapshot.MessagesProcessFailed)
	output += "\n"

	output += "# HELP notification_processing_time_avg_ms Tempo médio de processamento de mensagens em milissegundos\n"
	output += "# TYPE notification_processing_time_avg_ms gauge\n"
	output += fmt.Sprintf("notification_processing_time_avg_ms %d\n", snapshot.AverageProcessingTime.Milliseconds())
	output += "\n"

	output += "# HELP notification_uptime_seconds Tempo desde o início do serviço em segundos\n"
	output += "# TYPE notification_uptime_seconds gauge\n"
	output += fmt.Sprintf("notification_uptime_seconds %d\n", int64(time.Since(m.StartTime).Seconds()))
	output += "\n"

	return output
}

// Reset reseta todas as métricas (útil para testes)
func (m *Metrics) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.EmailsSentSuccess = 0
	m.EmailsSentFailure = 0
	m.MessagesOrderCreated = 0
	m.MessagesOrderUpdated = 0
	m.MessagesProcessFailed = 0
	m.TotalProcessingTime = 0
	m.ProcessedMessages = 0
	m.StartTime = time.Now()
}
