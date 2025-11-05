package proxy

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

// ProxyConfig armazena configurações do proxy
type ProxyConfig struct {
	TargetURL   string            // URL base do serviço de destino
	Timeout     time.Duration     // Timeout para requests
	MaxRetries  int               // Número máximo de tentativas
	RetryDelay  time.Duration     // Delay inicial entre tentativas
	ServiceName string            // Nome do serviço (para logging)
	CopyHeaders []string          // Headers para copiar (vazio = copiar todos)
	SkipHeaders []string          // Headers para não copiar
	AddHeaders  map[string]string // Headers adicionais
}

// ProxyHandler cria um handler Fiber que faz proxy para outro serviço
//
// O handler:
// 1. Extrai path e query params da request original
// 2. Constrói URL completo do serviço de destino
// 3. Copia headers relevantes (Authorization, Content-Type, etc)
// 4. Faz request HTTP ao serviço com timeout
// 5. Implementa retry logic com backoff exponencial
// 6. Copia response (status, headers, body) de volta ao cliente
// 7. Trata erros de conexão (503 Service Unavailable)
// 8. Adiciona header X-Gateway-Time com tempo de resposta
// 9. Faz logging de cada request
func ProxyHandler(config ProxyConfig) fiber.Handler {
	// Criar HTTP client com timeout
	client := &http.Client{
		Timeout: config.Timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// Não seguir redirects automaticamente
			return http.ErrUseLastResponse
		},
	}

	// Normalizar URL base (remover trailing slash)
	targetURL := strings.TrimSuffix(config.TargetURL, "/")

	return func(c *fiber.Ctx) error {
		startTime := time.Now()

		// Adicionar nome do serviço no contexto para uso nos middlewares (metrics, logging)
		c.Locals("service_name", config.ServiceName)

		// Construir URL completo do serviço de destino
		// Remove o prefixo /api/users ou /api/orders e mantém o resto do path
		fullPath := c.Path()

		// Extrair o path após o prefixo do serviço
		// Ex: /api/users/profile -> /profile
		//     /api/orders/123 -> /123
		var servicePath string
		if strings.HasPrefix(fullPath, "/api/users") {
			servicePath = strings.TrimPrefix(fullPath, "/api/users")
		} else if strings.HasPrefix(fullPath, "/api/orders") {
			servicePath = strings.TrimPrefix(fullPath, "/api/orders")
		} else {
			servicePath = fullPath
		}

		// Se não há path após o prefixo, usar /
		if servicePath == "" {
			servicePath = "/"
		}

		targetURLComplete := targetURL + servicePath

		// Adicionar query params se existirem
		if len(c.Request().URI().QueryString()) > 0 {
			targetURLComplete += "?" + string(c.Request().URI().QueryString())
		}

		// Tentar fazer o request com retry logic
		var lastErr error
		var resp *http.Response

		for attempt := 0; attempt <= config.MaxRetries; attempt++ {
			if attempt > 0 {
				// Backoff exponencial: 100ms, 200ms, 400ms, 800ms...
				backoff := config.RetryDelay * time.Duration(1<<uint(attempt-1))
				log.Printf("🔄 Retry %d/%d para %s após %v", attempt, config.MaxRetries, targetURLComplete, backoff)
				time.Sleep(backoff)
			}

			// Criar request
			req, err := createProxyRequest(c, targetURLComplete, config)
			if err != nil {
				lastErr = fmt.Errorf("erro ao criar request: %w", err)
				continue
			}

			// Fazer request ao serviço
			resp, err = client.Do(req)
			if err != nil {
				lastErr = fmt.Errorf("erro ao fazer request: %w", err)
				log.Printf("⚠️  Tentativa %d/%d falhou para %s %s: %v",
					attempt+1, config.MaxRetries+1, c.Method(), targetURLComplete, err)
				continue
			}

			// Se chegou aqui, request foi bem-sucedido
			lastErr = nil
			break
		}

		// Se todas as tentativas falharam
		if lastErr != nil {
			elapsed := time.Since(startTime)
			log.Printf("❌ Todas as tentativas falharam para %s %s (%.2fms): %v",
				c.Method(), targetURLComplete, elapsed.Seconds()*1000, lastErr)

			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"success": false,
				"error":   "Service temporarily unavailable",
				"message": fmt.Sprintf("%s is not responding", config.ServiceName),
			})
		}

		defer resp.Body.Close()

		// Calcular tempo de resposta
		elapsed := time.Since(startTime)
		elapsedMs := elapsed.Milliseconds()

		// Copiar status code
		c.Status(resp.StatusCode)

		// Copiar headers da response
		for key, values := range resp.Header {
			// Pular alguns headers que o Fiber já gerencia
			if shouldSkipResponseHeader(key) {
				continue
			}
			for _, value := range values {
				c.Response().Header.Add(key, value)
			}
		}

		// Adicionar header customizado com tempo de resposta
		c.Set("X-Gateway-Time", fmt.Sprintf("%dms", elapsedMs))
		c.Set("X-Gateway-Service", config.ServiceName)

		// Copiar body
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Printf("❌ Erro ao ler body da response: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"error":   "Failed to read response from service",
			})
		}

		// Log da request
		logProxyRequest(c, targetURLComplete, resp.StatusCode, elapsedMs, config.ServiceName)

		// Enviar response
		return c.Send(body)
	}
}

// createProxyRequest cria uma nova HTTP request copiando dados da request Fiber
func createProxyRequest(c *fiber.Ctx, targetURL string, config ProxyConfig) (*http.Request, error) {
	// Ler body da request original
	var bodyReader io.Reader
	if len(c.Body()) > 0 {
		bodyReader = bytes.NewReader(c.Body())
	}

	// Criar request
	req, err := http.NewRequest(c.Method(), targetURL, bodyReader)
	if err != nil {
		return nil, err
	}

	// Copiar headers relevantes da request original
	c.Request().Header.VisitAll(func(key, value []byte) {
		headerKey := string(key)
		headerValue := string(value)

		// Pular headers que não devem ser copiados
		if shouldSkipRequestHeader(headerKey, config.SkipHeaders) {
			return
		}

		// Se CopyHeaders está especificado, copiar apenas esses
		if len(config.CopyHeaders) > 0 {
			for _, allowed := range config.CopyHeaders {
				if strings.EqualFold(headerKey, allowed) {
					req.Header.Set(headerKey, headerValue)
					break
				}
			}
		} else {
			// Caso contrário, copiar todos exceto os da skip list
			req.Header.Set(headerKey, headerValue)
		}
	})

	// Adicionar headers customizados
	for key, value := range config.AddHeaders {
		req.Header.Set(key, value)
	}

	// Garantir que Content-Type seja copiado para POST/PUT/PATCH
	if c.Method() == "POST" || c.Method() == "PUT" || c.Method() == "PATCH" {
		if req.Header.Get("Content-Type") == "" {
			contentType := c.Get("Content-Type")
			if contentType != "" {
				req.Header.Set("Content-Type", contentType)
			}
		}
	}

	return req, nil
}

// shouldSkipRequestHeader verifica se um header deve ser pulado na request
func shouldSkipRequestHeader(header string, skipList []string) bool {
	header = strings.ToLower(header)

	// Headers que sempre devem ser pulados (gerenciados pelo Go)
	defaultSkip := []string{
		"connection",
		"keep-alive",
		"proxy-connection",
		"proxy-authenticate",
		"proxy-authorization",
		"te",
		"trailer",
		"transfer-encoding",
		"upgrade",
	}

	for _, skip := range defaultSkip {
		if header == skip {
			return true
		}
	}

	// Verificar skip list customizada
	for _, skip := range skipList {
		if strings.EqualFold(header, skip) {
			return true
		}
	}

	return false
}

// shouldSkipResponseHeader verifica se um header deve ser pulado na response
func shouldSkipResponseHeader(header string) bool {
	header = strings.ToLower(header)

	// Headers que sempre devem ser pulados (gerenciados pelo Fiber)
	skip := []string{
		"connection",
		"keep-alive",
		"proxy-connection",
		"proxy-authenticate",
		"proxy-authorization",
		"te",
		"trailer",
		"transfer-encoding",
		"upgrade",
		"content-length", // Fiber gerencia automaticamente
	}

	for _, s := range skip {
		if header == s {
			return true
		}
	}

	return false
}

// logProxyRequest faz log de uma request proxied
func logProxyRequest(c *fiber.Ctx, targetURL string, statusCode int, elapsedMs int64, serviceName string) {
	// Determinar emoji baseado no status code
	emoji := "✅"
	if statusCode >= 500 {
		emoji = "❌"
	} else if statusCode >= 400 {
		emoji = "⚠️"
	}

	log.Printf("%s [%s] %s %s → %s | Status: %d | Time: %dms | IP: %s",
		emoji,
		serviceName,
		c.Method(),
		c.Path(),
		targetURL,
		statusCode,
		elapsedMs,
		c.IP(),
	)
}

// NewDefaultProxyConfig cria uma configuração padrão para o proxy
func NewDefaultProxyConfig(targetURL, serviceName string) ProxyConfig {
	return ProxyConfig{
		TargetURL:   targetURL,
		Timeout:     30 * time.Second,
		MaxRetries:  3,
		RetryDelay:  100 * time.Millisecond,
		ServiceName: serviceName,
		CopyHeaders: []string{}, // Vazio = copiar todos
		SkipHeaders: []string{},
		AddHeaders:  map[string]string{},
	}
}
