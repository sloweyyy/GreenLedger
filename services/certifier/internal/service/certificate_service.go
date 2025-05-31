package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/sloweyyy/GreenLedger/services/certifier/internal/models"
	"github.com/sloweyyy/GreenLedger/services/certifier/internal/repository"
	"github.com/sloweyyy/GreenLedger/shared/logger"
)

// CertificateService handles certificate business logic
type CertificateService struct {
	certificateRepo *repository.CertificateRepository
	projectRepo     *repository.ProjectRepository
	logger          *logger.Logger
}

// NewCertificateService creates a new certificate service
func NewCertificateService(
	certificateRepo *repository.CertificateRepository,
	projectRepo *repository.ProjectRepository,
	logger *logger.Logger,
) *CertificateService {
	return &CertificateService{
		certificateRepo: certificateRepo,
		projectRepo:     projectRepo,
		logger:          logger,
	}
}

// IssueCertificateRequest represents a request to issue a certificate
type IssueCertificateRequest struct {
	UserID         string          `json:"user_id" binding:"required"`
	Type           string          `json:"type" binding:"required"`
	CarbonOffset   decimal.Decimal `json:"carbon_offset" binding:"required"`
	CreditsUsed    decimal.Decimal `json:"credits_used" binding:"required"`
	ProjectName    string          `json:"project_name" binding:"required"`
	Description    string          `json:"description"`
	VintageYear    int             `json:"vintage_year"`
	ExpirationDays int             `json:"expiration_days"`
}

// CertificateResponse represents a certificate in API responses
type CertificateResponse struct {
	ID                uuid.UUID       `json:"id"`
	UserID            string          `json:"user_id"`
	CertificateNumber string          `json:"certificate_number"`
	Type              string          `json:"type"`
	Status            string          `json:"status"`
	CarbonOffset      decimal.Decimal `json:"carbon_offset"`
	CreditsUsed       decimal.Decimal `json:"credits_used"`
	ProjectName       string          `json:"project_name"`
	ProjectType       string          `json:"project_type"`
	ProjectLocation   string          `json:"project_location"`
	VerificationBody  string          `json:"verification_body"`
	Standard          string          `json:"standard"`
	VintageYear       int             `json:"vintage_year"`
	SerialNumber      string          `json:"serial_number"`
	BlockchainTxHash  string          `json:"blockchain_tx_hash"`
	TokenID           string          `json:"token_id"`
	IssuedAt          *time.Time      `json:"issued_at"`
	ExpiresAt         *time.Time      `json:"expires_at"`
	CreatedAt         time.Time       `json:"created_at"`
}

// IssueCertificate issues a new carbon offset certificate
func (s *CertificateService) IssueCertificate(ctx context.Context, req *IssueCertificateRequest) (*CertificateResponse, error) {
	s.logger.LogInfo(ctx, "issuing certificate",
		logger.String("user_id", req.UserID),
		logger.String("project_name", req.ProjectName))

	// Validate request
	if err := s.validateIssueRequest(req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Get project information
	project, err := s.projectRepo.GetByName(ctx, req.ProjectName)
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	// Check if project has enough available credits
	if !project.CanIssueCredits(req.CreditsUsed) {
		return nil, fmt.Errorf("insufficient credits available in project")
	}

	// Generate certificate number and serial number
	certificateNumber := s.generateCertificateNumber(req.Type, project.Type)
	serialNumber := s.generateSerialNumber(project.Name, req.VintageYear)

	// Create certificate
	certificate := &models.Certificate{
		UserID:            req.UserID,
		CertificateNumber: certificateNumber,
		Type:              req.Type,
		Status:            models.CertificateStatusPending,
		CarbonOffset:      req.CarbonOffset,
		CreditsUsed:       req.CreditsUsed,
		ProjectName:       project.Name,
		ProjectType:       project.Type,
		ProjectLocation:   project.Location,
		VerificationBody:  project.VerificationBody,
		Standard:          project.Standard,
		VintageYear:       req.VintageYear,
		SerialNumber:      serialNumber,
	}

	// Set expiration if specified
	if req.ExpirationDays > 0 {
		expiresAt := time.Now().AddDate(0, 0, req.ExpirationDays)
		certificate.ExpiresAt = &expiresAt
	}

	// Save certificate
	if err := s.certificateRepo.Create(ctx, certificate); err != nil {
		return nil, fmt.Errorf("failed to create certificate: %w", err)
	}

	// Update project available credits
	if err := s.projectRepo.UpdateAvailableCredits(ctx, project.ID, req.CreditsUsed.InexactFloat64()); err != nil {
		s.logger.LogError(ctx, "failed to update project credits", err)
		// Note: In a real system, this should be handled with a transaction
	}

	// Issue the certificate (update status)
	now := time.Now().UTC()
	certificate.Status = models.CertificateStatusIssued
	certificate.IssuedAt = &now

	if err := s.certificateRepo.Update(ctx, certificate); err != nil {
		return nil, fmt.Errorf("failed to issue certificate: %w", err)
	}

	s.logger.LogInfo(ctx, "certificate issued successfully",
		logger.String("certificate_id", certificate.ID.String()),
		logger.String("certificate_number", certificate.CertificateNumber))

	return s.certificateToResponse(certificate), nil
}

// GetCertificate retrieves a certificate by ID
func (s *CertificateService) GetCertificate(ctx context.Context, certificateID uuid.UUID, userID string) (*CertificateResponse, error) {
	certificate, err := s.certificateRepo.GetByID(ctx, certificateID)
	if err != nil {
		return nil, fmt.Errorf("failed to get certificate: %w", err)
	}

	// Check if user owns the certificate (or is admin)
	if certificate.UserID != userID {
		return nil, fmt.Errorf("certificate not found")
	}

	return s.certificateToResponse(certificate), nil
}

// GetUserCertificates retrieves certificates for a user
func (s *CertificateService) GetUserCertificates(ctx context.Context, userID string, limit, offset int) ([]*CertificateResponse, int64, error) {
	certificates, total, err := s.certificateRepo.GetByUserID(ctx, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get user certificates: %w", err)
	}

	responses := make([]*CertificateResponse, len(certificates))
	for i, cert := range certificates {
		responses[i] = s.certificateToResponse(cert)
	}

	return responses, total, nil
}

// VerifyCertificate verifies a certificate by certificate number
func (s *CertificateService) VerifyCertificate(ctx context.Context, certificateNumber string) (*CertificateResponse, error) {
	certificate, err := s.certificateRepo.GetByCertificateNumber(ctx, certificateNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to verify certificate: %w", err)
	}

	// Check if certificate is valid
	if certificate.IsExpired() {
		return nil, fmt.Errorf("certificate has expired")
	}

	if certificate.IsRetired() {
		return nil, fmt.Errorf("certificate has been retired")
	}

	return s.certificateToResponse(certificate), nil
}

// RetireCertificate retires a certificate
func (s *CertificateService) RetireCertificate(ctx context.Context, certificateID uuid.UUID, userID string) error {
	certificate, err := s.certificateRepo.GetByID(ctx, certificateID)
	if err != nil {
		return fmt.Errorf("failed to get certificate: %w", err)
	}

	// Check if user owns the certificate
	if certificate.UserID != userID {
		return fmt.Errorf("certificate not found")
	}

	// Check if certificate can be retired
	if !certificate.CanTransfer() {
		return fmt.Errorf("certificate cannot be retired")
	}

	// Update certificate status
	now := time.Now().UTC()
	certificate.Status = models.CertificateStatusRetired
	certificate.RetiredAt = &now

	if err := s.certificateRepo.Update(ctx, certificate); err != nil {
		return fmt.Errorf("failed to retire certificate: %w", err)
	}

	s.logger.LogInfo(ctx, "certificate retired",
		logger.String("certificate_id", certificate.ID.String()),
		logger.String("user_id", userID))

	return nil
}

// validateIssueRequest validates a certificate issue request
func (s *CertificateService) validateIssueRequest(req *IssueCertificateRequest) error {
	// Validate certificate type
	validTypes := []string{
		models.CertificateTypeOffset,
		models.CertificateTypeReduction,
		models.CertificateTypeRemoval,
		models.CertificateTypeAvoidance,
	}

	isValidType := false
	for _, validType := range validTypes {
		if req.Type == validType {
			isValidType = true
			break
		}
	}
	if !isValidType {
		return fmt.Errorf("invalid certificate type: %s", req.Type)
	}

	// Validate amounts
	if req.CarbonOffset.LessThanOrEqual(decimal.Zero) {
		return fmt.Errorf("carbon offset must be positive")
	}

	if req.CreditsUsed.LessThanOrEqual(decimal.Zero) {
		return fmt.Errorf("credits used must be positive")
	}

	// Validate vintage year
	currentYear := time.Now().Year()
	if req.VintageYear < 1990 || req.VintageYear > currentYear {
		return fmt.Errorf("invalid vintage year: %d", req.VintageYear)
	}

	return nil
}

// generateCertificateNumber generates a unique certificate number
func (s *CertificateService) generateCertificateNumber(certType, projectType string) string {
	timestamp := time.Now().Unix()
	return fmt.Sprintf("GL-%s-%s-%d", certType, projectType, timestamp)
}

// generateSerialNumber generates a unique serial number
func (s *CertificateService) generateSerialNumber(projectName string, vintageYear int) string {
	timestamp := time.Now().Unix()
	return fmt.Sprintf("%s-%d-%d", projectName, vintageYear, timestamp)
}

// certificateToResponse converts a certificate model to response format
func (s *CertificateService) certificateToResponse(cert *models.Certificate) *CertificateResponse {
	return &CertificateResponse{
		ID:                cert.ID,
		UserID:            cert.UserID,
		CertificateNumber: cert.CertificateNumber,
		Type:              cert.Type,
		Status:            cert.Status,
		CarbonOffset:      cert.CarbonOffset,
		CreditsUsed:       cert.CreditsUsed,
		ProjectName:       cert.ProjectName,
		ProjectType:       cert.ProjectType,
		ProjectLocation:   cert.ProjectLocation,
		VerificationBody:  cert.VerificationBody,
		Standard:          cert.Standard,
		VintageYear:       cert.VintageYear,
		SerialNumber:      cert.SerialNumber,
		BlockchainTxHash:  cert.BlockchainTxHash,
		TokenID:           cert.TokenID,
		IssuedAt:          cert.IssuedAt,
		ExpiresAt:         cert.ExpiresAt,
		CreatedAt:         cert.CreatedAt,
	}
}
