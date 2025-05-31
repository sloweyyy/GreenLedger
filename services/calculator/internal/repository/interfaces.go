package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/sloweyyy/GreenLedger/services/calculator/internal/models"
)

// CalculationRepositoryInterface defines the interface for calculation repository
type CalculationRepositoryInterface interface {
	Create(ctx context.Context, calculation *models.Calculation) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Calculation, error)
	GetByUserID(ctx context.Context, userID string, limit, offset int) ([]*models.Calculation, int64, error)
	GetByUserIDAndDateRange(ctx context.Context, userID string, startDate, endDate time.Time, limit, offset int) ([]*models.Calculation, int64, error)
	Update(ctx context.Context, calculation *models.Calculation) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetUserStats(ctx context.Context, userID string, startDate, endDate time.Time) (*UserCalculationStats, error)
}

// EmissionFactorRepositoryInterface defines the interface for emission factor repository
type EmissionFactorRepositoryInterface interface {
	GetByActivityType(ctx context.Context, activityType string) ([]*models.EmissionFactor, error)
	GetByActivityTypeAndSubType(ctx context.Context, activityType, subType string) (*models.EmissionFactor, error)
	GetByActivityTypeAndLocation(ctx context.Context, activityType, location string) ([]*models.EmissionFactor, error)
	Create(ctx context.Context, factor *models.EmissionFactor) error
	Update(ctx context.Context, factor *models.EmissionFactor) error
	Delete(ctx context.Context, id string) error
	BulkCreate(ctx context.Context, factors []*models.EmissionFactor) error
	GetAll(ctx context.Context, activityType, location string, limit, offset int) ([]*models.EmissionFactor, int64, error)
}

// Ensure concrete types implement interfaces
var _ CalculationRepositoryInterface = (*CalculationRepository)(nil)
var _ EmissionFactorRepositoryInterface = (*EmissionFactorRepository)(nil)
