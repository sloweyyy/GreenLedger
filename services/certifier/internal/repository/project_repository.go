package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/sloweyyy/GreenLedger/services/certifier/internal/models"
	"github.com/sloweyyy/GreenLedger/shared/database"
	"github.com/sloweyyy/GreenLedger/shared/logger"
	"gorm.io/gorm"
)

// ProjectRepository handles certificate project data operations
type ProjectRepository struct {
	db     *database.PostgresDB
	logger *logger.Logger
}

// NewProjectRepository creates a new project repository
func NewProjectRepository(db *database.PostgresDB, logger *logger.Logger) *ProjectRepository {
	return &ProjectRepository{
		db:     db,
		logger: logger,
	}
}

// Create creates a new certificate project
func (r *ProjectRepository) Create(ctx context.Context, project *models.CertificateProject) error {
	if err := r.db.WithContext(ctx).Create(project).Error; err != nil {
		r.logger.LogError(ctx, "failed to create project", err,
			logger.String("project_name", project.Name))
		return fmt.Errorf("failed to create project: %w", err)
	}

	r.logger.LogInfo(ctx, "project created",
		logger.String("project_id", project.ID.String()),
		logger.String("project_name", project.Name))

	return nil
}

// GetByID retrieves a project by ID
func (r *ProjectRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.CertificateProject, error) {
	var project models.CertificateProject
	if err := r.db.WithContext(ctx).
		Preload("Certificates").
		First(&project, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("project not found")
		}
		r.logger.LogError(ctx, "failed to get project", err,
			logger.String("project_id", id.String()))
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	return &project, nil
}

// GetByName retrieves a project by name
func (r *ProjectRepository) GetByName(ctx context.Context, name string) (*models.CertificateProject, error) {
	var project models.CertificateProject
	if err := r.db.WithContext(ctx).
		First(&project, "name = ?", name).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("project not found")
		}
		r.logger.LogError(ctx, "failed to get project by name", err,
			logger.String("project_name", name))
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	return &project, nil
}

// GetAll retrieves all projects with pagination
func (r *ProjectRepository) GetAll(ctx context.Context, limit, offset int) ([]*models.CertificateProject, int64, error) {
	var projects []*models.CertificateProject
	var total int64

	// Get total count
	if err := r.db.WithContext(ctx).Model(&models.CertificateProject{}).
		Count(&total).Error; err != nil {
		r.logger.LogError(ctx, "failed to count projects", err)
		return nil, 0, fmt.Errorf("failed to count projects: %w", err)
	}

	// Get projects with pagination
	if err := r.db.WithContext(ctx).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&projects).Error; err != nil {
		r.logger.LogError(ctx, "failed to get projects", err)
		return nil, 0, fmt.Errorf("failed to get projects: %w", err)
	}

	return projects, total, nil
}

// GetActive retrieves active projects with pagination
func (r *ProjectRepository) GetActive(ctx context.Context, limit, offset int) ([]*models.CertificateProject, int64, error) {
	var projects []*models.CertificateProject
	var total int64

	// Get total count
	if err := r.db.WithContext(ctx).Model(&models.CertificateProject{}).
		Where("is_active = ?", true).
		Count(&total).Error; err != nil {
		r.logger.LogError(ctx, "failed to count active projects", err)
		return nil, 0, fmt.Errorf("failed to count active projects: %w", err)
	}

	// Get projects with pagination
	if err := r.db.WithContext(ctx).
		Where("is_active = ?", true).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&projects).Error; err != nil {
		r.logger.LogError(ctx, "failed to get active projects", err)
		return nil, 0, fmt.Errorf("failed to get active projects: %w", err)
	}

	return projects, total, nil
}

// GetByType retrieves projects by type with pagination
func (r *ProjectRepository) GetByType(ctx context.Context, projectType string, limit, offset int) ([]*models.CertificateProject, int64, error) {
	var projects []*models.CertificateProject
	var total int64

	// Get total count
	if err := r.db.WithContext(ctx).Model(&models.CertificateProject{}).
		Where("type = ?", projectType).
		Count(&total).Error; err != nil {
		r.logger.LogError(ctx, "failed to count projects by type", err,
			logger.String("project_type", projectType))
		return nil, 0, fmt.Errorf("failed to count projects: %w", err)
	}

	// Get projects with pagination
	if err := r.db.WithContext(ctx).
		Where("type = ?", projectType).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&projects).Error; err != nil {
		r.logger.LogError(ctx, "failed to get projects by type", err,
			logger.String("project_type", projectType))
		return nil, 0, fmt.Errorf("failed to get projects: %w", err)
	}

	return projects, total, nil
}

// Update updates a project
func (r *ProjectRepository) Update(ctx context.Context, project *models.CertificateProject) error {
	if err := r.db.WithContext(ctx).Save(project).Error; err != nil {
		r.logger.LogError(ctx, "failed to update project", err,
			logger.String("project_id", project.ID.String()))
		return fmt.Errorf("failed to update project: %w", err)
	}

	r.logger.LogInfo(ctx, "project updated",
		logger.String("project_id", project.ID.String()))

	return nil
}

// Delete deletes a project
func (r *ProjectRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Delete(&models.CertificateProject{}, "id = ?", id).Error; err != nil {
		r.logger.LogError(ctx, "failed to delete project", err,
			logger.String("project_id", id.String()))
		return fmt.Errorf("failed to delete project: %w", err)
	}

	r.logger.LogInfo(ctx, "project deleted",
		logger.String("project_id", id.String()))

	return nil
}

// UpdateAvailableCredits updates the available credits for a project
func (r *ProjectRepository) UpdateAvailableCredits(ctx context.Context, projectID uuid.UUID, creditsUsed float64) error {
	if err := r.db.WithContext(ctx).Model(&models.CertificateProject{}).
		Where("id = ?", projectID).
		Update("available_credits", gorm.Expr("available_credits - ?", creditsUsed)).Error; err != nil {
		r.logger.LogError(ctx, "failed to update available credits", err,
			logger.String("project_id", projectID.String()),
			logger.Float64("credits_used", creditsUsed))
		return fmt.Errorf("failed to update available credits: %w", err)
	}

	r.logger.LogInfo(ctx, "available credits updated",
		logger.String("project_id", projectID.String()),
		logger.Float64("credits_used", creditsUsed))

	return nil
}
