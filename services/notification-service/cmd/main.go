package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"pedidos-online/notification-service/internal/config"
	"pedidos-online/notification-service/internal/email"
	"pedidos-online/notification-service/internal/queue"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func main() {
	// Banner do serviço
	printBanner()

	// Carregar configurações
	cfg := config.LoadConfig()

	// Validar configurações
	if err := cfg.Validate(); err != nil {
		log.Fatalf("❌ Configuração inválida: %v", err)
	}

	// Inicializar Email Service
	emailService := email.NewEmailService(cfg.SMTP)

	// Testar conexão SMTP (opcional, mas recomendado)
	if err := emailService.TestConnection(); err != nil {
		log.Printf("⚠️  Aviso: Falha ao testar conexão SMTP: %v", err)
		log.Println("   O serviço continuará, mas e-mails podem falhar.")
	}

	// Inicializar RabbitMQ Consumer
	consumer, err := queue.NewConsumer(cfg.RabbitMQURL, emailService)
	if err != nil {
		log.Fatalf("❌ Erro ao inicializar RabbitMQ Consumer: %v", err)
	}
	defer consumer.Close()

	// Iniciar consumer em goroutine separada (não bloqueia)
	go func() {
		log.Println("🐰 Iniciando RabbitMQ Consumer...")
		if err := consumer.Start(); err != nil {
			log.Fatalf("❌ Erro no RabbitMQ Consumer: %v", err)
		}
	}()

	// Criar aplicação Fiber (servidor HTTP para health check)
	app := fiber.New(fiber.Config{
		AppName:               "Notification Service v1.0.0",
		DisableStartupMessage: false,
		ErrorHandler:          customErrorHandler,
	})

	// Middlewares
	app.Use(recover.New())
	app.Use(logger.New(logger.Config{
		Format:     "${time} | ${status} | ${latency} | ${method} ${path}\n",
		TimeFormat: "15:04:05",
		TimeZone:   "America/Sao_Paulo",
	}))
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,HEAD,OPTIONS",
		AllowHeaders: "Origin, Content-Type, Accept",
	}))

	// Rotas
	setupRoutes(app, cfg, consumer)

	// Canal para capturar sinais de shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	// Iniciar servidor HTTP em goroutine separada
	go func() {
		log.Printf("🚀 Notification Service rodando na porta %s", cfg.Port)
		log.Println("📧 Aguardando eventos de pedidos para enviar notificações...")
		if err := app.Listen(":" + cfg.Port); err != nil {
			log.Fatalf("❌ Erro ao iniciar servidor HTTP: %v", err)
		}
	}()

	// Aguardar sinal de shutdown
	<-quit
	log.Println("\n🛑 Recebido sinal de shutdown. Encerrando gracefully...")

	// Graceful shutdown do servidor HTTP
	log.Println("🔌 Fechando servidor HTTP...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := app.ShutdownWithContext(ctx); err != nil {
		log.Printf("⚠️  Erro ao encerrar servidor HTTP: %v", err)
	}

	// Fechar conexões RabbitMQ
	log.Println("🐰 Fechando RabbitMQ Consumer...")
	if err := consumer.Close(); err != nil {
		log.Printf("⚠️  Erro ao fechar RabbitMQ Consumer: %v", err)
	}

	log.Println("✅ Notification Service encerrado com sucesso")
}

// setupRoutes configura as rotas da aplicação
func setupRoutes(app *fiber.App, cfg *config.Config, consumer *queue.Consumer) {
	// Root endpoint
	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"service": cfg.ServiceName,
			"version": cfg.ServiceVersion,
			"status":  "running",
			"message": "Notification Service - Sistema de envio de notificações por e-mail",
		})
	})

	// Health check endpoint
	app.Get("/health", func(c *fiber.Ctx) error {
		// Verificar conexão RabbitMQ
		rabbitMQStatus := "ok"
		if !consumer.IsConnected() {
			rabbitMQStatus = "unhealthy"
			log.Printf("⚠️  RabbitMQ health check failed: not connected")
		}

		// Determinar status geral e código HTTP
		overallStatus := "healthy"
		statusCode := fiber.StatusOK

		if rabbitMQStatus == "unhealthy" {
			overallStatus = "unhealthy"
			statusCode = fiber.StatusServiceUnavailable // 503
		}

		return c.Status(statusCode).JSON(fiber.Map{
			"status":   overallStatus,
			"service":  cfg.ServiceName,
			"version":  cfg.ServiceVersion,
			"rabbitmq": rabbitMQStatus,
			"smtp": fiber.Map{
				"host": cfg.SMTP.Host,
				"port": cfg.SMTP.Port,
			},
		})
	})

	// Ready endpoint (Kubernetes readiness probe)
	app.Get("/ready", func(c *fiber.Ctx) error {
		if !consumer.IsConnected() {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"ready":   false,
				"message": "RabbitMQ não conectado",
			})
		}

		return c.JSON(fiber.Map{
			"ready": true,
		})
	})

	// Metrics endpoint (exemplo simples)
	app.Get("/metrics", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"service":            cfg.ServiceName,
			"version":            cfg.ServiceVersion,
			"rabbitmq_connected": consumer.IsConnected(),
			"uptime":             time.Since(startTime).String(),
		})
	})
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
║     📧  NOTIFICATION SERVICE - v1.0.0  📧            ║
║                                                       ║
║     Sistema de Notificações por E-mail                ║
║     Pedidos Online - Microserviços                    ║
║                                                       ║
╚═══════════════════════════════════════════════════════╝
`
	log.Println(banner)
}

// startTime armazena o momento em que o serviço foi iniciado
var startTime = time.Now()
