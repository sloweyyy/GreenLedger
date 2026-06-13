package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/sloweyyy/GreenLedger/services/reporting/internal/models"
	"github.com/sloweyyy/GreenLedger/services/reporting/internal/repository"
	"github.com/sloweyyy/GreenLedger/shared/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockReportingRepository is a mock implementation of ReportingRepository
type MockReportingRepository struct {
	mock.Mock
}

func (m *MockReportingRepository) GetTotalFootprint(ctx context.Context, userID string, startDate, endDate time.Time) (decimal.Decimal, int64, error) {
	args := m.Called(ctx, userID, startDate, endDate)
	return args.Get(0).(decimal.Decimal), args.Get(1).(int64), args.Error(2)
}

func (m *MockReportingRepository) GetFootprintByActivityType(ctx context.Context, userID string, startDate, endDate time.Time) ([]repository.ActivitySummary, error) {
	args := m.Called(ctx, userID, startDate, endDate)
	return args.Get(0).([]repository.ActivitySummary), args.Error(1)
}

func (m *MockReportingRepository) GetFootprintByMonth(ctx context.Context, userID string, startDate, endDate time.Time) ([]repository.MonthlyData, error) {
	args := m.Called(ctx, userID, startDate, endDate)
	return args.Get(0).([]repository.MonthlyData), args.Error(1)
}

func (m *MockReportingRepository) GetWalletBalance(ctx context.Context, userID string) (available, earned, spent decimal.Decimal, err error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(decimal.Decimal), args.Get(1).(decimal.Decimal), args.Get(2).(decimal.Decimal), args.Error(3)
}

func (m *MockReportingRepository) GetTransactionsBySource(ctx context.Context, userID string, startDate, endDate time.Time) ([]repository.TransactionSummary, error) {
	args := m.Called(ctx, userID, startDate, endDate)
	return args.Get(0).([]repository.TransactionSummary), args.Error(1)
}

func (m *MockReportingRepository) GetCreditsByMonth(ctx context.Context, userID string, startDate, endDate time.Time) ([]repository.MonthlyData, error) {
	args := m.Called(ctx, userID, startDate, endDate)
	return args.Get(0).([]repository.MonthlyData), args.Error(1)
}

func (m *MockReportingRepository) GetTopEarningActivities(ctx context.Context, userID string, startDate, endDate time.Time, limit int) ([]models.ActivitySummary, error) {
	args := m.Called(ctx, userID, startDate, endDate, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.ActivitySummary), args.Error(1)
}

func (m *MockReportingRepository) GetRecentTransactions(ctx context.Context, userID string, startDate, endDate time.Time, limit int) ([]models.TransactionSummary, error) {
	args := m.Called(ctx, userID, startDate, endDate, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.TransactionSummary), args.Error(1)
}

func (m *MockReportingRepository) GetActivityStats(ctx context.Context, userID string, startDate, endDate time.Time) (int64, []repository.DailyActivity, error) {
	args := m.Called(ctx, userID, startDate, endDate)
	if args.Get(1) == nil {
		return args.Get(0).(int64), nil, args.Error(2)
	}
	return args.Get(0).(int64), args.Get(1).([]repository.DailyActivity), args.Error(2)
}

func TestCollectSummaryData_MostLeastActiveDays(t *testing.T) {
	// Setup
	mockRepo := new(MockReportingRepository)
	log := logger.New("debug")
	collector := NewDatabaseDataCollectorWithRepo(mockRepo, log)
	ctx := context.Background()
	userID := "test-user"
	startDate := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2023, 1, 31, 0, 0, 0, 0, time.UTC)

	// Mock Footprint Data (needed for CollectSummaryData)
	mockRepo.On("GetTotalFootprint", ctx, userID, startDate, endDate).
		Return(decimal.NewFromFloat(100), int64(10), nil)
	mockRepo.On("GetFootprintByActivityType", ctx, userID, startDate, endDate).
		Return([]repository.ActivitySummary{}, nil)
	mockRepo.On("GetFootprintByMonth", ctx, userID, startDate, endDate).
		Return([]repository.MonthlyData{}, nil)

	// Mock Credits Data (needed for CollectSummaryData)
	mockRepo.On("GetWalletBalance", ctx, userID).
		Return(decimal.NewFromFloat(100), decimal.NewFromFloat(50), decimal.NewFromFloat(20), nil)
	mockRepo.On("GetTransactionsBySource", ctx, userID, startDate, endDate).
		Return([]repository.TransactionSummary{}, nil)
	mockRepo.On("GetCreditsByMonth", ctx, userID, startDate, endDate).
		Return([]repository.MonthlyData{}, nil)
	mockRepo.On("GetTopEarningActivities", ctx, userID, startDate, endDate, 10).
		Return(nil, nil)
	mockRepo.On("GetRecentTransactions", ctx, userID, startDate, endDate, 20).
		Return([]models.TransactionSummary{}, nil)

	// Test Case 1: Normal activity data
	day1 := startDate.AddDate(0, 0, 1) // 5 activities
	day2 := startDate.AddDate(0, 0, 2) // 2 activities
	day3 := startDate.AddDate(0, 0, 3) // 10 activities

	dailyStats := []repository.DailyActivity{
		{Day: day3, Count: 10}, // Most active (first in DESC sort)
		{Day: day1, Count: 5},
		{Day: day2, Count: 2}, // Least active (last in DESC sort)
	}

	mockRepo.On("GetActivityStats", ctx, userID, startDate, endDate).
		Return(int64(17), dailyStats, nil).Once()

	data, err := collector.CollectSummaryData(ctx, userID, startDate, endDate)
	assert.NoError(t, err)
	assert.Equal(t, day3, data.MostActiveDay)
	assert.Equal(t, day2, data.LeastActiveDay)
	assert.Equal(t, int64(17), data.TotalActivities)

	// Test Case 2: Tie breaking (equal activity counts)
	// day4 and day5 both have 8 activities.
	// The repository returns them sorted by count DESC, day ASC (mock simulates this).
	// Logic should pick day4 (earlier) as most active.
	// For least active, since both are "least" (equal count), it picks the last one in the list (day5).
	day4 := startDate.AddDate(0, 0, 4)
	day5 := startDate.AddDate(0, 0, 5)

	tieStats := []repository.DailyActivity{
		{Day: day4, Count: 8},
		{Day: day5, Count: 8},
	}

	mockRepo.On("GetActivityStats", ctx, userID, startDate, endDate).
		Return(int64(16), tieStats, nil).Once()

	dataTie, err := collector.CollectSummaryData(ctx, userID, startDate, endDate)
	assert.NoError(t, err)
	assert.Equal(t, day4, dataTie.MostActiveDay)  // Expecting earliest day (first in list)
	assert.Equal(t, day5, dataTie.LeastActiveDay) // Expecting latest day (last in list)

	// Test Case 3: Single day activity
	singleDayStats := []repository.DailyActivity{
		{Day: day1, Count: 1},
	}
	mockRepo.On("GetActivityStats", ctx, userID, startDate, endDate).
		Return(int64(1), singleDayStats, nil).Once()

	dataSingle, err := collector.CollectSummaryData(ctx, userID, startDate, endDate)
	assert.NoError(t, err)
	assert.Equal(t, day1, dataSingle.MostActiveDay)
	assert.Equal(t, day1, dataSingle.LeastActiveDay)

	// Test Case 4: No activity data
	mockRepo.On("GetActivityStats", ctx, userID, startDate, endDate).
		Return(int64(0), []repository.DailyActivity{}, nil).Once()

	dataEmpty, err := collector.CollectSummaryData(ctx, userID, startDate, endDate)
	assert.NoError(t, err)
	assert.Equal(t, startDate, dataEmpty.MostActiveDay)
	assert.Equal(t, endDate, dataEmpty.LeastActiveDay)
	assert.Equal(t, int64(0), dataEmpty.TotalActivities)

	// Test Case 5: Error fetching activity stats
	// We expect the error to be logged (suppressed) and fallback values to be used.
	mockRepo.On("GetActivityStats", ctx, userID, startDate, endDate).
		Return(int64(0), nil, errors.New("db error")).Once()

	dataErr, err := collector.CollectSummaryData(ctx, userID, startDate, endDate)
	assert.NoError(t, err)
	assert.Equal(t, startDate, dataErr.MostActiveDay) // Fallback
	assert.Equal(t, endDate, dataErr.LeastActiveDay)  // Fallback

	mockRepo.AssertExpectations(t)
}
