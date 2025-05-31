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
	"github.com/shopspring/decimal"
	"github.com/sloweyyy/GreenLedger/services/wallet/internal/handler"
	"github.com/sloweyyy/GreenLedger/services/wallet/internal/models"
	"github.com/sloweyyy/GreenLedger/services/wallet/internal/repository"
	"github.com/sloweyyy/GreenLedger/services/wallet/internal/service"
	"github.com/sloweyyy/GreenLedger/shared/config"
	"github.com/sloweyyy/GreenLedger/shared/database"
	sharedLogger "github.com/sloweyyy/GreenLedger/shared/logger"
	"github.com/sloweyyy/GreenLedger/shared/middleware"
)

// @title GreenLedger Wallet Service API
// @version 1.0
// @description Carbon credit wallet management service for GreenLedger
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.greenledger.com/support
// @contact.email truonglevinhphuc2006@gmail.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8083
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

	// Override database name for wallet service
	cfg.Database.DBName = "wallet_db"
	cfg.Server.Port = 8083
	cfg.Server.GRPCPort = 9083

	// Initialize logger
	logger := sharedLogger.New(cfg.Server.LogLevel).WithService("wallet")

	// Initialize database
	db, err := database.NewPostgresDB(&cfg.Database, logger)
	if err != nil {
		logger.LogError(context.Background(), "failed to connect to database", err)
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Run database migrations
	if err := db.Migrate(
		&models.Wallet{},
		&models.Transaction{},
		&models.TransactionBatch{},
		&models.CreditReservation{},
		&models.WalletSnapshot{},
	); err != nil {
		logger.LogError(context.Background(), "failed to run migrations", err)
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Initialize repositories
	walletRepo := repository.NewWalletRepository(db, logger)
	transactionRepo := repository.NewTransactionRepository(db, logger)

	// Initialize event publisher
	var eventPublisher service.EventPublisher
	if cfg.Server.Environment == "production" {
		eventPublisher = service.NewKafkaEventPublisher(cfg.Kafka.Brokers, logger)
	} else {
		eventPublisher = service.NewMockEventPublisher(logger)
	}

	// Initialize services
	walletService := service.NewWalletService(
		walletRepo,
		transactionRepo,
		eventPublisher,
		logger,
	)

	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware(cfg.Server.JWTSecret, logger)

	// Initialize handlers
	walletHandler := handler.NewWalletHandler(walletService, logger)

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
			"service": "wallet",
			"version": "1.0.0",
		})
	})

	// API routes
	v1 := router.Group("/api/v1")
	walletHandler.RegisterRoutes(v1, authMiddleware)

	// Create HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start event consumer for credit earned events
	go func() {
		if cfg.Server.Environment == "production" {
			consumer := service.NewEventConsumer(cfg.Kafka.Brokers, "wallet-service", logger)
			defer consumer.Close()

			err := consumer.ConsumeEvents(context.Background(), func(ctx context.Context, event interface{}) error {
				switch e := event.(type) {
				case *service.CreditEarnedEvent:
					// Credit user's wallet when they earn credits from activities
					req := &service.CreditBalanceRequest{
						UserID:      e.UserID,
						Amount:      decimal.NewFromFloat(e.CreditsEarned),
						Source:      "eco_activity",
						Description: fmt.Sprintf("Credits earned from %s: %s", e.ActivityType, e.Description),
						ReferenceID: e.ActivityID,
					}

					_, err := walletService.CreditBalance(ctx, req)
					if err != nil {
						logger.LogError(ctx, "failed to credit wallet from activity", err,
							sharedLogger.String("user_id", e.UserID),
							sharedLogger.String("activity_id", e.ActivityID))
						return err
					}

					logger.LogInfo(ctx, "wallet credited from eco activity",
						sharedLogger.String("user_id", e.UserID),
						sharedLogger.String("activity_id", e.ActivityID),
						sharedLogger.Float64("credits", e.CreditsEarned))

					return nil
				default:
					logger.LogWarn(ctx, "unknown event type received")
					return nil
				}
			})

			if err != nil {
				logger.LogError(context.Background(), "event consumer error", err)
			}
		}
	}()

	// Start server in a goroutine
	go func() {
		logger.LogInfo(context.Background(), "starting wallet service",
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

	logger.LogInfo(context.Background(), "shutting down wallet service")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.LogError(context.Background(), "server forced to shutdown", err)
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	// Close event publisher
	if kafkaPublisher, ok := eventPublisher.(*service.KafkaEventPublisher); ok {
		kafkaPublisher.Close()
	}

	logger.LogInfo(context.Background(), "wallet service stopped")
}
