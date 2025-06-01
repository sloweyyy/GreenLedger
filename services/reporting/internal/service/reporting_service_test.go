package service

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/sloweyyy/GreenLedger/services/reporting/internal/models"
)

func TestReportModel_IsCompleted(t *testing.T) {
	report := &models.Report{
		ID:     uuid.New(),
		Status: models.ReportStatusCompleted,
	}

	if !report.IsCompleted() {
		t.Error("Expected report to be completed")
	}
}

func TestReportModel_IsPending(t *testing.T) {
	report := &models.Report{
		ID:     uuid.New(),
		Status: models.ReportStatusPending,
	}

	if !report.IsPending() {
		t.Error("Expected report to be pending")
	}
}

func TestReportModel_IsExpired(t *testing.T) {
	pastTime := time.Now().Add(-1 * time.Hour)
	report := &models.Report{
		ID:        uuid.New(),
		ExpiresAt: &pastTime,
	}

	if !report.IsExpired() {
		t.Error("Expected report to be expired")
	}
}

func TestReportSchedule_ShouldRun(t *testing.T) {
	pastTime := time.Now().Add(-1 * time.Hour)
	schedule := &models.ReportSchedule{
		ID:       uuid.New(),
		IsActive: true,
		NextRun:  &pastTime,
	}

	if !schedule.ShouldRun() {
		t.Error("Expected schedule to run")
	}
}

func TestFootprintReportData_Creation(t *testing.T) {
	data := &models.FootprintReportData{
		UserID:            "test-user-123",
		TotalCO2Kg:        decimal.NewFromFloat(100.5),
		TotalCalculations: 10,
		StartDate:         time.Now().AddDate(0, -1, 0),
		EndDate:           time.Now(),
	}

	if data.UserID != "test-user-123" {
		t.Errorf("Expected UserID 'test-user-123', got %s", data.UserID)
	}

	if !data.TotalCO2Kg.Equal(decimal.NewFromFloat(100.5)) {
		t.Errorf("Expected TotalCO2Kg 100.5, got %s", data.TotalCO2Kg)
	}

	if data.TotalCalculations != 10 {
		t.Errorf("Expected TotalCalculations 10, got %d", data.TotalCalculations)
	}
}
