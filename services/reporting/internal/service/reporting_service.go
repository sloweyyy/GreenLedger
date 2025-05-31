package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sloweyyy/GreenLedger/services/reporting/internal/models"
	"github.com/sloweyyy/GreenLedger/services/reporting/internal/repository"
	"github.com/sloweyyy/GreenLedger/shared/logger"
)

// ReportingService handles report generation and management
type ReportingService struct {
	reportRepo     *repository.ReportRepository
	dataCollector  DataCollector
	reportRenderer ReportRenderer
	logger         *logger.Logger
}

// NewReportingService creates a new reporting service
func NewReportingService(
	reportRepo *repository.ReportRepository,
	dataCollector DataCollector,
	reportRenderer ReportRenderer,
	logger *logger.Logger,
) *ReportingService {
	return &ReportingService{
		reportRepo:     reportRepo,
		dataCollector:  dataCollector,
		reportRenderer: reportRenderer,
		logger:         logger,
	}
}

// GenerateReportRequest represents a request to generate a report
type GenerateReportRequest struct {
	UserID      string                 `json:"user_id" binding:"required"`
	Type        string                 `json:"type" binding:"required"`
	Format      string                 `json:"format" binding:"required"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	StartDate   time.Time              `json:"start_date" binding:"required"`
	EndDate     time.Time              `json:"end_date" binding:"required"`
	Parameters  map[string]interface{} `json:"parameters"`
}

// ReportResponse represents a report in API responses
type ReportResponse struct {
	ID          uuid.UUID  `json:"id"`
	UserID      string     `json:"user_id"`
	Type        string     `json:"type"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Format      string     `json:"format"`
	Status      string     `json:"status"`
	FilePath    string     `json:"file_path,omitempty"`
	FileSize    int64      `json:"file_size,omitempty"`
	StartDate   time.Time  `json:"start_date"`
	EndDate     time.Time  `json:"end_date"`
	GeneratedAt *time.Time `json:"generated_at"`
	ExpiresAt   *time.Time `json:"expires_at"`
	CreatedAt   time.Time  `json:"created_at"`
}

// DataCollector interface for collecting report data
type DataCollector interface {
	CollectFootprintData(ctx context.Context, userID string, startDate, endDate time.Time) (*models.FootprintReportData, error)
	CollectCreditsData(ctx context.Context, userID string, startDate, endDate time.Time) (*models.CreditsReportData, error)
	CollectSummaryData(ctx context.Context, userID string, startDate, endDate time.Time) (*models.SummaryReportData, error)
}

// ReportRenderer interface for rendering reports
type ReportRenderer interface {
	RenderPDF(ctx context.Context, reportType string, data interface{}) ([]byte, error)
	RenderJSON(ctx context.Context, data interface{}) ([]byte, error)
	RenderCSV(ctx context.Context, reportType string, data interface{}) ([]byte, error)
}

// GenerateReport generates a new report
func (s *ReportingService) GenerateReport(ctx context.Context, req *GenerateReportRequest) (*ReportResponse, error) {
	s.logger.LogInfo(ctx, "generating report",
		logger.String("user_id", req.UserID),
		logger.String("type", req.Type),
		logger.String("format", req.Format))

	// Validate request
	if err := s.validateReportRequest(req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Create report record
	report := &models.Report{
		UserID:      req.UserID,
		Type:        req.Type,
		Title:       req.Title,
		Description: req.Description,
		Format:      req.Format,
		Status:      models.ReportStatusPending,
		StartDate:   req.StartDate,
		EndDate:     req.EndDate,
	}

	if req.Title == "" {
		report.Title = s.generateDefaultTitle(req.Type, req.StartDate, req.EndDate)
	}

	// Set expiration (30 days from now)
	expiresAt := time.Now().AddDate(0, 0, 30)
	report.ExpiresAt = &expiresAt

	// Serialize parameters
	if req.Parameters != nil {
		parametersJSON, err := json.Marshal(req.Parameters)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize parameters: %w", err)
		}
		report.Parameters = string(parametersJSON)
	}

	// Save report
	if err := s.reportRepo.Create(ctx, report); err != nil {
		return nil, fmt.Errorf("failed to create report: %w", err)
	}

	// Generate report asynchronously
	go s.generateReportAsync(context.Background(), report)

	return s.reportToResponse(report), nil
}

// GetReport retrieves a report by ID
func (s *ReportingService) GetReport(ctx context.Context, reportID uuid.UUID, userID string) (*ReportResponse, error) {
	report, err := s.reportRepo.GetByID(ctx, reportID)
	if err != nil {
		return nil, fmt.Errorf("failed to get report: %w", err)
	}

	// Check if user owns the report
	if report.UserID != userID {
		return nil, fmt.Errorf("report not found")
	}

	return s.reportToResponse(report), nil
}

// GetUserReports retrieves reports for a user
func (s *ReportingService) GetUserReports(ctx context.Context, userID string, limit, offset int) ([]*ReportResponse, int64, error) {
	reports, total, err := s.reportRepo.GetByUserID(ctx, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get user reports: %w", err)
	}

	responses := make([]*ReportResponse, len(reports))
	for i, report := range reports {
		responses[i] = s.reportToResponse(report)
	}

	return responses, total, nil
}

// DeleteReport deletes a report
func (s *ReportingService) DeleteReport(ctx context.Context, reportID uuid.UUID, userID string) error {
	report, err := s.reportRepo.GetByID(ctx, reportID)
	if err != nil {
		return fmt.Errorf("failed to get report: %w", err)
	}

	// Check if user owns the report
	if report.UserID != userID {
		return fmt.Errorf("report not found")
	}

	// Delete report file if exists
	if report.FilePath != "" {
		// TODO: Delete file from storage
	}

	// Delete report record
	if err := s.reportRepo.Delete(ctx, reportID); err != nil {
		return fmt.Errorf("failed to delete report: %w", err)
	}

	s.logger.LogInfo(ctx, "report deleted",
		logger.String("report_id", reportID.String()),
		logger.String("user_id", userID))

	return nil
}

// generateReportAsync generates the report content asynchronously
func (s *ReportingService) generateReportAsync(ctx context.Context, report *models.Report) {
	s.logger.LogInfo(ctx, "starting async report generation",
		logger.String("report_id", report.ID.String()))

	// Update status to generating
	report.Status = models.ReportStatusGenerating
	if err := s.reportRepo.Update(ctx, report); err != nil {
		s.logger.LogError(ctx, "failed to update report status", err)
		return
	}

	// Collect data based on report type
	var data interface{}
	var err error

	switch report.Type {
	case models.ReportTypeFootprint:
		data, err = s.dataCollector.CollectFootprintData(ctx, report.UserID, report.StartDate, report.EndDate)
	case models.ReportTypeCredits:
		data, err = s.dataCollector.CollectCreditsData(ctx, report.UserID, report.StartDate, report.EndDate)
	case models.ReportTypeSummary:
		data, err = s.dataCollector.CollectSummaryData(ctx, report.UserID, report.StartDate, report.EndDate)
	default:
		err = fmt.Errorf("unsupported report type: %s", report.Type)
	}

	if err != nil {
		s.logger.LogError(ctx, "failed to collect report data", err,
			logger.String("report_id", report.ID.String()))
		report.Status = models.ReportStatusFailed
		s.reportRepo.Update(ctx, report)
		return
	}

	// Render report based on format
	var content []byte
	switch report.Format {
	case models.ReportFormatPDF:
		content, err = s.reportRenderer.RenderPDF(ctx, report.Type, data)
	case models.ReportFormatJSON:
		content, err = s.reportRenderer.RenderJSON(ctx, data)
	case models.ReportFormatCSV:
		content, err = s.reportRenderer.RenderCSV(ctx, report.Type, data)
	default:
		err = fmt.Errorf("unsupported report format: %s", report.Format)
	}

	if err != nil {
		s.logger.LogError(ctx, "failed to render report", err,
			logger.String("report_id", report.ID.String()))
		report.Status = models.ReportStatusFailed
		s.reportRepo.Update(ctx, report)
		return
	}

	// Save report file
	filePath := fmt.Sprintf("reports/%s/%s.%s", report.UserID, report.ID.String(), report.Format)
	// TODO: Save content to file storage (S3, local filesystem, etc.)

	// Update report with file information
	now := time.Now().UTC()
	report.Status = models.ReportStatusCompleted
	report.FilePath = filePath
	report.FileSize = int64(len(content))
	report.GeneratedAt = &now

	if err := s.reportRepo.Update(ctx, report); err != nil {
		s.logger.LogError(ctx, "failed to update report", err)
		return
	}

	s.logger.LogInfo(ctx, "report generated successfully",
		logger.String("report_id", report.ID.String()),
		logger.String("file_path", filePath),
		logger.Int("file_size", len(content)))
}

// validateReportRequest validates a report generation request
func (s *ReportingService) validateReportRequest(req *GenerateReportRequest) error {
	// Validate report type
	validTypes := []string{
		models.ReportTypeFootprint,
		models.ReportTypeCredits,
		models.ReportTypeActivities,
		models.ReportTypeTransactions,
		models.ReportTypeSummary,
	}

	isValidType := false
	for _, validType := range validTypes {
		if req.Type == validType {
			isValidType = true
			break
		}
	}
	if !isValidType {
		return fmt.Errorf("invalid report type: %s", req.Type)
	}

	// Validate report format
	validFormats := []string{
		models.ReportFormatPDF,
		models.ReportFormatJSON,
		models.ReportFormatCSV,
	}

	isValidFormat := false
	for _, validFormat := range validFormats {
		if req.Format == validFormat {
			isValidFormat = true
			break
		}
	}
	if !isValidFormat {
		return fmt.Errorf("invalid report format: %s", req.Format)
	}

	// Validate date range
	if req.EndDate.Before(req.StartDate) {
		return fmt.Errorf("end date must be after start date")
	}

	// Validate date range is not too large (max 1 year)
	if req.EndDate.Sub(req.StartDate) > 365*24*time.Hour {
		return fmt.Errorf("date range cannot exceed 1 year")
	}

	return nil
}

// generateDefaultTitle generates a default title for a report
func (s *ReportingService) generateDefaultTitle(reportType string, startDate, endDate time.Time) string {
	var typeTitle string
	switch reportType {
	case models.ReportTypeFootprint:
		typeTitle = "Carbon Footprint Report"
	case models.ReportTypeCredits:
		typeTitle = "Carbon Credits Report"
	case models.ReportTypeActivities:
		typeTitle = "Activities Report"
	case models.ReportTypeTransactions:
		typeTitle = "Transactions Report"
	case models.ReportTypeSummary:
		typeTitle = "Summary Report"
	default:
		typeTitle = "Report"
	}

	return fmt.Sprintf("%s (%s to %s)",
		typeTitle,
		startDate.Format("2006-01-02"),
		endDate.Format("2006-01-02"))
}

// reportToResponse converts a report model to response format
func (s *ReportingService) reportToResponse(report *models.Report) *ReportResponse {
	return &ReportResponse{
		ID:          report.ID,
		UserID:      report.UserID,
		Type:        report.Type,
		Title:       report.Title,
		Description: report.Description,
		Format:      report.Format,
		Status:      report.Status,
		FilePath:    report.FilePath,
		FileSize:    report.FileSize,
		StartDate:   report.StartDate,
		EndDate:     report.EndDate,
		GeneratedAt: report.GeneratedAt,
		ExpiresAt:   report.ExpiresAt,
		CreatedAt:   report.CreatedAt,
	}
}
