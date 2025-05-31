package service

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/sloweyyy/GreenLedger/services/reporting/internal/models"
	"github.com/sloweyyy/GreenLedger/shared/database"
	"github.com/sloweyyy/GreenLedger/shared/logger"
)

// DatabaseDataCollector implements DataCollector using database queries
type DatabaseDataCollector struct {
	calculatorDB *database.PostgresDB
	trackerDB    *database.PostgresDB
	walletDB     *database.PostgresDB
	logger       *logger.Logger
}

// NewDatabaseDataCollector creates a new database data collector
func NewDatabaseDataCollector(
	calculatorDB *database.PostgresDB,
	trackerDB *database.PostgresDB,
	walletDB *database.PostgresDB,
	logger *logger.Logger,
) *DatabaseDataCollector {
	return &DatabaseDataCollector{
		calculatorDB: calculatorDB,
		trackerDB:    trackerDB,
		walletDB:     walletDB,
		logger:       logger,
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
	var totalCO2 sql.NullFloat64
	var totalCalculations sql.NullInt64

	query := `
		SELECT 
			COALESCE(SUM(total_co2_kg), 0) as total_co2,
			COUNT(*) as total_calculations
		FROM calculations 
		WHERE user_id = $1 AND created_at >= $2 AND created_at <= $3
	`

	err := c.calculatorDB.WithContext(ctx).Raw(query, userID, startDate, endDate).
		Row().Scan(&totalCO2, &totalCalculations)
	if err != nil {
		return nil, fmt.Errorf("failed to get total footprint: %w", err)
	}

	data.TotalCO2Kg = decimal.NewFromFloat(totalCO2.Float64)
	data.TotalCalculations = totalCalculations.Int64

	// Calculate average per day
	days := endDate.Sub(startDate).Hours() / 24
	if days > 0 {
		data.AveragePerDay = data.TotalCO2Kg.Div(decimal.NewFromFloat(days))
	}

	// Get CO2 by activity type
	activityQuery := `
		SELECT 
			a.activity_type,
			COALESCE(SUM(a.co2_kg), 0) as total_co2,
			COUNT(*) as count
		FROM activities a
		JOIN calculations c ON a.calculation_id = c.id
		WHERE c.user_id = $1 AND c.created_at >= $2 AND c.created_at <= $3
		GROUP BY a.activity_type
		ORDER BY total_co2 DESC
	`

	rows, err := c.calculatorDB.WithContext(ctx).Raw(activityQuery, userID, startDate, endDate).Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to get activity breakdown: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var activityType string
		var totalCO2 sql.NullFloat64
		var count sql.NullInt64

		if err := rows.Scan(&activityType, &totalCO2, &count); err != nil {
			continue
		}

		co2Amount := decimal.NewFromFloat(totalCO2.Float64)
		data.ByActivityType[activityType] = co2Amount

		// Add to top activities
		if len(data.TopActivities) < 10 {
			data.TopActivities = append(data.TopActivities, models.ActivitySummary{
				ActivityType:       activityType,
				Count:              count.Int64,
				TotalCO2:           co2Amount,
				AveragePerActivity: co2Amount.Div(decimal.NewFromInt(count.Int64)),
			})
		}
	}

	// Get CO2 by month
	monthQuery := `
		SELECT 
			DATE_TRUNC('month', created_at) as month,
			COALESCE(SUM(total_co2_kg), 0) as total_co2
		FROM calculations 
		WHERE user_id = $1 AND created_at >= $2 AND created_at <= $3
		GROUP BY DATE_TRUNC('month', created_at)
		ORDER BY month
	`

	monthRows, err := c.calculatorDB.WithContext(ctx).Raw(monthQuery, userID, startDate, endDate).Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to get monthly breakdown: %w", err)
	}
	defer monthRows.Close()

	for monthRows.Next() {
		var month time.Time
		var totalCO2 sql.NullFloat64

		if err := monthRows.Scan(&month, &totalCO2); err != nil {
			continue
		}

		monthKey := month.Format("2006-01")
		data.ByMonth[monthKey] = decimal.NewFromFloat(totalCO2.Float64)
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
	var availableCredits sql.NullFloat64
	var totalEarned sql.NullFloat64
	var totalSpent sql.NullFloat64

	walletQuery := `
		SELECT 
			COALESCE(available_credits, 0) as available_credits,
			COALESCE(total_earned, 0) as total_earned,
			COALESCE(total_spent, 0) as total_spent
		FROM wallets 
		WHERE user_id = $1
	`

	err := c.walletDB.WithContext(ctx).Raw(walletQuery, userID).
		Row().Scan(&availableCredits, &totalEarned, &totalSpent)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to get wallet data: %w", err)
	}

	data.CurrentBalance = decimal.NewFromFloat(availableCredits.Float64)
	data.TotalCreditsEarned = decimal.NewFromFloat(totalEarned.Float64)
	data.TotalCreditsSpent = decimal.NewFromFloat(totalSpent.Float64)

	// Get transaction count and breakdown by source
	transactionQuery := `
		SELECT 
			COUNT(*) as total_transactions,
			source,
			COALESCE(SUM(CASE WHEN type IN ('credit_earned', 'transfer_in', 'refund', 'bonus') THEN amount ELSE 0 END), 0) as credits_earned
		FROM transactions 
		WHERE user_id = $1 AND created_at >= $2 AND created_at <= $3 AND status = 'completed'
		GROUP BY source
	`

	rows, err := c.walletDB.WithContext(ctx).Raw(transactionQuery, userID, startDate, endDate).Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction data: %w", err)
	}
	defer rows.Close()

	var totalTransactions int64
	for rows.Next() {
		var count sql.NullInt64
		var source sql.NullString
		var creditsEarned sql.NullFloat64

		if err := rows.Scan(&count, &source, &creditsEarned); err != nil {
			continue
		}

		totalTransactions += count.Int64
		if source.Valid {
			data.BySource[source.String] = decimal.NewFromFloat(creditsEarned.Float64)
		}
	}
	data.TotalTransactions = totalTransactions

	// Get credits by month
	monthQuery := `
		SELECT 
			DATE_TRUNC('month', created_at) as month,
			COALESCE(SUM(CASE WHEN type IN ('credit_earned', 'transfer_in', 'refund', 'bonus') THEN amount ELSE 0 END), 0) as credits_earned
		FROM transactions 
		WHERE user_id = $1 AND created_at >= $2 AND created_at <= $3 AND status = 'completed'
		GROUP BY DATE_TRUNC('month', created_at)
		ORDER BY month
	`

	monthRows, err := c.walletDB.WithContext(ctx).Raw(monthQuery, userID, startDate, endDate).Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to get monthly credits: %w", err)
	}
	defer monthRows.Close()

	for monthRows.Next() {
		var month time.Time
		var creditsEarned sql.NullFloat64

		if err := monthRows.Scan(&month, &creditsEarned); err != nil {
			continue
		}

		monthKey := month.Format("2006-01")
		data.ByMonth[monthKey] = decimal.NewFromFloat(creditsEarned.Float64)
	}

	// Get top earning activities from tracker service
	if c.trackerDB != nil {
		activityQuery := `
			SELECT 
				at.name as activity_type,
				COUNT(*) as count,
				COALESCE(SUM(ea.credits_earned), 0) as total_credits
			FROM eco_activities ea
			JOIN activity_types at ON ea.activity_type_id = at.id
			WHERE ea.user_id = $1 AND ea.created_at >= $2 AND ea.created_at <= $3 AND ea.is_verified = true
			GROUP BY at.name
			ORDER BY total_credits DESC
			LIMIT 10
		`

		activityRows, err := c.trackerDB.WithContext(ctx).Raw(activityQuery, userID, startDate, endDate).Rows()
		if err == nil {
			defer activityRows.Close()

			for activityRows.Next() {
				var activityType string
				var count sql.NullInt64
				var totalCredits sql.NullFloat64

				if err := activityRows.Scan(&activityType, &count, &totalCredits); err != nil {
					continue
				}

				credits := decimal.NewFromFloat(totalCredits.Float64)
				data.TopEarningActivities = append(data.TopEarningActivities, models.ActivitySummary{
					ActivityType:       activityType,
					Count:              count.Int64,
					TotalCredits:       credits,
					AveragePerActivity: credits.Div(decimal.NewFromInt(count.Int64)),
				})
			}
		}
	}

	// Get recent transactions
	recentQuery := `
		SELECT id, type, amount, description, created_at
		FROM transactions 
		WHERE user_id = $1 AND created_at >= $2 AND created_at <= $3 AND status = 'completed'
		ORDER BY created_at DESC
		LIMIT 20
	`

	recentRows, err := c.walletDB.WithContext(ctx).Raw(recentQuery, userID, startDate, endDate).Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to get recent transactions: %w", err)
	}
	defer recentRows.Close()

	for recentRows.Next() {
		var id string
		var txType string
		var amount sql.NullFloat64
		var description string
		var createdAt time.Time

		if err := recentRows.Scan(&id, &txType, &amount, &description, &createdAt); err != nil {
			continue
		}

		txID, _ := uuid.Parse(id)
		data.RecentTransactions = append(data.RecentTransactions, models.TransactionSummary{
			ID:          txID,
			Type:        txType,
			Amount:      decimal.NewFromFloat(amount.Float64),
			Description: description,
			CreatedAt:   createdAt,
		})
	}

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

	// Get activity count from tracker service
	var totalActivities int64
	if c.trackerDB != nil {
		activityCountQuery := `
			SELECT COUNT(*) 
			FROM eco_activities 
			WHERE user_id = $1 AND created_at >= $2 AND created_at <= $3
		`
		c.trackerDB.WithContext(ctx).Raw(activityCountQuery, userID, startDate, endDate).
			Row().Scan(&totalActivities)
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
	}

	// TODO: Calculate most/least active days
	data.MostActiveDay = startDate
	data.LeastActiveDay = endDate

	return data, nil
}
