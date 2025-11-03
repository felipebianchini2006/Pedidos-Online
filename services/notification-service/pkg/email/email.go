package email

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"net/smtp"
	"pedidos-online/notification-service/internal/model"
	"strings"
	"time"
)

// EmailService define a interface para envio de e-mails
type EmailService interface {
	// SendOrderConfirmation envia e-mail de confirmação de pedido
	SendOrderConfirmation(orderID string, userEmail string, items []model.OrderItem, totalAmount float64, address model.Address) error

	// SendOrderStatusUpdate envia e-mail de atualização de status
	SendOrderStatusUpdate(orderID string, userEmail string, oldStatus, newStatus string) error
}

// SMTPEmailService implementa EmailService usando SMTP
type SMTPEmailService struct {
	smtpHost     string
	smtpPort     int
	smtpUser     string
	smtpPassword string
	fromEmail    string
	maxRetries   int
}

// SMTPConfig armazena configurações SMTP
type SMTPConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	From     string
}

// NewSMTPEmailService cria uma nova instância de SMTPEmailService
//
// Parâmetros:
//   - config: Configurações SMTP
//
// Retorno:
//   - EmailService: Interface do serviço de e-mail
//
// Exemplo:
//
//	config := &email.SMTPConfig{
//	    Host:     "smtp.gmail.com",
//	    Port:     587,
//	    User:     "user@gmail.com",
//	    Password: "app-password",
//	    From:     "noreply@pedidosonline.com",
//	}
//	emailService := email.NewSMTPEmailService(config)
func NewSMTPEmailService(config *SMTPConfig) EmailService {
	log.Printf("📧 Inicializando SMTP Email Service: %s:%d", config.Host, config.Port)
	return &SMTPEmailService{
		smtpHost:     config.Host,
		smtpPort:     config.Port,
		smtpUser:     config.User,
		smtpPassword: config.Password,
		fromEmail:    config.From,
		maxRetries:   3,
	}
}

// SendOrderConfirmation envia e-mail de confirmação de pedido
//
// Parâmetros:
//   - orderID: ID do pedido
//   - userEmail: E-mail do usuário
//   - items: Lista de itens do pedido
//   - totalAmount: Valor total do pedido
//   - address: Endereço de entrega
//
// Retorno:
//   - error: Erro ao enviar e-mail (se houver)
func (s *SMTPEmailService) SendOrderConfirmation(orderID string, userEmail string, items []model.OrderItem, totalAmount float64, address model.Address) error {
	log.Printf("📧 Enviando e-mail de confirmação de pedido: %s para %s", orderID, userEmail)

	// Validar e-mail
	if userEmail == "" {
		return fmt.Errorf("e-mail do usuário não pode estar vazio")
	}

	// Assunto do e-mail
	subject := fmt.Sprintf("Pedido #%s confirmado!", orderID)

	// Gerar HTML
	htmlBody, err := s.generateOrderConfirmationHTML(orderID, items, totalAmount, address)
	if err != nil {
		return fmt.Errorf("erro ao gerar HTML de confirmação: %w", err)
	}

	// Gerar texto plano (fallback)
	plainBody := s.generateOrderConfirmationPlainText(orderID, items, totalAmount, address)

	// Enviar e-mail com retry
	return s.sendEmailWithRetry(userEmail, subject, plainBody, htmlBody)
}

// SendOrderStatusUpdate envia e-mail de atualização de status
//
// Parâmetros:
//   - orderID: ID do pedido
//   - userEmail: E-mail do usuário
//   - oldStatus: Status anterior
//   - newStatus: Novo status
//
// Retorno:
//   - error: Erro ao enviar e-mail (se houver)
func (s *SMTPEmailService) SendOrderStatusUpdate(orderID string, userEmail string, oldStatus, newStatus string) error {
	log.Printf("📧 Enviando e-mail de atualização de status: %s (%s -> %s) para %s", orderID, oldStatus, newStatus, userEmail)

	// Validar e-mail
	if userEmail == "" {
		return fmt.Errorf("e-mail do usuário não pode estar vazio")
	}

	// Assunto do e-mail
	statusMessage := getStatusMessage(newStatus)
	subject := fmt.Sprintf("Pedido #%s: %s", orderID, statusMessage)

	// Gerar HTML
	htmlBody, err := s.generateOrderStatusUpdateHTML(orderID, oldStatus, newStatus)
	if err != nil {
		return fmt.Errorf("erro ao gerar HTML de atualização: %w", err)
	}

	// Gerar texto plano (fallback)
	plainBody := s.generateOrderStatusUpdatePlainText(orderID, oldStatus, newStatus)

	// Enviar e-mail com retry
	return s.sendEmailWithRetry(userEmail, subject, plainBody, htmlBody)
}

// sendEmailWithRetry envia e-mail com retry em caso de falha temporária
func (s *SMTPEmailService) sendEmailWithRetry(to, subject, plainBody, htmlBody string) error {
	var lastErr error

	for attempt := 1; attempt <= s.maxRetries; attempt++ {
		err := s.sendEmail(to, subject, plainBody, htmlBody)
		if err == nil {
			log.Printf("✅ E-mail enviado com sucesso para %s (tentativa %d/%d)", to, attempt, s.maxRetries)
			return nil
		}

		lastErr = err
		log.Printf("⚠️  Tentativa %d/%d falhou ao enviar e-mail para %s: %v", attempt, s.maxRetries, to, err)

		// Se for erro permanente (autenticação, e-mail inválido), não tentar novamente
		if isPermanentError(err) {
			log.Printf("❌ Erro permanente detectado, não tentando novamente: %v", err)
			return err
		}

		// Aguardar antes de tentar novamente (backoff exponencial)
		if attempt < s.maxRetries {
			waitTime := time.Duration(attempt*attempt) * time.Second
			log.Printf("⏳ Aguardando %v antes da próxima tentativa...", waitTime)
			time.Sleep(waitTime)
		}
	}

	return fmt.Errorf("falha ao enviar e-mail após %d tentativas: %w", s.maxRetries, lastErr)
}

// sendEmail envia um e-mail via SMTP
func (s *SMTPEmailService) sendEmail(to, subject, plainBody, htmlBody string) error {
	// Criar mensagem MIME multipart
	message := s.buildMIMEMessage(to, subject, plainBody, htmlBody)

	// Configurar autenticação SMTP
	auth := smtp.PlainAuth("", s.smtpUser, s.smtpPassword, s.smtpHost)

	// Endereço do servidor SMTP
	addr := fmt.Sprintf("%s:%d", s.smtpHost, s.smtpPort)

	// Enviar e-mail
	err := smtp.SendMail(addr, auth, s.fromEmail, []string{to}, []byte(message))
	if err != nil {
		return fmt.Errorf("erro ao enviar e-mail via SMTP: %w", err)
	}

	return nil
}

// buildMIMEMessage constrói uma mensagem MIME multipart (texto + HTML)
func (s *SMTPEmailService) buildMIMEMessage(to, subject, plainBody, htmlBody string) string {
	boundary := "boundary_pedidos_online"

	var msg strings.Builder

	// Headers
	msg.WriteString(fmt.Sprintf("From: %s\r\n", s.fromEmail))
	msg.WriteString(fmt.Sprintf("To: %s\r\n", to))
	msg.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	msg.WriteString("MIME-Version: 1.0\r\n")
	msg.WriteString(fmt.Sprintf("Content-Type: multipart/alternative; boundary=\"%s\"\r\n", boundary))
	msg.WriteString("\r\n")

	// Parte 1: Texto plano
	msg.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	msg.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
	msg.WriteString("Content-Transfer-Encoding: 8bit\r\n")
	msg.WriteString("\r\n")
	msg.WriteString(plainBody)
	msg.WriteString("\r\n\r\n")

	// Parte 2: HTML
	msg.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	msg.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
	msg.WriteString("Content-Transfer-Encoding: 8bit\r\n")
	msg.WriteString("\r\n")
	msg.WriteString(htmlBody)
	msg.WriteString("\r\n\r\n")

	// Finalizar
	msg.WriteString(fmt.Sprintf("--%s--\r\n", boundary))

	return msg.String()
}

// generateOrderConfirmationHTML gera HTML do e-mail de confirmação
func (s *SMTPEmailService) generateOrderConfirmationHTML(orderID string, items []model.OrderItem, totalAmount float64, address model.Address) (string, error) {
	tmpl := `
<!DOCTYPE html>
<html lang="pt-BR">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Pedido Confirmado</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
            line-height: 1.6;
            color: #333;
            background-color: #f4f4f4;
            margin: 0;
            padding: 0;
        }
        .container {
            max-width: 600px;
            margin: 20px auto;
            background-color: #ffffff;
            border-radius: 8px;
            overflow: hidden;
            box-shadow: 0 2px 10px rgba(0, 0, 0, 0.1);
        }
        .header {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            padding: 30px 20px;
            text-align: center;
        }
        .header h1 {
            margin: 0;
            font-size: 28px;
            font-weight: 600;
        }
        .header p {
            margin: 10px 0 0 0;
            font-size: 16px;
            opacity: 0.9;
        }
        .content {
            padding: 30px 20px;
        }
        .order-id {
            background-color: #f8f9fa;
            border-left: 4px solid #667eea;
            padding: 15px;
            margin-bottom: 25px;
            border-radius: 4px;
        }
        .order-id strong {
            color: #667eea;
            font-size: 18px;
        }
        .section-title {
            font-size: 18px;
            font-weight: 600;
            color: #333;
            margin-bottom: 15px;
            padding-bottom: 10px;
            border-bottom: 2px solid #e9ecef;
        }
        .items-table {
            width: 100%;
            border-collapse: collapse;
            margin-bottom: 20px;
        }
        .items-table th {
            background-color: #f8f9fa;
            padding: 12px;
            text-align: left;
            font-weight: 600;
            color: #495057;
            border-bottom: 2px solid #dee2e6;
        }
        .items-table td {
            padding: 12px;
            border-bottom: 1px solid #e9ecef;
        }
        .items-table tr:last-child td {
            border-bottom: none;
        }
        .item-name {
            font-weight: 500;
            color: #333;
        }
        .item-quantity {
            color: #6c757d;
            text-align: center;
        }
        .item-price, .item-subtotal {
            text-align: right;
            color: #333;
        }
        .total-section {
            background-color: #f8f9fa;
            padding: 20px;
            margin: 25px 0;
            border-radius: 4px;
            text-align: right;
        }
        .total-label {
            font-size: 18px;
            color: #495057;
            margin-bottom: 5px;
        }
        .total-amount {
            font-size: 32px;
            font-weight: 700;
            color: #28a745;
        }
        .address-section {
            background-color: #fff3cd;
            border: 1px solid #ffc107;
            border-radius: 4px;
            padding: 20px;
            margin-top: 25px;
        }
        .address-section .section-title {
            color: #856404;
            border-bottom-color: #ffc107;
        }
        .address-details {
            color: #856404;
            line-height: 1.8;
        }
        .footer {
            background-color: #f8f9fa;
            padding: 20px;
            text-align: center;
            color: #6c757d;
            font-size: 14px;
        }
        .footer p {
            margin: 5px 0;
        }
        .footer a {
            color: #667eea;
            text-decoration: none;
        }
        .emoji {
            font-size: 24px;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1><span class="emoji">✅</span> Pedido Confirmado!</h1>
            <p>Recebemos seu pedido e já estamos preparando tudo para você</p>
        </div>
        
        <div class="content">
            <div class="order-id">
                <strong>Pedido #{{.OrderID}}</strong>
            </div>

            <h2 class="section-title"><span class="emoji">📦</span> Itens do Pedido</h2>
            <table class="items-table">
                <thead>
                    <tr>
                        <th>Produto</th>
                        <th style="text-align: center;">Qtd.</th>
                        <th style="text-align: right;">Preço Unit.</th>
                        <th style="text-align: right;">Subtotal</th>
                    </tr>
                </thead>
                <tbody>
                    {{range .Items}}
                    <tr>
                        <td class="item-name">{{.ProductName}}</td>
                        <td class="item-quantity">{{.Quantity}}</td>
                        <td class="item-price">{{formatCurrency .Price}}</td>
                        <td class="item-subtotal">{{formatCurrency .Subtotal}}</td>
                    </tr>
                    {{end}}
                </tbody>
            </table>

            <div class="total-section">
                <div class="total-label">Valor Total</div>
                <div class="total-amount">{{formatCurrency .TotalAmount}}</div>
            </div>

            <div class="address-section">
                <h2 class="section-title"><span class="emoji">📍</span> Endereço de Entrega</h2>
                <div class="address-details">
                    <strong>{{.Address.Street}}, {{.Address.Number}}</strong><br>
                    {{if .Address.Complement}}{{.Address.Complement}}<br>{{end}}
                    {{.Address.City}} - {{.Address.State}}<br>
                    CEP: {{.Address.ZipCode}}
                </div>
            </div>
        </div>

        <div class="footer">
            <p><strong>Pedidos Online</strong></p>
            <p>Você receberá atualizações sobre o status do seu pedido por e-mail.</p>
            <p>Este é um e-mail automático, por favor não responda.</p>
        </div>
    </div>
</body>
</html>
`

	// Preparar dados para o template
	type TemplateItem struct {
		ProductName string
		Quantity    int
		Price       float64
		Subtotal    float64
	}

	var templateItems []TemplateItem
	for _, item := range items {
		templateItems = append(templateItems, TemplateItem{
			ProductName: item.ProductName,
			Quantity:    item.Quantity,
			Price:       item.Price,
			Subtotal:    item.Price * float64(item.Quantity),
		})
	}

	data := struct {
		OrderID     string
		Items       []TemplateItem
		TotalAmount float64
		Address     model.Address
	}{
		OrderID:     orderID,
		Items:       templateItems,
		TotalAmount: totalAmount,
		Address:     address,
	}

	// Criar template com função personalizada
	funcMap := template.FuncMap{
		"formatCurrency": formatCurrency,
	}

	t, err := template.New("orderConfirmation").Funcs(funcMap).Parse(tmpl)
	if err != nil {
		return "", fmt.Errorf("erro ao fazer parse do template: %w", err)
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("erro ao executar template: %w", err)
	}

	return buf.String(), nil
}

// generateOrderConfirmationPlainText gera versão texto plano do e-mail de confirmação
func (s *SMTPEmailService) generateOrderConfirmationPlainText(orderID string, items []model.OrderItem, totalAmount float64, address model.Address) string {
	var buf strings.Builder

	buf.WriteString("✅ PEDIDO CONFIRMADO!\n\n")
	buf.WriteString(fmt.Sprintf("Pedido: #%s\n\n", orderID))
	buf.WriteString("📦 ITENS DO PEDIDO:\n")
	buf.WriteString("----------------------------------------\n")

	for _, item := range items {
		subtotal := item.Price * float64(item.Quantity)
		buf.WriteString(fmt.Sprintf("%s\n", item.ProductName))
		buf.WriteString(fmt.Sprintf("  Quantidade: %d x %s = %s\n\n", item.Quantity, formatCurrency(item.Price), formatCurrency(subtotal)))
	}

	buf.WriteString("----------------------------------------\n")
	buf.WriteString(fmt.Sprintf("TOTAL: %s\n\n", formatCurrency(totalAmount)))

	buf.WriteString("📍 ENDEREÇO DE ENTREGA:\n")
	buf.WriteString(fmt.Sprintf("%s, %s\n", address.Street, address.Number))
	if address.Complement != "" {
		buf.WriteString(fmt.Sprintf("%s\n", address.Complement))
	}
	buf.WriteString(fmt.Sprintf("%s - %s\n", address.City, address.State))
	buf.WriteString(fmt.Sprintf("CEP: %s\n\n", address.ZipCode))

	buf.WriteString("Você receberá atualizações sobre o status do seu pedido por e-mail.\n\n")
	buf.WriteString("Pedidos Online\n")
	buf.WriteString("Este é um e-mail automático, por favor não responda.\n")

	return buf.String()
}

// generateOrderStatusUpdateHTML gera HTML do e-mail de atualização de status
func (s *SMTPEmailService) generateOrderStatusUpdateHTML(orderID, oldStatus, newStatus string) (string, error) {
	tmpl := `
<!DOCTYPE html>
<html lang="pt-BR">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Atualização de Pedido</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
            line-height: 1.6;
            color: #333;
            background-color: #f4f4f4;
            margin: 0;
            padding: 0;
        }
        .container {
            max-width: 600px;
            margin: 20px auto;
            background-color: #ffffff;
            border-radius: 8px;
            overflow: hidden;
            box-shadow: 0 2px 10px rgba(0, 0, 0, 0.1);
        }
        .header {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            padding: 30px 20px;
            text-align: center;
        }
        .header h1 {
            margin: 0;
            font-size: 28px;
            font-weight: 600;
        }
        .header p {
            margin: 10px 0 0 0;
            font-size: 16px;
            opacity: 0.9;
        }
        .content {
            padding: 30px 20px;
        }
        .order-id {
            background-color: #f8f9fa;
            border-left: 4px solid #667eea;
            padding: 15px;
            margin-bottom: 25px;
            border-radius: 4px;
        }
        .order-id strong {
            color: #667eea;
            font-size: 18px;
        }
        .status-box {
            background-color: {{.StatusColor}};
            border-radius: 8px;
            padding: 25px;
            text-align: center;
            margin: 25px 0;
        }
        .status-icon {
            font-size: 48px;
            margin-bottom: 10px;
        }
        .status-title {
            font-size: 24px;
            font-weight: 600;
            color: #333;
            margin-bottom: 10px;
        }
        .status-description {
            font-size: 16px;
            color: #495057;
        }
        .old-status {
            color: #6c757d;
            font-size: 14px;
            margin-top: 10px;
        }
        .timeline {
            margin: 30px 0;
        }
        .timeline-item {
            display: flex;
            align-items: center;
            padding: 15px;
            border-left: 3px solid #e9ecef;
            margin-left: 20px;
        }
        .timeline-item.active {
            border-left-color: #667eea;
            background-color: #f8f9fa;
            border-radius: 4px;
        }
        .timeline-icon {
            font-size: 24px;
            margin-right: 15px;
        }
        .timeline-text {
            font-size: 16px;
        }
        .timeline-item.active .timeline-text {
            font-weight: 600;
            color: #667eea;
        }
        .info-box {
            background-color: #e7f3ff;
            border-left: 4px solid #2196F3;
            padding: 15px;
            margin: 25px 0;
            border-radius: 4px;
        }
        .footer {
            background-color: #f8f9fa;
            padding: 20px;
            text-align: center;
            color: #6c757d;
            font-size: 14px;
        }
        .footer p {
            margin: 5px 0;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>📦 Atualização do Pedido</h1>
            <p>Há novidades sobre o seu pedido!</p>
        </div>
        
        <div class="content">
            <div class="order-id">
                <strong>Pedido #{{.OrderID}}</strong>
            </div>

            <div class="status-box">
                <div class="status-icon">{{.StatusEmoji}}</div>
                <div class="status-title">{{.StatusTitle}}</div>
                <div class="status-description">{{.StatusMessage}}</div>
                <div class="old-status">Status anterior: {{.OldStatusName}}</div>
            </div>

            {{if .ShowTimeline}}
            <div class="timeline">
                <h3 style="margin-bottom: 20px; color: #333;">📋 Acompanhe seu pedido:</h3>
                {{range .Timeline}}
                <div class="timeline-item {{.Class}}">
                    <div class="timeline-icon">{{.Icon}}</div>
                    <div class="timeline-text">{{.Text}}</div>
                </div>
                {{end}}
            </div>
            {{end}}

            {{if .SpecialMessage}}
            <div class="info-box">
                {{.SpecialMessage}}
            </div>
            {{end}}
        </div>

        <div class="footer">
            <p><strong>Pedidos Online</strong></p>
            <p>Você receberá novas atualizações por e-mail.</p>
            <p>Este é um e-mail automático, por favor não responda.</p>
        </div>
    </div>
</body>
</html>
`

	// Preparar dados do status
	statusInfo := getStatusInfo(newStatus)
	oldStatusInfo := getStatusInfo(oldStatus)

	// Criar timeline
	type TimelineItem struct {
		Icon  string
		Text  string
		Class string
	}

	timeline := []TimelineItem{
		{Icon: "⏳", Text: "Pedido recebido", Class: ""},
		{Icon: "✅", Text: "Pedido confirmado", Class: ""},
		{Icon: "👨‍🍳", Text: "Pedido em preparação", Class: ""},
		{Icon: "🚚", Text: "Pedido enviado", Class: ""},
		{Icon: "✅", Text: "Pedido entregue", Class: ""},
	}

	// Marcar status atual como ativo
	statusOrder := map[string]int{
		"pending":   0,
		"confirmed": 1,
		"preparing": 2,
		"shipped":   3,
		"delivered": 4,
	}

	currentIndex := statusOrder[newStatus]
	for i := 0; i <= currentIndex && i < len(timeline); i++ {
		if i == currentIndex {
			timeline[i].Class = "active"
		}
	}

	// Mensagem especial
	specialMessage := ""
	if newStatus == "delivered" {
		specialMessage = "🎉 <strong>Pedido entregue com sucesso!</strong><br>Esperamos que você tenha gostado. Obrigado pela preferência!"
	} else if newStatus == "cancelled" {
		specialMessage = "❌ <strong>Pedido cancelado.</strong><br>Se você tiver dúvidas, entre em contato conosco."
	}

	showTimeline := newStatus != "cancelled"

	data := struct {
		OrderID        string
		StatusEmoji    string
		StatusTitle    string
		StatusMessage  string
		StatusColor    string
		OldStatusName  string
		Timeline       []TimelineItem
		ShowTimeline   bool
		SpecialMessage string
	}{
		OrderID:        orderID,
		StatusEmoji:    statusInfo.Emoji,
		StatusTitle:    statusInfo.Title,
		StatusMessage:  statusInfo.Message,
		StatusColor:    statusInfo.Color,
		OldStatusName:  oldStatusInfo.Title,
		Timeline:       timeline,
		ShowTimeline:   showTimeline,
		SpecialMessage: specialMessage,
	}

	t, err := template.New("statusUpdate").Parse(tmpl)
	if err != nil {
		return "", fmt.Errorf("erro ao fazer parse do template: %w", err)
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("erro ao executar template: %w", err)
	}

	return buf.String(), nil
}

// generateOrderStatusUpdatePlainText gera versão texto plano do e-mail de atualização
func (s *SMTPEmailService) generateOrderStatusUpdatePlainText(orderID, oldStatus, newStatus string) string {
	var buf strings.Builder

	statusInfo := getStatusInfo(newStatus)
	oldStatusInfo := getStatusInfo(oldStatus)

	buf.WriteString("📦 ATUALIZAÇÃO DO PEDIDO\n\n")
	buf.WriteString(fmt.Sprintf("Pedido: #%s\n\n", orderID))
	buf.WriteString(fmt.Sprintf("%s %s\n", statusInfo.Emoji, statusInfo.Title))
	buf.WriteString(fmt.Sprintf("%s\n\n", statusInfo.Message))
	buf.WriteString(fmt.Sprintf("Status anterior: %s\n\n", oldStatusInfo.Title))

	if newStatus == "delivered" {
		buf.WriteString("🎉 Pedido entregue com sucesso!\n")
		buf.WriteString("Esperamos que você tenha gostado. Obrigado pela preferência!\n\n")
	} else if newStatus == "cancelled" {
		buf.WriteString("❌ Pedido cancelado.\n")
		buf.WriteString("Se você tiver dúvidas, entre em contato conosco.\n\n")
	}

	buf.WriteString("Você receberá novas atualizações por e-mail.\n\n")
	buf.WriteString("Pedidos Online\n")
	buf.WriteString("Este é um e-mail automático, por favor não responda.\n")

	return buf.String()
}

// StatusInfo armazena informações de apresentação de um status
type StatusInfo struct {
	Emoji   string
	Title   string
	Message string
	Color   string
}

// getStatusInfo retorna informações de apresentação para um status
func getStatusInfo(status string) StatusInfo {
	statusMap := map[string]StatusInfo{
		"pending": {
			Emoji:   "⏳",
			Title:   "Pedido Recebido",
			Message: "Seu pedido foi recebido e está aguardando confirmação.",
			Color:   "#fff3cd",
		},
		"confirmed": {
			Emoji:   "✅",
			Title:   "Pedido Confirmado",
			Message: "Seu pedido foi confirmado e em breve será preparado.",
			Color:   "#d4edda",
		},
		"preparing": {
			Emoji:   "👨‍🍳",
			Title:   "Pedido em Preparação",
			Message: "Seu pedido está sendo preparado com todo cuidado.",
			Color:   "#d1ecf1",
		},
		"shipped": {
			Emoji:   "🚚",
			Title:   "Pedido Enviado",
			Message: "Seu pedido foi enviado e está a caminho!",
			Color:   "#cfe2ff",
		},
		"delivered": {
			Emoji:   "✅",
			Title:   "Pedido Entregue",
			Message: "Seu pedido foi entregue com sucesso!",
			Color:   "#d4edda",
		},
		"cancelled": {
			Emoji:   "❌",
			Title:   "Pedido Cancelado",
			Message: "Seu pedido foi cancelado.",
			Color:   "#f8d7da",
		},
	}

	if info, ok := statusMap[status]; ok {
		return info
	}

	return StatusInfo{
		Emoji:   "📦",
		Title:   "Status Atualizado",
		Message: "Seu pedido teve uma atualização.",
		Color:   "#e2e3e5",
	}
}

// getStatusMessage retorna mensagem curta para um status
func getStatusMessage(status string) string {
	messages := map[string]string{
		"pending":   "Pedido Recebido",
		"confirmed": "Pedido Confirmado",
		"preparing": "Em Preparação",
		"shipped":   "Enviado para Entrega",
		"delivered": "Entregue",
		"cancelled": "Cancelado",
	}

	if msg, ok := messages[status]; ok {
		return msg
	}

	return "Status Atualizado"
}

// formatCurrency formata um valor float64 para moeda brasileira (R$ X.XXX,XX)
//
// Parâmetros:
//   - value: Valor a ser formatado
//
// Retorno:
//   - string: Valor formatado (ex: "R$ 1.234,56")
//
// Exemplo:
//
//	formatCurrency(1234.56) // Retorna "R$ 1.234,56"
//	formatCurrency(99.9)    // Retorna "R$ 99,90"
func formatCurrency(value float64) string {
	// Converter para centavos
	cents := int64(value * 100)

	// Separar reais e centavos
	reais := cents / 100
	centavos := cents % 100

	// Formatar reais com separador de milhares
	reaisStr := formatThousands(reais)

	// Retornar formatado
	return fmt.Sprintf("R$ %s,%02d", reaisStr, centavos)
}

// formatThousands formata um número com separador de milhares
func formatThousands(n int64) string {
	if n < 1000 {
		return fmt.Sprintf("%d", n)
	}

	// Converter para string
	str := fmt.Sprintf("%d", n)

	// Adicionar pontos a cada 3 dígitos (da direita para esquerda)
	var result strings.Builder
	for i, char := range str {
		if i > 0 && (len(str)-i)%3 == 0 {
			result.WriteRune('.')
		}
		result.WriteRune(char)
	}

	return result.String()
}

// isPermanentError verifica se um erro é permanente (não deve tentar novamente)
func isPermanentError(err error) bool {
	errStr := strings.ToLower(err.Error())

	// Erros permanentes comuns
	permanentErrors := []string{
		"authentication failed",
		"invalid credentials",
		"bad username or password",
		"mailbox unavailable",
		"user unknown",
		"invalid recipient",
		"recipient address rejected",
		"550", // Erro SMTP de caixa de correio não encontrada
		"553", // Erro SMTP de endereço não permitido
	}

	for _, permErr := range permanentErrors {
		if strings.Contains(errStr, permErr) {
			return true
		}
	}

	return false
}
