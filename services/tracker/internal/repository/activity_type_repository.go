package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/sloweyyy/GreenLedger/services/tracker/internal/models"
	"github.com/sloweyyy/GreenLedger/shared/database"
	"github.com/sloweyyy/GreenLedger/shared/logger"
	"gorm.io/gorm"
)

// ActivityTypeRepository handles activity type data operations
type ActivityTypeRepository struct {
	db     *database.PostgresDB
	logger *logger.Logger
}

// NewActivityTypeRepository creates a new activity type repository
func NewActivityTypeRepository(db *database.PostgresDB, logger *logger.Logger) *ActivityTypeRepository {
	return &ActivityTypeRepository{
		db:     db,
		logger: logger,
	}
}

// Create creates a new activity type
func (r *ActivityTypeRepository) Create(ctx context.Context, activityType *models.ActivityType) error {
	err := r.db.WithContext(ctx).Create(activityType).Error
	if err != nil {
		r.logger.LogError(ctx, "failed to create activity type", err,
			logger.String("name", activityType.Name))
		return fmt.Errorf("failed to create activity type: %w", err)
	}

	return nil
}

// GetByID retrieves an activity type by ID
func (r *ActivityTypeRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.ActivityType, error) {
	var activityType models.ActivityType
	
	err := r.db.WithContext(ctx).
		Preload("CreditRules").
		First(&activityType, "id = ?", id).Error
	
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, database.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get activity type: %w", err)
	}

	return &activityType, nil
}

// GetByName retrieves an activity type by name
func (r *ActivityTypeRepository) GetByName(ctx context.Context, name string) (*models.ActivityType, error) {
	var activityType models.ActivityType
	
	err := r.db.WithContext(ctx).
		Preload("CreditRules").
		First(&activityType, "name = ?", name).Error
	
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, database.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get activity type: %w", err)
	}

	return &activityType, nil
}

// GetAll retrieves all activity types
func (r *ActivityTypeRepository) GetAll(ctx context.Context) ([]*models.ActivityType, error) {
	var activityTypes []*models.ActivityType
	
	err := r.db.WithContext(ctx).
		Preload("CreditRules").
		Where("is_active = true").
		Order("category, name").
		Find(&activityTypes).Error
	
	if err != nil {
		return nil, fmt.Errorf("failed to get activity types: %w", err)
	}

	return activityTypes, nil
}

// GetByCategory retrieves activity types by category
func (r *ActivityTypeRepository) GetByCategory(ctx context.Context, category string) ([]*models.ActivityType, error) {
	var activityTypes []*models.ActivityType
	
	err := r.db.WithContext(ctx).
		Preload("CreditRules").
		Where("category = ? AND is_active = true", category).
		Order("name").
		Find(&activityTypes).Error
	
	if err != nil {
		return nil, fmt.Errorf("failed to get activity types by category: %w", err)
	}

	return activityTypes, nil
}

// Update updates an activity type
func (r *ActivityTypeRepository) Update(ctx context.Context, activityType *models.ActivityType) error {
	err := r.db.WithContext(ctx).Save(activityType).Error
	if err != nil {
		r.logger.LogError(ctx, "failed to update activity type", err,
			logger.String("id", activityType.ID.String()))
		return fmt.Errorf("failed to update activity type: %w", err)
	}

	return nil
}

// Delete deletes an activity type
func (r *ActivityTypeRepository) Delete(ctx context.Context, id uuid.UUID) error {
	err := r.db.WithContext(ctx).Delete(&models.ActivityType{}, "id = ?", id).Error
	if err != nil {
		r.logger.LogError(ctx, "failed to delete activity type", err,
			logger.String("id", id.String()))
		return fmt.Errorf("failed to delete activity type: %w", err)
	}

	return nil
}

// BulkCreate creates multiple activity types
func (r *ActivityTypeRepository) BulkCreate(ctx context.Context, activityTypes []*models.ActivityType) error {
	if len(activityTypes) == 0 {
		return nil
	}

	err := r.db.WithContext(ctx).CreateInBatches(activityTypes, 100).Error
	if err != nil {
		r.logger.LogError(ctx, "failed to bulk create activity types", err)
		return fmt.Errorf("failed to bulk create activity types: %w", err)
	}

	return nil
}

// CreditRuleRepository handles credit rule data operations
type CreditRuleRepository struct {
	db     *database.PostgresDB
	logger *logger.Logger
}

// NewCreditRuleRepository creates a new credit rule repository
func NewCreditRuleRepository(db *database.PostgresDB, logger *logger.Logger) *CreditRuleRepository {
	return &CreditRuleRepository{
		db:     db,
		logger: logger,
	}
}

// Create creates a new credit rule
func (r *CreditRuleRepository) Create(ctx context.Context, rule *models.CreditRule) error {
	err := r.db.WithContext(ctx).Create(rule).Error
	if err != nil {
		r.logger.LogError(ctx, "failed to create credit rule", err,
			logger.String("name", rule.Name))
		return fmt.Errorf("failed to create credit rule: %w", err)
	}

	return nil
}

// GetByID retrieves a credit rule by ID
func (r *CreditRuleRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.CreditRule, error) {
	var rule models.CreditRule
	
	err := r.db.WithContext(ctx).First(&rule, "id = ?", id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, database.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get credit rule: %w", err)
	}

	return &rule, nil
}

// GetActiveRulesByActivityType retrieves active credit rules for an activity type
func (r *CreditRuleRepository) GetActiveRulesByActivityType(ctx context.Context, activityTypeID uuid.UUID) ([]*models.CreditRule, error) {
	var rules []*models.CreditRule
	
	err := r.db.WithContext(ctx).
		Where("activity_type_id = ? AND is_active = true AND (valid_to IS NULL OR valid_to > NOW())", activityTypeID).
		Order("credits_per_unit DESC").
		Find(&rules).Error
	
	if err != nil {
		return nil, fmt.Errorf("failed to get credit rules: %w", err)
	}

	return rules, nil
}

// Update updates a credit rule
func (r *CreditRuleRepository) Update(ctx context.Context, rule *models.CreditRule) error {
	err := r.db.WithContext(ctx).Save(rule).Error
	if err != nil {
		return fmt.Errorf("failed to update credit rule: %w", err)
	}

	return nil
}

// Delete deletes a credit rule
func (r *CreditRuleRepository) Delete(ctx context.Context, id uuid.UUID) error {
	err := r.db.WithContext(ctx).Delete(&models.CreditRule{}, "id = ?", id).Error
	if err != nil {
		return fmt.Errorf("failed to delete credit rule: %w", err)
	}

	return nil
}
