package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/greenledger/services/tracker/internal/models"
	"github.com/greenledger/services/tracker/internal/service"
	"github.com/greenledger/shared/database"
	"github.com/greenledger/shared/logger"
	"gorm.io/gorm"
)

// ActivityRepository handles eco-activity data operations
type ActivityRepository struct {
	db     *database.PostgresDB
	logger *logger.Logger
}

// NewActivityRepository creates a new activity repository
func NewActivityRepository(db *database.PostgresDB, logger *logger.Logger) *ActivityRepository {
	return &ActivityRepository{
		db:     db,
		logger: logger,
	}
}

// Create creates a new eco-activity
func (r *ActivityRepository) Create(ctx context.Context, activity *models.EcoActivity) error {
	err := r.db.WithContext(ctx).Create(activity).Error
	if err != nil {
		r.logger.LogError(ctx, "failed to create activity", err,
			logger.String("user_id", activity.UserID))
		return fmt.Errorf("failed to create activity: %w", err)
	}

	r.logger.LogInfo(ctx, "activity created successfully",
		logger.String("activity_id", activity.ID.String()),
		logger.String("user_id", activity.UserID))

	return nil
}

// GetByID retrieves an activity by ID
func (r *ActivityRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.EcoActivity, error) {
	var activity models.EcoActivity
	
	err := r.db.WithContext(ctx).
		Preload("ActivityType").
		First(&activity, "id = ?", id).Error
	
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, database.ErrNotFound
		}
		r.logger.LogError(ctx, "failed to get activity by ID", err,
			logger.String("activity_id", id.String()))
		return nil, fmt.Errorf("failed to get activity: %w", err)
	}

	return &activity, nil
}

// GetByUserID retrieves activities for a specific user
func (r *ActivityRepository) GetByUserID(ctx context.Context, userID string, limit, offset int) ([]*models.EcoActivity, int64, error) {
	var activities []*models.EcoActivity
	var total int64

	// Get total count
	if err := r.db.WithContext(ctx).Model(&models.EcoActivity{}).
		Where("user_id = ?", userID).Count(&total).Error; err != nil {
		r.logger.LogError(ctx, "failed to count activities", err,
			logger.String("user_id", userID))
		return nil, 0, fmt.Errorf("failed to count activities: %w", err)
	}

	// Get activities with activity types
	err := r.db.WithContext(ctx).
		Preload("ActivityType").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&activities).Error

	if err != nil {
		r.logger.LogError(ctx, "failed to get activities by user ID", err,
			logger.String("user_id", userID))
		return nil, 0, fmt.Errorf("failed to get activities: %w", err)
	}

	return activities, total, nil
}

// GetByUserIDAndDateRange retrieves activities for a user within a date range
func (r *ActivityRepository) GetByUserIDAndDateRange(ctx context.Context, userID string, startDate, endDate time.Time, limit, offset int) ([]*models.EcoActivity, int64, error) {
	var activities []*models.EcoActivity
	var total int64

	query := r.db.WithContext(ctx).Model(&models.EcoActivity{}).
		Where("user_id = ? AND created_at >= ? AND created_at <= ?", userID, startDate, endDate)

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		r.logger.LogError(ctx, "failed to count activities in date range", err,
			logger.String("user_id", userID))
		return nil, 0, fmt.Errorf("failed to count activities: %w", err)
	}

	// Get activities
	err := r.db.WithContext(ctx).
		Preload("ActivityType").
		Where("user_id = ? AND created_at >= ? AND created_at <= ?", userID, startDate, endDate).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&activities).Error

	if err != nil {
		r.logger.LogError(ctx, "failed to get activities by date range", err,
			logger.String("user_id", userID))
		return nil, 0, fmt.Errorf("failed to get activities: %w", err)
	}

	return activities, total, nil
}

// Update updates an activity
func (r *ActivityRepository) Update(ctx context.Context, activity *models.EcoActivity) error {
	err := r.db.WithContext(ctx).Save(activity).Error
	if err != nil {
		r.logger.LogError(ctx, "failed to update activity", err,
			logger.String("activity_id", activity.ID.String()))
		return fmt.Errorf("failed to update activity: %w", err)
	}

	return nil
}

// Delete deletes an activity
func (r *ActivityRepository) Delete(ctx context.Context, id uuid.UUID) error {
	err := r.db.WithContext(ctx).Delete(&models.EcoActivity{}, "id = ?", id).Error
	if err != nil {
		r.logger.LogError(ctx, "failed to delete activity", err,
			logger.String("activity_id", id.String()))
		return fmt.Errorf("failed to delete activity: %w", err)
	}

	return nil
}

// GetUnverifiedActivities retrieves activities that require verification
func (r *ActivityRepository) GetUnverifiedActivities(ctx context.Context, limit, offset int) ([]*models.EcoActivity, int64, error) {
	var activities []*models.EcoActivity
	var total int64

	// Get total count
	if err := r.db.WithContext(ctx).Model(&models.EcoActivity{}).
		Where("is_verified = false").Count(&total).Error; err != nil {
		r.logger.LogError(ctx, "failed to count unverified activities", err)
		return nil, 0, fmt.Errorf("failed to count unverified activities: %w", err)
	}

	// Get activities
	err := r.db.WithContext(ctx).
		Preload("ActivityType").
		Where("is_verified = false").
		Order("created_at ASC").
		Limit(limit).
		Offset(offset).
		Find(&activities).Error

	if err != nil {
		r.logger.LogError(ctx, "failed to get unverified activities", err)
		return nil, 0, fmt.Errorf("failed to get unverified activities: %w", err)
	}

	return activities, total, nil
}

// GetUserStats retrieves activity statistics for a user
func (r *ActivityRepository) GetUserStats(ctx context.Context, userID string, startDate, endDate time.Time) (*service.UserActivityStats, error) {
	var result struct {
		TotalActivities    int64   `gorm:"column:total_activities"`
		TotalCreditsEarned float64 `gorm:"column:total_credits_earned"`
		TotalDuration      int     `gorm:"column:total_duration"`
		TotalDistance      float64 `gorm:"column:total_distance"`
	}

	err := r.db.WithContext(ctx).
		Model(&models.EcoActivity{}).
		Select(`
			COUNT(*) as total_activities,
			COALESCE(SUM(credits_earned), 0) as total_credits_earned,
			COALESCE(SUM(duration), 0) as total_duration,
			COALESCE(SUM(distance), 0) as total_distance
		`).
		Where("user_id = ? AND created_at >= ? AND created_at <= ?", userID, startDate, endDate).
		Scan(&result).Error

	if err != nil {
		r.logger.LogError(ctx, "failed to get user stats", err,
			logger.String("user_id", userID))
		return nil, fmt.Errorf("failed to get user stats: %w", err)
	}

	stats := &service.UserActivityStats{
		UserID:             userID,
		TotalActivities:    result.TotalActivities,
		TotalCreditsEarned: result.TotalCreditsEarned,
		TotalDuration:      result.TotalDuration,
		TotalDistance:      result.TotalDistance,
		StartDate:          startDate,
		EndDate:            endDate,
	}

	return stats, nil
}

// GetActivitiesByType retrieves activities by activity type
func (r *ActivityRepository) GetActivitiesByType(ctx context.Context, activityTypeID uuid.UUID, limit, offset int) ([]*models.EcoActivity, int64, error) {
	var activities []*models.EcoActivity
	var total int64

	// Get total count
	if err := r.db.WithContext(ctx).Model(&models.EcoActivity{}).
		Where("activity_type_id = ?", activityTypeID).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count activities by type: %w", err)
	}

	// Get activities
	err := r.db.WithContext(ctx).
		Preload("ActivityType").
		Where("activity_type_id = ?", activityTypeID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&activities).Error

	if err != nil {
		return nil, 0, fmt.Errorf("failed to get activities by type: %w", err)
	}

	return activities, total, nil
}

// GetRecentActivities retrieves recent activities across all users
func (r *ActivityRepository) GetRecentActivities(ctx context.Context, limit int) ([]*models.EcoActivity, error) {
	var activities []*models.EcoActivity

	err := r.db.WithContext(ctx).
		Preload("ActivityType").
		Where("is_verified = true").
		Order("created_at DESC").
		Limit(limit).
		Find(&activities).Error

	if err != nil {
		r.logger.LogError(ctx, "failed to get recent activities", err)
		return nil, fmt.Errorf("failed to get recent activities: %w", err)
	}

	return activities, nil
}
