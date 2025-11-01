package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	_ "github.com/jackc/pgx/v5/stdlib"

	"pedidos-online/user-service/internal/config"
	"pedidos-online/user-service/internal/handler"
	"pedidos-online/user-service/internal/middleware"
	"pedidos-online/user-service/internal/repository"
	"pedidos-online/user-service/internal/service"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Connect to PostgreSQL
	db, err := connectDatabase(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	log.Println("✅ Connected to PostgreSQL successfully")

	// Initialize Fiber app
	app := fiber.New(fiber.Config{
		AppName:           "User Service v1.0",
		EnablePrintRoutes: cfg.IsDevelopment(),
		ErrorHandler:      errorHandler,
	})

	// Setup middlewares
	setupMiddlewares(app, cfg)

	// Setup routes
	setupRoutes(app, db, cfg)

	// Start server in goroutine
	go func() {
		addr := fmt.Sprintf(":%s", cfg.Port)
		log.Printf("🚀 Server starting on http://localhost%s", addr)
		if err := app.Listen(addr); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Graceful shutdown
	gracefulShutdown(app, db)
}

// connectDatabase establishes a connection to PostgreSQL with connection pooling
func connectDatabase(cfg *config.Config) (*sql.DB, error) {
	dsn := cfg.GetDSN()

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(cfg.DBMaxOpenConns)
	db.SetMaxIdleConns(cfg.DBMaxIdleConns)
	db.SetConnMaxLifetime(cfg.DBConnMaxLifetime)

	// Verify connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

// setupMiddlewares configures all middlewares for the application
func setupMiddlewares(app *fiber.App, cfg *config.Config) {
	// Recover middleware - recovers from panics
	app.Use(recover.New(recover.Config{
		EnableStackTrace: cfg.IsDevelopment(),
	}))

	// Logger middleware - logs all requests
	if cfg.IsDevelopment() {
		app.Use(logger.New(logger.Config{
			Format:     "[${time}] ${status} - ${method} ${path} (${latency})\n",
			TimeFormat: "15:04:05",
			TimeZone:   "Local",
		}))
	} else {
		app.Use(logger.New(logger.Config{
			Format:     "${time} | ${status} | ${latency} | ${ip} | ${method} | ${path} | ${error}\n",
			TimeFormat: "2006-01-02 15:04:05",
			TimeZone:   "UTC",
		}))
	}

	// CORS middleware - handles cross-origin requests
	origins := strings.Split(cfg.CORSOrigins, ",")
	for i, origin := range origins {
		origins[i] = strings.TrimSpace(origin)
	}

	app.Use(cors.New(cors.Config{
		AllowOrigins:     strings.Join(origins, ","),
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization",
		AllowCredentials: true,
		ExposeHeaders:    "Content-Length",
		MaxAge:           86400, // 24 hours
	}))
}

// setupRoutes configures all routes for the application
func setupRoutes(app *fiber.App, db *sql.DB, cfg *config.Config) {
	// Initialize layers
	userRepo := repository.NewUserRepository(db)
	userService := service.NewUserService(userRepo, cfg.JWTSecret, cfg.JWTExpiration)
	userHandler := handler.NewUserHandler(userService)

	// Create auth middleware (usando JWT secret diretamente)
	authMiddleware := middleware.AuthMiddleware(cfg.JWTSecret)

	// Register user routes
	handler.RegisterRoutes(app, userHandler, authMiddleware)

	// Health check endpoint
	app.Get("/health", func(c *fiber.Ctx) error {
		// Check database connection
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		if err := db.PingContext(ctx); err != nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"success": false,
				"status":  "unhealthy",
				"error":   "database connection failed",
			})
		}

		return c.JSON(fiber.Map{
			"success": true,
			"status":  "healthy",
			"service": "user-service",
			"version": "1.0.0",
		})
	})

	// Ready check endpoint (more strict than health)
	app.Get("/ready", func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		// Check if database is ready
		var result int
		if err := db.QueryRowContext(ctx, "SELECT 1").Scan(&result); err != nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"success": false,
				"ready":   false,
				"error":   "database not ready",
			})
		}

		return c.JSON(fiber.Map{
			"success": true,
			"ready":   true,
			"service": "user-service",
		})
	})

	// 404 handler
	app.Use(func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"error":   "route not found",
			"message": fmt.Sprintf("Cannot %s %s", c.Method(), c.Path()),
		})
	})
}

// errorHandler is a custom error handler for Fiber
func errorHandler(c *fiber.Ctx, err error) error {
	// Default status code
	code := fiber.StatusInternalServerError

	// Check if it's a Fiber error
	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
	}

	// Log error
	log.Printf("ERROR: [%d] %s - %v", code, c.Path(), err)

	// Send error response
	return c.Status(code).JSON(fiber.Map{
		"success": false,
		"error":   err.Error(),
	})
}

// gracefulShutdown handles graceful shutdown of the application
func gracefulShutdown(app *fiber.App, db *sql.DB) {
	// Create channel to listen for interrupt signals
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	// Block until signal is received
	<-quit

	log.Println("🛑 Shutting down server...")

	// Shutdown server with timeout
	if err := app.ShutdownWithTimeout(10 * time.Second); err != nil {
		log.Printf("Error during server shutdown: %v", err)
	}

	// Close database connection
	if err := db.Close(); err != nil {
		log.Printf("Error closing database: %v", err)
	}

	log.Println("✅ Server stopped gracefully")
}
