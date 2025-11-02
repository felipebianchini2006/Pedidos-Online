package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
)

// Config armazena todas as configurações do Notification Service
type Config struct {
	Port           string
	RabbitMQURL    string
	SMTP           SMTPConfig
	ServiceName    string
	ServiceVersion string
}

// SMTPConfig armazena as configurações do servidor SMTP
type SMTPConfig struct {
	Host     string // Servidor SMTP (ex: smtp.gmail.com)
	Port     int    // Porta SMTP (ex: 587 para TLS)
	User     string // Usuário para autenticação
	Password string // Senha ou App Password
	From     string // E-mail remetente (ex: noreply@pedidosonline.com)
}

// LoadConfig carrega configurações das variáveis de ambiente
//
// Variáveis obrigatórias:
//   - PORT: Porta do servidor HTTP (padrão: 8003)
//   - RABBITMQ_URL: URL de conexão do RabbitMQ
//   - SMTP_HOST: Servidor SMTP
//   - SMTP_PORT: Porta SMTP
//   - SMTP_USER: Usuário SMTP
//   - SMTP_PASSWORD: Senha SMTP
//   - EMAIL_FROM: E-mail remetente
//
// Exemplo de uso:
//
//	cfg := config.LoadConfig()
func LoadConfig() *Config {
	log.Println("📋 Carregando configurações do Notification Service...")

	// Configurações obrigatórias
	rabbitMQURL := getEnv("RABBITMQ_URL", "")
	if rabbitMQURL == "" {
		log.Fatal("❌ RABBITMQ_URL não configurado")
	}

	smtpHost := getEnv("SMTP_HOST", "")
	if smtpHost == "" {
		log.Fatal("❌ SMTP_HOST não configurado")
	}

	smtpUser := getEnv("SMTP_USER", "")
	if smtpUser == "" {
		log.Fatal("❌ SMTP_USER não configurado")
	}

	smtpPassword := getEnv("SMTP_PASSWORD", "")
	if smtpPassword == "" {
		log.Fatal("❌ SMTP_PASSWORD não configurado")
	}

	emailFrom := getEnv("EMAIL_FROM", "")
	if emailFrom == "" {
		log.Fatal("❌ EMAIL_FROM não configurado")
	}

	// Porta SMTP (padrão: 587 para TLS)
	smtpPortStr := getEnv("SMTP_PORT", "587")
	smtpPort, err := strconv.Atoi(smtpPortStr)
	if err != nil {
		log.Printf("⚠️  SMTP_PORT inválido (%s), usando padrão 587", smtpPortStr)
		smtpPort = 587
	}

	cfg := &Config{
		Port:        getEnv("PORT", "8003"),
		RabbitMQURL: rabbitMQURL,
		SMTP: SMTPConfig{
			Host:     smtpHost,
			Port:     smtpPort,
			User:     smtpUser,
			Password: smtpPassword,
			From:     emailFrom,
		},
		ServiceName:    "notification-service",
		ServiceVersion: "1.0.0",
	}

	log.Println("✅ Configurações carregadas com sucesso")
	log.Printf("   - Porta: %s", cfg.Port)
	log.Printf("   - RabbitMQ: %s", maskPassword(cfg.RabbitMQURL))
	log.Printf("   - SMTP Host: %s:%d", cfg.SMTP.Host, cfg.SMTP.Port)
	log.Printf("   - SMTP User: %s", cfg.SMTP.User)
	log.Printf("   - Email From: %s", cfg.SMTP.From)

	return cfg
}

// getEnv retorna o valor de uma variável de ambiente ou um valor padrão
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// maskPassword mascara a senha em URLs (ex: amqp://user:***@host:5672/)
func maskPassword(url string) string {
	// Simplificação: apenas para logging seguro
	// Implementação completa usaria regex para mascarar senha na URL
	if len(url) > 30 {
		return url[:15] + "***" + url[len(url)-10:]
	}
	return "***"
}

// Validate valida se todas as configurações obrigatórias estão presentes
func (c *Config) Validate() error {
	if c.RabbitMQURL == "" {
		return fmt.Errorf("RABBITMQ_URL é obrigatório")
	}

	if c.SMTP.Host == "" {
		return fmt.Errorf("SMTP_HOST é obrigatório")
	}

	if c.SMTP.Port == 0 {
		return fmt.Errorf("SMTP_PORT é obrigatório")
	}

	if c.SMTP.User == "" {
		return fmt.Errorf("SMTP_USER é obrigatório")
	}

	if c.SMTP.Password == "" {
		return fmt.Errorf("SMTP_PASSWORD é obrigatório")
	}

	if c.SMTP.From == "" {
		return fmt.Errorf("EMAIL_FROM é obrigatório")
	}

	return nil
}
