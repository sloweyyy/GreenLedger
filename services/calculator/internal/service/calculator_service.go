package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/greenledger/services/calculator/internal/models"
	"github.com/greenledger/services/calculator/internal/repository"
	"github.com/greenledger/shared/logger"
)

// CalculatorService handles carbon footprint calculations
type CalculatorService struct {
	calculationRepo    *repository.CalculationRepository
	emissionFactorRepo *repository.EmissionFactorRepository
	logger             *logger.Logger
}

// NewCalculatorService creates a new calculator service
func NewCalculatorService(
	calculationRepo *repository.CalculationRepository,
	emissionFactorRepo *repository.EmissionFactorRepository,
	logger *logger.Logger,
) *CalculatorService {
	return &CalculatorService{
		calculationRepo:    calculationRepo,
		emissionFactorRepo: emissionFactorRepo,
		logger:             logger,
	}
}

// CalculateFootprintRequest represents a calculation request
type CalculateFootprintRequest struct {
	UserID     string                `json:"user_id" binding:"required"`
	Activities []ActivityDataRequest `json:"activities" binding:"required,min=1"`
}

// ActivityDataRequest represents activity data for calculation
type ActivityDataRequest struct {
	ActivityType string                 `json:"activity_type" binding:"required"`
	Data         map[string]interface{} `json:"data" binding:"required"`
}

// CalculateFootprintResponse represents a calculation response
type CalculateFootprintResponse struct {
	CalculationID   uuid.UUID        `json:"calculation_id"`
	TotalCO2Kg      float64          `json:"total_co2_kg"`
	ActivityResults []ActivityResult `json:"activity_results"`
	CalculatedAt    time.Time        `json:"calculated_at"`
}

// ActivityResult represents the result of an activity calculation
type ActivityResult struct {
	ActivityType   string                 `json:"activity_type"`
	CO2Kg          float64                `json:"co2_kg"`
	EmissionFactor float64                `json:"emission_factor"`
	FactorSource   string                 `json:"factor_source"`
	ActivityData   map[string]interface{} `json:"activity_data"`
}

// CalculateFootprint calculates carbon footprint for given activities
func (s *CalculatorService) CalculateFootprint(ctx context.Context, req *CalculateFootprintRequest) (*CalculateFootprintResponse, error) {
	s.logger.LogInfo(ctx, "starting footprint calculation",
		logger.String("user_id", req.UserID),
		logger.Int("activity_count", len(req.Activities)))

	var totalCO2 float64
	var activityResults []ActivityResult
	var activities []models.Activity

	calculationID := uuid.New()

	// Calculate each activity
	for i, activityReq := range req.Activities {
		result, err := s.calculateActivity(ctx, activityReq)
		if err != nil {
			s.logger.LogError(ctx, "failed to calculate activity", err,
				logger.String("user_id", req.UserID),
				logger.Int("activity_index", i),
				logger.String("activity_type", activityReq.ActivityType))
			return nil, fmt.Errorf("failed to calculate activity %d: %w", i, err)
		}

		totalCO2 += result.CO2Kg
		activityResults = append(activityResults, *result)

		// Convert activity data to JSON
		activityDataJSON, err := json.Marshal(result.ActivityData)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal activity data: %w", err)
		}

		// Create activity model
		activity := models.Activity{
			CalculationID:  calculationID,
			ActivityType:   result.ActivityType,
			CO2Kg:          result.CO2Kg,
			EmissionFactor: result.EmissionFactor,
			FactorSource:   result.FactorSource,
			ActivityData:   string(activityDataJSON),
		}
		activities = append(activities, activity)
	}

	// Create calculation record
	calculation := &models.Calculation{
		ID:         calculationID,
		UserID:     req.UserID,
		TotalCO2Kg: totalCO2,
		Activities: activities,
	}

	// Save to database
	if err := s.calculationRepo.Create(ctx, calculation); err != nil {
		s.logger.LogError(ctx, "failed to save calculation", err,
			logger.String("user_id", req.UserID),
			logger.String("calculation_id", calculationID.String()))
		return nil, fmt.Errorf("failed to save calculation: %w", err)
	}

	response := &CalculateFootprintResponse{
		CalculationID:   calculationID,
		TotalCO2Kg:      totalCO2,
		ActivityResults: activityResults,
		CalculatedAt:    time.Now().UTC(),
	}

	s.logger.LogInfo(ctx, "footprint calculation completed",
		logger.String("user_id", req.UserID),
		logger.String("calculation_id", calculationID.String()),
		logger.Float64("total_co2_kg", totalCO2))

	return response, nil
}

// calculateActivity calculates CO2 emissions for a single activity
func (s *CalculatorService) calculateActivity(ctx context.Context, req ActivityDataRequest) (*ActivityResult, error) {
	switch req.ActivityType {
	case models.ActivityTypeVehicleTravel:
		return s.calculateVehicleTravel(ctx, req.Data)
	case models.ActivityTypeElectricity:
		return s.calculateElectricity(ctx, req.Data)
	case models.ActivityTypePurchase:
		return s.calculatePurchase(ctx, req.Data)
	case models.ActivityTypeFlight:
		return s.calculateFlight(ctx, req.Data)
	case models.ActivityTypeHeating:
		return s.calculateHeating(ctx, req.Data)
	default:
		return nil, fmt.Errorf("unsupported activity type: %s", req.ActivityType)
	}
}

// calculateVehicleTravel calculates emissions for vehicle travel
func (s *CalculatorService) calculateVehicleTravel(ctx context.Context, data map[string]interface{}) (*ActivityResult, error) {
	vehicleType, ok := data["vehicle_type"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid vehicle_type")
	}

	distanceKm, ok := data["distance_km"].(float64)
	if !ok {
		return nil, fmt.Errorf("missing or invalid distance_km")
	}

	// Get emission factor
	factor, err := s.emissionFactorRepo.GetByActivityTypeAndSubType(ctx, models.ActivityTypeVehicleTravel, vehicleType)
	if err != nil {
		return nil, fmt.Errorf("failed to get emission factor for vehicle type %s: %w", vehicleType, err)
	}

	// Calculate CO2 emissions (factor is typically in kg CO2 per km)
	co2Kg := distanceKm * factor.FactorCO2

	return &ActivityResult{
		ActivityType:   models.ActivityTypeVehicleTravel,
		CO2Kg:          co2Kg,
		EmissionFactor: factor.FactorCO2,
		FactorSource:   factor.Source,
		ActivityData:   data,
	}, nil
}

// calculateElectricity calculates emissions for electricity usage
func (s *CalculatorService) calculateElectricity(ctx context.Context, data map[string]interface{}) (*ActivityResult, error) {
	kwhUsage, ok := data["kwh_usage"].(float64)
	if !ok {
		return nil, fmt.Errorf("missing or invalid kwh_usage")
	}

	location, _ := data["location"].(string)

	// Get emission factor (try location-specific first, then default)
	factors, err := s.emissionFactorRepo.GetByActivityTypeAndLocation(ctx, models.ActivityTypeElectricity, location)
	if err != nil || len(factors) == 0 {
		return nil, fmt.Errorf("failed to get emission factor for electricity in location %s: %w", location, err)
	}

	factor := factors[0] // Use the first (most specific) factor

	// Calculate CO2 emissions (factor is typically in kg CO2 per kWh)
	co2Kg := kwhUsage * factor.FactorCO2

	return &ActivityResult{
		ActivityType:   models.ActivityTypeElectricity,
		CO2Kg:          co2Kg,
		EmissionFactor: factor.FactorCO2,
		FactorSource:   factor.Source,
		ActivityData:   data,
	}, nil
}

// calculatePurchase calculates emissions for purchases
func (s *CalculatorService) calculatePurchase(ctx context.Context, data map[string]interface{}) (*ActivityResult, error) {
	category, ok := data["category"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid category")
	}

	priceUSD, ok := data["price_usd"].(float64)
	if !ok {
		return nil, fmt.Errorf("missing or invalid price_usd")
	}

	// Get emission factor
	factor, err := s.emissionFactorRepo.GetByActivityTypeAndSubType(ctx, models.ActivityTypePurchase, category)
	if err != nil {
		return nil, fmt.Errorf("failed to get emission factor for purchase category %s: %w", category, err)
	}

	// Calculate CO2 emissions (factor is typically in kg CO2 per USD)
	co2Kg := priceUSD * factor.FactorCO2

	return &ActivityResult{
		ActivityType:   models.ActivityTypePurchase,
		CO2Kg:          co2Kg,
		EmissionFactor: factor.FactorCO2,
		FactorSource:   factor.Source,
		ActivityData:   data,
	}, nil
}

// calculateFlight calculates emissions for flights
func (s *CalculatorService) calculateFlight(ctx context.Context, data map[string]interface{}) (*ActivityResult, error) {
	departureAirport, ok := data["departure_airport"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid departure_airport")
	}

	arrivalAirport, ok := data["arrival_airport"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid arrival_airport")
	}

	flightClass, _ := data["flight_class"].(string)
	if flightClass == "" {
		flightClass = models.FlightClassEconomy
	}

	isRoundTrip, _ := data["is_round_trip"].(bool)

	// Calculate distance (simplified - in real implementation, use airport coordinates)
	distance := s.calculateFlightDistance(departureAirport, arrivalAirport)

	// Get emission factor based on flight class
	factor, err := s.emissionFactorRepo.GetByActivityTypeAndSubType(ctx, models.ActivityTypeFlight, flightClass)
	if err != nil {
		return nil, fmt.Errorf("failed to get emission factor for flight class %s: %w", flightClass, err)
	}

	// Calculate CO2 emissions (factor is typically in kg CO2 per km)
	co2Kg := distance * factor.FactorCO2
	if isRoundTrip {
		co2Kg *= 2
	}

	return &ActivityResult{
		ActivityType:   models.ActivityTypeFlight,
		CO2Kg:          co2Kg,
		EmissionFactor: factor.FactorCO2,
		FactorSource:   factor.Source,
		ActivityData:   data,
	}, nil
}

// calculateHeating calculates emissions for heating
func (s *CalculatorService) calculateHeating(ctx context.Context, data map[string]interface{}) (*ActivityResult, error) {
	fuelType, ok := data["fuel_type"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid fuel_type")
	}

	consumption, ok := data["consumption"].(float64)
	if !ok {
		return nil, fmt.Errorf("missing or invalid consumption")
	}

	// Get emission factor
	factor, err := s.emissionFactorRepo.GetByActivityTypeAndSubType(ctx, models.ActivityTypeHeating, fuelType)
	if err != nil {
		return nil, fmt.Errorf("failed to get emission factor for heating fuel %s: %w", fuelType, err)
	}

	// Calculate CO2 emissions
	co2Kg := consumption * factor.FactorCO2

	return &ActivityResult{
		ActivityType:   models.ActivityTypeHeating,
		CO2Kg:          co2Kg,
		EmissionFactor: factor.FactorCO2,
		FactorSource:   factor.Source,
		ActivityData:   data,
	}, nil
}

// calculateFlightDistance calculates distance between airports (simplified)
func (s *CalculatorService) calculateFlightDistance(departure, arrival string) float64 {
	// This is a simplified implementation
	// In a real system, you would use actual airport coordinates and great circle distance

	// Sample distances for common routes (in km)
	routes := map[string]float64{
		"LAX-JFK": 3944,
		"JFK-LAX": 3944,
		"LHR-JFK": 5541,
		"JFK-LHR": 5541,
		"NRT-LAX": 8815,
		"LAX-NRT": 8815,
		"SFO-NYC": 4139,
		"NYC-SFO": 4139,
		"LHR-CDG": 344,
		"CDG-LHR": 344,
	}

	key := departure + "-" + arrival
	if distance, exists := routes[key]; exists {
		return distance
	}

	// Estimate based on airport codes (very simplified)
	// In reality, you'd use a proper airport database and haversine formula
	return 1000 // Default distance for unknown routes
}

// GetCalculationHistory retrieves calculation history for a user
func (s *CalculatorService) GetCalculationHistory(ctx context.Context, userID string, startDate, endDate *time.Time, limit, offset int) ([]*models.Calculation, int64, error) {
	if startDate != nil && endDate != nil {
		return s.calculationRepo.GetByUserIDAndDateRange(ctx, userID, *startDate, *endDate, limit, offset)
	}
	return s.calculationRepo.GetByUserID(ctx, userID, limit, offset)
}

// GetCalculationByID retrieves a specific calculation
func (s *CalculatorService) GetCalculationByID(ctx context.Context, id uuid.UUID) (*models.Calculation, error) {
	return s.calculationRepo.GetByID(ctx, id)
}
