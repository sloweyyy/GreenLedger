package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Calculation represents a carbon footprint calculation
type Calculation struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID      string    `gorm:"not null;index" json:"user_id"`
	TotalCO2Kg  float64   `gorm:"not null" json:"total_co2_kg"`
	Activities  []Activity `gorm:"foreignKey:CalculationID" json:"activities"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Activity represents an individual activity in a calculation
type Activity struct {
	ID             uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CalculationID  uuid.UUID `gorm:"type:uuid;not null;index" json:"calculation_id"`
	ActivityType   string    `gorm:"not null" json:"activity_type"`
	CO2Kg          float64   `gorm:"not null" json:"co2_kg"`
	EmissionFactor float64   `gorm:"not null" json:"emission_factor"`
	FactorSource   string    `gorm:"not null" json:"factor_source"`
	ActivityData   string    `gorm:"type:jsonb" json:"activity_data"` // JSON data
	CreatedAt      time.Time `json:"created_at"`
}

// EmissionFactor represents emission factors for different activities
type EmissionFactor struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ActivityType string    `gorm:"not null;index" json:"activity_type"`
	SubType      string    `gorm:"not null;index" json:"sub_type"`
	FactorCO2    float64   `gorm:"not null" json:"factor_co2_per_unit"`
	Unit         string    `gorm:"not null" json:"unit"`
	Source       string    `gorm:"not null" json:"source"`
	Location     string    `gorm:"index" json:"location"`
	LastUpdated  time.Time `json:"last_updated"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// VehicleActivityData represents vehicle travel activity data
type VehicleActivityData struct {
	VehicleType       string  `json:"vehicle_type"`
	DistanceKm        float64 `json:"distance_km"`
	FuelEfficiencyL   float64 `json:"fuel_efficiency_l_per_100km,omitempty"`
}

// ElectricityActivityData represents electricity usage activity data
type ElectricityActivityData struct {
	KwhUsage float64 `json:"kwh_usage"`
	Location string  `json:"location"`
}

// PurchaseActivityData represents purchase activity data
type PurchaseActivityData struct {
	Category string  `json:"category"`
	Quantity float64 `json:"quantity"`
	PriceUSD float64 `json:"price_usd"`
}

// FlightActivityData represents flight activity data
type FlightActivityData struct {
	DepartureAirport string `json:"departure_airport"`
	ArrivalAirport   string `json:"arrival_airport"`
	IsRoundTrip      bool   `json:"is_round_trip"`
	FlightClass      string `json:"flight_class"`
}

// HeatingActivityData represents heating activity data
type HeatingActivityData struct {
	FuelType    string  `json:"fuel_type"`
	Consumption float64 `json:"consumption"`
	Unit        string  `json:"unit"`
}

// BeforeCreate hook for Calculation
func (c *Calculation) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}

// BeforeCreate hook for Activity
func (a *Activity) BeforeCreate(tx *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return nil
}

// BeforeCreate hook for EmissionFactor
func (e *EmissionFactor) BeforeCreate(tx *gorm.DB) error {
	if e.ID == uuid.Nil {
		e.ID = uuid.New()
	}
	return nil
}

// TableName returns the table name for Calculation
func (Calculation) TableName() string {
	return "calculations"
}

// TableName returns the table name for Activity
func (Activity) TableName() string {
	return "activities"
}

// TableName returns the table name for EmissionFactor
func (EmissionFactor) TableName() string {
	return "emission_factors"
}

// Activity type constants
const (
	ActivityTypeVehicleTravel   = "vehicle_travel"
	ActivityTypeElectricity     = "electricity_usage"
	ActivityTypePurchase        = "purchase"
	ActivityTypeFlight          = "flight"
	ActivityTypeHeating         = "heating"
)

// Vehicle type constants
const (
	VehicleTypeCarGasoline = "car_gasoline"
	VehicleTypeCarDiesel   = "car_diesel"
	VehicleTypeCarElectric = "car_electric"
	VehicleTypeCarHybrid   = "car_hybrid"
	VehicleTypeMotorcycle  = "motorcycle"
	VehicleTypeBus         = "bus"
	VehicleTypeTrain       = "train"
)

// Purchase category constants
const (
	PurchaseCategoryFood        = "food"
	PurchaseCategoryClothing    = "clothing"
	PurchaseCategoryElectronics = "electronics"
	PurchaseCategoryFurniture   = "furniture"
	PurchaseCategoryOther       = "other"
)

// Flight class constants
const (
	FlightClassEconomy  = "economy"
	FlightClassBusiness = "business"
	FlightClassFirst    = "first"
)

// Heating fuel type constants
const (
	HeatingFuelNaturalGas = "natural_gas"
	HeatingFuelOil        = "oil"
	HeatingFuelElectric   = "electric"
	HeatingFuelPropane    = "propane"
)
