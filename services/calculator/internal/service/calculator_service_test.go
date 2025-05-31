package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/greenledger/services/calculator/internal/models"
	"github.com/greenledger/shared/logger"
)

// MockCalculationRepository is a mock implementation of CalculationRepository
type MockCalculationRepository struct {
	mock.Mock
}

func (m *MockCalculationRepository) Create(ctx context.Context, calculation *models.Calculation) error {
	args := m.Called(ctx, calculation)
	return args.Error(0)
}

func (m *MockCalculationRepository) GetByID(ctx context.Context, id string) (*models.Calculation, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*models.Calculation), args.Error(1)
}

func (m *MockCalculationRepository) GetByUserID(ctx context.Context, userID string, limit, offset int) ([]*models.Calculation, int64, error) {
	args := m.Called(ctx, userID, limit, offset)
	return args.Get(0).([]*models.Calculation), args.Get(1).(int64), args.Error(2)
}

func (m *MockCalculationRepository) GetByUserIDAndDateRange(ctx context.Context, userID string, startDate, endDate time.Time, limit, offset int) ([]*models.Calculation, int64, error) {
	args := m.Called(ctx, userID, startDate, endDate, limit, offset)
	return args.Get(0).([]*models.Calculation), args.Get(1).(int64), args.Error(2)
}

// MockEmissionFactorRepository is a mock implementation of EmissionFactorRepository
type MockEmissionFactorRepository struct {
	mock.Mock
}

func (m *MockEmissionFactorRepository) GetByActivityType(ctx context.Context, activityType string) ([]*models.EmissionFactor, error) {
	args := m.Called(ctx, activityType)
	return args.Get(0).([]*models.EmissionFactor), args.Error(1)
}

func (m *MockEmissionFactorRepository) GetByActivityTypeAndSubType(ctx context.Context, activityType, subType string) (*models.EmissionFactor, error) {
	args := m.Called(ctx, activityType, subType)
	return args.Get(0).(*models.EmissionFactor), args.Error(1)
}

func (m *MockEmissionFactorRepository) GetByActivityTypeAndLocation(ctx context.Context, activityType, location string) ([]*models.EmissionFactor, error) {
	args := m.Called(ctx, activityType, location)
	return args.Get(0).([]*models.EmissionFactor), args.Error(1)
}

func TestCalculatorService_CalculateVehicleTravel(t *testing.T) {
	// Setup
	mockCalcRepo := new(MockCalculationRepository)
	mockFactorRepo := new(MockEmissionFactorRepository)
	logger := logger.New("debug")
	service := NewCalculatorService(mockCalcRepo, mockFactorRepo, logger)

	ctx := context.Background()

	// Mock emission factor
	emissionFactor := &models.EmissionFactor{
		ActivityType: models.ActivityTypeVehicleTravel,
		SubType:      models.VehicleTypeCarGasoline,
		FactorCO2:    0.21, // kg CO2 per km
		Unit:         "km",
		Source:       "EPA 2023",
	}

	mockFactorRepo.On("GetByActivityTypeAndSubType", ctx, models.ActivityTypeVehicleTravel, models.VehicleTypeCarGasoline).
		Return(emissionFactor, nil)

	// Test data
	activityData := map[string]interface{}{
		"vehicle_type": models.VehicleTypeCarGasoline,
		"distance_km":  100.0,
	}

	// Execute
	result, err := service.calculateVehicleTravel(ctx, activityData)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, models.ActivityTypeVehicleTravel, result.ActivityType)
	assert.Equal(t, 21.0, result.CO2Kg) // 100 km * 0.21 kg/km = 21 kg CO2
	assert.Equal(t, 0.21, result.EmissionFactor)
	assert.Equal(t, "EPA 2023", result.FactorSource)

	mockFactorRepo.AssertExpectations(t)
}

func TestCalculatorService_CalculateElectricity(t *testing.T) {
	// Setup
	mockCalcRepo := new(MockCalculationRepository)
	mockFactorRepo := new(MockEmissionFactorRepository)
	logger := logger.New("debug")
	service := NewCalculatorService(mockCalcRepo, mockFactorRepo, logger)

	ctx := context.Background()

	// Mock emission factor
	emissionFactor := &models.EmissionFactor{
		ActivityType: models.ActivityTypeElectricity,
		SubType:      "grid",
		FactorCO2:    0.5, // kg CO2 per kWh
		Unit:         "kWh",
		Source:       "IEA 2023",
		Location:     "US",
	}

	mockFactorRepo.On("GetByActivityTypeAndLocation", ctx, models.ActivityTypeElectricity, "US").
		Return([]*models.EmissionFactor{emissionFactor}, nil)

	// Test data
	activityData := map[string]interface{}{
		"kwh_usage": 50.0,
		"location":  "US",
	}

	// Execute
	result, err := service.calculateElectricity(ctx, activityData)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, models.ActivityTypeElectricity, result.ActivityType)
	assert.Equal(t, 25.0, result.CO2Kg) // 50 kWh * 0.5 kg/kWh = 25 kg CO2
	assert.Equal(t, 0.5, result.EmissionFactor)
	assert.Equal(t, "IEA 2023", result.FactorSource)

	mockFactorRepo.AssertExpectations(t)
}

func TestCalculatorService_CalculateFootprint(t *testing.T) {
	// Setup
	mockCalcRepo := new(MockCalculationRepository)
	mockFactorRepo := new(MockEmissionFactorRepository)
	logger := logger.New("debug")
	service := NewCalculatorService(mockCalcRepo, mockFactorRepo, logger)

	ctx := context.Background()

	// Mock emission factors
	vehicleEmissionFactor := &models.EmissionFactor{
		ActivityType: models.ActivityTypeVehicleTravel,
		SubType:      models.VehicleTypeCarGasoline,
		FactorCO2:    0.21,
		Unit:         "km",
		Source:       "EPA 2023",
	}

	electricityEmissionFactor := &models.EmissionFactor{
		ActivityType: models.ActivityTypeElectricity,
		SubType:      "grid",
		FactorCO2:    0.5,
		Unit:         "kWh",
		Source:       "IEA 2023",
		Location:     "US",
	}

	mockFactorRepo.On("GetByActivityTypeAndSubType", ctx, models.ActivityTypeVehicleTravel, models.VehicleTypeCarGasoline).
		Return(vehicleEmissionFactor, nil)
	mockFactorRepo.On("GetByActivityTypeAndLocation", ctx, models.ActivityTypeElectricity, "US").
		Return([]*models.EmissionFactor{electricityEmissionFactor}, nil)

	mockCalcRepo.On("Create", ctx, mock.AnythingOfType("*models.Calculation")).
		Return(nil)

	// Test request
	req := &CalculateFootprintRequest{
		UserID: "test-user-123",
		Activities: []ActivityDataRequest{
			{
				ActivityType: models.ActivityTypeVehicleTravel,
				Data: map[string]interface{}{
					"vehicle_type": models.VehicleTypeCarGasoline,
					"distance_km":  100.0,
				},
			},
			{
				ActivityType: models.ActivityTypeElectricity,
				Data: map[string]interface{}{
					"kwh_usage": 50.0,
					"location":  "US",
				},
			},
		},
	}

	// Execute
	response, err := service.CalculateFootprint(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, 46.0, response.TotalCO2Kg) // 21 + 25 = 46 kg CO2
	assert.Len(t, response.ActivityResults, 2)
	assert.Equal(t, 21.0, response.ActivityResults[0].CO2Kg)
	assert.Equal(t, 25.0, response.ActivityResults[1].CO2Kg)

	mockFactorRepo.AssertExpectations(t)
	mockCalcRepo.AssertExpectations(t)
}

func TestCalculatorService_CalculateFootprint_InvalidActivityType(t *testing.T) {
	// Setup
	mockCalcRepo := new(MockCalculationRepository)
	mockFactorRepo := new(MockEmissionFactorRepository)
	logger := logger.New("debug")
	service := NewCalculatorService(mockCalcRepo, mockFactorRepo, logger)

	ctx := context.Background()

	// Test request with invalid activity type
	req := &CalculateFootprintRequest{
		UserID: "test-user-123",
		Activities: []ActivityDataRequest{
			{
				ActivityType: "invalid_activity",
				Data: map[string]interface{}{
					"some_data": "value",
				},
			},
		},
	}

	// Execute
	response, err := service.CalculateFootprint(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "unsupported activity type")
}
