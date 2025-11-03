package middleware

import (
	"encoding/json"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
)

// LogEntry representa uma entrada de log estruturada
type LogEntry struct {
	Timestamp  string `json:"timestamp"`
	Method     string `json:"method"`
	Path       string `json:"path"`
	StatusCode int    `json:"status_code"`
	Latency    int64  `json:"latency_ms"`
	IP         string `json:"ip"`
	UserAgent  string `json:"user_agent"`
	Error      string `json:"error,omitempty"`
}

// NewLoggerMiddleware cria um middleware de logging estruturado
//
// O middleware registra:
// - Timestamp da request
// - Método HTTP (GET, POST, etc)
// - Path da request
// - Status code da response
// - Latência em milissegundos
// - IP do cliente
// - User-Agent
// - Erros (se houver)
//
// Formato: JSON para facilitar parsing por ferramentas de log aggregation
func NewLoggerMiddleware(detailedLog bool) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Capturar tempo de início
		start := time.Now()

		// Processar request
		err := c.Next()

		// Calcular latência
		latency := time.Since(start)

		// Criar entrada de log
		entry := LogEntry{
			Timestamp:  start.Format(time.RFC3339),
			Method:     c.Method(),
			Path:       c.Path(),
			StatusCode: c.Response().StatusCode(),
			Latency:    latency.Milliseconds(),
			IP:         c.IP(),
			UserAgent:  c.Get("User-Agent"),
		}

		// Adicionar erro se houver
		if err != nil {
			entry.Error = err.Error()
		}

		// Log em formato JSON se detailedLog estiver ativado
		if detailedLog {
			logJSON, _ := json.Marshal(entry)
			log.Printf("%s", string(logJSON))
		} else {
			// Log simplificado
			emoji := getStatusEmoji(entry.StatusCode)
			log.Printf("%s %s %s | %d | %dms | %s",
				emoji,
				entry.Method,
				entry.Path,
				entry.StatusCode,
				entry.Latency,
				entry.IP,
			)
		}

		return err
	}
}

// getStatusEmoji retorna um emoji baseado no status code
func getStatusEmoji(statusCode int) string {
	switch {
	case statusCode >= 500:
		return "❌"
	case statusCode >= 400:
		return "⚠️"
	case statusCode >= 300:
		return "🔄"
	case statusCode >= 200:
		return "✅"
	default:
		return "ℹ️"
	}
}

// NewSimpleLoggerMiddleware cria um middleware de logging simples (não-JSON)
func NewSimpleLoggerMiddleware() fiber.Handler {
	return NewLoggerMiddleware(false)
}

// NewDetailedLoggerMiddleware cria um middleware de logging detalhado (JSON)
func NewDetailedLoggerMiddleware() fiber.Handler {
	return NewLoggerMiddleware(true)
}
