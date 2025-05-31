package service

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/sloweyyy/GreenLedger/services/reporting/internal/models"
	"github.com/sloweyyy/GreenLedger/shared/logger"
	"github.com/jung-kurt/gofpdf"
	"github.com/shopspring/decimal"
	"github.com/wcharczuk/go-chart/v2"
)

// PDFReportRenderer implements ReportRenderer for PDF format
type PDFReportRenderer struct {
	logger *logger.Logger
}

// NewPDFReportRenderer creates a new PDF report renderer
func NewPDFReportRenderer(logger *logger.Logger) *PDFReportRenderer {
	return &PDFReportRenderer{
		logger: logger,
	}
}

// RenderPDF renders a report as PDF
func (r *PDFReportRenderer) RenderPDF(ctx context.Context, reportType string, data interface{}) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 16)

	switch reportType {
	case models.ReportTypeFootprint:
		return r.renderFootprintPDF(pdf, data.(*models.FootprintReportData))
	case models.ReportTypeCredits:
		return r.renderCreditsPDF(pdf, data.(*models.CreditsReportData))
	case models.ReportTypeSummary:
		return r.renderSummaryPDF(pdf, data.(*models.SummaryReportData))
	default:
		return nil, fmt.Errorf("unsupported report type for PDF: %s", reportType)
	}
}

// RenderJSON renders data as JSON
func (r *PDFReportRenderer) RenderJSON(ctx context.Context, data interface{}) ([]byte, error) {
	return json.MarshalIndent(data, "", "  ")
}

// RenderCSV renders data as CSV
func (r *PDFReportRenderer) RenderCSV(ctx context.Context, reportType string, data interface{}) ([]byte, error) {
	var buffer bytes.Buffer
	writer := csv.NewWriter(&buffer)

	switch reportType {
	case models.ReportTypeFootprint:
		return r.renderFootprintCSV(writer, data.(*models.FootprintReportData))
	case models.ReportTypeCredits:
		return r.renderCreditsCSV(writer, data.(*models.CreditsReportData))
	case models.ReportTypeSummary:
		return r.renderSummaryCSV(writer, data.(*models.SummaryReportData))
	default:
		return nil, fmt.Errorf("unsupported report type for CSV: %s", reportType)
	}
}

// renderFootprintPDF renders carbon footprint data as PDF
func (r *PDFReportRenderer) renderFootprintPDF(pdf *gofpdf.Fpdf, data *models.FootprintReportData) ([]byte, error) {
	// Title
	pdf.Cell(190, 10, "Carbon Footprint Report")
	pdf.Ln(15)

	// Report period
	pdf.SetFont("Arial", "", 12)
	pdf.Cell(190, 8, fmt.Sprintf("Period: %s to %s", 
		data.StartDate.Format("2006-01-02"), 
		data.EndDate.Format("2006-01-02")))
	pdf.Ln(10)

	// Summary section
	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(190, 8, "Summary")
	pdf.Ln(10)

	pdf.SetFont("Arial", "", 11)
	pdf.Cell(190, 6, fmt.Sprintf("Total CO2 Emissions: %.2f kg", data.TotalCO2Kg))
	pdf.Ln(6)
	pdf.Cell(190, 6, fmt.Sprintf("Total Calculations: %d", data.TotalCalculations))
	pdf.Ln(6)
	pdf.Cell(190, 6, fmt.Sprintf("Average per Day: %.2f kg", data.AveragePerDay))
	pdf.Ln(15)

	// Activity breakdown
	if len(data.ByActivityType) > 0 {
		pdf.SetFont("Arial", "B", 14)
		pdf.Cell(190, 8, "Breakdown by Activity Type")
		pdf.Ln(10)

		pdf.SetFont("Arial", "", 11)
		for activityType, co2 := range data.ByActivityType {
			pdf.Cell(190, 6, fmt.Sprintf("%s: %.2f kg CO2", activityType, co2))
			pdf.Ln(6)
		}
		pdf.Ln(10)
	}

	// Top activities
	if len(data.TopActivities) > 0 {
		pdf.SetFont("Arial", "B", 14)
		pdf.Cell(190, 8, "Top Activities")
		pdf.Ln(10)

		pdf.SetFont("Arial", "", 11)
		for i, activity := range data.TopActivities {
			if i >= 5 { // Limit to top 5 for PDF
				break
			}
			pdf.Cell(190, 6, fmt.Sprintf("%s: %.2f kg CO2 (%d times)", 
				activity.ActivityType, activity.TotalCO2, activity.Count))
			pdf.Ln(6)
		}
	}

	var buffer bytes.Buffer
	err := pdf.Output(&buffer)
	return buffer.Bytes(), err
}

// renderCreditsPDF renders carbon credits data as PDF
func (r *PDFReportRenderer) renderCreditsPDF(pdf *gofpdf.Fpdf, data *models.CreditsReportData) ([]byte, error) {
	// Title
	pdf.Cell(190, 10, "Carbon Credits Report")
	pdf.Ln(15)

	// Report period
	pdf.SetFont("Arial", "", 12)
	pdf.Cell(190, 8, fmt.Sprintf("Period: %s to %s", 
		data.StartDate.Format("2006-01-02"), 
		data.EndDate.Format("2006-01-02")))
	pdf.Ln(10)

	// Summary section
	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(190, 8, "Summary")
	pdf.Ln(10)

	pdf.SetFont("Arial", "", 11)
	pdf.Cell(190, 6, fmt.Sprintf("Current Balance: %.2f credits", data.CurrentBalance))
	pdf.Ln(6)
	pdf.Cell(190, 6, fmt.Sprintf("Total Earned: %.2f credits", data.TotalCreditsEarned))
	pdf.Ln(6)
	pdf.Cell(190, 6, fmt.Sprintf("Total Spent: %.2f credits", data.TotalCreditsSpent))
	pdf.Ln(6)
	pdf.Cell(190, 6, fmt.Sprintf("Total Transactions: %d", data.TotalTransactions))
	pdf.Ln(15)

	// Credits by source
	if len(data.BySource) > 0 {
		pdf.SetFont("Arial", "B", 14)
		pdf.Cell(190, 8, "Credits by Source")
		pdf.Ln(10)

		pdf.SetFont("Arial", "", 11)
		for source, credits := range data.BySource {
			pdf.Cell(190, 6, fmt.Sprintf("%s: %.2f credits", source, credits))
			pdf.Ln(6)
		}
		pdf.Ln(10)
	}

	// Top earning activities
	if len(data.TopEarningActivities) > 0 {
		pdf.SetFont("Arial", "B", 14)
		pdf.Cell(190, 8, "Top Earning Activities")
		pdf.Ln(10)

		pdf.SetFont("Arial", "", 11)
		for i, activity := range data.TopEarningActivities {
			if i >= 5 { // Limit to top 5 for PDF
				break
			}
			pdf.Cell(190, 6, fmt.Sprintf("%s: %.2f credits (%d times)", 
				activity.ActivityType, activity.TotalCredits, activity.Count))
			pdf.Ln(6)
		}
	}

	var buffer bytes.Buffer
	err := pdf.Output(&buffer)
	return buffer.Bytes(), err
}

// renderSummaryPDF renders summary data as PDF
func (r *PDFReportRenderer) renderSummaryPDF(pdf *gofpdf.Fpdf, data *models.SummaryReportData) ([]byte, error) {
	// Title
	pdf.Cell(190, 10, "Summary Report")
	pdf.Ln(15)

	// Report period
	pdf.SetFont("Arial", "", 12)
	pdf.Cell(190, 8, fmt.Sprintf("Period: %s to %s", 
		data.StartDate.Format("2006-01-02"), 
		data.EndDate.Format("2006-01-02")))
	pdf.Ln(15)

	// Environmental Impact
	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(190, 8, "Environmental Impact")
	pdf.Ln(10)

	pdf.SetFont("Arial", "", 11)
	pdf.Cell(190, 6, fmt.Sprintf("Total CO2 Emissions: %.2f kg", data.TotalCO2Kg))
	pdf.Ln(6)
	pdf.Cell(190, 6, fmt.Sprintf("Average CO2 per Day: %.2f kg", data.AverageCO2PerDay))
	pdf.Ln(6)
	pdf.Cell(190, 6, fmt.Sprintf("Total Calculations: %d", data.TotalCalculations))
	pdf.Ln(15)

	// Carbon Credits
	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(190, 8, "Carbon Credits")
	pdf.Ln(10)

	pdf.SetFont("Arial", "", 11)
	pdf.Cell(190, 6, fmt.Sprintf("Current Balance: %.2f credits", data.CurrentBalance))
	pdf.Ln(6)
	pdf.Cell(190, 6, fmt.Sprintf("Total Earned: %.2f credits", data.TotalCreditsEarned))
	pdf.Ln(6)
	pdf.Cell(190, 6, fmt.Sprintf("Total Spent: %.2f credits", data.TotalCreditsSpent))
	pdf.Ln(6)
	pdf.Cell(190, 6, fmt.Sprintf("Average Credits per Day: %.2f", data.AverageCreditsPerDay))
	pdf.Ln(15)

	// Activity Summary
	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(190, 8, "Activity Summary")
	pdf.Ln(10)

	pdf.SetFont("Arial", "", 11)
	pdf.Cell(190, 6, fmt.Sprintf("Total Eco Activities: %d", data.TotalActivities))
	pdf.Ln(6)
	pdf.Cell(190, 6, fmt.Sprintf("Total Transactions: %d", data.TotalTransactions))
	pdf.Ln(6)

	var buffer bytes.Buffer
	err := pdf.Output(&buffer)
	return buffer.Bytes(), err
}

// renderFootprintCSV renders carbon footprint data as CSV
func (r *PDFReportRenderer) renderFootprintCSV(writer *csv.Writer, data *models.FootprintReportData) ([]byte, error) {
	// Write headers
	writer.Write([]string{"Metric", "Value", "Unit"})
	
	// Write summary data
	writer.Write([]string{"Total CO2", data.TotalCO2Kg.String(), "kg"})
	writer.Write([]string{"Total Calculations", strconv.FormatInt(data.TotalCalculations, 10), "count"})
	writer.Write([]string{"Average per Day", data.AveragePerDay.String(), "kg/day"})
	
	// Write empty row
	writer.Write([]string{})
	
	// Write activity breakdown
	writer.Write([]string{"Activity Type", "CO2 Emissions", "Unit"})
	for activityType, co2 := range data.ByActivityType {
		writer.Write([]string{activityType, co2.String(), "kg"})
	}

	writer.Flush()
	return []byte{}, writer.Error()
}

// renderCreditsCSV renders carbon credits data as CSV
func (r *PDFReportRenderer) renderCreditsCSV(writer *csv.Writer, data *models.CreditsReportData) ([]byte, error) {
	// Write headers
	writer.Write([]string{"Metric", "Value", "Unit"})
	
	// Write summary data
	writer.Write([]string{"Current Balance", data.CurrentBalance.String(), "credits"})
	writer.Write([]string{"Total Earned", data.TotalCreditsEarned.String(), "credits"})
	writer.Write([]string{"Total Spent", data.TotalCreditsSpent.String(), "credits"})
	writer.Write([]string{"Total Transactions", strconv.FormatInt(data.TotalTransactions, 10), "count"})
	
	// Write empty row
	writer.Write([]string{})
	
	// Write credits by source
	writer.Write([]string{"Source", "Credits Earned", "Unit"})
	for source, credits := range data.BySource {
		writer.Write([]string{source, credits.String(), "credits"})
	}

	writer.Flush()
	return []byte{}, writer.Error()
}

// renderSummaryCSV renders summary data as CSV
func (r *PDFReportRenderer) renderSummaryCSV(writer *csv.Writer, data *models.SummaryReportData) ([]byte, error) {
	// Write headers
	writer.Write([]string{"Category", "Metric", "Value", "Unit"})
	
	// Environmental Impact
	writer.Write([]string{"Environmental", "Total CO2", data.TotalCO2Kg.String(), "kg"})
	writer.Write([]string{"Environmental", "Average CO2 per Day", data.AverageCO2PerDay.String(), "kg/day"})
	writer.Write([]string{"Environmental", "Total Calculations", strconv.FormatInt(data.TotalCalculations, 10), "count"})
	
	// Carbon Credits
	writer.Write([]string{"Credits", "Current Balance", data.CurrentBalance.String(), "credits"})
	writer.Write([]string{"Credits", "Total Earned", data.TotalCreditsEarned.String(), "credits"})
	writer.Write([]string{"Credits", "Total Spent", data.TotalCreditsSpent.String(), "credits"})
	writer.Write([]string{"Credits", "Average per Day", data.AverageCreditsPerDay.String(), "credits/day"})
	
	// Activities
	writer.Write([]string{"Activities", "Total Eco Activities", strconv.FormatInt(data.TotalActivities, 10), "count"})
	writer.Write([]string{"Activities", "Total Transactions", strconv.FormatInt(data.TotalTransactions, 10), "count"})

	writer.Flush()
	return []byte{}, writer.Error()
}
