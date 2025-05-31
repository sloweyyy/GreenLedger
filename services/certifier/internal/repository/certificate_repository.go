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

// CertificateRepository handles certificate data operations
type CertificateRepository struct {
	db     *database.PostgresDB
	logger *logger.Logger
}

// NewCertificateRepository creates a new certificate repository
func NewCertificateRepository(db *database.PostgresDB, logger *logger.Logger) *CertificateRepository {
	return &CertificateRepository{
		db:     db,
		logger: logger,
	}
}

// Create creates a new certificate
func (r *CertificateRepository) Create(ctx context.Context, certificate *models.Certificate) error {
	if err := r.db.WithContext(ctx).Create(certificate).Error; err != nil {
		r.logger.LogError(ctx, "failed to create certificate", err,
			logger.String("user_id", certificate.UserID))
		return fmt.Errorf("failed to create certificate: %w", err)
	}

	r.logger.LogInfo(ctx, "certificate created",
		logger.String("certificate_id", certificate.ID.String()),
		logger.String("user_id", certificate.UserID))

	return nil
}

// GetByID retrieves a certificate by ID
func (r *CertificateRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Certificate, error) {
	var certificate models.Certificate
	if err := r.db.WithContext(ctx).
		Preload("Verifications").
		Preload("Transfers").
		First(&certificate, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("certificate not found")
		}
		r.logger.LogError(ctx, "failed to get certificate", err,
			logger.String("certificate_id", id.String()))
		return nil, fmt.Errorf("failed to get certificate: %w", err)
	}

	return &certificate, nil
}

// GetByUserID retrieves certificates for a user
func (r *CertificateRepository) GetByUserID(ctx context.Context, userID string, limit, offset int) ([]*models.Certificate, int64, error) {
	var certificates []*models.Certificate
	var total int64

	// Get total count
	if err := r.db.WithContext(ctx).Model(&models.Certificate{}).
		Where("user_id = ?", userID).
		Count(&total).Error; err != nil {
		r.logger.LogError(ctx, "failed to count certificates", err,
			logger.String("user_id", userID))
		return nil, 0, fmt.Errorf("failed to count certificates: %w", err)
	}

	// Get certificates with pagination
	if err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&certificates).Error; err != nil {
		r.logger.LogError(ctx, "failed to get certificates", err,
			logger.String("user_id", userID))
		return nil, 0, fmt.Errorf("failed to get certificates: %w", err)
	}

	return certificates, total, nil
}

// GetByCertificateNumber retrieves a certificate by certificate number
func (r *CertificateRepository) GetByCertificateNumber(ctx context.Context, certificateNumber string) (*models.Certificate, error) {
	var certificate models.Certificate
	if err := r.db.WithContext(ctx).
		Preload("Verifications").
		Preload("Transfers").
		First(&certificate, "certificate_number = ?", certificateNumber).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("certificate not found")
		}
		r.logger.LogError(ctx, "failed to get certificate by number", err,
			logger.String("certificate_number", certificateNumber))
		return nil, fmt.Errorf("failed to get certificate: %w", err)
	}

	return &certificate, nil
}

// Update updates a certificate
func (r *CertificateRepository) Update(ctx context.Context, certificate *models.Certificate) error {
	if err := r.db.WithContext(ctx).Save(certificate).Error; err != nil {
		r.logger.LogError(ctx, "failed to update certificate", err,
			logger.String("certificate_id", certificate.ID.String()))
		return fmt.Errorf("failed to update certificate: %w", err)
	}

	r.logger.LogInfo(ctx, "certificate updated",
		logger.String("certificate_id", certificate.ID.String()))

	return nil
}

// Delete deletes a certificate
func (r *CertificateRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Delete(&models.Certificate{}, "id = ?", id).Error; err != nil {
		r.logger.LogError(ctx, "failed to delete certificate", err,
			logger.String("certificate_id", id.String()))
		return fmt.Errorf("failed to delete certificate: %w", err)
	}

	r.logger.LogInfo(ctx, "certificate deleted",
		logger.String("certificate_id", id.String()))

	return nil
}

// GetByStatus retrieves certificates by status
func (r *CertificateRepository) GetByStatus(ctx context.Context, status string, limit, offset int) ([]*models.Certificate, int64, error) {
	var certificates []*models.Certificate
	var total int64

	// Get total count
	if err := r.db.WithContext(ctx).Model(&models.Certificate{}).
		Where("status = ?", status).
		Count(&total).Error; err != nil {
		r.logger.LogError(ctx, "failed to count certificates by status", err,
			logger.String("status", status))
		return nil, 0, fmt.Errorf("failed to count certificates: %w", err)
	}

	// Get certificates with pagination
	if err := r.db.WithContext(ctx).
		Where("status = ?", status).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&certificates).Error; err != nil {
		r.logger.LogError(ctx, "failed to get certificates by status", err,
			logger.String("status", status))
		return nil, 0, fmt.Errorf("failed to get certificates: %w", err)
	}

	return certificates, total, nil
}

// CreateVerification creates a certificate verification
func (r *CertificateRepository) CreateVerification(ctx context.Context, verification *models.CertificateVerification) error {
	if err := r.db.WithContext(ctx).Create(verification).Error; err != nil {
		r.logger.LogError(ctx, "failed to create verification", err,
			logger.String("certificate_id", verification.CertificateID.String()))
		return fmt.Errorf("failed to create verification: %w", err)
	}

	r.logger.LogInfo(ctx, "verification created",
		logger.String("verification_id", verification.ID.String()),
		logger.String("certificate_id", verification.CertificateID.String()))

	return nil
}

// CreateTransfer creates a certificate transfer
func (r *CertificateRepository) CreateTransfer(ctx context.Context, transfer *models.CertificateTransfer) error {
	if err := r.db.WithContext(ctx).Create(transfer).Error; err != nil {
		r.logger.LogError(ctx, "failed to create transfer", err,
			logger.String("certificate_id", transfer.CertificateID.String()))
		return fmt.Errorf("failed to create transfer: %w", err)
	}

	r.logger.LogInfo(ctx, "transfer created",
		logger.String("transfer_id", transfer.ID.String()),
		logger.String("certificate_id", transfer.CertificateID.String()))

	return nil
}

// GetTransfersByUserID retrieves transfers for a user
func (r *CertificateRepository) GetTransfersByUserID(ctx context.Context, userID string, limit, offset int) ([]*models.CertificateTransfer, int64, error) {
	var transfers []*models.CertificateTransfer
	var total int64

	// Get total count
	if err := r.db.WithContext(ctx).Model(&models.CertificateTransfer{}).
		Where("from_user_id = ? OR to_user_id = ?", userID, userID).
		Count(&total).Error; err != nil {
		r.logger.LogError(ctx, "failed to count transfers", err,
			logger.String("user_id", userID))
		return nil, 0, fmt.Errorf("failed to count transfers: %w", err)
	}

	// Get transfers with pagination
	if err := r.db.WithContext(ctx).
		Preload("Certificate").
		Where("from_user_id = ? OR to_user_id = ?", userID, userID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&transfers).Error; err != nil {
		r.logger.LogError(ctx, "failed to get transfers", err,
			logger.String("user_id", userID))
		return nil, 0, fmt.Errorf("failed to get transfers: %w", err)
	}

	return transfers, total, nil
}
