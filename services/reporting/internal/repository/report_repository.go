package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/sloweyyy/GreenLedger/services/reporting/internal/models"
	"github.com/sloweyyy/GreenLedger/shared/database"
	"github.com/sloweyyy/GreenLedger/shared/logger"
	"gorm.io/gorm"
)

// ReportRepository handles report data operations
type ReportRepository struct {
	db     *database.PostgresDB
	logger *logger.Logger
}

// NewReportRepository creates a new report repository
func NewReportRepository(db *database.PostgresDB, logger *logger.Logger) *ReportRepository {
	return &ReportRepository{
		db:     db,
		logger: logger,
	}
}

// Create creates a new report
func (r *ReportRepository) Create(ctx context.Context, report *models.Report) error {
	err := r.db.DB.WithContext(ctx).Create(report).Error
	if err != nil {
		r.logger.LogError(ctx, "failed to create report", err,
			logger.String("user_id", report.UserID),
			logger.String("type", report.Type))
		return fmt.Errorf("failed to create report: %w", err)
	}

	r.logger.LogInfo(ctx, "report created",
		logger.String("report_id", report.ID.String()),
		logger.String("user_id", report.UserID))

	return nil
}

// GetByID retrieves a report by ID
func (r *ReportRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Report, error) {
	var report models.Report

	err := r.db.DB.WithContext(ctx).First(&report, "id = ?", id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("report not found")
		}
		r.logger.LogError(ctx, "failed to get report by ID", err,
			logger.String("report_id", id.String()))
		return nil, fmt.Errorf("failed to get report: %w", err)
	}

	return &report, nil
}

// GetByUserID retrieves reports for a user with pagination
func (r *ReportRepository) GetByUserID(ctx context.Context, userID string, limit, offset int) ([]*models.Report, int64, error) {
	var reports []*models.Report
	var total int64

	// Get total count
	if err := r.db.DB.WithContext(ctx).Model(&models.Report{}).
		Where("user_id = ?", userID).Count(&total).Error; err != nil {
		r.logger.LogError(ctx, "failed to count user reports", err,
			logger.String("user_id", userID))
		return nil, 0, fmt.Errorf("failed to count reports: %w", err)
	}

	// Get reports
	err := r.db.DB.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&reports).Error

	if err != nil {
		r.logger.LogError(ctx, "failed to get user reports", err,
			logger.String("user_id", userID))
		return nil, 0, fmt.Errorf("failed to get reports: %w", err)
	}

	return reports, total, nil
}

// Update updates a report
func (r *ReportRepository) Update(ctx context.Context, report *models.Report) error {
	err := r.db.DB.WithContext(ctx).Save(report).Error
	if err != nil {
		r.logger.LogError(ctx, "failed to update report", err,
			logger.String("report_id", report.ID.String()))
		return fmt.Errorf("failed to update report: %w", err)
	}

	return nil
}

// Delete deletes a report
func (r *ReportRepository) Delete(ctx context.Context, id uuid.UUID) error {
	err := r.db.DB.WithContext(ctx).Delete(&models.Report{}, "id = ?", id).Error
	if err != nil {
		r.logger.LogError(ctx, "failed to delete report", err,
			logger.String("report_id", id.String()))
		return fmt.Errorf("failed to delete report: %w", err)
	}

	return nil
}

// GetExpiredReports retrieves reports that have expired
func (r *ReportRepository) GetExpiredReports(ctx context.Context, limit int) ([]*models.Report, error) {
	var reports []*models.Report

	err := r.db.DB.WithContext(ctx).
		Where("expires_at < NOW() AND status = ?", models.ReportStatusCompleted).
		Order("expires_at ASC").
		Limit(limit).
		Find(&reports).Error

	if err != nil {
		r.logger.LogError(ctx, "failed to get expired reports", err)
		return nil, fmt.Errorf("failed to get expired reports: %w", err)
	}

	return reports, nil
}

// GetReportsByStatus retrieves reports by status
func (r *ReportRepository) GetReportsByStatus(ctx context.Context, status string, limit, offset int) ([]*models.Report, int64, error) {
	var reports []*models.Report
	var total int64

	// Get total count
	if err := r.db.DB.WithContext(ctx).Model(&models.Report{}).
		Where("status = ?", status).Count(&total).Error; err != nil {
		r.logger.LogError(ctx, "failed to count reports by status", err,
			logger.String("status", status))
		return nil, 0, fmt.Errorf("failed to count reports: %w", err)
	}

	// Get reports
	err := r.db.DB.WithContext(ctx).
		Where("status = ?", status).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&reports).Error

	if err != nil {
		r.logger.LogError(ctx, "failed to get reports by status", err,
			logger.String("status", status))
		return nil, 0, fmt.Errorf("failed to get reports: %w", err)
	}

	return reports, total, nil
}

// GetReportsByType retrieves reports by type for a user
func (r *ReportRepository) GetReportsByType(ctx context.Context, userID, reportType string, limit, offset int) ([]*models.Report, int64, error) {
	var reports []*models.Report
	var total int64

	// Get total count
	if err := r.db.DB.WithContext(ctx).Model(&models.Report{}).
		Where("user_id = ? AND type = ?", userID, reportType).Count(&total).Error; err != nil {
		r.logger.LogError(ctx, "failed to count reports by type", err,
			logger.String("user_id", userID),
			logger.String("type", reportType))
		return nil, 0, fmt.Errorf("failed to count reports: %w", err)
	}

	// Get reports
	err := r.db.DB.WithContext(ctx).
		Where("user_id = ? AND type = ?", userID, reportType).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&reports).Error

	if err != nil {
		r.logger.LogError(ctx, "failed to get reports by type", err,
			logger.String("user_id", userID),
			logger.String("type", reportType))
		return nil, 0, fmt.Errorf("failed to get reports: %w", err)
	}

	return reports, total, nil
}

// GetUserReportStats retrieves report statistics for a user
func (r *ReportRepository) GetUserReportStats(ctx context.Context, userID string) (*models.ReportStats, error) {
	var stats models.ReportStats

	// Get total reports count
	if err := r.db.DB.WithContext(ctx).Model(&models.Report{}).
		Where("user_id = ?", userID).Count(&stats.TotalReports).Error; err != nil {
		return nil, fmt.Errorf("failed to count total reports: %w", err)
	}

	// Get completed reports count
	if err := r.db.DB.WithContext(ctx).Model(&models.Report{}).
		Where("user_id = ? AND status = ?", userID, models.ReportStatusCompleted).
		Count(&stats.CompletedReports).Error; err != nil {
		return nil, fmt.Errorf("failed to count completed reports: %w", err)
	}

	// Get pending reports count
	if err := r.db.DB.WithContext(ctx).Model(&models.Report{}).
		Where("user_id = ? AND status IN (?)", userID, []string{models.ReportStatusPending, models.ReportStatusGenerating}).
		Count(&stats.PendingReports).Error; err != nil {
		return nil, fmt.Errorf("failed to count pending reports: %w", err)
	}

	// Get failed reports count
	if err := r.db.DB.WithContext(ctx).Model(&models.Report{}).
		Where("user_id = ? AND status = ?", userID, models.ReportStatusFailed).
		Count(&stats.FailedReports).Error; err != nil {
		return nil, fmt.Errorf("failed to count failed reports: %w", err)
	}

	stats.UserID = userID

	return &stats, nil
}

// CleanupExpiredReports removes expired report records
func (r *ReportRepository) CleanupExpiredReports(ctx context.Context, batchSize int) (int64, error) {
	result := r.db.DB.WithContext(ctx).
		Where("expires_at < NOW() AND status = ?", models.ReportStatusCompleted).
		Limit(batchSize).
		Delete(&models.Report{})

	if result.Error != nil {
		r.logger.LogError(ctx, "failed to cleanup expired reports", result.Error)
		return 0, fmt.Errorf("failed to cleanup expired reports: %w", result.Error)
	}

	r.logger.LogInfo(ctx, "cleaned up expired reports",
		logger.String("count", fmt.Sprintf("%d", result.RowsAffected)))

	return result.RowsAffected, nil
}
