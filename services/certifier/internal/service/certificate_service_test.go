package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/sloweyyy/GreenLedger/services/certifier/internal/models"
)

// MockCertificateRepository implements the repository interface for testing
type MockCertificateRepository struct {
	certificates map[uuid.UUID]*models.Certificate
}

func NewMockCertificateRepository() *MockCertificateRepository {
	return &MockCertificateRepository{
		certificates: make(map[uuid.UUID]*models.Certificate),
	}
}

func (m *MockCertificateRepository) Create(ctx context.Context, certificate *models.Certificate) error {
	if certificate.ID == uuid.Nil {
		certificate.ID = uuid.New()
	}
	certificate.CreatedAt = time.Now()
	certificate.UpdatedAt = time.Now()
	m.certificates[certificate.ID] = certificate
	return nil
}

func (m *MockCertificateRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Certificate, error) {
	if cert, exists := m.certificates[id]; exists {
		return cert, nil
	}
	return nil, nil
}

func (m *MockCertificateRepository) GetByUserID(ctx context.Context, userID string, limit, offset int) ([]*models.Certificate, int64, error) {
	var result []*models.Certificate
	for _, cert := range m.certificates {
		if cert.UserID == userID {
			result = append(result, cert)
		}
	}
	return result, int64(len(result)), nil
}

func (m *MockCertificateRepository) Update(ctx context.Context, certificate *models.Certificate) error {
	certificate.UpdatedAt = time.Now()
	m.certificates[certificate.ID] = certificate
	return nil
}

func (m *MockCertificateRepository) Delete(ctx context.Context, id uuid.UUID) error {
	delete(m.certificates, id)
	return nil
}

func (m *MockCertificateRepository) GetByCertificateNumber(ctx context.Context, certificateNumber string) (*models.Certificate, error) {
	for _, cert := range m.certificates {
		if cert.CertificateNumber == certificateNumber {
			return cert, nil
		}
	}
	return nil, nil
}

func (m *MockCertificateRepository) GetByStatus(ctx context.Context, status string, limit, offset int) ([]*models.Certificate, int64, error) {
	var result []*models.Certificate
	for _, cert := range m.certificates {
		if cert.Status == status {
			result = append(result, cert)
		}
	}
	return result, int64(len(result)), nil
}

func (m *MockCertificateRepository) CreateVerification(ctx context.Context, verification *models.CertificateVerification) error {
	return nil
}

func (m *MockCertificateRepository) CreateTransfer(ctx context.Context, transfer *models.CertificateTransfer) error {
	return nil
}

func (m *MockCertificateRepository) GetTransfersByUserID(ctx context.Context, userID string, limit, offset int) ([]*models.CertificateTransfer, int64, error) {
	return []*models.CertificateTransfer{}, 0, nil
}

// MockProjectRepository implements the project repository interface for testing
type MockProjectRepository struct {
	projects map[uuid.UUID]*models.CertificateProject
}

func NewMockProjectRepository() *MockProjectRepository {
	return &MockProjectRepository{
		projects: make(map[uuid.UUID]*models.CertificateProject),
	}
}

func (m *MockProjectRepository) Create(ctx context.Context, project *models.CertificateProject) error {
	if project.ID == uuid.Nil {
		project.ID = uuid.New()
	}
	project.CreatedAt = time.Now()
	project.UpdatedAt = time.Now()
	m.projects[project.ID] = project
	return nil
}

func (m *MockProjectRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.CertificateProject, error) {
	if project, exists := m.projects[id]; exists {
		return project, nil
	}
	return nil, nil
}

func (m *MockProjectRepository) GetByName(ctx context.Context, name string) (*models.CertificateProject, error) {
	for _, project := range m.projects {
		if project.Name == name {
			return project, nil
		}
	}
	return nil, nil
}

func (m *MockProjectRepository) GetAll(ctx context.Context, limit, offset int) ([]*models.CertificateProject, int64, error) {
	var result []*models.CertificateProject
	for _, project := range m.projects {
		result = append(result, project)
	}
	return result, int64(len(result)), nil
}

func (m *MockProjectRepository) GetActive(ctx context.Context, limit, offset int) ([]*models.CertificateProject, int64, error) {
	var result []*models.CertificateProject
	for _, project := range m.projects {
		if project.IsActive {
			result = append(result, project)
		}
	}
	return result, int64(len(result)), nil
}

func (m *MockProjectRepository) GetByType(ctx context.Context, projectType string, limit, offset int) ([]*models.CertificateProject, int64, error) {
	var result []*models.CertificateProject
	for _, project := range m.projects {
		if project.Type == projectType {
			result = append(result, project)
		}
	}
	return result, int64(len(result)), nil
}

func (m *MockProjectRepository) Update(ctx context.Context, project *models.CertificateProject) error {
	project.UpdatedAt = time.Now()
	m.projects[project.ID] = project
	return nil
}

func (m *MockProjectRepository) Delete(ctx context.Context, id uuid.UUID) error {
	delete(m.projects, id)
	return nil
}

func (m *MockProjectRepository) UpdateAvailableCredits(ctx context.Context, projectID uuid.UUID, creditsUsed float64) error {
	if project, exists := m.projects[projectID]; exists {
		project.AvailableCredits = project.AvailableCredits.Sub(decimal.NewFromFloat(creditsUsed))
		m.projects[projectID] = project
	}
	return nil
}

func TestCertificateModel_Creation(t *testing.T) {
	certificate := &models.Certificate{
		ID:                uuid.New(),
		UserID:            "test-user-123",
		CertificateNumber: "GL-OFFSET-RENEWABLE-123456",
		Type:              models.CertificateTypeOffset,
		Status:            models.CertificateStatusPending,
		CarbonOffset:      decimal.NewFromFloat(50.5),
		CreditsUsed:       decimal.NewFromFloat(50.5),
		ProjectName:       "Test Solar Project",
		ProjectType:       models.ProjectTypeRenewable,
		ProjectLocation:   "California, USA",
		VintageYear:       2023,
		SerialNumber:      "SOLAR-2023-123456",
	}

	if certificate.UserID != "test-user-123" {
		t.Errorf("Expected UserID 'test-user-123', got %s", certificate.UserID)
	}

	if certificate.Type != models.CertificateTypeOffset {
		t.Errorf("Expected Type %s, got %s", models.CertificateTypeOffset, certificate.Type)
	}

	if !certificate.CarbonOffset.Equal(decimal.NewFromFloat(50.5)) {
		t.Errorf("Expected CarbonOffset 50.5, got %s", certificate.CarbonOffset)
	}

	if certificate.Status != models.CertificateStatusPending {
		t.Errorf("Expected Status %s, got %s", models.CertificateStatusPending, certificate.Status)
	}

	if certificate.ProjectType != models.ProjectTypeRenewable {
		t.Errorf("Expected ProjectType %s, got %s", models.ProjectTypeRenewable, certificate.ProjectType)
	}
}

func TestCertificateModel_IsIssued(t *testing.T) {
	issuedCert := &models.Certificate{
		ID:     uuid.New(),
		Status: models.CertificateStatusIssued,
	}

	verifiedCert := &models.Certificate{
		ID:     uuid.New(),
		Status: models.CertificateStatusVerified,
	}

	pendingCert := &models.Certificate{
		ID:     uuid.New(),
		Status: models.CertificateStatusPending,
	}

	if !issuedCert.IsIssued() {
		t.Error("Expected issued certificate to be issued")
	}

	if !verifiedCert.IsIssued() {
		t.Error("Expected verified certificate to be issued")
	}

	if pendingCert.IsIssued() {
		t.Error("Expected pending certificate to not be issued")
	}
}

func TestCertificateModel_CanTransfer(t *testing.T) {
	transferableCert := &models.Certificate{
		ID:     uuid.New(),
		Status: models.CertificateStatusIssued,
	}

	retiredCert := &models.Certificate{
		ID:     uuid.New(),
		Status: models.CertificateStatusRetired,
	}

	expiredTime := time.Now().Add(-1 * time.Hour)
	expiredCert := &models.Certificate{
		ID:        uuid.New(),
		Status:    models.CertificateStatusIssued,
		ExpiresAt: &expiredTime,
	}

	if !transferableCert.CanTransfer() {
		t.Error("Expected issued certificate to be transferable")
	}

	if retiredCert.CanTransfer() {
		t.Error("Expected retired certificate to not be transferable")
	}

	if expiredCert.CanTransfer() {
		t.Error("Expected expired certificate to not be transferable")
	}
}

func TestCertificateProjectModel_CanIssueCredits(t *testing.T) {
	activeProject := &models.CertificateProject{
		ID:               uuid.New(),
		IsActive:         true,
		AvailableCredits: decimal.NewFromFloat(100.0),
	}

	inactiveProject := &models.CertificateProject{
		ID:               uuid.New(),
		IsActive:         false,
		AvailableCredits: decimal.NewFromFloat(100.0),
	}

	insufficientProject := &models.CertificateProject{
		ID:               uuid.New(),
		IsActive:         true,
		AvailableCredits: decimal.NewFromFloat(10.0),
	}

	requestAmount := decimal.NewFromFloat(50.0)

	if !activeProject.CanIssueCredits(requestAmount) {
		t.Error("Expected active project with sufficient credits to allow issuance")
	}

	if inactiveProject.CanIssueCredits(requestAmount) {
		t.Error("Expected inactive project to not allow issuance")
	}

	if insufficientProject.CanIssueCredits(requestAmount) {
		t.Error("Expected project with insufficient credits to not allow issuance")
	}
}
