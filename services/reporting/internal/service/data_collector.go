package service

import (
	"context"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"github.com/sloweyyy/GreenLedger/services/reporting/internal/models"
	"github.com/sloweyyy/GreenLedger/services/reporting/internal/repository"
	"github.com/sloweyyy/GreenLedger/shared/database"
	"github.com/sloweyyy/GreenLedger/shared/logger"
)

// DatabaseDataCollector implements DataCollector using database queries
type DatabaseDataCollector struct {
	repo   repository.ReportingRepository
	logger *logger.Logger
}

// NewDatabaseDataCollector creates a new database data collector
func NewDatabaseDataCollector(
	calculatorDB *database.PostgresDB,
	trackerDB *database.PostgresDB,
	walletDB *database.PostgresDB,
	logger *logger.Logger,
) *DatabaseDataCollector {
	return &DatabaseDataCollector{
		repo:   repository.NewPostgresReportingRepository(calculatorDB, trackerDB, walletDB),
		logger: logger,
	}
}

// NewDatabaseDataCollectorWithRepo creates a new database data collector with injected repository
func NewDatabaseDataCollectorWithRepo(
	repo repository.ReportingRepository,
	logger *logger.Logger,
) *DatabaseDataCollector {
	return &DatabaseDataCollector{
		repo:   repo,
		logger: logger,
	}
}

// CollectFootprintData collects carbon footprint data for a user
func (c *DatabaseDataCollector) CollectFootprintData(ctx context.Context, userID string, startDate, endDate time.Time) (*models.FootprintReportData, error) {
	c.logger.LogInfo(ctx, "collecting footprint data",
		logger.String("user_id", userID),
		logger.String("start_date", startDate.Format("2006-01-02")),
		logger.String("end_date", endDate.Format("2006-01-02")))

	data := &models.FootprintReportData{
		UserID:         userID,
		StartDate:      startDate,
		EndDate:        endDate,
		ByActivityType: make(map[string]decimal.Decimal),
		ByMonth:        make(map[string]decimal.Decimal),
		TopActivities:  make([]models.ActivitySummary, 0),
	}

	// Get total CO2 and calculation count
	totalCO2, totalCalculations, err := c.repo.GetTotalFootprint(ctx, userID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get total footprint: %w", err)
	}

	data.TotalCO2Kg = totalCO2
	data.TotalCalculations = totalCalculations

	// Calculate average per day
	days := endDate.Sub(startDate).Hours() / 24
	if days > 0 {
		data.AveragePerDay = data.TotalCO2Kg.Div(decimal.NewFromFloat(days))
	}

	// Get CO2 by activity type
	activitySummaries, err := c.repo.GetFootprintByActivityType(ctx, userID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get activity breakdown: %w", err)
	}

	for _, summary := range activitySummaries {
		data.ByActivityType[summary.ActivityType] = summary.TotalCO2

		// Add to top activities
		if len(data.TopActivities) < 10 {
			averagePerActivity := decimal.Zero
			if summary.Count > 0 {
				averagePerActivity = summary.TotalCO2.Div(decimal.NewFromInt(summary.Count))
			}

			data.TopActivities = append(data.TopActivities, models.ActivitySummary{
				ActivityType:       summary.ActivityType,
				Count:              summary.Count,
				TotalCO2:           summary.TotalCO2,
				AveragePerActivity: averagePerActivity,
			})
		}
	}

	// Get CO2 by month
	monthlyData, err := c.repo.GetFootprintByMonth(ctx, userID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get monthly breakdown: %w", err)
	}

	for _, md := range monthlyData {
		monthKey := md.Month.Format("2006-01")
		data.ByMonth[monthKey] = md.Value
	}

	// TODO: Calculate comparison to average (would need global statistics)
	data.ComparisonToAverage = decimal.Zero

	return data, nil
}

// CollectCreditsData collects carbon credits data for a user
func (c *DatabaseDataCollector) CollectCreditsData(ctx context.Context, userID string, startDate, endDate time.Time) (*models.CreditsReportData, error) {
	c.logger.LogInfo(ctx, "collecting credits data",
		logger.String("user_id", userID))

	data := &models.CreditsReportData{
		UserID:               userID,
		StartDate:            startDate,
		EndDate:              endDate,
		BySource:             make(map[string]decimal.Decimal),
		ByMonth:              make(map[string]decimal.Decimal),
		TopEarningActivities: make([]models.ActivitySummary, 0),
		RecentTransactions:   make([]models.TransactionSummary, 0),
	}

	// Get current wallet balance
	avail, earn, spent, err := c.repo.GetWalletBalance(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get wallet data: %w", err)
	}

	data.CurrentBalance = avail
	data.TotalCreditsEarned = earn
	data.TotalCreditsSpent = spent

	// Get transaction count and breakdown by source
	txSummaries, err := c.repo.GetTransactionsBySource(ctx, userID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction data: %w", err)
	}

	var totalTransactions int64
	for _, summary := range txSummaries {
		totalTransactions += summary.TotalCount
		data.BySource[summary.Source] = summary.CreditsEarned
	}
	data.TotalTransactions = totalTransactions

	// Get credits by month
	monthlyData, err := c.repo.GetCreditsByMonth(ctx, userID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get monthly credits: %w", err)
	}

	for _, md := range monthlyData {
		monthKey := md.Month.Format("2006-01")
		data.ByMonth[monthKey] = md.Value
	}

	// Get top earning activities from tracker service
	topActivities, err := c.repo.GetTopEarningActivities(ctx, userID, startDate, endDate, 10)
	if err == nil && topActivities != nil {
		data.TopEarningActivities = topActivities
	} else if err != nil {
		c.logger.LogError(ctx, "failed to get top earning activities", err,
			logger.String("user_id", userID))
	}

	// Get recent transactions
	recentTx, err := c.repo.GetRecentTransactions(ctx, userID, startDate, endDate, 20)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent transactions: %w", err)
	}
	data.RecentTransactions = recentTx

	return data, nil
}

// CollectSummaryData collects summary data for a user
func (c *DatabaseDataCollector) CollectSummaryData(ctx context.Context, userID string, startDate, endDate time.Time) (*models.SummaryReportData, error) {
	c.logger.LogInfo(ctx, "collecting summary data",
		logger.String("user_id", userID))

	// Collect footprint and credits data
	footprintData, err := c.CollectFootprintData(ctx, userID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to collect footprint data: %w", err)
	}

	creditsData, err := c.CollectCreditsData(ctx, userID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to collect credits data: %w", err)
	}

	// Get activity count and daily stats from tracker service
	totalActivities, dailyActivities, err := c.repo.GetActivityStats(ctx, userID, startDate, endDate)
	if err != nil {
		// Log error but continue with 0 activities
		c.logger.LogError(ctx, "failed to get activity stats", err,
			logger.String("user_id", userID))
	}

	// Calculate averages
	days := endDate.Sub(startDate).Hours() / 24
	var averageCO2PerDay, averageCreditsPerDay decimal.Decimal
	if days > 0 {
		averageCO2PerDay = footprintData.TotalCO2Kg.Div(decimal.NewFromFloat(days))
		averageCreditsPerDay = creditsData.TotalCreditsEarned.Div(decimal.NewFromFloat(days))
	}

	data := &models.SummaryReportData{
		UserID:               userID,
		TotalCO2Kg:           footprintData.TotalCO2Kg,
		TotalCreditsEarned:   creditsData.TotalCreditsEarned,
		TotalCreditsSpent:    creditsData.TotalCreditsSpent,
		CurrentBalance:       creditsData.CurrentBalance,
		TotalActivities:      totalActivities,
		TotalCalculations:    footprintData.TotalCalculations,
		TotalTransactions:    creditsData.TotalTransactions,
		AverageCO2PerDay:     averageCO2PerDay,
		AverageCreditsPerDay: averageCreditsPerDay,
		StartDate:            startDate,
		EndDate:              endDate,
		MostActiveDay:        startDate,
		LeastActiveDay:       endDate,
	}

	// Calculate most/least active days from daily stats.
	// Note: dailyActivities only includes days where at least one activity occurred.
	// When there is no activity at all in the period, dailyActivities will be empty
	// and the defaults (Start/End date) set above will be used.
	//
	// As a consequence, both MostActiveDay and LeastActiveDay are calculated
	// *only among days with at least one activity*. Days with zero activities
	// are not considered when determining these values.
	if len(dailyActivities) > 0 {
		// Most active day is the first one (ordered by count DESC).
		// In case of a tie, the sort order 'day ASC' ensures the earliest day is picked.
		data.MostActiveDay = dailyActivities[0].Day

		// Least active day is the last one (ordered by count DESC).
		// This represents the least active day *among days with activity*.
		// In case of a tie for the lowest count, the sort order 'day ASC' means
		// the latest day with that count (the last element) is selected.
		data.LeastActiveDay = dailyActivities[len(dailyActivities)-1].Day
	}

	return data, nil
}
