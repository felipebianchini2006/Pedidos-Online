package metrics

import (
	"testing"
	"time"
)

// TestGetMetrics testa a criação da instância global
func TestGetMetrics(t *testing.T) {
	m1 := GetMetrics()
	m2 := GetMetrics()

	if m1 != m2 {
		t.Error("GetMetrics deve retornar a mesma instância (singleton)")
	}

	if m1 == nil {
		t.Fatal("GetMetrics retornou nil")
	}
}

// TestIncrementEmailSuccess testa incremento de e-mails com sucesso
func TestIncrementEmailSuccess(t *testing.T) {
	m := &Metrics{}

	if m.EmailsSentSuccess != 0 {
		t.Errorf("Contador inicial deveria ser 0, got %d", m.EmailsSentSuccess)
	}

	m.IncrementEmailSuccess()
	if m.EmailsSentSuccess != 1 {
		t.Errorf("Após incremento, esperado 1, got %d", m.EmailsSentSuccess)
	}

	m.IncrementEmailSuccess()
	m.IncrementEmailSuccess()
	if m.EmailsSentSuccess != 3 {
		t.Errorf("Após 3 incrementos, esperado 3, got %d", m.EmailsSentSuccess)
	}
}

// TestIncrementEmailFailure testa incremento de e-mails com falha
func TestIncrementEmailFailure(t *testing.T) {
	m := &Metrics{}

	m.IncrementEmailFailure()
	if m.EmailsSentFailure != 1 {
		t.Errorf("Esperado 1, got %d", m.EmailsSentFailure)
	}
}

// TestIncrementOrderCreated testa incremento de eventos order.created
func TestIncrementOrderCreated(t *testing.T) {
	m := &Metrics{}

	m.IncrementOrderCreated()
	if m.MessagesOrderCreated != 1 {
		t.Errorf("Esperado 1, got %d", m.MessagesOrderCreated)
	}
}

// TestIncrementOrderUpdated testa incremento de eventos order.updated
func TestIncrementOrderUpdated(t *testing.T) {
	m := &Metrics{}

	m.IncrementOrderUpdated()
	if m.MessagesOrderUpdated != 1 {
		t.Errorf("Esperado 1, got %d", m.MessagesOrderUpdated)
	}
}

// TestIncrementProcessFailed testa incremento de mensagens com falha
func TestIncrementProcessFailed(t *testing.T) {
	m := &Metrics{}

	m.IncrementProcessFailed()
	if m.MessagesProcessFailed != 1 {
		t.Errorf("Esperado 1, got %d", m.MessagesProcessFailed)
	}
}

// TestRecordProcessingTime testa registro de tempo de processamento
func TestRecordProcessingTime(t *testing.T) {
	m := &Metrics{}

	m.RecordProcessingTime(100 * time.Millisecond)
	if m.ProcessedMessages != 1 {
		t.Errorf("Esperado 1 mensagem processada, got %d", m.ProcessedMessages)
	}

	if m.TotalProcessingTime != 100*time.Millisecond {
		t.Errorf("Tempo total esperado 100ms, got %v", m.TotalProcessingTime)
	}

	m.RecordProcessingTime(200 * time.Millisecond)
	if m.ProcessedMessages != 2 {
		t.Errorf("Esperado 2 mensagens processadas, got %d", m.ProcessedMessages)
	}

	if m.TotalProcessingTime != 300*time.Millisecond {
		t.Errorf("Tempo total esperado 300ms, got %v", m.TotalProcessingTime)
	}
}

// TestGetAverageProcessingTime testa cálculo de tempo médio
func TestGetAverageProcessingTime(t *testing.T) {
	m := &Metrics{}

	// Caso 1: Sem mensagens processadas
	avg := m.GetAverageProcessingTime()
	if avg != 0 {
		t.Errorf("Média deveria ser 0 quando não há mensagens, got %v", avg)
	}

	// Caso 2: Uma mensagem
	m.RecordProcessingTime(100 * time.Millisecond)
	avg = m.GetAverageProcessingTime()
	if avg != 100*time.Millisecond {
		t.Errorf("Média esperada 100ms, got %v", avg)
	}

	// Caso 3: Múltiplas mensagens
	m.RecordProcessingTime(200 * time.Millisecond)
	m.RecordProcessingTime(300 * time.Millisecond)
	avg = m.GetAverageProcessingTime()
	expected := 200 * time.Millisecond // (100 + 200 + 300) / 3
	if avg != expected {
		t.Errorf("Média esperada %v, got %v", expected, avg)
	}
}

// TestGetSnapshot testa criação de snapshot
func TestGetSnapshot(t *testing.T) {
	m := &Metrics{
		StartTime: time.Now().Add(-1 * time.Hour),
	}

	m.IncrementEmailSuccess()
	m.IncrementEmailSuccess()
	m.IncrementEmailFailure()
	m.IncrementOrderCreated()
	m.IncrementOrderUpdated()
	m.RecordProcessingTime(100 * time.Millisecond)

	snapshot := m.GetSnapshot()

	if snapshot.EmailsSentSuccess != 2 {
		t.Errorf("EmailsSentSuccess: esperado 2, got %d", snapshot.EmailsSentSuccess)
	}

	if snapshot.EmailsSentFailure != 1 {
		t.Errorf("EmailsSentFailure: esperado 1, got %d", snapshot.EmailsSentFailure)
	}

	if snapshot.MessagesOrderCreated != 1 {
		t.Errorf("MessagesOrderCreated: esperado 1, got %d", snapshot.MessagesOrderCreated)
	}

	if snapshot.MessagesOrderUpdated != 1 {
		t.Errorf("MessagesOrderUpdated: esperado 1, got %d", snapshot.MessagesOrderUpdated)
	}

	if snapshot.TotalMessages != 2 {
		t.Errorf("TotalMessages: esperado 2, got %d", snapshot.TotalMessages)
	}

	if snapshot.AverageProcessingTime != 100*time.Millisecond {
		t.Errorf("AverageProcessingTime: esperado 100ms, got %v", snapshot.AverageProcessingTime)
	}

	if snapshot.Uptime == "" {
		t.Error("Uptime está vazio")
	}
}

// TestFormatPrometheus testa formatação Prometheus
func TestFormatPrometheus(t *testing.T) {
	m := &Metrics{
		StartTime: time.Now(),
	}

	m.IncrementEmailSuccess()
	m.IncrementEmailFailure()
	m.IncrementOrderCreated()

	output := m.FormatPrometheus()

	// Verificar se contém métricas esperadas
	expectedStrings := []string{
		"# HELP notification_emails_sent_total",
		"# TYPE notification_emails_sent_total counter",
		"notification_emails_sent_total{status=\"success\"} 1",
		"notification_emails_sent_total{status=\"failure\"} 1",
		"notification_messages_processed_total{event_type=\"order.created\"} 1",
		"# HELP notification_processing_time_avg_ms",
		"# TYPE notification_processing_time_avg_ms gauge",
	}

	for _, expected := range expectedStrings {
		if !contains(output, expected) {
			t.Errorf("Output Prometheus não contém: %s", expected)
		}
	}
}

// TestReset testa reset de métricas
func TestReset(t *testing.T) {
	m := &Metrics{}

	// Adicionar algumas métricas
	m.IncrementEmailSuccess()
	m.IncrementEmailFailure()
	m.IncrementOrderCreated()
	m.RecordProcessingTime(100 * time.Millisecond)

	// Verificar que métricas foram setadas
	if m.EmailsSentSuccess == 0 {
		t.Error("EmailsSentSuccess deveria ser > 0 antes do reset")
	}

	// Reset
	m.Reset()

	// Verificar que tudo foi zerado
	if m.EmailsSentSuccess != 0 {
		t.Errorf("EmailsSentSuccess deveria ser 0 após reset, got %d", m.EmailsSentSuccess)
	}

	if m.EmailsSentFailure != 0 {
		t.Errorf("EmailsSentFailure deveria ser 0 após reset, got %d", m.EmailsSentFailure)
	}

	if m.MessagesOrderCreated != 0 {
		t.Errorf("MessagesOrderCreated deveria ser 0 após reset, got %d", m.MessagesOrderCreated)
	}

	if m.TotalProcessingTime != 0 {
		t.Errorf("TotalProcessingTime deveria ser 0 após reset, got %v", m.TotalProcessingTime)
	}

	if m.ProcessedMessages != 0 {
		t.Errorf("ProcessedMessages deveria ser 0 após reset, got %d", m.ProcessedMessages)
	}
}

// TestConcurrentAccess testa acesso concorrente (thread-safety)
func TestConcurrentAccess(t *testing.T) {
	m := &Metrics{}

	// Executar incrementos concorrentes
	done := make(chan bool)
	iterations := 1000

	go func() {
		for i := 0; i < iterations; i++ {
			m.IncrementEmailSuccess()
		}
		done <- true
	}()

	go func() {
		for i := 0; i < iterations; i++ {
			m.IncrementEmailFailure()
		}
		done <- true
	}()

	go func() {
		for i := 0; i < iterations; i++ {
			m.RecordProcessingTime(time.Millisecond)
		}
		done <- true
	}()

	// Aguardar todas as goroutines
	<-done
	<-done
	<-done

	// Verificar contadores
	if m.EmailsSentSuccess != int64(iterations) {
		t.Errorf("EmailsSentSuccess: esperado %d, got %d", iterations, m.EmailsSentSuccess)
	}

	if m.EmailsSentFailure != int64(iterations) {
		t.Errorf("EmailsSentFailure: esperado %d, got %d", iterations, m.EmailsSentFailure)
	}

	if m.ProcessedMessages != int64(iterations) {
		t.Errorf("ProcessedMessages: esperado %d, got %d", iterations, m.ProcessedMessages)
	}
}

// Helper function para verificar se string contém substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// BenchmarkIncrementEmailSuccess benchmarca incremento de e-mails
func BenchmarkIncrementEmailSuccess(b *testing.B) {
	m := &Metrics{}
	for i := 0; i < b.N; i++ {
		m.IncrementEmailSuccess()
	}
}

// BenchmarkGetSnapshot benchmarca criação de snapshot
func BenchmarkGetSnapshot(b *testing.B) {
	m := &Metrics{StartTime: time.Now()}
	m.IncrementEmailSuccess()
	m.RecordProcessingTime(100 * time.Millisecond)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = m.GetSnapshot()
	}
}

// BenchmarkFormatPrometheus benchmarca formatação Prometheus
func BenchmarkFormatPrometheus(b *testing.B) {
	m := &Metrics{StartTime: time.Now()}
	m.IncrementEmailSuccess()
	m.IncrementOrderCreated()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = m.FormatPrometheus()
	}
}
