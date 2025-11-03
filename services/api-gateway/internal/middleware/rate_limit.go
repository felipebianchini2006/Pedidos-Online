package middleware

import (
	"fmt"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
)

// RateLimiter implementa rate limiting por IP usando in-memory cache
type RateLimiter struct {
	mu            sync.RWMutex
	requests      map[string]*requestCounter
	limit         int           // Número máximo de requests
	window        time.Duration // Janela de tempo
	cleanupTicker *time.Ticker
	stopCleanup   chan bool
}

// requestCounter armazena contador de requests para um IP
type requestCounter struct {
	count     int
	resetTime time.Time
}

// NewRateLimiter cria um novo rate limiter
//
// Parâmetros:
// - limit: Número máximo de requests permitidos por janela
// - window: Duração da janela (ex: 1 minuto)
//
// Exemplo: NewRateLimiter(100, time.Minute) = 100 requests por minuto
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		requests:    make(map[string]*requestCounter),
		limit:       limit,
		window:      window,
		stopCleanup: make(chan bool),
	}

	// Iniciar goroutine de limpeza (remove entradas expiradas a cada minuto)
	rl.cleanupTicker = time.NewTicker(time.Minute)
	go rl.cleanup()

	return rl
}

// Middleware retorna um handler Fiber para rate limiting
func (rl *RateLimiter) Middleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		ip := c.IP()

		// Verificar se o IP excedeu o limite
		allowed, remaining, resetTime := rl.Allow(ip)

		// Adicionar headers de rate limit
		c.Set("X-RateLimit-Limit", fmt.Sprintf("%d", rl.limit))
		c.Set("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
		c.Set("X-RateLimit-Reset", fmt.Sprintf("%d", resetTime.Unix()))

		if !allowed {
			// Limite excedido - retornar 429 Too Many Requests
			retryAfter := time.Until(resetTime).Seconds()
			c.Set("Retry-After", fmt.Sprintf("%.0f", retryAfter))

			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"success": false,
				"error":   "Rate limit exceeded",
				"message": fmt.Sprintf("Too many requests. Please try again in %.0f seconds.", retryAfter),
				"limit":   rl.limit,
				"window":  rl.window.String(),
			})
		}

		// Permitir request
		return c.Next()
	}
}

// Allow verifica se uma request do IP é permitida
// Retorna: (allowed, remaining, resetTime)
func (rl *RateLimiter) Allow(ip string) (bool, int, time.Time) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()

	// Buscar contador existente
	counter, exists := rl.requests[ip]

	// Se não existe ou expirou, criar novo
	if !exists || now.After(counter.resetTime) {
		counter = &requestCounter{
			count:     0,
			resetTime: now.Add(rl.window),
		}
		rl.requests[ip] = counter
	}

	// Incrementar contador
	counter.count++

	// Calcular remaining
	remaining := rl.limit - counter.count
	if remaining < 0 {
		remaining = 0
	}

	// Verificar se excedeu o limite
	allowed := counter.count <= rl.limit

	return allowed, remaining, counter.resetTime
}

// cleanup remove entradas expiradas do cache periodicamente
func (rl *RateLimiter) cleanup() {
	for {
		select {
		case <-rl.cleanupTicker.C:
			rl.removeExpired()
		case <-rl.stopCleanup:
			rl.cleanupTicker.Stop()
			return
		}
	}
}

// removeExpired remove entradas expiradas do cache
func (rl *RateLimiter) removeExpired() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	for ip, counter := range rl.requests {
		if now.After(counter.resetTime) {
			delete(rl.requests, ip)
		}
	}
}

// Stop para a goroutine de limpeza
func (rl *RateLimiter) Stop() {
	close(rl.stopCleanup)
}

// GetStats retorna estatísticas do rate limiter
func (rl *RateLimiter) GetStats() map[string]interface{} {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	return map[string]interface{}{
		"total_ips": len(rl.requests),
		"limit":     rl.limit,
		"window":    rl.window.String(),
	}
}

// NewRateLimitMiddleware cria um middleware de rate limiting com configuração padrão
//
// Parâmetros:
// - requestsPerMinute: Número de requests permitidos por minuto
//
// Exemplo: NewRateLimitMiddleware(100) = 100 requests por minuto por IP
func NewRateLimitMiddleware(requestsPerMinute int) fiber.Handler {
	limiter := NewRateLimiter(requestsPerMinute, time.Minute)
	return limiter.Middleware()
}
