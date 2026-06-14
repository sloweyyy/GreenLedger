// Package database provides PostgreSQL connection, migration and health-check helpers built on GORM.
package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/glebarez/sqlite"
	_ "github.com/lib/pq" // registers the PostgreSQL database/sql driver
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"

	"github.com/sloweyyy/GreenLedger/shared/config"
	"github.com/sloweyyy/GreenLedger/shared/logger"
)

// PostgresDB wraps gorm.DB with additional functionality
// Note: Name kept as PostgresDB to avoid breaking changes in all services,
// but it now supports multiple database drivers.
type PostgresDB struct {
	*gorm.DB
	config *config.DatabaseConfig
	logger *logger.Logger
}

// NewPostgresDB creates a new database connection (PostgreSQL or SQLite)
func NewPostgresDB(cfg *config.DatabaseConfig, log *logger.Logger) (*PostgresDB, error) {
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

	var dialector gorm.Dialector

	if cfg.Type == "sqlite" {
		// Use the configured path. If DBName is set (like in service specific configs),
		// we might want to append it, but the DB_PATH usually points to a directory
		// in the setup script: "export DB_PATH=../../data/${service_name}.db"
		// The config.LoadConfig() defaults path to "./data/greenledger.db".
		// Services override this via env vars.

		// If Path ends in .db, use it directly. Otherwise treat as directory and append DBName.
		dbPath := cfg.Path
		if dbPath == "" {
			dbPath = cfg.DBName + ".db"
		}

		dialector = sqlite.Open(dbPath)
		if log != nil {
			log.LogInfo(context.Background(), fmt.Sprintf("Using SQLite database at: %s", dbPath))
		}
	} else {
		// Default to PostgreSQL
		dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s",
			cfg.Host, cfg.User, cfg.Password, cfg.DBName, cfg.Port, cfg.SSLMode)
		dialector = postgres.Open(dsn)
	}

	db, err := gorm.Open(dialector, gormConfig)
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
	return p.AutoMigrate(models...)
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
