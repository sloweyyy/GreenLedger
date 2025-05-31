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
	"github.com/sloweyyy/GreenLedger/services/reporting/internal/handler"
	"github.com/sloweyyy/GreenLedger/services/reporting/internal/models"
	"github.com/sloweyyy/GreenLedger/services/reporting/internal/repository"
	"github.com/sloweyyy/GreenLedger/services/reporting/internal/service"
	"github.com/sloweyyy/GreenLedger/shared/config"
	"github.com/sloweyyy/GreenLedger/shared/database"
	sharedLogger "github.com/sloweyyy/GreenLedger/shared/logger"
	"github.com/sloweyyy/GreenLedger/shared/middleware"
)

// @title GreenLedger Reporting Service API
// @version 1.0
// @description Report generation and analytics service for GreenLedger
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.greenledger.com/support
// @contact.email support@greenledger.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8085
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

	// Override database name for reporting service
	cfg.Database.DBName = "reporting_db"
	cfg.Server.Port = 8085
	cfg.Server.GRPCPort = 9085

	// Initialize logger
	logger := sharedLogger.New(cfg.Server.LogLevel).WithService("reporting")

	// Initialize main database
	db, err := database.NewPostgresDB(&cfg.Database, logger)
	if err != nil {
		logger.LogError(context.Background(), "failed to connect to database", err)
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize external service databases for data collection
	calculatorDBConfig := cfg.Database
	calculatorDBConfig.DBName = "calculator_db"
	calculatorDB, err := database.NewPostgresDB(&calculatorDBConfig, logger)
	if err != nil {
		logger.LogWarn(context.Background(), "failed to connect to calculator database",
			sharedLogger.String("error", err.Error()))
		calculatorDB = nil
	}

	trackerDBConfig := cfg.Database
	trackerDBConfig.DBName = "tracker_db"
	trackerDB, err := database.NewPostgresDB(&trackerDBConfig, logger)
	if err != nil {
		logger.LogWarn(context.Background(), "failed to connect to tracker database",
			sharedLogger.String("error", err.Error()))
		trackerDB = nil
	}

	walletDBConfig := cfg.Database
	walletDBConfig.DBName = "wallet_db"
	walletDB, err := database.NewPostgresDB(&walletDBConfig, logger)
	if err != nil {
		logger.LogWarn(context.Background(), "failed to connect to wallet database",
			sharedLogger.String("error", err.Error()))
		walletDB = nil
	}

	// Run database migrations
	if err := db.Migrate(
		&models.Report{},
		&models.ReportSchedule{},
		&models.ReportTemplate{},
		&models.ReportData{},
	); err != nil {
		logger.LogError(context.Background(), "failed to run migrations", err)
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Initialize repositories
	reportRepo := repository.NewReportRepository(db, logger)

	// Initialize services
	dataCollector := service.NewDatabaseDataCollector(
		calculatorDB,
		trackerDB,
		walletDB,
		logger,
	)

	reportRenderer := service.NewPDFReportRenderer(logger)

	reportingService := service.NewReportingService(
		reportRepo,
		dataCollector,
		reportRenderer,
		logger,
	)

	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware(cfg.Server.JWTSecret, logger)

	// Initialize handlers
	reportingHandler := handler.NewReportingHandler(reportingService, logger)

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
			"service": "reporting",
			"version": "1.0.0",
		})
	})

	// API routes
	v1 := router.Group("/api/v1")
	reportingHandler.RegisterRoutes(v1, authMiddleware)

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
		logger.LogInfo(context.Background(), "starting reporting service",
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

	logger.LogInfo(context.Background(), "shutting down reporting service")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.LogError(context.Background(), "server forced to shutdown", err)
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	logger.LogInfo(context.Background(), "reporting service stopped")
}
