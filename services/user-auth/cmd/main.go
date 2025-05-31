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
	"github.com/greenledger/services/user-auth/internal/handler"
	"github.com/greenledger/services/user-auth/internal/models"
	"github.com/greenledger/services/user-auth/internal/repository"
	"github.com/greenledger/services/user-auth/internal/service"
	"github.com/greenledger/shared/config"
	"github.com/greenledger/shared/database"
	"github.com/greenledger/shared/logger"
	"github.com/greenledger/shared/middleware"
)

// @title GreenLedger User Authentication Service API
// @version 1.0
// @description User authentication and authorization service for GreenLedger
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.greenledger.com/support
// @contact.email truonglevinhphuc2006@gmail.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8084
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

	// Override database name for user-auth service
	cfg.Database.DBName = "userauth_db"
	cfg.Server.Port = 8084
	cfg.Server.GRPCPort = 9084

	// Initialize logger
	logger := logger.New(cfg.Server.LogLevel).WithService("user-auth")

	// Initialize database
	db, err := database.NewPostgresDB(&cfg.Database, logger)
	if err != nil {
		logger.LogError(context.Background(), "failed to connect to database", err)
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Run database migrations
	if err := db.Migrate(
		&models.User{},
		&models.Role{},
		&models.Permission{},
		&models.Session{},
		&models.UserProfile{},
		&models.PasswordResetToken{},
		&models.EmailVerificationToken{},
	); err != nil {
		logger.LogError(context.Background(), "failed to run migrations", err)
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Initialize repositories
	userRepo := repository.NewUserRepository(db, logger)
	sessionRepo := repository.NewSessionRepository(db, logger)
	roleRepo := repository.NewRoleRepository(db, logger)
	permissionRepo := repository.NewPermissionRepository(db, logger)

	// Initialize services
	authService := service.NewAuthService(userRepo, sessionRepo, roleRepo, cfg.Server.JWTSecret, logger)
	userService := service.NewUserService(userRepo, logger)

	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware(cfg.Server.JWTSecret, logger)

	// Initialize handlers
	authHandler := handler.NewAuthHandler(authService, userService, logger)

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
			"service": "user-auth",
			"version": "1.0.0",
		})
	})

	// API routes
	v1 := router.Group("/api/v1")
	authHandler.RegisterRoutes(v1, authMiddleware)

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
		logger.LogInfo(context.Background(), "starting user-auth service",
			logger.Int("port", cfg.Server.Port),
			logger.String("environment", cfg.Server.Environment))

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.LogError(context.Background(), "failed to start server", err)
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Initialize default roles and permissions
	go func() {
		if err := initializeRolesAndPermissions(context.Background(), roleRepo, permissionRepo, logger); err != nil {
			logger.LogError(context.Background(), "failed to initialize roles and permissions", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.LogInfo(context.Background(), "shutting down user-auth service")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.LogError(context.Background(), "server forced to shutdown", err)
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	logger.LogInfo(context.Background(), "user-auth service stopped")
}

// initializeRolesAndPermissions initializes default roles and permissions
func initializeRolesAndPermissions(ctx context.Context, roleRepo *repository.RoleRepository, permissionRepo *repository.PermissionRepository, logger *logger.Logger) error {
	// Check if roles already exist
	roles, err := roleRepo.List(ctx)
	if err != nil {
		return fmt.Errorf("failed to check existing roles: %w", err)
	}

	if len(roles) > 0 {
		logger.LogInfo(ctx, "roles and permissions already initialized")
		return nil
	}

	logger.LogInfo(ctx, "initializing default roles and permissions")

	// Create default permissions
	defaultPermissions := []*models.Permission{
		// User permissions
		{Name: "user:read", Resource: models.ResourceUser, Action: models.PermissionRead, Description: "Read user data"},
		{Name: "user:write", Resource: models.ResourceUser, Action: models.PermissionWrite, Description: "Write user data"},
		{Name: "user:delete", Resource: models.ResourceUser, Action: models.PermissionDelete, Description: "Delete user data"},
		{Name: "user:admin", Resource: models.ResourceUser, Action: models.PermissionAdmin, Description: "Admin user operations"},

		// Calculation permissions
		{Name: "calculation:read", Resource: models.ResourceCalculation, Action: models.PermissionRead, Description: "Read calculations"},
		{Name: "calculation:write", Resource: models.ResourceCalculation, Action: models.PermissionWrite, Description: "Create calculations"},
		{Name: "calculation:delete", Resource: models.ResourceCalculation, Action: models.PermissionDelete, Description: "Delete calculations"},

		// Wallet permissions
		{Name: "wallet:read", Resource: models.ResourceWallet, Action: models.PermissionRead, Description: "Read wallet data"},
		{Name: "wallet:write", Resource: models.ResourceWallet, Action: models.PermissionWrite, Description: "Manage wallet"},

		// Report permissions
		{Name: "report:read", Resource: models.ResourceReport, Action: models.PermissionRead, Description: "Read reports"},
		{Name: "report:write", Resource: models.ResourceReport, Action: models.PermissionWrite, Description: "Create reports"},

		// Certificate permissions
		{Name: "certificate:read", Resource: models.ResourceCertificate, Action: models.PermissionRead, Description: "Read certificates"},
		{Name: "certificate:write", Resource: models.ResourceCertificate, Action: models.PermissionWrite, Description: "Issue certificates"},
	}

	if err := permissionRepo.BulkCreate(ctx, defaultPermissions); err != nil {
		return fmt.Errorf("failed to create default permissions: %w", err)
	}

	// Create default roles
	adminRole := &models.Role{
		Name:        models.RoleAdmin,
		Description: "Administrator with full access",
	}

	userRole := &models.Role{
		Name:        models.RoleUser,
		Description: "Regular user with basic access",
	}

	moderatorRole := &models.Role{
		Name:        models.RoleModerator,
		Description: "Moderator with limited admin access",
	}

	// Create roles
	if err := roleRepo.Create(ctx, adminRole); err != nil {
		return fmt.Errorf("failed to create admin role: %w", err)
	}

	if err := roleRepo.Create(ctx, userRole); err != nil {
		return fmt.Errorf("failed to create user role: %w", err)
	}

	if err := roleRepo.Create(ctx, moderatorRole); err != nil {
		return fmt.Errorf("failed to create moderator role: %w", err)
	}

	// Assign permissions to admin role (all permissions)
	for _, permission := range defaultPermissions {
		if err := roleRepo.AssignPermission(ctx, adminRole.ID, permission.ID); err != nil {
			logger.LogError(ctx, "failed to assign permission to admin role", err,
				logger.String("permission", permission.Name))
		}
	}

	// Assign basic permissions to user role
	userPermissions := []string{
		"user:read", "user:write",
		"calculation:read", "calculation:write",
		"wallet:read", "wallet:write",
		"report:read",
		"certificate:read",
	}

	for _, permName := range userPermissions {
		for _, permission := range defaultPermissions {
			if permission.Name == permName {
				if err := roleRepo.AssignPermission(ctx, userRole.ID, permission.ID); err != nil {
					logger.LogError(ctx, "failed to assign permission to user role", err,
						logger.String("permission", permission.Name))
				}
				break
			}
		}
	}

	logger.LogInfo(ctx, "default roles and permissions initialized successfully",
		logger.Int("permissions_count", len(defaultPermissions)),
		logger.Int("roles_count", 3))

	return nil
}
