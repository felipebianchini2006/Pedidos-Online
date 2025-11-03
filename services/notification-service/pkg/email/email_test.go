package email

import (
	"pedidos-online/notification-service/internal/model"
	"strings"
	"testing"
)

// TestFormatCurrency testa a formatação de moeda
func TestFormatCurrency(t *testing.T) {
	tests := []struct {
		name     string
		value    float64
		expected string
	}{
		{
			name:     "Valor simples",
			value:    10.50,
			expected: "R$ 10,50",
		},
		{
			name:     "Valor com milhares",
			value:    1234.56,
			expected: "R$ 1.234,56",
		},
		{
			name:     "Valor grande",
			value:    1234567.89,
			expected: "R$ 1.234.567,88", // Arredondamento de ponto flutuante
		},
		{
			name:     "Valor zero",
			value:    0.00,
			expected: "R$ 0,00",
		},
		{
			name:     "Valor com um centavo",
			value:    99.01,
			expected: "R$ 99,01",
		},
		{
			name:     "Valor arredondado",
			value:    99.90,
			expected: "R$ 99,90",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatCurrency(tt.value)
			if result != tt.expected {
				t.Errorf("formatCurrency(%v) = %v, esperado %v", tt.value, result, tt.expected)
			}
		})
	}
}

// TestGetStatusMessage testa as mensagens de status
func TestGetStatusMessage(t *testing.T) {
	tests := []struct {
		name     string
		status   string
		expected string
	}{
		{
			name:     "Status pending",
			status:   "pending",
			expected: "Pedido Recebido",
		},
		{
			name:     "Status confirmed",
			status:   "confirmed",
			expected: "Pedido Confirmado",
		},
		{
			name:     "Status preparing",
			status:   "preparing",
			expected: "Em Preparação",
		},
		{
			name:     "Status shipped",
			status:   "shipped",
			expected: "Enviado para Entrega",
		},
		{
			name:     "Status delivered",
			status:   "delivered",
			expected: "Entregue",
		},
		{
			name:     "Status cancelled",
			status:   "cancelled",
			expected: "Cancelado",
		},
		{
			name:     "Status desconhecido",
			status:   "unknown",
			expected: "Status Atualizado",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getStatusMessage(tt.status)
			if result != tt.expected {
				t.Errorf("getStatusMessage(%v) = %v, esperado %v", tt.status, result, tt.expected)
			}
		})
	}
}

// TestGetStatusInfo testa as informações completas de status
func TestGetStatusInfo(t *testing.T) {
	tests := []struct {
		name          string
		status        string
		expectedEmoji string
		expectedTitle string
	}{
		{
			name:          "Status pending",
			status:        "pending",
			expectedEmoji: "⏳",
			expectedTitle: "Pedido Recebido",
		},
		{
			name:          "Status confirmed",
			status:        "confirmed",
			expectedEmoji: "✅",
			expectedTitle: "Pedido Confirmado",
		},
		{
			name:          "Status preparing",
			status:        "preparing",
			expectedEmoji: "👨‍🍳",
			expectedTitle: "Pedido em Preparação",
		},
		{
			name:          "Status shipped",
			status:        "shipped",
			expectedEmoji: "🚚",
			expectedTitle: "Pedido Enviado",
		},
		{
			name:          "Status delivered",
			status:        "delivered",
			expectedEmoji: "✅",
			expectedTitle: "Pedido Entregue",
		},
		{
			name:          "Status cancelled",
			status:        "cancelled",
			expectedEmoji: "❌",
			expectedTitle: "Pedido Cancelado",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := getStatusInfo(tt.status)
			if info.Emoji != tt.expectedEmoji {
				t.Errorf("getStatusInfo(%v).Emoji = %v, esperado %v", tt.status, info.Emoji, tt.expectedEmoji)
			}
			if info.Title != tt.expectedTitle {
				t.Errorf("getStatusInfo(%v).Title = %v, esperado %v", tt.status, info.Title, tt.expectedTitle)
			}
			if info.Message == "" {
				t.Errorf("getStatusInfo(%v).Message está vazio", tt.status)
			}
			if info.Color == "" {
				t.Errorf("getStatusInfo(%v).Color está vazio", tt.status)
			}
		})
	}
}

// TestIsPermanentError testa a detecção de erros permanentes
func TestIsPermanentError(t *testing.T) {
	tests := []struct {
		name       string
		errMessage string
		expected   bool
	}{
		{
			name:       "Erro de autenticação",
			errMessage: "535 Authentication failed",
			expected:   true,
		},
		{
			name:       "Credenciais inválidas",
			errMessage: "Invalid credentials",
			expected:   true,
		},
		{
			name:       "Usuário desconhecido",
			errMessage: "550 User unknown",
			expected:   true,
		},
		{
			name:       "Endereço rejeitado",
			errMessage: "553 Recipient address rejected",
			expected:   true,
		},
		{
			name:       "Erro temporário",
			errMessage: "421 Service not available, try again later",
			expected:   false,
		},
		{
			name:       "Erro de conexão",
			errMessage: "connection timeout",
			expected:   false,
		},
		{
			name:       "Erro de rede",
			errMessage: "dial tcp: connection refused",
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Criar erro fake
			err := &fakeError{message: tt.errMessage}
			result := isPermanentError(err)
			if result != tt.expected {
				t.Errorf("isPermanentError(%v) = %v, esperado %v", tt.errMessage, result, tt.expected)
			}
		})
	}
}

// TestGenerateOrderConfirmationPlainText testa geração de texto plano de confirmação
func TestGenerateOrderConfirmationPlainText(t *testing.T) {
	service := &SMTPEmailService{
		fromEmail: "test@example.com",
	}

	items := []model.OrderItem{
		{
			ProductID:   "prod-1",
			ProductName: "Pizza Margherita",
			Quantity:    2,
			Price:       35.90,
		},
		{
			ProductID:   "prod-2",
			ProductName: "Coca-Cola 2L",
			Quantity:    1,
			Price:       8.50,
		},
	}

	address := model.Address{
		Street:     "Rua das Flores",
		Number:     "123",
		City:       "São Paulo",
		State:      "SP",
		ZipCode:    "01234-567",
		Complement: "Apto 45",
	}

	result := service.generateOrderConfirmationPlainText("ABC123", items, 80.30, address)

	// Verificar conteúdo esperado
	expectedStrings := []string{
		"PEDIDO CONFIRMADO",
		"#ABC123",
		"Pizza Margherita",
		"Coca-Cola 2L",
		"R$ 80,30",
		"Rua das Flores, 123",
		"Apto 45",
		"São Paulo - SP",
		"CEP: 01234-567",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(result, expected) {
			t.Errorf("Texto plano não contém '%s'", expected)
		}
	}
}

// TestGenerateOrderStatusUpdatePlainText testa geração de texto plano de atualização
func TestGenerateOrderStatusUpdatePlainText(t *testing.T) {
	service := &SMTPEmailService{
		fromEmail: "test@example.com",
	}

	result := service.generateOrderStatusUpdatePlainText("ABC123", "pending", "confirmed")

	// Verificar conteúdo esperado
	expectedStrings := []string{
		"ATUALIZAÇÃO DO PEDIDO",
		"#ABC123",
		"Pedido Confirmado",
		"Status anterior: Pedido Recebido",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(result, expected) {
			t.Errorf("Texto plano não contém '%s'", expected)
		}
	}
}

// TestGenerateOrderConfirmationHTML testa geração de HTML de confirmação
func TestGenerateOrderConfirmationHTML(t *testing.T) {
	service := &SMTPEmailService{
		fromEmail: "test@example.com",
	}

	items := []model.OrderItem{
		{
			ProductID:   "prod-1",
			ProductName: "Pizza Margherita",
			Quantity:    2,
			Price:       35.90,
		},
	}

	address := model.Address{
		Street:  "Rua das Flores",
		Number:  "123",
		City:    "São Paulo",
		State:   "SP",
		ZipCode: "01234-567",
	}

	result, err := service.generateOrderConfirmationHTML("ABC123", items, 71.80, address)
	if err != nil {
		t.Fatalf("Erro ao gerar HTML: %v", err)
	}

	// Verificar se é HTML válido
	if !strings.Contains(result, "<!DOCTYPE html>") {
		t.Error("HTML não contém DOCTYPE")
	}

	// Verificar conteúdo esperado
	expectedStrings := []string{
		"Pedido Confirmado",
		"#ABC123",
		"Pizza Margherita",
		"R$ 71,80",
		"Rua das Flores, 123",
		"São Paulo - SP",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(result, expected) {
			t.Errorf("HTML não contém '%s'", expected)
		}
	}
}

// TestGenerateOrderStatusUpdateHTML testa geração de HTML de atualização
func TestGenerateOrderStatusUpdateHTML(t *testing.T) {
	service := &SMTPEmailService{
		fromEmail: "test@example.com",
	}

	result, err := service.generateOrderStatusUpdateHTML("ABC123", "pending", "confirmed")
	if err != nil {
		t.Fatalf("Erro ao gerar HTML: %v", err)
	}

	// Verificar se é HTML válido
	if !strings.Contains(result, "<!DOCTYPE html>") {
		t.Error("HTML não contém DOCTYPE")
	}

	// Verificar conteúdo esperado
	expectedStrings := []string{
		"Atualização do Pedido",
		"#ABC123",
		"Pedido Confirmado",
		"✅",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(result, expected) {
			t.Errorf("HTML não contém '%s'", expected)
		}
	}
}

// TestBuildMIMEMessage testa construção de mensagem MIME
func TestBuildMIMEMessage(t *testing.T) {
	service := &SMTPEmailService{
		fromEmail: "from@example.com",
	}

	plainBody := "Texto plano"
	htmlBody := "<html><body>HTML</body></html>"

	result := service.buildMIMEMessage("to@example.com", "Test Subject", plainBody, htmlBody)

	// Verificar headers
	expectedHeaders := []string{
		"From: from@example.com",
		"To: to@example.com",
		"Subject: Test Subject",
		"MIME-Version: 1.0",
		"Content-Type: multipart/alternative",
	}

	for _, header := range expectedHeaders {
		if !strings.Contains(result, header) {
			t.Errorf("Mensagem MIME não contém header '%s'", header)
		}
	}

	// Verificar partes
	if !strings.Contains(result, "Texto plano") {
		t.Error("Mensagem MIME não contém texto plano")
	}
	if !strings.Contains(result, "<html><body>HTML</body></html>") {
		t.Error("Mensagem MIME não contém HTML")
	}
}

// TestNewSMTPEmailService testa criação do serviço
func TestNewSMTPEmailService(t *testing.T) {
	config := &SMTPConfig{
		Host:     "smtp.example.com",
		Port:     587,
		User:     "user@example.com",
		Password: "password123",
		From:     "noreply@example.com",
	}

	service := NewSMTPEmailService(config)

	if service == nil {
		t.Fatal("NewSMTPEmailService retornou nil")
	}

	// Verificar se é do tipo correto
	smtpService, ok := service.(*SMTPEmailService)
	if !ok {
		t.Fatal("NewSMTPEmailService não retornou *SMTPEmailService")
	}

	// Verificar configurações
	if smtpService.smtpHost != config.Host {
		t.Errorf("smtpHost = %v, esperado %v", smtpService.smtpHost, config.Host)
	}
	if smtpService.smtpPort != config.Port {
		t.Errorf("smtpPort = %v, esperado %v", smtpService.smtpPort, config.Port)
	}
	if smtpService.smtpUser != config.User {
		t.Errorf("smtpUser = %v, esperado %v", smtpService.smtpUser, config.User)
	}
	if smtpService.fromEmail != config.From {
		t.Errorf("fromEmail = %v, esperado %v", smtpService.fromEmail, config.From)
	}
	if smtpService.maxRetries != 3 {
		t.Errorf("maxRetries = %v, esperado 3", smtpService.maxRetries)
	}
}

// TestSendOrderConfirmation_EmptyEmail testa validação de e-mail vazio
func TestSendOrderConfirmation_EmptyEmail(t *testing.T) {
	service := &SMTPEmailService{
		fromEmail: "test@example.com",
	}

	items := []model.OrderItem{
		{ProductName: "Teste", Quantity: 1, Price: 10.0},
	}

	address := model.Address{
		Street: "Rua Teste",
		Number: "123",
		City:   "São Paulo",
		State:  "SP",
	}

	err := service.SendOrderConfirmation("ABC123", "", items, 10.0, address)

	if err == nil {
		t.Error("Esperava erro para e-mail vazio, mas não recebeu")
	}

	if !strings.Contains(err.Error(), "e-mail do usuário não pode estar vazio") {
		t.Errorf("Erro inesperado: %v", err)
	}
}

// TestSendOrderStatusUpdate_EmptyEmail testa validação de e-mail vazio
func TestSendOrderStatusUpdate_EmptyEmail(t *testing.T) {
	service := &SMTPEmailService{
		fromEmail: "test@example.com",
	}

	err := service.SendOrderStatusUpdate("ABC123", "", "pending", "confirmed")

	if err == nil {
		t.Error("Esperava erro para e-mail vazio, mas não recebeu")
	}

	if !strings.Contains(err.Error(), "e-mail do usuário não pode estar vazio") {
		t.Errorf("Erro inesperado: %v", err)
	}
}

// fakeError implementa error para testes
type fakeError struct {
	message string
}

func (e *fakeError) Error() string {
	return e.message
}

// BenchmarkFormatCurrency benchmark para formatação de moeda
func BenchmarkFormatCurrency(b *testing.B) {
	for i := 0; i < b.N; i++ {
		formatCurrency(1234567.89)
	}
}

// BenchmarkGenerateOrderConfirmationHTML benchmark para geração de HTML
func BenchmarkGenerateOrderConfirmationHTML(b *testing.B) {
	service := &SMTPEmailService{
		fromEmail: "test@example.com",
	}

	items := []model.OrderItem{
		{ProductName: "Pizza", Quantity: 2, Price: 35.90},
		{ProductName: "Coca-Cola", Quantity: 1, Price: 8.50},
	}

	address := model.Address{
		Street:  "Rua Teste",
		Number:  "123",
		City:    "São Paulo",
		State:   "SP",
		ZipCode: "12345-678",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = service.generateOrderConfirmationHTML("ABC123", items, 80.30, address)
	}
}
