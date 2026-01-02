package repository

import (
	"context"
	"time"

	"github.com/shopspring/decimal"
	"github.com/sloweyyy/GreenLedger/services/reporting/internal/models"
)

// DailyActivity represents aggregated activity count for a day
type DailyActivity struct {
	Day   time.Time
	Count int64
}

// ActivitySummary represents aggregated activity data by type
type ActivitySummary struct {
	ActivityType string
	TotalCO2     decimal.Decimal
	Count        int64
}

// MonthlyData represents aggregated data for a month
type MonthlyData struct {
	Month time.Time
	Value decimal.Decimal
}

// TransactionSummary represents transaction aggregation
type TransactionSummary struct {
	TotalCount    int64
	Source        string
	CreditsEarned decimal.Decimal
}

// ReportingRepository defines the interface for data access
type ReportingRepository interface {
	// Calculation/Footprint methods
	GetTotalFootprint(ctx context.Context, userID string, startDate, endDate time.Time) (decimal.Decimal, int64, error)
	GetFootprintByActivityType(ctx context.Context, userID string, startDate, endDate time.Time) ([]ActivitySummary, error)
	GetFootprintByMonth(ctx context.Context, userID string, startDate, endDate time.Time) ([]MonthlyData, error)

	// Wallet/Credits methods
	GetWalletBalance(ctx context.Context, userID string) (available, earned, spent decimal.Decimal, err error)
	GetTransactionsBySource(ctx context.Context, userID string, startDate, endDate time.Time) ([]TransactionSummary, error)
	GetCreditsByMonth(ctx context.Context, userID string, startDate, endDate time.Time) ([]MonthlyData, error)
	GetRecentTransactions(ctx context.Context, userID string, startDate, endDate time.Time, limit int) ([]models.TransactionSummary, error)

	// Tracker/Activity methods
	GetTopEarningActivities(ctx context.Context, userID string, startDate, endDate time.Time, limit int) ([]models.ActivitySummary, error)
	GetActivityStats(ctx context.Context, userID string, startDate, endDate time.Time) (totalCount int64, dailyStats []DailyActivity, err error)
}
