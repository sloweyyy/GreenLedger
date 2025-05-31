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
	"github.com/sloweyyy/GreenLedger/services/tracker/internal/handler"
	"github.com/sloweyyy/GreenLedger/services/tracker/internal/models"
	"github.com/sloweyyy/GreenLedger/services/tracker/internal/repository"
	"github.com/sloweyyy/GreenLedger/services/tracker/internal/service"
	"github.com/sloweyyy/GreenLedger/shared/config"
	"github.com/sloweyyy/GreenLedger/shared/database"
	"github.com/sloweyyy/GreenLedger/shared/logger"
	"github.com/sloweyyy/GreenLedger/shared/middleware"
)

// @title GreenLedger Activity Tracker Service API
// @version 1.0
// @description Activity tracking service for eco-friendly activities in GreenLedger
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.greenledger.com/support
// @contact.email truonglevinhphuc2006@gmail.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8082
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

	// Override database name for tracker service
	cfg.Database.DBName = "tracker_db"
	cfg.Server.Port = 8082
	cfg.Server.GRPCPort = 9082

	// Initialize logger
	logger := logger.New(cfg.Server.LogLevel).WithService("tracker")

	// Initialize database
	db, err := database.NewPostgresDB(&cfg.Database, logger)
	if err != nil {
		logger.LogError(context.Background(), "failed to connect to database", err)
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Run database migrations
	if err := db.Migrate(
		&models.EcoActivity{},
		&models.ActivityType{},
		&models.CreditRule{},
		&models.ActivityChallenge{},
		&models.ChallengeParticipant{},
		&models.IoTDevice{},
	); err != nil {
		logger.LogError(context.Background(), "failed to run migrations", err)
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Initialize repositories
	activityRepo := repository.NewActivityRepository(db, logger)
	activityTypeRepo := repository.NewActivityTypeRepository(db, logger)
	creditRuleRepo := repository.NewCreditRuleRepository(db, logger)

	// Initialize event publisher
	var eventPublisher service.EventPublisher
	if cfg.Server.Environment == "production" {
		eventPublisher = service.NewKafkaEventPublisher(cfg.Kafka.Brokers, logger)
	} else {
		eventPublisher = service.NewMockEventPublisher(logger)
	}

	// Initialize services
	trackerService := service.NewTrackerService(
		activityRepo,
		activityTypeRepo,
		creditRuleRepo,
		eventPublisher,
		logger,
	)

	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware(cfg.Server.JWTSecret, logger)

	// Initialize handlers
	trackerHandler := handler.NewTrackerHandler(trackerService, logger)

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
			"service": "tracker",
			"version": "1.0.0",
		})
	})

	// API routes
	v1 := router.Group("/api/v1")
	trackerHandler.RegisterRoutes(v1, authMiddleware)

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
		logger.LogInfo(context.Background(), "starting tracker service",
			logger.Int("port", cfg.Server.Port),
			logger.String("environment", cfg.Server.Environment))

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.LogError(context.Background(), "failed to start server", err)
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Initialize default activity types
	go func() {
		if err := initializeActivityTypes(context.Background(), activityTypeRepo, creditRuleRepo, logger); err != nil {
			logger.LogError(context.Background(), "failed to initialize activity types", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.LogInfo(context.Background(), "shutting down tracker service")

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

	logger.LogInfo(context.Background(), "tracker service stopped")
}

// initializeActivityTypes initializes default activity types and credit rules
func initializeActivityTypes(ctx context.Context, activityTypeRepo *repository.ActivityTypeRepository, creditRuleRepo *repository.CreditRuleRepository, logger *logger.Logger) error {
	// Check if activity types already exist
	activityTypes, err := activityTypeRepo.GetAll(ctx)
	if err != nil {
		return fmt.Errorf("failed to check existing activity types: %w", err)
	}

	if len(activityTypes) > 0 {
		logger.LogInfo(ctx, "activity types already initialized")
		return nil
	}

	logger.LogInfo(ctx, "initializing default activity types")

	defaultActivityTypes := []*models.ActivityType{
		// Transport activities
		{
			Name:                 models.ActivityBiking,
			Category:             models.CategoryTransport,
			Description:          "Cycling instead of using motorized transport",
			Icon:                 "🚴",
			BaseCreditsPerUnit:   0.5,
			Unit:                 "km",
			IsActive:             true,
			RequiresVerification: false,
		},
		{
			Name:                 models.ActivityWalking,
			Category:             models.CategoryTransport,
			Description:          "Walking instead of using motorized transport",
			Icon:                 "🚶",
			BaseCreditsPerUnit:   0.3,
			Unit:                 "km",
			IsActive:             true,
			RequiresVerification: false,
		},
		{
			Name:                 models.ActivityPublicTransit,
			Category:             models.CategoryTransport,
			Description:          "Using public transportation",
			Icon:                 "🚌",
			BaseCreditsPerUnit:   0.2,
			Unit:                 "km",
			IsActive:             true,
			RequiresVerification: false,
		},
		{
			Name:                 models.ActivityCarPooling,
			Category:             models.CategoryTransport,
			Description:          "Sharing rides with others",
			Icon:                 "🚗",
			BaseCreditsPerUnit:   0.1,
			Unit:                 "km",
			IsActive:             true,
			RequiresVerification: false,
		},

		// Energy activities
		{
			Name:                 models.ActivitySolarEnergy,
			Category:             models.CategoryEnergy,
			Description:          "Using solar energy",
			Icon:                 "☀️",
			BaseCreditsPerUnit:   1.0,
			Unit:                 "kWh",
			IsActive:             true,
			RequiresVerification: true,
		},

		// Waste activities
		{
			Name:                 models.ActivityRecycling,
			Category:             models.CategoryWaste,
			Description:          "Recycling materials",
			Icon:                 "♻️",
			BaseCreditsPerUnit:   0.1,
			Unit:                 "kg",
			IsActive:             true,
			RequiresVerification: false,
		},
		{
			Name:                 models.ActivityComposting,
			Category:             models.CategoryWaste,
			Description:          "Composting organic waste",
			Icon:                 "🌱",
			BaseCreditsPerUnit:   0.2,
			Unit:                 "kg",
			IsActive:             true,
			RequiresVerification: false,
		},

		// Nature activities
		{
			Name:                 models.ActivityTreePlanting,
			Category:             models.CategoryNature,
			Description:          "Planting trees",
			Icon:                 "🌳",
			BaseCreditsPerUnit:   5.0,
			Unit:                 "units",
			IsActive:             true,
			RequiresVerification: true,
		},

		// Consumption activities
		{
			Name:                 models.ActivityLocalShopping,
			Category:             models.CategoryConsumption,
			Description:          "Shopping locally to reduce transport emissions",
			Icon:                 "🛒",
			BaseCreditsPerUnit:   1.0,
			Unit:                 "units",
			IsActive:             true,
			RequiresVerification: false,
		},
		{
			Name:                 models.ActivityVegetarianMeal,
			Category:             models.CategoryConsumption,
			Description:          "Eating vegetarian meals",
			Icon:                 "🥗",
			BaseCreditsPerUnit:   0.5,
			Unit:                 "units",
			IsActive:             true,
			RequiresVerification: false,
		},
	}

	if err := activityTypeRepo.BulkCreate(ctx, defaultActivityTypes); err != nil {
		return fmt.Errorf("failed to create default activity types: %w", err)
	}

	logger.LogInfo(ctx, "default activity types initialized successfully",
		logger.Int("count", len(defaultActivityTypes)))

	return nil
}
