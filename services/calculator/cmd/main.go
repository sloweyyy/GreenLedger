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
	"github.com/sloweyyy/GreenLedger/services/calculator/internal/handler"
	"github.com/sloweyyy/GreenLedger/services/calculator/internal/models"
	"github.com/sloweyyy/GreenLedger/services/calculator/internal/repository"
	"github.com/sloweyyy/GreenLedger/services/calculator/internal/service"
	"github.com/sloweyyy/GreenLedger/shared/config"
	"github.com/sloweyyy/GreenLedger/shared/database"
	sharedLogger "github.com/sloweyyy/GreenLedger/shared/logger"
	"github.com/sloweyyy/GreenLedger/shared/middleware"
)

// @title GreenLedger Calculator Service API
// @version 1.0
// @description Carbon footprint calculation service for GreenLedger
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.greenledger.com/support
// @contact.email truonglevinhphuc2006@gmail.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8081
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

	// Override database name for calculator service
	cfg.Database.DBName = "calculator_db"
	cfg.Server.Port = 8081
	cfg.Server.GRPCPort = 9081

	// Initialize logger
	logger := sharedLogger.New(cfg.Server.LogLevel).WithService("calculator")

	// Initialize database
	db, err := database.NewPostgresDB(&cfg.Database, logger)
	if err != nil {
		logger.LogError(context.Background(), "failed to connect to database", err)
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Run database migrations
	if err := db.Migrate(&models.Calculation{}, &models.Activity{}, &models.EmissionFactor{}); err != nil {
		logger.LogError(context.Background(), "failed to run migrations", err)
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Initialize repositories
	calculationRepo := repository.NewCalculationRepository(db, logger)
	emissionFactorRepo := repository.NewEmissionFactorRepository(db, logger)

	// Initialize services
	calculatorService := service.NewCalculatorService(calculationRepo, emissionFactorRepo, logger)

	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware(cfg.Server.JWTSecret, logger)

	// Initialize handlers
	calculatorHandler := handler.NewCalculatorHandler(calculatorService, logger)

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
			"service": "calculator",
			"version": "1.0.0",
		})
	})

	// API routes
	v1 := router.Group("/api/v1")
	calculatorHandler.RegisterRoutes(v1, authMiddleware)

	// Swagger documentation
	// router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

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
		logger.LogInfo(context.Background(), "starting calculator service",
			sharedLogger.Int("port", cfg.Server.Port),
			sharedLogger.String("environment", cfg.Server.Environment))

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.LogError(context.Background(), "failed to start server", err)
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Initialize default emission factors
	go func() {
		if err := initializeEmissionFactors(context.Background(), emissionFactorRepo, logger); err != nil {
			logger.LogError(context.Background(), "failed to initialize emission factors", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.LogInfo(context.Background(), "shutting down calculator service")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.LogError(context.Background(), "server forced to shutdown", err)
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	logger.LogInfo(context.Background(), "calculator service stopped")
}

// initializeEmissionFactors initializes default emission factors
func initializeEmissionFactors(ctx context.Context, repo *repository.EmissionFactorRepository, logger *sharedLogger.Logger) error {
	// Check if emission factors already exist
	factors, _, err := repo.GetAll(ctx, "", "", 1, 0)
	if err != nil {
		return fmt.Errorf("failed to check existing emission factors: %w", err)
	}

	if len(factors) > 0 {
		logger.LogInfo(ctx, "emission factors already initialized")
		return nil
	}

	logger.LogInfo(ctx, "initializing default emission factors")

	defaultFactors := []*models.EmissionFactor{
		// Vehicle travel factors (kg CO2 per km)
		{ActivityType: models.ActivityTypeVehicleTravel, SubType: models.VehicleTypeCarGasoline, FactorCO2: 0.21, Unit: "km", Source: "EPA 2023", Location: ""},
		{ActivityType: models.ActivityTypeVehicleTravel, SubType: models.VehicleTypeCarDiesel, FactorCO2: 0.17, Unit: "km", Source: "EPA 2023", Location: ""},
		{ActivityType: models.ActivityTypeVehicleTravel, SubType: models.VehicleTypeCarElectric, FactorCO2: 0.05, Unit: "km", Source: "EPA 2023", Location: ""},
		{ActivityType: models.ActivityTypeVehicleTravel, SubType: models.VehicleTypeCarHybrid, FactorCO2: 0.12, Unit: "km", Source: "EPA 2023", Location: ""},
		{ActivityType: models.ActivityTypeVehicleTravel, SubType: models.VehicleTypeMotorcycle, FactorCO2: 0.14, Unit: "km", Source: "EPA 2023", Location: ""},
		{ActivityType: models.ActivityTypeVehicleTravel, SubType: models.VehicleTypeBus, FactorCO2: 0.08, Unit: "km", Source: "EPA 2023", Location: ""},
		{ActivityType: models.ActivityTypeVehicleTravel, SubType: models.VehicleTypeTrain, FactorCO2: 0.04, Unit: "km", Source: "EPA 2023", Location: ""},

		// Electricity factors (kg CO2 per kWh) - varies by location
		{ActivityType: models.ActivityTypeElectricity, SubType: "grid", FactorCO2: 0.5, Unit: "kWh", Source: "IEA 2023", Location: "US"},
		{ActivityType: models.ActivityTypeElectricity, SubType: "grid", FactorCO2: 0.3, Unit: "kWh", Source: "IEA 2023", Location: "EU"},
		{ActivityType: models.ActivityTypeElectricity, SubType: "grid", FactorCO2: 0.7, Unit: "kWh", Source: "IEA 2023", Location: "CN"},
		{ActivityType: models.ActivityTypeElectricity, SubType: "grid", FactorCO2: 0.45, Unit: "kWh", Source: "IEA 2023", Location: ""}, // Global average

		// Purchase factors (kg CO2 per USD)
		{ActivityType: models.ActivityTypePurchase, SubType: models.PurchaseCategoryFood, FactorCO2: 0.5, Unit: "USD", Source: "DEFRA 2023", Location: ""},
		{ActivityType: models.ActivityTypePurchase, SubType: models.PurchaseCategoryClothing, FactorCO2: 0.8, Unit: "USD", Source: "DEFRA 2023", Location: ""},
		{ActivityType: models.ActivityTypePurchase, SubType: models.PurchaseCategoryElectronics, FactorCO2: 0.3, Unit: "USD", Source: "DEFRA 2023", Location: ""},
		{ActivityType: models.ActivityTypePurchase, SubType: models.PurchaseCategoryFurniture, FactorCO2: 0.4, Unit: "USD", Source: "DEFRA 2023", Location: ""},
		{ActivityType: models.ActivityTypePurchase, SubType: models.PurchaseCategoryOther, FactorCO2: 0.35, Unit: "USD", Source: "DEFRA 2023", Location: ""},

		// Flight factors (kg CO2 per km per passenger)
		{ActivityType: models.ActivityTypeFlight, SubType: models.FlightClassEconomy, FactorCO2: 0.15, Unit: "km", Source: "ICAO 2023", Location: ""},
		{ActivityType: models.ActivityTypeFlight, SubType: models.FlightClassBusiness, FactorCO2: 0.25, Unit: "km", Source: "ICAO 2023", Location: ""},
		{ActivityType: models.ActivityTypeFlight, SubType: models.FlightClassFirst, FactorCO2: 0.35, Unit: "km", Source: "ICAO 2023", Location: ""},

		// Heating factors (kg CO2 per unit)
		{ActivityType: models.ActivityTypeHeating, SubType: models.HeatingFuelNaturalGas, FactorCO2: 2.0, Unit: "m3", Source: "EPA 2023", Location: ""},
		{ActivityType: models.ActivityTypeHeating, SubType: models.HeatingFuelOil, FactorCO2: 2.7, Unit: "L", Source: "EPA 2023", Location: ""},
		{ActivityType: models.ActivityTypeHeating, SubType: models.HeatingFuelElectric, FactorCO2: 0.5, Unit: "kWh", Source: "EPA 2023", Location: ""},
		{ActivityType: models.ActivityTypeHeating, SubType: models.HeatingFuelPropane, FactorCO2: 1.5, Unit: "L", Source: "EPA 2023", Location: ""},
	}

	// Set timestamps
	now := time.Now().UTC()
	for _, factor := range defaultFactors {
		factor.LastUpdated = now
	}

	if err := repo.BulkCreate(ctx, defaultFactors); err != nil {
		return fmt.Errorf("failed to create default emission factors: %w", err)
	}

	logger.LogInfo(ctx, "default emission factors initialized successfully",
		sharedLogger.Int("count", len(defaultFactors)))

	return nil
}
