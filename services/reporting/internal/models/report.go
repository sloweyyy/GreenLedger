package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// Report represents a generated report
type Report struct {
	ID          uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID      string     `gorm:"not null;index" json:"user_id"`
	Type        string     `gorm:"not null;index" json:"type"`
	Title       string     `gorm:"not null" json:"title"`
	Description string     `json:"description"`
	Format      string     `gorm:"not null" json:"format"`
	Status      string     `gorm:"not null;default:'pending'" json:"status"`
	FilePath    string     `json:"file_path"`
	FileSize    int64      `json:"file_size"`
	Parameters  string     `gorm:"type:jsonb" json:"parameters"`
	StartDate   time.Time  `gorm:"not null" json:"start_date"`
	EndDate     time.Time  `gorm:"not null" json:"end_date"`
	GeneratedAt *time.Time `json:"generated_at"`
	ExpiresAt   *time.Time `json:"expires_at"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// ReportSchedule represents a scheduled report
type ReportSchedule struct {
	ID          uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID      string     `gorm:"not null;index" json:"user_id"`
	Name        string     `gorm:"not null" json:"name"`
	Description string     `json:"description"`
	ReportType  string     `gorm:"not null" json:"report_type"`
	Format      string     `gorm:"not null" json:"format"`
	Schedule    string     `gorm:"not null" json:"schedule"` // cron expression
	Parameters  string     `gorm:"type:jsonb" json:"parameters"`
	IsActive    bool       `gorm:"default:true" json:"is_active"`
	LastRun     *time.Time `json:"last_run"`
	NextRun     *time.Time `json:"next_run"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`

	// Relationships
	Reports []Report `gorm:"foreignKey:UserID;references:UserID" json:"reports,omitempty"`
}

// ReportTemplate represents a report template
type ReportTemplate struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Name        string    `gorm:"uniqueIndex;not null" json:"name"`
	Type        string    `gorm:"not null;index" json:"type"`
	Title       string    `gorm:"not null" json:"title"`
	Description string    `json:"description"`
	Template    string    `gorm:"type:text;not null" json:"template"`
	Parameters  string    `gorm:"type:jsonb" json:"parameters"`
	IsActive    bool      `gorm:"default:true" json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ReportData represents aggregated data for reports
type ReportData struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ReportID  uuid.UUID `gorm:"type:uuid;not null;index" json:"report_id"`
	DataType  string    `gorm:"not null" json:"data_type"`
	Data      string    `gorm:"type:jsonb;not null" json:"data"`
	CreatedAt time.Time `json:"created_at"`

	// Relationship
	Report Report `gorm:"foreignKey:ReportID" json:"-"`
}

// BeforeCreate hooks
func (r *Report) BeforeCreate(tx *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return nil
}

func (rs *ReportSchedule) BeforeCreate(tx *gorm.DB) error {
	if rs.ID == uuid.Nil {
		rs.ID = uuid.New()
	}
	return nil
}

func (rt *ReportTemplate) BeforeCreate(tx *gorm.DB) error {
	if rt.ID == uuid.Nil {
		rt.ID = uuid.New()
	}
	return nil
}

func (rd *ReportData) BeforeCreate(tx *gorm.DB) error {
	if rd.ID == uuid.Nil {
		rd.ID = uuid.New()
	}
	return nil
}

// Table names
func (Report) TableName() string         { return "reports" }
func (ReportSchedule) TableName() string { return "report_schedules" }
func (ReportTemplate) TableName() string { return "report_templates" }
func (ReportData) TableName() string     { return "report_data" }

// Report types
const (
	ReportTypeFootprint    = "footprint"
	ReportTypeCredits      = "credits"
	ReportTypeActivities   = "activities"
	ReportTypeTransactions = "transactions"
	ReportTypeSummary      = "summary"
	ReportTypeComparison   = "comparison"
	ReportTypeLeaderboard  = "leaderboard"
)

// Report formats
const (
	ReportFormatPDF  = "pdf"
	ReportFormatJSON = "json"
	ReportFormatCSV  = "csv"
	ReportFormatXLSX = "xlsx"
)

// Report statuses
const (
	ReportStatusPending    = "pending"
	ReportStatusGenerating = "generating"
	ReportStatusCompleted  = "completed"
	ReportStatusFailed     = "failed"
	ReportStatusExpired    = "expired"
)

// Helper methods for Report
func (r *Report) IsCompleted() bool {
	return r.Status == ReportStatusCompleted
}

func (r *Report) IsPending() bool {
	return r.Status == ReportStatusPending
}

func (r *Report) IsExpired() bool {
	return r.ExpiresAt != nil && time.Now().After(*r.ExpiresAt)
}

// Helper methods for ReportSchedule
func (rs *ReportSchedule) ShouldRun() bool {
	if !rs.IsActive {
		return false
	}
	if rs.NextRun == nil {
		return true
	}
	return time.Now().After(*rs.NextRun)
}

// FootprintReportData represents carbon footprint report data
type FootprintReportData struct {
	UserID              string                     `json:"user_id"`
	TotalCO2Kg          decimal.Decimal            `json:"total_co2_kg"`
	TotalCalculations   int64                      `json:"total_calculations"`
	AveragePerDay       decimal.Decimal            `json:"average_per_day"`
	ByActivityType      map[string]decimal.Decimal `json:"by_activity_type"`
	ByMonth             map[string]decimal.Decimal `json:"by_month"`
	TopActivities       []ActivitySummary          `json:"top_activities"`
	ComparisonToAverage decimal.Decimal            `json:"comparison_to_average"`
	StartDate           time.Time                  `json:"start_date"`
	EndDate             time.Time                  `json:"end_date"`
}

// CreditsReportData represents carbon credits report data
type CreditsReportData struct {
	UserID               string                     `json:"user_id"`
	TotalCreditsEarned   decimal.Decimal            `json:"total_credits_earned"`
	TotalCreditsSpent    decimal.Decimal            `json:"total_credits_spent"`
	CurrentBalance       decimal.Decimal            `json:"current_balance"`
	TotalTransactions    int64                      `json:"total_transactions"`
	BySource             map[string]decimal.Decimal `json:"by_source"`
	ByMonth              map[string]decimal.Decimal `json:"by_month"`
	TopEarningActivities []ActivitySummary          `json:"top_earning_activities"`
	RecentTransactions   []TransactionSummary       `json:"recent_transactions"`
	StartDate            time.Time                  `json:"start_date"`
	EndDate              time.Time                  `json:"end_date"`
}

// ActivitySummary represents a summary of an activity
type ActivitySummary struct {
	ActivityType       string          `json:"activity_type"`
	Count              int64           `json:"count"`
	TotalCO2           decimal.Decimal `json:"total_co2"`
	TotalCredits       decimal.Decimal `json:"total_credits"`
	AveragePerActivity decimal.Decimal `json:"average_per_activity"`
}

// TransactionSummary represents a summary of a transaction
type TransactionSummary struct {
	ID          uuid.UUID       `json:"id"`
	Type        string          `json:"type"`
	Amount      decimal.Decimal `json:"amount"`
	Description string          `json:"description"`
	CreatedAt   time.Time       `json:"created_at"`
}

// SummaryReportData represents overall summary report data
type SummaryReportData struct {
	UserID               string          `json:"user_id"`
	TotalCO2Kg           decimal.Decimal `json:"total_co2_kg"`
	TotalCreditsEarned   decimal.Decimal `json:"total_credits_earned"`
	TotalCreditsSpent    decimal.Decimal `json:"total_credits_spent"`
	CurrentBalance       decimal.Decimal `json:"current_balance"`
	TotalActivities      int64           `json:"total_activities"`
	TotalCalculations    int64           `json:"total_calculations"`
	TotalTransactions    int64           `json:"total_transactions"`
	AverageCO2PerDay     decimal.Decimal `json:"average_co2_per_day"`
	AverageCreditsPerDay decimal.Decimal `json:"average_credits_per_day"`
	MostActiveDay        time.Time       `json:"most_active_day"`
	LeastActiveDay       time.Time       `json:"least_active_day"`
	StartDate            time.Time       `json:"start_date"`
	EndDate              time.Time       `json:"end_date"`
}

// ReportStats represents report statistics for a user
type ReportStats struct {
	UserID           string `json:"user_id"`
	TotalReports     int64  `json:"total_reports"`
	CompletedReports int64  `json:"completed_reports"`
	PendingReports   int64  `json:"pending_reports"`
	FailedReports    int64  `json:"failed_reports"`
}
