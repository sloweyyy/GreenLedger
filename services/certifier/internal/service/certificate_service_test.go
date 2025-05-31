package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/sloweyyy/GreenLedger/services/certifier/internal/models"
	"github.com/sloweyyy/GreenLedger/shared/logger"
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

func (m *MockCertificateRepository) GetByUserID(ctx context.Context, userID string, limit, offset int) ([]*models.Certificate, error) {
	var result []*models.Certificate
	for _, cert := range m.certificates {
		if cert.UserID == userID {
			result = append(result, cert)
		}
	}
	return result, nil
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

func (m *MockProjectRepository) GetAll(ctx context.Context, limit, offset int) ([]*models.CertificateProject, error) {
	var result []*models.CertificateProject
	for _, project := range m.projects {
		result = append(result, project)
	}
	return result, nil
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

func TestCertificateService_CreateCertificate(t *testing.T) {
	// Setup
	mockCertRepo := NewMockCertificateRepository()
	mockProjectRepo := NewMockProjectRepository()
	logger := logger.New("debug")
	
	service := NewCertificateService(mockCertRepo, mockProjectRepo, logger)
	
	// Create a test project first
	project := &models.CertificateProject{
		Name:             "Test Solar Project",
		Type:             "renewable_energy",
		Location:         "California, USA",
		TotalCredits:     decimal.NewFromInt(1000),
		AvailableCredits: decimal.NewFromInt(1000),
		PricePerCredit:   decimal.NewFromFloat(10.50),
		IsActive:         true,
		StartDate:        time.Now().AddDate(-1, 0, 0),
		EndDate:          time.Now().AddDate(1, 0, 0),
	}
	
	err := mockProjectRepo.Create(context.Background(), project)
	if err != nil {
		t.Fatalf("Failed to create test project: %v", err)
	}
	
	// Test certificate creation
	request := CreateCertificateRequest{
		UserID:       "test-user-123",
		Type:         models.CertificateTypeOffset,
		CarbonOffset: decimal.NewFromFloat(50.5),
		CreditsUsed:  decimal.NewFromFloat(50.5),
		ProjectName:  "Test Solar Project",
	}
	
	ctx := context.Background()
	certificate, err := service.CreateCertificate(ctx, request)
	
	// Assertions
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	if certificate == nil {
		t.Fatal("Expected certificate to be created, got nil")
	}
	
	if certificate.UserID != request.UserID {
		t.Errorf("Expected UserID %s, got %s", request.UserID, certificate.UserID)
	}
	
	if certificate.Type != request.Type {
		t.Errorf("Expected Type %s, got %s", request.Type, certificate.Type)
	}
	
	if !certificate.CarbonOffset.Equal(request.CarbonOffset) {
		t.Errorf("Expected CarbonOffset %s, got %s", request.CarbonOffset, certificate.CarbonOffset)
	}
	
	if certificate.CertificateNumber == "" {
		t.Error("Expected CertificateNumber to be generated")
	}
	
	if certificate.Status != "pending" {
		t.Errorf("Expected Status 'pending', got %s", certificate.Status)
	}
}

func TestCertificateService_GetCertificateByID(t *testing.T) {
	// Setup
	mockCertRepo := NewMockCertificateRepository()
	mockProjectRepo := NewMockProjectRepository()
	logger := logger.New("debug")
	
	service := NewCertificateService(mockCertRepo, mockProjectRepo, logger)
	
	// Create a test certificate
	certificate := &models.Certificate{
		ID:                uuid.New(),
		UserID:            "test-user-123",
		CertificateNumber: "CERT-TEST-001",
		Type:              models.CertificateTypeOffset,
		Status:            "pending",
		CarbonOffset:      decimal.NewFromFloat(25.5),
		CreditsUsed:       decimal.NewFromFloat(25.5),
		ProjectName:       "Test Project",
		ProjectType:       "renewable_energy",
	}
	
	err := mockCertRepo.Create(context.Background(), certificate)
	if err != nil {
		t.Fatalf("Failed to create test certificate: %v", err)
	}
	
	// Test getting certificate by ID
	ctx := context.Background()
	result, err := service.GetCertificateByID(ctx, certificate.ID)
	
	// Assertions
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	if result == nil {
		t.Fatal("Expected certificate to be found, got nil")
	}
	
	if result.ID != certificate.ID {
		t.Errorf("Expected ID %s, got %s", certificate.ID, result.ID)
	}
	
	if result.UserID != certificate.UserID {
		t.Errorf("Expected UserID %s, got %s", certificate.UserID, result.UserID)
	}
}

func TestCertificateService_GetCertificateByID_NotFound(t *testing.T) {
	// Setup
	mockCertRepo := NewMockCertificateRepository()
	mockProjectRepo := NewMockProjectRepository()
	logger := logger.New("debug")
	
	service := NewCertificateService(mockCertRepo, mockProjectRepo, logger)
	
	// Test getting non-existent certificate
	ctx := context.Background()
	nonExistentID := uuid.New()
	result, err := service.GetCertificateByID(ctx, nonExistentID)
	
	// Assertions
	if err == nil {
		t.Fatal("Expected error for non-existent certificate, got nil")
	}
	
	if result != nil {
		t.Error("Expected nil result for non-existent certificate")
	}
}
