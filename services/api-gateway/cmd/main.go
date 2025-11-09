package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"pedidos-online/api-gateway/internal/config"
	"pedidos-online/api-gateway/internal/middleware"
	"pedidos-online/api-gateway/internal/proxy"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/joho/godotenv"
)

func main() {
	// Banner do serviço
	printBanner()

	// Carregar variáveis de ambiente do arquivo .env
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️  Arquivo .env não encontrado, usando variáveis de ambiente do sistema")
	}

	// Carregar configurações
	cfg := config.LoadConfig()

	// Validar configurações
	if err := cfg.Validate(); err != nil {
		log.Fatalf("❌ Configuração inválida: %v", err)
	}

	// Criar aplicação Fiber
	app := fiber.New(fiber.Config{
		AppName:               "API Gateway v1.0.0",
		DisableStartupMessage: false,
		ErrorHandler:          customErrorHandler,
		ReadTimeout:           30 * time.Second,
		WriteTimeout:          30 * time.Second,
		IdleTimeout:           120 * time.Second,
	})

	// Middlewares globais
	setupMiddlewares(app, cfg)

	// Configurar rotas
	setupRoutes(app, cfg)

	// Canal para capturar sinais de shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	// Iniciar servidor HTTP em goroutine separada
	go func() {
		log.Printf("🚀 API Gateway rodando na porta %s", cfg.Port)
		log.Printf("📡 Roteando requisições para:")
		log.Printf("   - User Service: %s", cfg.UserServiceURL)
		log.Printf("   - Order Service: %s", cfg.OrderServiceURL)
		if err := app.Listen(":" + cfg.Port); err != nil {
			log.Fatalf("❌ Erro ao iniciar servidor HTTP: %v", err)
		}
	}()

	// Aguardar sinal de shutdown
	<-quit
	log.Println("\n🛑 Recebido sinal de shutdown. Encerrando gracefully...")

	// Graceful shutdown do servidor HTTP
	log.Println("🔌 Fechando servidor HTTP...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := app.ShutdownWithContext(ctx); err != nil {
		log.Printf("⚠️  Erro ao encerrar servidor HTTP: %v", err)
	}

	log.Println("✅ API Gateway encerrado com sucesso")
}

// setupMiddlewares configura middlewares globais
func setupMiddlewares(app *fiber.App, cfg *config.Config) {
	// Recover middleware (recupera de panics)
	app.Use(recover.New())

	// CORS middleware
	if cfg.EnableCORS {
		corsMiddleware := middleware.NewDefaultCORSMiddleware(cfg.AllowedOrigins)
		app.Use(corsMiddleware)
		log.Printf("✅ CORS habilitado: %v", cfg.AllowedOrigins)
	}

	// Logger middleware - Log estruturado de todas as requests
	app.Use(middleware.NewDetailedLoggerMiddleware())
	log.Println("✅ Logger middleware habilitado (formato JSON)")

	// Rate limit middleware
	if cfg.EnableRateLimit {
		rateLimitMiddleware := middleware.NewRateLimitMiddleware(cfg.RateLimitPerMin)
		app.Use(rateLimitMiddleware)
		log.Printf("✅ Rate limiting habilitado: %d req/min por IP", cfg.RateLimitPerMin)
	}
}

// setupRoutes configura as rotas da aplicação
func setupRoutes(app *fiber.App, cfg *config.Config) {
	// Root endpoint
	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"service": cfg.ServiceName,
			"version": cfg.ServiceVersion,
			"status":  "running",
			"message": "API Gateway - Centralized routing for microservices",
			"endpoints": fiber.Map{
				"health":        "/health",
				"user_service":  "/api/users/*",
				"order_service": "/api/orders/*",
			},
		})
	})

	// Health check endpoint
	app.Get("/health", healthCheckHandler(cfg))

	// Criar proxy handlers
	userServiceProxy := proxy.ProxyHandler(proxy.ProxyConfig{
		TargetURL:   cfg.UserServiceURL,
		Timeout:     time.Duration(cfg.RequestTimeout) * time.Second,
		MaxRetries:  cfg.MaxRetries,
		RetryDelay:  100 * time.Millisecond,
		ServiceName: "user-service",
	})

	orderServiceProxy := proxy.ProxyHandler(proxy.ProxyConfig{
		TargetURL:   cfg.OrderServiceURL,
		Timeout:     time.Duration(cfg.RequestTimeout) * time.Second,
		MaxRetries:  cfg.MaxRetries,
		RetryDelay:  100 * time.Millisecond,
		ServiceName: "order-service",
	})

	// Rotas de proxy
	app.All("/api/users/*", userServiceProxy)
	app.All("/api/orders/*", orderServiceProxy)

	log.Println("✅ Rotas configuradas:")
	log.Println("   GET  / → Gateway info")
	log.Println("   GET  /health → Health check")
	log.Println("   ALL  /api/users/* → User Service")
	log.Println("   ALL  /api/orders/* → Order Service")
}

// healthCheckHandler retorna um handler para health check
func healthCheckHandler(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		services := make(map[string]string)

		// Verificar User Service
		userServiceStatus := checkServiceHealth(cfg.UserServiceURL + "/health")
		services["user-service"] = userServiceStatus

		// Verificar Order Service
		orderServiceStatus := checkServiceHealth(cfg.OrderServiceURL + "/health")
		services["order-service"] = orderServiceStatus

		// Determinar status geral
		overallStatus := "healthy"
		statusCode := fiber.StatusOK

		if userServiceStatus != "ok" || orderServiceStatus != "ok" {
			overallStatus = "degraded"
		}

		if userServiceStatus == "unreachable" && orderServiceStatus == "unreachable" {
			overallStatus = "unhealthy"
			statusCode = fiber.StatusServiceUnavailable
		}

		return c.Status(statusCode).JSON(fiber.Map{
			"status":   overallStatus,
			"gateway":  "ok",
			"services": services,
		})
	}
}

// checkServiceHealth verifica se um serviço está saudável
func checkServiceHealth(healthURL string) string {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get(healthURL)
	if err != nil {
		return "unreachable"
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return "ok"
	}

	return fmt.Sprintf("unhealthy (status: %d)", resp.StatusCode)
}

// customErrorHandler trata erros globais da aplicação
func customErrorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError

	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
	}

	log.Printf("❌ Erro HTTP %d: %v", code, err)

	return c.Status(code).JSON(fiber.Map{
		"success": false,
		"error":   err.Error(),
	})
}

// printBanner imprime banner do serviço
func printBanner() {
	banner := `
╔═══════════════════════════════════════════════════════╗
║                                                       ║
║         🌐  API GATEWAY - v1.0.0  🌐                 ║
║                                                       ║
║     Centralized Routing for Microservices             ║
║     Pedidos Online - API Gateway                      ║
║                                                       ║
╚═══════════════════════════════════════════════════════╝
`
	log.Println(banner)
}
