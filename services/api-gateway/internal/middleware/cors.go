package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

// CORSConfig armazena configurações do CORS
type CORSConfig struct {
	AllowedOrigins   []string // Origens permitidas
	AllowedMethods   []string // Métodos permitidos
	AllowedHeaders   []string // Headers permitidos
	AllowCredentials bool     // Permitir envio de credenciais (cookies)
	MaxAge           int      // Tempo de cache do preflight (segundos)
}

// NewCORSMiddleware cria um middleware CORS configurável
//
// Configurações:
// - AllowedOrigins: Lista de origens permitidas (usar ["*"] para permitir todas)
// - AllowedMethods: GET, POST, PUT, DELETE, PATCH, OPTIONS
// - AllowedHeaders: Authorization, Content-Type, X-Requested-With, etc
// - AllowCredentials: true para permitir cookies e headers de autenticação
// - MaxAge: Tempo que o browser pode cachear a response do preflight
func NewCORSMiddleware(config CORSConfig) fiber.Handler {
	// Valores padrão
	if len(config.AllowedMethods) == 0 {
		config.AllowedMethods = []string{
			"GET",
			"POST",
			"PUT",
			"DELETE",
			"PATCH",
			"OPTIONS",
			"HEAD",
		}
	}

	if len(config.AllowedHeaders) == 0 {
		config.AllowedHeaders = []string{
			"Origin",
			"Content-Type",
			"Accept",
			"Authorization",
			"X-Requested-With",
			"X-CSRF-Token",
		}
	}

	if config.MaxAge == 0 {
		config.MaxAge = 3600 // 1 hora
	}

	// Configurar CORS do Fiber
	corsConfig := cors.Config{
		AllowOrigins:     strings.Join(config.AllowedOrigins, ","),
		AllowMethods:     strings.Join(config.AllowedMethods, ","),
		AllowHeaders:     strings.Join(config.AllowedHeaders, ","),
		AllowCredentials: config.AllowCredentials,
		MaxAge:           config.MaxAge,
	}

	return cors.New(corsConfig)
}

// NewDefaultCORSMiddleware cria um middleware CORS com configuração padrão segura
func NewDefaultCORSMiddleware(allowedOrigins []string) fiber.Handler {
	// Se usar wildcard (*), não pode ter AllowCredentials=true
	allowCredentials := true
	if len(allowedOrigins) == 1 && allowedOrigins[0] == "*" {
		allowCredentials = false
	}

	return NewCORSMiddleware(CORSConfig{
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowedHeaders:   []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowCredentials: allowCredentials,
		MaxAge:           3600,
	})
}
