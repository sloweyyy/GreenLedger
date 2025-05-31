package repository

import (
	"context"
	"fmt"

	"github.com/greenledger/services/calculator/internal/models"
	"github.com/greenledger/shared/database"
	"github.com/greenledger/shared/logger"
	"gorm.io/gorm"
)

// EmissionFactorRepository handles emission factor data operations
type EmissionFactorRepository struct {
	db     *database.PostgresDB
	logger *logger.Logger
}

// NewEmissionFactorRepository creates a new emission factor repository
func NewEmissionFactorRepository(db *database.PostgresDB, logger *logger.Logger) *EmissionFactorRepository {
	return &EmissionFactorRepository{
		db:     db,
		logger: logger,
	}
}

// GetByActivityType retrieves emission factors by activity type
func (r *EmissionFactorRepository) GetByActivityType(ctx context.Context, activityType string) ([]*models.EmissionFactor, error) {
	var factors []*models.EmissionFactor

	err := r.db.WithContext(ctx).
		Where("activity_type = ?", activityType).
		Order("sub_type").
		Find(&factors).Error

	if err != nil {
		r.logger.LogError(ctx, "failed to get emission factors by activity type", err,
			logger.String("activity_type", activityType))
		return nil, fmt.Errorf("failed to get emission factors: %w", err)
	}

	return factors, nil
}

// GetByActivityTypeAndSubType retrieves a specific emission factor
func (r *EmissionFactorRepository) GetByActivityTypeAndSubType(ctx context.Context, activityType, subType string) (*models.EmissionFactor, error) {
	var factor models.EmissionFactor

	err := r.db.WithContext(ctx).
		Where("activity_type = ? AND sub_type = ?", activityType, subType).
		First(&factor).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, database.ErrNotFound
		}
		r.logger.LogError(ctx, "failed to get emission factor", err,
			logger.String("activity_type", activityType),
			logger.String("sub_type", subType))
		return nil, fmt.Errorf("failed to get emission factor: %w", err)
	}

	return &factor, nil
}

// GetByActivityTypeAndLocation retrieves emission factors by activity type and location
func (r *EmissionFactorRepository) GetByActivityTypeAndLocation(ctx context.Context, activityType, location string) ([]*models.EmissionFactor, error) {
	var factors []*models.EmissionFactor

	query := r.db.WithContext(ctx).Where("activity_type = ?", activityType)
	
	if location != "" {
		// Try to find location-specific factors first, then fall back to global
		query = query.Where("location = ? OR location = '' OR location IS NULL", location).
			Order("CASE WHEN location = ? THEN 0 ELSE 1 END, sub_type", location)
	} else {
		query = query.Order("sub_type")
	}

	err := query.Find(&factors).Error
	if err != nil {
		r.logger.LogError(ctx, "failed to get emission factors by location", err,
			logger.String("activity_type", activityType),
			logger.String("location", location))
		return nil, fmt.Errorf("failed to get emission factors: %w", err)
	}

	return factors, nil
}

// Create creates a new emission factor
func (r *EmissionFactorRepository) Create(ctx context.Context, factor *models.EmissionFactor) error {
	err := r.db.WithContext(ctx).Create(factor).Error
	if err != nil {
		r.logger.LogError(ctx, "failed to create emission factor", err,
			logger.String("activity_type", factor.ActivityType),
			logger.String("sub_type", factor.SubType))
		return fmt.Errorf("failed to create emission factor: %w", err)
	}

	r.logger.LogInfo(ctx, "emission factor created successfully",
		logger.String("factor_id", factor.ID.String()),
		logger.String("activity_type", factor.ActivityType),
		logger.String("sub_type", factor.SubType))

	return nil
}

// Update updates an emission factor
func (r *EmissionFactorRepository) Update(ctx context.Context, factor *models.EmissionFactor) error {
	err := r.db.WithContext(ctx).Save(factor).Error
	if err != nil {
		r.logger.LogError(ctx, "failed to update emission factor", err,
			logger.String("factor_id", factor.ID.String()))
		return fmt.Errorf("failed to update emission factor: %w", err)
	}

	r.logger.LogInfo(ctx, "emission factor updated successfully",
		logger.String("factor_id", factor.ID.String()))

	return nil
}

// Delete deletes an emission factor
func (r *EmissionFactorRepository) Delete(ctx context.Context, id string) error {
	err := r.db.WithContext(ctx).Delete(&models.EmissionFactor{}, "id = ?", id).Error
	if err != nil {
		r.logger.LogError(ctx, "failed to delete emission factor", err,
			logger.String("factor_id", id))
		return fmt.Errorf("failed to delete emission factor: %w", err)
	}

	r.logger.LogInfo(ctx, "emission factor deleted successfully",
		logger.String("factor_id", id))

	return nil
}

// BulkCreate creates multiple emission factors
func (r *EmissionFactorRepository) BulkCreate(ctx context.Context, factors []*models.EmissionFactor) error {
	if len(factors) == 0 {
		return nil
	}

	err := r.db.WithContext(ctx).CreateInBatches(factors, 100).Error
	if err != nil {
		r.logger.LogError(ctx, "failed to bulk create emission factors", err,
			logger.Int("count", len(factors)))
		return fmt.Errorf("failed to bulk create emission factors: %w", err)
	}

	r.logger.LogInfo(ctx, "emission factors bulk created successfully",
		logger.Int("count", len(factors)))

	return nil
}

// GetAll retrieves all emission factors with optional filtering
func (r *EmissionFactorRepository) GetAll(ctx context.Context, activityType, location string, limit, offset int) ([]*models.EmissionFactor, int64, error) {
	var factors []*models.EmissionFactor
	var total int64

	query := r.db.WithContext(ctx).Model(&models.EmissionFactor{})

	// Apply filters
	if activityType != "" {
		query = query.Where("activity_type = ?", activityType)
	}
	if location != "" {
		query = query.Where("location = ? OR location = '' OR location IS NULL", location)
	}

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		r.logger.LogError(ctx, "failed to count emission factors", err)
		return nil, 0, fmt.Errorf("failed to count emission factors: %w", err)
	}

	// Get factors
	err := query.Order("activity_type, sub_type").
		Limit(limit).
		Offset(offset).
		Find(&factors).Error

	if err != nil {
		r.logger.LogError(ctx, "failed to get all emission factors", err)
		return nil, 0, fmt.Errorf("failed to get emission factors: %w", err)
	}

	return factors, total, nil
}
