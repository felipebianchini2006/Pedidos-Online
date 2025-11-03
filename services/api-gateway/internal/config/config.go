package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

// Config armazena todas as configurações do API Gateway
type Config struct {
	Port              string   // Porta do servidor (padrão: 8000)
	UserServiceURL    string   // URL do User Service
	OrderServiceURL   string   // URL do Order Service
	AllowedOrigins    []string // Origens permitidas para CORS
	RateLimitPerMin   int      // Limite de requisições por minuto por IP
	RequestTimeout    int      // Timeout para requests aos serviços (segundos)
	MaxRetries        int      // Número máximo de tentativas
	ServiceName       string   // Nome do serviço
	ServiceVersion    string   // Versão do serviço
	EnableRateLimit   bool     // Habilitar rate limiting
	EnableCORS        bool     // Habilitar CORS
	EnableDetailedLog bool     // Habilitar logging detalhado
}

// LoadConfig carrega configurações das variáveis de ambiente
//
// Variáveis de ambiente:
//   - PORT: Porta do servidor (padrão: 8000)
//   - USER_SERVICE_URL: URL do User Service (obrigatório)
//   - ORDER_SERVICE_URL: URL do Order Service (obrigatório)
//   - ALLOWED_ORIGINS: Origens permitidas separadas por vírgula (padrão: *)
//   - RATE_LIMIT_PER_MIN: Limite de requisições por minuto (padrão: 100)
//   - REQUEST_TIMEOUT: Timeout em segundos (padrão: 30)
//   - MAX_RETRIES: Número máximo de tentativas (padrão: 3)
//   - ENABLE_RATE_LIMIT: Habilitar rate limiting (padrão: true)
//   - ENABLE_CORS: Habilitar CORS (padrão: true)
//   - ENABLE_DETAILED_LOG: Habilitar logging detalhado (padrão: true)
func LoadConfig() *Config {
	log.Println("📋 Carregando configurações do API Gateway...")

	// Configurações obrigatórias
	userServiceURL := getEnv("USER_SERVICE_URL", "")
	if userServiceURL == "" {
		log.Fatal("❌ USER_SERVICE_URL não configurado")
	}

	orderServiceURL := getEnv("ORDER_SERVICE_URL", "")
	if orderServiceURL == "" {
		log.Fatal("❌ ORDER_SERVICE_URL não configurado")
	}

	// Configurações opcionais
	port := getEnv("PORT", "8000")
	allowedOriginsStr := getEnv("ALLOWED_ORIGINS", "*")
	allowedOrigins := parseAllowedOrigins(allowedOriginsStr)

	rateLimitPerMin := getEnvAsInt("RATE_LIMIT_PER_MIN", 100)
	requestTimeout := getEnvAsInt("REQUEST_TIMEOUT", 30)
	maxRetries := getEnvAsInt("MAX_RETRIES", 3)

	enableRateLimit := getEnvAsBool("ENABLE_RATE_LIMIT", true)
	enableCORS := getEnvAsBool("ENABLE_CORS", true)
	enableDetailedLog := getEnvAsBool("ENABLE_DETAILED_LOG", true)

	cfg := &Config{
		Port:              port,
		UserServiceURL:    userServiceURL,
		OrderServiceURL:   orderServiceURL,
		AllowedOrigins:    allowedOrigins,
		RateLimitPerMin:   rateLimitPerMin,
		RequestTimeout:    requestTimeout,
		MaxRetries:        maxRetries,
		ServiceName:       "api-gateway",
		ServiceVersion:    "1.0.0",
		EnableRateLimit:   enableRateLimit,
		EnableCORS:        enableCORS,
		EnableDetailedLog: enableDetailedLog,
	}

	// Log das configurações
	log.Println("✅ Configurações carregadas com sucesso")
	log.Printf("   - Porta: %s", cfg.Port)
	log.Printf("   - User Service: %s", cfg.UserServiceURL)
	log.Printf("   - Order Service: %s", cfg.OrderServiceURL)
	log.Printf("   - Allowed Origins: %v", cfg.AllowedOrigins)
	log.Printf("   - Rate Limit: %d req/min", cfg.RateLimitPerMin)
	log.Printf("   - Request Timeout: %ds", cfg.RequestTimeout)
	log.Printf("   - Max Retries: %d", cfg.MaxRetries)
	log.Printf("   - Rate Limit Enabled: %v", cfg.EnableRateLimit)
	log.Printf("   - CORS Enabled: %v", cfg.EnableCORS)
	log.Printf("   - Detailed Log Enabled: %v", cfg.EnableDetailedLog)

	return cfg
}

// Validate valida se todas as configurações obrigatórias estão presentes
func (c *Config) Validate() error {
	if c.UserServiceURL == "" {
		return fmt.Errorf("USER_SERVICE_URL é obrigatório")
	}

	if c.OrderServiceURL == "" {
		return fmt.Errorf("ORDER_SERVICE_URL é obrigatório")
	}

	if c.Port == "" {
		return fmt.Errorf("PORT é obrigatório")
	}

	if c.RateLimitPerMin <= 0 {
		return fmt.Errorf("RATE_LIMIT_PER_MIN deve ser maior que 0")
	}

	if c.RequestTimeout <= 0 {
		return fmt.Errorf("REQUEST_TIMEOUT deve ser maior que 0")
	}

	if c.MaxRetries < 0 {
		return fmt.Errorf("MAX_RETRIES não pode ser negativo")
	}

	return nil
}

// getEnv retorna o valor de uma variável de ambiente ou um valor padrão
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// getEnvAsInt retorna o valor de uma variável de ambiente como inteiro
func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		log.Printf("⚠️  Valor inválido para %s (%s), usando padrão %d", key, valueStr, defaultValue)
		return defaultValue
	}

	return value
}

// getEnvAsBool retorna o valor de uma variável de ambiente como booleano
func getEnvAsBool(key string, defaultValue bool) bool {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.ParseBool(valueStr)
	if err != nil {
		log.Printf("⚠️  Valor inválido para %s (%s), usando padrão %v", key, valueStr, defaultValue)
		return defaultValue
	}

	return value
}

// parseAllowedOrigins parseia a string de origens permitidas
func parseAllowedOrigins(originsStr string) []string {
	if originsStr == "*" {
		return []string{"*"}
	}

	origins := strings.Split(originsStr, ",")
	result := make([]string, 0, len(origins))

	for _, origin := range origins {
		trimmed := strings.TrimSpace(origin)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}

	if len(result) == 0 {
		return []string{"*"}
	}

	return result
}
