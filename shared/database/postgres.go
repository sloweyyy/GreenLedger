package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/sloweyyy/GreenLedger/shared/config"
	"github.com/sloweyyy/GreenLedger/shared/logger"
	_ "github.com/lib/pq"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

// PostgresDB wraps gorm.DB with additional functionality
type PostgresDB struct {
	*gorm.DB
	config *config.DatabaseConfig
	logger *logger.Logger
}

// NewPostgresDB creates a new PostgreSQL database connection
func NewPostgresDB(cfg *config.DatabaseConfig, log *logger.Logger) (*PostgresDB, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s",
		cfg.Host, cfg.User, cfg.Password, cfg.DBName, cfg.Port, cfg.SSLMode)

	// Configure GORM logger
	gormLogLevel := gormLogger.Silent
	if log != nil {
		gormLogLevel = gormLogger.Info
	}

	gormConfig := &gorm.Config{
		Logger: gormLogger.Default.LogMode(gormLogLevel),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	}

	db, err := gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Get underlying sql.DB for connection pool configuration
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Configure connection pool
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &PostgresDB{
		DB:     db,
		config: cfg,
		logger: log,
	}, nil
}

// Close closes the database connection
func (p *PostgresDB) Close() error {
	sqlDB, err := p.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// Ping checks if the database connection is alive
func (p *PostgresDB) Ping(ctx context.Context) error {
	sqlDB, err := p.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.PingContext(ctx)
}

// WithTransaction executes a function within a database transaction
func (p *PostgresDB) WithTransaction(ctx context.Context, fn func(*gorm.DB) error) error {
	return p.DB.WithContext(ctx).Transaction(fn)
}

// GetStats returns database connection statistics
func (p *PostgresDB) GetStats() sql.DBStats {
	sqlDB, _ := p.DB.DB()
	return sqlDB.Stats()
}

// Migrate runs database migrations for the given models
func (p *PostgresDB) Migrate(models ...interface{}) error {
	return p.DB.AutoMigrate(models...)
}

// HealthCheck performs a health check on the database
func (p *PostgresDB) HealthCheck(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var result int
	err := p.DB.WithContext(ctx).Raw("SELECT 1").Scan(&result).Error
	if err != nil {
		return fmt.Errorf("database health check failed: %w", err)
	}

	if result != 1 {
		return fmt.Errorf("unexpected health check result: %d", result)
	}

	return nil
}

// BeginTx starts a new transaction with the given options
func (p *PostgresDB) BeginTx(ctx context.Context, opts *sql.TxOptions) *gorm.DB {
	return p.DB.WithContext(ctx).Begin(opts)
}

// Repository provides common database operations
type Repository struct {
	db     *PostgresDB
	logger *logger.Logger
}

// NewRepository creates a new repository instance
func NewRepository(db *PostgresDB, logger *logger.Logger) *Repository {
	return &Repository{
		db:     db,
		logger: logger,
	}
}

// Create creates a new record
func (r *Repository) Create(ctx context.Context, model interface{}) error {
	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		r.logger.LogError(ctx, "failed to create record", err,
			logger.String("model", fmt.Sprintf("%T", model)))
		return err
	}
	return nil
}

// GetByID retrieves a record by ID
func (r *Repository) GetByID(ctx context.Context, model interface{}, id interface{}) error {
	if err := r.db.WithContext(ctx).First(model, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return ErrNotFound
		}
		r.logger.LogError(ctx, "failed to get record by ID", err,
			logger.String("model", fmt.Sprintf("%T", model)),
			logger.Any("id", id))
		return err
	}
	return nil
}

// Update updates a record
func (r *Repository) Update(ctx context.Context, model interface{}) error {
	if err := r.db.WithContext(ctx).Save(model).Error; err != nil {
		r.logger.LogError(ctx, "failed to update record", err,
			logger.String("model", fmt.Sprintf("%T", model)))
		return err
	}
	return nil
}

// Delete deletes a record
func (r *Repository) Delete(ctx context.Context, model interface{}) error {
	if err := r.db.WithContext(ctx).Delete(model).Error; err != nil {
		r.logger.LogError(ctx, "failed to delete record", err,
			logger.String("model", fmt.Sprintf("%T", model)))
		return err
	}
	return nil
}

// Custom errors
var (
	ErrNotFound = fmt.Errorf("record not found")
)
