package email

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"pedidos-online/notification-service/internal/config"
	"pedidos-online/notification-service/internal/model"

	"gopkg.in/gomail.v2"
)

// EmailService gerencia o envio de e-mails via SMTP
type EmailService struct {
	smtpConfig config.SMTPConfig
	dialer     *gomail.Dialer
}

// NewEmailService cria uma nova instância do EmailService
//
// Parâmetros:
//   - smtpConfig: Configurações SMTP (host, port, user, password, from)
//
// Retorno:
//   - *EmailService: Instância do serviço de e-mail
//
// Exemplo:
//
//	emailService := email.NewEmailService(cfg.SMTP)
func NewEmailService(smtpConfig config.SMTPConfig) *EmailService {
	// Configurar dialer do gomail
	dialer := gomail.NewDialer(
		smtpConfig.Host,
		smtpConfig.Port,
		smtpConfig.User,
		smtpConfig.Password,
	)

	log.Printf("📧 Email Service inicializado: %s:%d", smtpConfig.Host, smtpConfig.Port)

	return &EmailService{
		smtpConfig: smtpConfig,
		dialer:     dialer,
	}
}

// SendOrderConfirmation envia e-mail de confirmação de pedido
//
// Parâmetros:
//   - event: Evento de pedido criado
//
// Retorno:
//   - error: Erro ao enviar e-mail (se houver)
//
// Template: HTML com detalhes do pedido (itens, total, endereço)
func (s *EmailService) SendOrderConfirmation(event *model.OrderCreatedEvent) error {
	log.Printf("📧 Enviando e-mail de confirmação de pedido: %s", event.OrderID)

	// Determinar destinatário
	to := event.UserEmail
	if to == "" {
		log.Printf("⚠️  UserEmail não disponível para pedido %s", event.OrderID)
		return fmt.Errorf("user_email não disponível no evento")
	}

	// Subject do e-mail
	subject := fmt.Sprintf("✅ Pedido #%s Confirmado - Pedidos Online", event.OrderID[:8])

	// Gerar corpo do e-mail em HTML
	body, err := s.generateOrderConfirmationHTML(event)
	if err != nil {
		return fmt.Errorf("erro ao gerar HTML de confirmação: %w", err)
	}

	// Enviar e-mail
	if err := s.sendEmail(to, subject, body); err != nil {
		return fmt.Errorf("erro ao enviar e-mail de confirmação: %w", err)
	}

	log.Printf("✅ E-mail de confirmação enviado para %s (pedido %s)", to, event.OrderID)
	return nil
}

// SendOrderStatusUpdate envia e-mail de atualização de status do pedido
//
// Parâmetros:
//   - event: Evento de atualização de status
//
// Retorno:
//   - error: Erro ao enviar e-mail (se houver)
//
// Template: HTML com informações da mudança de status
func (s *EmailService) SendOrderStatusUpdate(event *model.OrderUpdatedEvent) error {
	log.Printf("📧 Enviando e-mail de atualização de status: %s (%s -> %s)",
		event.OrderID, event.OldStatus, event.NewStatus)

	// Determinar destinatário
	to := event.UserEmail
	if to == "" {
		log.Printf("⚠️  UserEmail não disponível para pedido %s", event.OrderID)
		return fmt.Errorf("user_email não disponível no evento")
	}

	// Subject do e-mail
	statusDesc := model.GetStatusDescription(event.NewStatus)
	subject := fmt.Sprintf("📦 Pedido #%s - Status: %s", event.OrderID[:8], statusDesc)

	// Gerar corpo do e-mail em HTML
	body, err := s.generateOrderStatusUpdateHTML(event)
	if err != nil {
		return fmt.Errorf("erro ao gerar HTML de atualização: %w", err)
	}

	// Enviar e-mail
	if err := s.sendEmail(to, subject, body); err != nil {
		return fmt.Errorf("erro ao enviar e-mail de atualização: %w", err)
	}

	log.Printf("✅ E-mail de atualização enviado para %s (pedido %s)", to, event.OrderID)
	return nil
}

// sendEmail envia um e-mail via SMTP usando gomail
func (s *EmailService) sendEmail(to, subject, htmlBody string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", s.smtpConfig.From)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", htmlBody)

	// Enviar usando o dialer configurado
	if err := s.dialer.DialAndSend(m); err != nil {
		return fmt.Errorf("erro ao enviar e-mail via SMTP: %w", err)
	}

	return nil
}

// generateOrderConfirmationHTML gera HTML de confirmação de pedido
func (s *EmailService) generateOrderConfirmationHTML(event *model.OrderCreatedEvent) (string, error) {
	tmpl := `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background-color: #4CAF50; color: white; padding: 20px; text-align: center; border-radius: 5px 5px 0 0; }
        .content { background-color: #f9f9f9; padding: 20px; border: 1px solid #ddd; }
        .order-info { background-color: white; padding: 15px; margin: 15px 0; border-radius: 5px; }
        .item { padding: 10px; border-bottom: 1px solid #eee; }
        .item:last-child { border-bottom: none; }
        .total { font-size: 1.2em; font-weight: bold; color: #4CAF50; text-align: right; padding: 15px 0; }
        .address { background-color: #fffbea; padding: 15px; margin: 15px 0; border-left: 4px solid #ffc107; }
        .footer { text-align: center; padding: 20px; color: #777; font-size: 0.9em; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>✅ Pedido Confirmado!</h1>
        </div>
        <div class="content">
            <p>Olá! Seu pedido foi confirmado com sucesso.</p>
            
            <div class="order-info">
                <p><strong>Número do Pedido:</strong> #{{.OrderID}}</p>
                <p><strong>Data:</strong> {{.CreatedAt}}</p>
            </div>

            <h3>📦 Itens do Pedido:</h3>
            {{range .Items}}
            <div class="item">
                <strong>{{.ProductName}}</strong><br>
                Quantidade: {{.Quantity}} x R$ {{printf "%.2f" .Price}} = R$ {{printf "%.2f" .Total}}
            </div>
            {{end}}

            <div class="total">
                Total: R$ {{printf "%.2f" .TotalAmount}}
            </div>

            <div class="address">
                <h4>📍 Endereço de Entrega:</h4>
                <p>{{.FormattedAddress}}</p>
            </div>

            <p>Você receberá atualizações sobre o status do seu pedido por e-mail.</p>
        </div>
        <div class="footer">
            <p>Pedidos Online - Sistema de Pedidos</p>
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
		Total       float64
	}

	var items []TemplateItem
	for _, item := range event.Items {
		items = append(items, TemplateItem{
			ProductName: item.ProductName,
			Quantity:    item.Quantity,
			Price:       item.Price,
			Total:       item.CalculateItemTotal(),
		})
	}

	data := struct {
		OrderID          string
		CreatedAt        string
		Items            []TemplateItem
		TotalAmount      float64
		FormattedAddress string
	}{
		OrderID:          event.OrderID,
		CreatedAt:        event.CreatedAt.Format("02/01/2006 15:04"),
		Items:            items,
		TotalAmount:      event.TotalAmount,
		FormattedAddress: event.Address.FormatAddress(),
	}

	// Executar template
	t, err := template.New("orderConfirmation").Parse(tmpl)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// generateOrderStatusUpdateHTML gera HTML de atualização de status
func (s *EmailService) generateOrderStatusUpdateHTML(event *model.OrderUpdatedEvent) (string, error) {
	tmpl := `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background-color: #2196F3; color: white; padding: 20px; text-align: center; border-radius: 5px 5px 0 0; }
        .content { background-color: #f9f9f9; padding: 20px; border: 1px solid #ddd; }
        .status-box { background-color: white; padding: 20px; margin: 20px 0; text-align: center; border-radius: 5px; }
        .status { font-size: 1.5em; color: #4CAF50; font-weight: bold; }
        .timeline { margin: 20px 0; }
        .timeline-item { padding: 10px 0; border-left: 3px solid #ddd; padding-left: 20px; }
        .timeline-item.active { border-left-color: #4CAF50; color: #4CAF50; font-weight: bold; }
        .footer { text-align: center; padding: 20px; color: #777; font-size: 0.9em; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>📦 Atualização do Pedido</h1>
        </div>
        <div class="content">
            <p>Olá! Há uma atualização sobre o seu pedido.</p>
            
            <div class="status-box">
                <p><strong>Pedido:</strong> #{{.OrderID}}</p>
                <p class="status">{{.StatusIcon}} {{.NewStatusDescription}}</p>
                <p style="color: #777; font-size: 0.9em;">
                    Status anterior: {{.OldStatusDescription}}
                </p>
            </div>

            <div class="timeline">
                <h4>📋 Timeline do Pedido:</h4>
                {{range .Timeline}}
                <div class="timeline-item {{.Class}}">
                    {{.Icon}} {{.Description}}
                </div>
                {{end}}
            </div>

            <p><strong>Atualizado em:</strong> {{.UpdatedAt}}</p>

            {{if eq .NewStatus "delivered"}}
            <div style="background-color: #e8f5e9; padding: 15px; margin: 15px 0; border-radius: 5px;">
                <p><strong>🎉 Pedido Entregue!</strong></p>
                <p>Esperamos que você tenha gostado. Obrigado pela preferência!</p>
            </div>
            {{end}}

            {{if eq .NewStatus "cancelled"}}
            <div style="background-color: #ffebee; padding: 15px; margin: 15px 0; border-radius: 5px;">
                <p><strong>❌ Pedido Cancelado</strong></p>
                <p>Se você tiver dúvidas, entre em contato conosco.</p>
            </div>
            {{end}}
        </div>
        <div class="footer">
            <p>Pedidos Online - Sistema de Pedidos</p>
            <p>Este é um e-mail automático, por favor não responda.</p>
        </div>
    </div>
</body>
</html>
`

	// Determinar ícone do status
	statusIcons := map[string]string{
		"pending":   "⏳",
		"confirmed": "✅",
		"preparing": "👨‍🍳",
		"shipped":   "🚚",
		"delivered": "✅",
		"cancelled": "❌",
	}

	statusIcon := statusIcons[event.NewStatus]
	if statusIcon == "" {
		statusIcon = "📦"
	}

	// Timeline completo de status
	type TimelineItem struct {
		Icon        string
		Description string
		Class       string
	}

	allStatuses := []struct {
		status string
		icon   string
		desc   string
	}{
		{"pending", "⏳", "Pendente"},
		{"confirmed", "✅", "Confirmado"},
		{"preparing", "👨‍🍳", "Em Preparação"},
		{"shipped", "🚚", "Enviado"},
		{"delivered", "✅", "Entregue"},
	}

	var timeline []TimelineItem
	reached := false
	for _, s := range allStatuses {
		class := ""
		if s.status == event.NewStatus {
			class = "active"
			reached = true
		} else if !reached {
			class = "active"
		}

		timeline = append(timeline, TimelineItem{
			Icon:        s.icon,
			Description: s.desc,
			Class:       class,
		})
	}

	// Se foi cancelado, mostrar separadamente
	if event.NewStatus == "cancelled" {
		timeline = []TimelineItem{
			{Icon: "❌", Description: "Cancelado", Class: "active"},
		}
	}

	data := struct {
		OrderID              string
		OldStatus            string
		NewStatus            string
		OldStatusDescription string
		NewStatusDescription string
		StatusIcon           string
		UpdatedAt            string
		Timeline             []TimelineItem
	}{
		OrderID:              event.OrderID,
		OldStatus:            event.OldStatus,
		NewStatus:            event.NewStatus,
		OldStatusDescription: model.GetStatusDescription(event.OldStatus),
		NewStatusDescription: model.GetStatusDescription(event.NewStatus),
		StatusIcon:           statusIcon,
		UpdatedAt:            event.UpdatedAt.Format("02/01/2006 15:04"),
		Timeline:             timeline,
	}

	// Executar template
	t, err := template.New("orderStatusUpdate").Parse(tmpl)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// TestConnection testa a conexão SMTP
func (s *EmailService) TestConnection() error {
	log.Println("🧪 Testando conexão SMTP...")

	// Tentar conectar ao servidor SMTP
	conn, err := s.dialer.Dial()
	if err != nil {
		return fmt.Errorf("erro ao conectar ao servidor SMTP: %w", err)
	}
	defer conn.Close()

	log.Println("✅ Conexão SMTP testada com sucesso")
	return nil
}
