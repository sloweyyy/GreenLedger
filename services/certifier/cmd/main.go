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

	"github.com/gin-gonic/gin"
	"github.com/sloweyyy/GreenLedger/services/certifier/internal/handler"
	"github.com/sloweyyy/GreenLedger/services/certifier/internal/models"
	"github.com/sloweyyy/GreenLedger/services/certifier/internal/repository"
	"github.com/sloweyyy/GreenLedger/services/certifier/internal/service"
	"github.com/sloweyyy/GreenLedger/shared/config"
	"github.com/sloweyyy/GreenLedger/shared/database"
	sharedLogger "github.com/sloweyyy/GreenLedger/shared/logger"
	"github.com/sloweyyy/GreenLedger/shared/middleware"
)

// @title GreenLedger Certificate Service API
// @version 1.0
// @description Carbon offset certificate management service for GreenLedger
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.greenledger.com/support
// @contact.email support@greenledger.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8086
// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Override database name for certificate service
	cfg.Database.DBName = "certifier_db"
	cfg.Server.Port = 8086
	cfg.Server.GRPCPort = 9086

	// Initialize logger
	logger := sharedLogger.New(cfg.Server.LogLevel).WithService("certifier")

	// Initialize database
	db, err := database.NewPostgresDB(&cfg.Database, logger)
	if err != nil {
		logger.LogError(context.Background(), "failed to connect to database", err)
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Run database migrations
	if err := db.Migrate(
		&models.Certificate{},
		&models.CertificateVerification{},
		&models.CertificateTransfer{},
		&models.CertificateTemplate{},
		&models.CertificateProject{},
	); err != nil {
		logger.LogError(context.Background(), "failed to run migrations", err)
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Initialize repositories
	certificateRepo := repository.NewCertificateRepository(db, logger)
	projectRepo := repository.NewProjectRepository(db, logger)

	// Initialize services
	certificateService := service.NewCertificateService(
		certificateRepo,
		projectRepo,
		logger,
	)

	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware(cfg.Server.JWTSecret, logger)

	// Initialize handlers
	certificateHandler := handler.NewCertificateHandler(certificateService, logger)

	// Setup Gin router
	if cfg.Server.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.RequestLogger(logger))
	router.Use(middleware.CORS())

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		if err := db.HealthCheck(c.Request.Context()); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "unhealthy",
				"error":  err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": "certifier",
			"version": "1.0.0",
		})
	})

	// API routes
	v1 := router.Group("/api/v1")
	certificateHandler.RegisterRoutes(v1, authMiddleware)

	// Create HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		logger.LogInfo(context.Background(), "starting certificate service",
			sharedLogger.Int("port", cfg.Server.Port),
			sharedLogger.String("environment", cfg.Server.Environment))

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.LogError(context.Background(), "failed to start server", err)
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.LogInfo(context.Background(), "shutting down certificate service")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.LogError(context.Background(), "server forced to shutdown", err)
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	logger.LogInfo(context.Background(), "certificate service stopped")
}
