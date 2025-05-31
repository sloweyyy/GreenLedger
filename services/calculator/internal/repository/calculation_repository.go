package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/greenledger/services/calculator/internal/models"
	"github.com/greenledger/shared/database"
	"github.com/greenledger/shared/logger"
	"gorm.io/gorm"
)

// CalculationRepository handles calculation data operations
type CalculationRepository struct {
	db     *database.PostgresDB
	logger *logger.Logger
}

// NewCalculationRepository creates a new calculation repository
func NewCalculationRepository(db *database.PostgresDB, logger *logger.Logger) *CalculationRepository {
	return &CalculationRepository{
		db:     db,
		logger: logger,
	}
}

// Create creates a new calculation with activities
func (r *CalculationRepository) Create(ctx context.Context, calculation *models.Calculation) error {
	return r.db.WithTransaction(ctx, func(tx *gorm.DB) error {
		// Create calculation
		if err := tx.Create(calculation).Error; err != nil {
			r.logger.LogError(ctx, "failed to create calculation", err,
				logger.String("user_id", calculation.UserID))
			return fmt.Errorf("failed to create calculation: %w", err)
		}

		r.logger.LogInfo(ctx, "calculation created successfully",
			logger.String("calculation_id", calculation.ID.String()),
			logger.String("user_id", calculation.UserID),
			logger.Float64("total_co2_kg", calculation.TotalCO2Kg))

		return nil
	})
}

// GetByID retrieves a calculation by ID
func (r *CalculationRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Calculation, error) {
	var calculation models.Calculation
	
	err := r.db.WithContext(ctx).
		Preload("Activities").
		First(&calculation, "id = ?", id).Error
	
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, database.ErrNotFound
		}
		r.logger.LogError(ctx, "failed to get calculation by ID", err,
			logger.String("calculation_id", id.String()))
		return nil, fmt.Errorf("failed to get calculation: %w", err)
	}

	return &calculation, nil
}

// GetByUserID retrieves calculations for a specific user
func (r *CalculationRepository) GetByUserID(ctx context.Context, userID string, limit, offset int) ([]*models.Calculation, int64, error) {
	var calculations []*models.Calculation
	var total int64

	// Get total count
	if err := r.db.WithContext(ctx).Model(&models.Calculation{}).
		Where("user_id = ?", userID).Count(&total).Error; err != nil {
		r.logger.LogError(ctx, "failed to count calculations", err,
			logger.String("user_id", userID))
		return nil, 0, fmt.Errorf("failed to count calculations: %w", err)
	}

	// Get calculations with activities
	err := r.db.WithContext(ctx).
		Preload("Activities").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&calculations).Error

	if err != nil {
		r.logger.LogError(ctx, "failed to get calculations by user ID", err,
			logger.String("user_id", userID))
		return nil, 0, fmt.Errorf("failed to get calculations: %w", err)
	}

	return calculations, total, nil
}

// GetByUserIDAndDateRange retrieves calculations for a user within a date range
func (r *CalculationRepository) GetByUserIDAndDateRange(ctx context.Context, userID string, startDate, endDate time.Time, limit, offset int) ([]*models.Calculation, int64, error) {
	var calculations []*models.Calculation
	var total int64

	query := r.db.WithContext(ctx).Model(&models.Calculation{}).
		Where("user_id = ? AND created_at >= ? AND created_at <= ?", userID, startDate, endDate)

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		r.logger.LogError(ctx, "failed to count calculations in date range", err,
			logger.String("user_id", userID),
			logger.String("start_date", startDate.Format(time.RFC3339)),
			logger.String("end_date", endDate.Format(time.RFC3339)))
		return nil, 0, fmt.Errorf("failed to count calculations: %w", err)
	}

	// Get calculations with activities
	err := r.db.WithContext(ctx).
		Preload("Activities").
		Where("user_id = ? AND created_at >= ? AND created_at <= ?", userID, startDate, endDate).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&calculations).Error

	if err != nil {
		r.logger.LogError(ctx, "failed to get calculations by date range", err,
			logger.String("user_id", userID))
		return nil, 0, fmt.Errorf("failed to get calculations: %w", err)
	}

	return calculations, total, nil
}

// Update updates a calculation
func (r *CalculationRepository) Update(ctx context.Context, calculation *models.Calculation) error {
	err := r.db.WithContext(ctx).Save(calculation).Error
	if err != nil {
		r.logger.LogError(ctx, "failed to update calculation", err,
			logger.String("calculation_id", calculation.ID.String()))
		return fmt.Errorf("failed to update calculation: %w", err)
	}

	r.logger.LogInfo(ctx, "calculation updated successfully",
		logger.String("calculation_id", calculation.ID.String()))

	return nil
}

// Delete deletes a calculation and its activities
func (r *CalculationRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithTransaction(ctx, func(tx *gorm.DB) error {
		// Delete activities first
		if err := tx.Where("calculation_id = ?", id).Delete(&models.Activity{}).Error; err != nil {
			r.logger.LogError(ctx, "failed to delete activities", err,
				logger.String("calculation_id", id.String()))
			return fmt.Errorf("failed to delete activities: %w", err)
		}

		// Delete calculation
		if err := tx.Delete(&models.Calculation{}, "id = ?", id).Error; err != nil {
			r.logger.LogError(ctx, "failed to delete calculation", err,
				logger.String("calculation_id", id.String()))
			return fmt.Errorf("failed to delete calculation: %w", err)
		}

		r.logger.LogInfo(ctx, "calculation deleted successfully",
			logger.String("calculation_id", id.String()))

		return nil
	})
}

// GetUserStats retrieves calculation statistics for a user
func (r *CalculationRepository) GetUserStats(ctx context.Context, userID string, startDate, endDate time.Time) (*UserCalculationStats, error) {
	var stats UserCalculationStats
	
	// Get total calculations and CO2
	var result struct {
		TotalCalculations int64   `gorm:"column:total_calculations"`
		TotalCO2Kg        float64 `gorm:"column:total_co2_kg"`
		AvgCO2Kg          float64 `gorm:"column:avg_co2_kg"`
	}

	err := r.db.WithContext(ctx).
		Model(&models.Calculation{}).
		Select("COUNT(*) as total_calculations, COALESCE(SUM(total_co2_kg), 0) as total_co2_kg, COALESCE(AVG(total_co2_kg), 0) as avg_co2_kg").
		Where("user_id = ? AND created_at >= ? AND created_at <= ?", userID, startDate, endDate).
		Scan(&result).Error

	if err != nil {
		r.logger.LogError(ctx, "failed to get user stats", err,
			logger.String("user_id", userID))
		return nil, fmt.Errorf("failed to get user stats: %w", err)
	}

	stats.UserID = userID
	stats.TotalCalculations = result.TotalCalculations
	stats.TotalCO2Kg = result.TotalCO2Kg
	stats.AverageCO2Kg = result.AvgCO2Kg
	stats.StartDate = startDate
	stats.EndDate = endDate

	return &stats, nil
}

// UserCalculationStats represents calculation statistics for a user
type UserCalculationStats struct {
	UserID            string    `json:"user_id"`
	TotalCalculations int64     `json:"total_calculations"`
	TotalCO2Kg        float64   `json:"total_co2_kg"`
	AverageCO2Kg      float64   `json:"average_co2_kg"`
	StartDate         time.Time `json:"start_date"`
	EndDate           time.Time `json:"end_date"`
}
