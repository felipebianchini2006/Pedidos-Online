package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

// Config armazena todas as configurações da aplicação
type Config struct {
	// Porta do servidor HTTP
	Port string

	// MongoDB
	MongoURI     string // URI de conexão completa (ex: mongodb://localhost:27017)
	MongoDB      string // Nome do banco de dados
	MongoTimeout int    // Timeout para operações (segundos)

	// JWT
	JWTSecret string // Secret compartilhado com User Service

	// RabbitMQ
	RabbitMQURL string // URL de conexão (ex: amqp://guest:guest@localhost:5672/)

	// User Service
	UserServiceURL string // URL base do User Service (ex: http://user-service:8001)
}

// LoadConfig carrega as configurações de variáveis de ambiente
// Retorna erro se campos obrigatórios não estiverem definidos
func LoadConfig() (*Config, error) {
	// Tentar carregar .env (ignora erro se não existir em produção)
	_ = godotenv.Load()

	config := &Config{
		Port:           getEnv("PORT", "8002"),
		MongoURI:       getEnv("MONGO_URI", "mongodb://localhost:27017"),
		MongoDB:        getEnv("MONGO_DATABASE", "orders_db"),
		MongoTimeout:   getEnvAsInt("MONGO_TIMEOUT", 10),
		JWTSecret:      os.Getenv("JWT_SECRET"),
		RabbitMQURL:    getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
		UserServiceURL: getEnv("USER_SERVICE_URL", "http://localhost:8001"),
	}

	// Validar campos obrigatórios
	if err := config.validate(); err != nil {
		return nil, err
	}

	return config, nil
}

// validate verifica se todos os campos obrigatórios estão preenchidos
func (c *Config) validate() error {
	if c.Port == "" {
		return fmt.Errorf("PORT é obrigatório")
	}

	if c.MongoURI == "" {
		return fmt.Errorf("MONGO_URI é obrigatório")
	}

	if c.MongoDB == "" {
		return fmt.Errorf("MONGO_DATABASE é obrigatório")
	}

	if c.JWTSecret == "" {
		return fmt.Errorf("JWT_SECRET é obrigatório (deve ser o mesmo do User Service)")
	}

	if c.RabbitMQURL == "" {
		return fmt.Errorf("RABBITMQ_URL é obrigatório")
	}

	return nil
}

// getEnv obtém variável de ambiente ou retorna valor padrão
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// getEnvAsInt obtém variável de ambiente como int ou retorna valor padrão
func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}

	var value int
	_, err := fmt.Sscanf(valueStr, "%d", &value)
	if err != nil {
		return defaultValue
	}

	return value
}
