package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"pedidos-online/order-service/internal/config"
	"pedidos-online/order-service/internal/handler"
	"pedidos-online/order-service/internal/middleware"
	"pedidos-online/order-service/internal/queue"
	"pedidos-online/order-service/internal/repository"
)

func main() {
	// Banner
	printBanner()

	// Carregar configurações
	log.Println("📋 Carregando configurações...")
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("❌ Erro ao carregar configurações: %v", err)
	}
	log.Println("✅ Configurações carregadas")

	// Conectar ao MongoDB
	log.Println("🍃 Conectando ao MongoDB...")
	mongoClient, err := connectMongoDB(cfg)
	if err != nil {
		log.Fatalf("❌ Erro ao conectar ao MongoDB: %v", err)
	}
	defer disconnectMongoDB(mongoClient)
	log.Println("✅ MongoDB conectado")

	// Obter database
	db := mongoClient.Database(cfg.MongoDB)

	// Criar repository
	orderRepo := repository.NewOrderRepository(db)

	// Criar índices
	log.Println("🔧 Criando índices MongoDB...")
	if err := orderRepo.CreateIndexes(context.Background()); err != nil {
		log.Printf("⚠️  Erro ao criar índices: %v", err)
	} else {
		log.Println("✅ Índices criados")
	}

	// Conectar ao RabbitMQ
	log.Println("🐰 Conectando ao RabbitMQ...")
	publisher, err := queue.NewPublisher(cfg.RabbitMQURL)
	if err != nil {
		log.Fatalf("❌ Erro ao conectar ao RabbitMQ: %v", err)
	}
	defer publisher.Close()

	// Criar handler
	orderHandler := handler.NewOrderHandler(orderRepo, publisher)

	// Inicializar Fiber
	app := fiber.New(fiber.Config{
		AppName:      "Order Service v1.0.0",
		ErrorHandler: customErrorHandler,
	})

	// Middlewares globais
	app.Use(recover.New()) // Recupera de panics
	app.Use(logger.New(logger.Config{
		Format: "[${time}] ${status} - ${method} ${path} (${latency})\n",
	}))
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders: "Origin,Content-Type,Accept,Authorization",
	}))

	// Rotas públicas
	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"service": "Order Service",
			"version": "1.0.0",
			"status":  "running",
		})
	})

	app.Get("/health", func(c *fiber.Ctx) error {
		// Verificar MongoDB
		mongoStatus := "healthy"
		if err := mongoClient.Ping(c.Context(), nil); err != nil {
			mongoStatus = "unhealthy"
		}

		// Verificar RabbitMQ
		rabbitMQStatus := "healthy"
		if !publisher.IsConnected() {
			rabbitMQStatus = "unhealthy"
		}

		overallStatus := "healthy"
		if mongoStatus == "unhealthy" || rabbitMQStatus == "unhealthy" {
			overallStatus = "unhealthy"
		}

		return c.JSON(fiber.Map{
			"status":   overallStatus,
			"service":  "order-service",
			"version":  "1.0.0",
			"mongodb":  mongoStatus,
			"rabbitmq": rabbitMQStatus,
		})
	})

	// Grupo de rotas da API v1 (protegidas com autenticação)
	api := app.Group("/api/v1")

	// Aplicar middleware de autenticação
	api.Use(middleware.AuthMiddleware(cfg.JWTSecret))

	// Rotas de pedidos
	orders := api.Group("/orders")
	orders.Post("/", orderHandler.CreateOrder)                // Criar pedido
	orders.Get("/", orderHandler.GetOrders)                   // Listar pedidos do usuário
	orders.Get("/:id", orderHandler.GetOrderByID)             // Obter pedido específico
	orders.Put("/:id/status", orderHandler.UpdateOrderStatus) // Atualizar status

	// Iniciar servidor em goroutine
	go func() {
		addr := fmt.Sprintf(":%s", cfg.Port)
		log.Printf("🚀 Order Service rodando na porta %s", cfg.Port)
		log.Printf("📍 Endpoints disponíveis:")
		log.Printf("   GET  /               - Info do serviço")
		log.Printf("   GET  /health         - Health check")
		log.Printf("   POST /api/v1/orders  - Criar pedido (autenticado)")
		log.Printf("   GET  /api/v1/orders  - Listar pedidos (autenticado)")
		log.Printf("   GET  /api/v1/orders/:id - Obter pedido (autenticado)")
		log.Printf("   PUT  /api/v1/orders/:id/status - Atualizar status")

		if err := app.Listen(addr); err != nil {
			log.Fatalf("❌ Erro ao iniciar servidor: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Println("\n🛑 Desligando Order Service...")

	// Shutdown do Fiber
	log.Println("🔌 Fechando servidor HTTP...")
	if err := app.Shutdown(); err != nil {
		log.Printf("⚠️  Erro ao fechar servidor: %v", err)
	}

	log.Println("✅ Order Service desligado com sucesso")
}

// connectMongoDB estabelece conexão com MongoDB
func connectMongoDB(cfg *config.Config) (*mongo.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.MongoTimeout)*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(cfg.MongoURI)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, err
	}

	// Verificar conexão
	if err := client.Ping(ctx, nil); err != nil {
		return nil, err
	}

	return client, nil
}

// disconnectMongoDB fecha a conexão com MongoDB
func disconnectMongoDB(client *mongo.Client) {
	if client != nil {
		log.Println("🔌 Fechando conexão MongoDB...")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := client.Disconnect(ctx); err != nil {
			log.Printf("⚠️  Erro ao desconectar MongoDB: %v", err)
		} else {
			log.Println("✅ MongoDB desconectado")
		}
	}
}

// customErrorHandler manipula erros do Fiber
func customErrorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError

	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
	}

	return c.Status(code).JSON(fiber.Map{
		"success": false,
		"error":   err.Error(),
	})
}

// printBanner exibe o banner do serviço
func printBanner() {
	banner := `
╔═══════════════════════════════════════════════════════════╗
║                                                           ║
║            📦 ORDER SERVICE - Pedidos Online             ║
║                      Version 1.0.0                        ║
║                                                           ║
║  Microserviço de gerenciamento de pedidos                ║
║  MongoDB + RabbitMQ + JWT Authentication                 ║
║                                                           ║
╚═══════════════════════════════════════════════════════════╝
`
	fmt.Println(banner)
}
