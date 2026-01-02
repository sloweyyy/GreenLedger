package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/sloweyyy/GreenLedger/services/reporting/internal/models"
	"github.com/sloweyyy/GreenLedger/shared/database"
)

// PostgresReportingRepository implements ReportingRepository using PostgresDB
type PostgresReportingRepository struct {
	calculatorDB *database.PostgresDB
	trackerDB    *database.PostgresDB
	walletDB     *database.PostgresDB
}

// NewPostgresReportingRepository creates a new postgres repository
func NewPostgresReportingRepository(
	calculatorDB *database.PostgresDB,
	trackerDB *database.PostgresDB,
	walletDB *database.PostgresDB,
) *PostgresReportingRepository {
	return &PostgresReportingRepository{
		calculatorDB: calculatorDB,
		trackerDB:    trackerDB,
		walletDB:     walletDB,
	}
}

// GetTotalFootprint gets total CO2 and calculation count
func (r *PostgresReportingRepository) GetTotalFootprint(ctx context.Context, userID string, startDate, endDate time.Time) (decimal.Decimal, int64, error) {
	var totalCO2 sql.NullFloat64
	var totalCalculations sql.NullInt64

	query := `
		SELECT
			COALESCE(SUM(total_co2_kg), 0) as total_co2,
			COUNT(*) as total_calculations
		FROM calculations
		WHERE user_id = $1 AND created_at >= $2 AND created_at <= $3
	`

	err := r.calculatorDB.WithContext(ctx).Raw(query, userID, startDate, endDate).
		Row().Scan(&totalCO2, &totalCalculations)
	if err != nil {
		return decimal.Zero, 0, err
	}

	return decimal.NewFromFloat(totalCO2.Float64), totalCalculations.Int64, nil
}

// GetFootprintByActivityType gets footprint breakdown by activity type
func (r *PostgresReportingRepository) GetFootprintByActivityType(ctx context.Context, userID string, startDate, endDate time.Time) ([]ActivitySummary, error) {
	query := `
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

	rows, err := r.calculatorDB.WithContext(ctx).Raw(query, userID, startDate, endDate).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var summaries []ActivitySummary
	for rows.Next() {
		var activityType string
		var totalCO2 sql.NullFloat64
		var count sql.NullInt64

		if err := rows.Scan(&activityType, &totalCO2, &count); err != nil {
			continue
		}

		summaries = append(summaries, ActivitySummary{
			ActivityType: activityType,
			TotalCO2:     decimal.NewFromFloat(totalCO2.Float64),
			Count:        count.Int64,
		})
	}

	return summaries, nil
}

// GetFootprintByMonth gets footprint breakdown by month
func (r *PostgresReportingRepository) GetFootprintByMonth(ctx context.Context, userID string, startDate, endDate time.Time) ([]MonthlyData, error) {
	query := `
		SELECT
			DATE_TRUNC('month', created_at) as month,
			COALESCE(SUM(total_co2_kg), 0) as total_co2
		FROM calculations
		WHERE user_id = $1 AND created_at >= $2 AND created_at <= $3
		GROUP BY DATE_TRUNC('month', created_at)
		ORDER BY month
	`

	rows, err := r.calculatorDB.WithContext(ctx).Raw(query, userID, startDate, endDate).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var data []MonthlyData
	for rows.Next() {
		var month time.Time
		var totalCO2 sql.NullFloat64

		if err := rows.Scan(&month, &totalCO2); err != nil {
			continue
		}

		data = append(data, MonthlyData{
			Month: month,
			Value: decimal.NewFromFloat(totalCO2.Float64),
		})
	}

	return data, nil
}

// GetWalletBalance gets current wallet balance
func (r *PostgresReportingRepository) GetWalletBalance(ctx context.Context, userID string) (available, earned, spent decimal.Decimal, err error) {
	var avail, earn, sp sql.NullFloat64

	query := `
		SELECT
			COALESCE(available_credits, 0) as available_credits,
			COALESCE(total_earned, 0) as total_earned,
			COALESCE(total_spent, 0) as total_spent
		FROM wallets
		WHERE user_id = $1
	`

	err = r.walletDB.WithContext(ctx).Raw(query, userID).
		Row().Scan(&avail, &earn, &sp)
	if err != nil {
		if err == sql.ErrNoRows {
			return decimal.Zero, decimal.Zero, decimal.Zero, nil
		}
		return decimal.Zero, decimal.Zero, decimal.Zero, err
	}

	return decimal.NewFromFloat(avail.Float64), decimal.NewFromFloat(earn.Float64), decimal.NewFromFloat(sp.Float64), nil
}

// GetTransactionsBySource gets transaction breakdown by source
func (r *PostgresReportingRepository) GetTransactionsBySource(ctx context.Context, userID string, startDate, endDate time.Time) ([]TransactionSummary, error) {
	query := `
		SELECT
			COUNT(*) as total_transactions,
			source,
			COALESCE(SUM(CASE WHEN type IN ('credit_earned', 'transfer_in', 'refund', 'bonus') THEN amount ELSE 0 END), 0) as credits_earned
		FROM transactions
		WHERE user_id = $1 AND created_at >= $2 AND created_at <= $3 AND status = 'completed'
		GROUP BY source
	`

	rows, err := r.walletDB.WithContext(ctx).Raw(query, userID, startDate, endDate).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var summaries []TransactionSummary
	for rows.Next() {
		var count sql.NullInt64
		var source sql.NullString
		var creditsEarned sql.NullFloat64

		if err := rows.Scan(&count, &source, &creditsEarned); err != nil {
			continue
		}

		if source.Valid {
			summaries = append(summaries, TransactionSummary{
				TotalCount:    count.Int64,
				Source:        source.String,
				CreditsEarned: decimal.NewFromFloat(creditsEarned.Float64),
			})
		}
	}

	return summaries, nil
}

// GetCreditsByMonth gets credits earned by month
func (r *PostgresReportingRepository) GetCreditsByMonth(ctx context.Context, userID string, startDate, endDate time.Time) ([]MonthlyData, error) {
	query := `
		SELECT
			DATE_TRUNC('month', created_at) as month,
			COALESCE(SUM(CASE WHEN type IN ('credit_earned', 'transfer_in', 'refund', 'bonus') THEN amount ELSE 0 END), 0) as credits_earned
		FROM transactions
		WHERE user_id = $1 AND created_at >= $2 AND created_at <= $3 AND status = 'completed'
		GROUP BY DATE_TRUNC('month', created_at)
		ORDER BY month
	`

	rows, err := r.walletDB.WithContext(ctx).Raw(query, userID, startDate, endDate).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var data []MonthlyData
	for rows.Next() {
		var month time.Time
		var creditsEarned sql.NullFloat64

		if err := rows.Scan(&month, &creditsEarned); err != nil {
			continue
		}

		data = append(data, MonthlyData{
			Month: month,
			Value: decimal.NewFromFloat(creditsEarned.Float64),
		})
	}

	return data, nil
}

// GetTopEarningActivities gets top earning activities from tracker service
func (r *PostgresReportingRepository) GetTopEarningActivities(ctx context.Context, userID string, startDate, endDate time.Time, limit int) ([]models.ActivitySummary, error) {
	if r.trackerDB == nil {
		return nil, nil
	}

	query := `
		SELECT
			at.name as activity_type,
			COUNT(*) as count,
			COALESCE(SUM(ea.credits_earned), 0) as total_credits
		FROM eco_activities ea
		JOIN activity_types at ON ea.activity_type_id = at.id
		WHERE ea.user_id = $1 AND ea.created_at >= $2 AND ea.created_at <= $3 AND ea.is_verified = true
		GROUP BY at.name
		ORDER BY total_credits DESC
		LIMIT $4
	`

	rows, err := r.trackerDB.WithContext(ctx).Raw(query, userID, startDate, endDate, limit).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var summaries []models.ActivitySummary
	for rows.Next() {
		var activityType string
		var count sql.NullInt64
		var totalCredits sql.NullFloat64

		if err := rows.Scan(&activityType, &count, &totalCredits); err != nil {
			continue
		}

		credits := decimal.NewFromFloat(totalCredits.Float64)
		averagePerActivity := decimal.Zero
		if count.Int64 > 0 {
			averagePerActivity = credits.Div(decimal.NewFromInt(count.Int64))
		}

		summaries = append(summaries, models.ActivitySummary{
			ActivityType:       activityType,
			Count:              count.Int64,
			TotalCredits:       credits,
			AveragePerActivity: averagePerActivity,
		})
	}

	return summaries, nil
}

// GetRecentTransactions gets recent transactions
func (r *PostgresReportingRepository) GetRecentTransactions(ctx context.Context, userID string, startDate, endDate time.Time, limit int) ([]models.TransactionSummary, error) {
	query := `
		SELECT id, type, amount, description, created_at
		FROM transactions
		WHERE user_id = $1 AND created_at >= $2 AND created_at <= $3 AND status = 'completed'
		ORDER BY created_at DESC
		LIMIT $4
	`

	rows, err := r.walletDB.WithContext(ctx).Raw(query, userID, startDate, endDate, limit).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []models.TransactionSummary
	for rows.Next() {
		var id string
		var txType string
		var amount sql.NullFloat64
		var description string
		var createdAt time.Time

		if err := rows.Scan(&id, &txType, &amount, &description, &createdAt); err != nil {
			continue
		}

		txID, _ := uuid.Parse(id)
		transactions = append(transactions, models.TransactionSummary{
			ID:          txID,
			Type:        txType,
			Amount:      decimal.NewFromFloat(amount.Float64),
			Description: description,
			CreatedAt:   createdAt,
		})
	}

	return transactions, nil
}

// GetActivityStats gets total activity count and daily breakdown
func (r *PostgresReportingRepository) GetActivityStats(ctx context.Context, userID string, startDate, endDate time.Time) (int64, []DailyActivity, error) {
	if r.trackerDB == nil {
		return 0, nil, nil
	}

	// Combined query to get total count and daily breakdown.
	// We use a single aggregation query to get daily counts, and then derive the
	// total count in code by summing the per-day counts.

	query := `
		SELECT
			DATE_TRUNC('day', created_at) as day,
			COUNT(*) as count
		FROM eco_activities
		WHERE user_id = $1 AND created_at >= $2 AND created_at <= $3
		GROUP BY DATE_TRUNC('day', created_at)
		ORDER BY count DESC, day ASC
	`

	type dailyRow struct {
		Day   time.Time `gorm:"column:day"`
		Count int64     `gorm:"column:count"`
	}

	var rows []dailyRow
	err := r.trackerDB.WithContext(ctx).Raw(query, userID, startDate, endDate).Scan(&rows).Error
	if err != nil {
		return 0, nil, err
	}

	var totalCount int64
	var dailyActivities []DailyActivity

	for _, row := range rows {
		totalCount += row.Count
		dailyActivities = append(dailyActivities, DailyActivity{
			Day:   row.Day,
			Count: row.Count,
		})
	}

	return totalCount, dailyActivities, nil
}
