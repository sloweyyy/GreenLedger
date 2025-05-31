package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// EcoActivity represents an eco-friendly activity
type EcoActivity struct {
	ID             uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID         string     `gorm:"not null;index" json:"user_id"`
	ActivityTypeID uuid.UUID  `gorm:"type:uuid;not null;index" json:"activity_type_id"`
	Description    string     `gorm:"not null" json:"description"`
	Duration       int        `gorm:"not null" json:"duration"` // in minutes
	Distance       float64    `json:"distance"`                 // in kilometers (for transport activities)
	Quantity       float64    `json:"quantity"`                 // generic quantity field
	Unit           string     `json:"unit"`
	Location       string     `json:"location"`
	CreditsEarned  float64    `gorm:"not null;default:0" json:"credits_earned"`
	IsVerified     bool       `gorm:"default:false" json:"is_verified"`
	VerifiedAt     *time.Time `json:"verified_at"`
	VerifiedBy     string     `json:"verified_by"`
	Source         string     `gorm:"not null" json:"source"`        // manual, iot, webhook, etc.
	SourceData     string     `gorm:"type:jsonb" json:"source_data"` // Original data from source
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`

	// Relationships
	ActivityType ActivityType `gorm:"foreignKey:ActivityTypeID" json:"activity_type,omitempty"`
}

// ActivityType represents types of eco-friendly activities
type ActivityType struct {
	ID                   uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Name                 string    `gorm:"uniqueIndex;not null" json:"name"`
	Category             string    `gorm:"not null;index" json:"category"`
	Description          string    `json:"description"`
	Icon                 string    `json:"icon"`
	BaseCreditsPerUnit   float64   `gorm:"not null" json:"base_credits_per_unit"`
	Unit                 string    `gorm:"not null" json:"unit"`
	IsActive             bool      `gorm:"default:true" json:"is_active"`
	RequiresVerification bool      `gorm:"default:false" json:"requires_verification"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`

	// Relationships
	CreditRules []CreditRule `gorm:"foreignKey:ActivityTypeID" json:"credit_rules,omitempty"`
}

// CreditRule represents rules for calculating credits for activities
type CreditRule struct {
	ID             uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ActivityTypeID uuid.UUID  `gorm:"type:uuid;not null;index" json:"activity_type_id"`
	Name           string     `gorm:"not null" json:"name"`
	Description    string     `json:"description"`
	MinValue       float64    `json:"min_value"`
	MaxValue       float64    `json:"max_value"`
	CreditsPerUnit float64    `gorm:"not null" json:"credits_per_unit"`
	Multiplier     float64    `gorm:"default:1" json:"multiplier"`
	IsActive       bool       `gorm:"default:true" json:"is_active"`
	ValidFrom      time.Time  `json:"valid_from"`
	ValidTo        *time.Time `json:"valid_to"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`

	// Relationship
	ActivityType ActivityType `gorm:"foreignKey:ActivityTypeID" json:"-"`
}

// ActivityChallenge represents challenges for eco-activities
type ActivityChallenge struct {
	ID            uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Name          string    `gorm:"not null" json:"name"`
	Description   string    `json:"description"`
	StartDate     time.Time `gorm:"not null" json:"start_date"`
	EndDate       time.Time `gorm:"not null" json:"end_date"`
	TargetValue   float64   `gorm:"not null" json:"target_value"`
	TargetUnit    string    `gorm:"not null" json:"target_unit"`
	RewardCredits float64   `gorm:"not null" json:"reward_credits"`
	IsActive      bool      `gorm:"default:true" json:"is_active"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`

	// Relationships
	Participants []ChallengeParticipant `gorm:"foreignKey:ChallengeID" json:"participants,omitempty"`
}

// ChallengeParticipant represents a user's participation in a challenge
type ChallengeParticipant struct {
	ID          uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ChallengeID uuid.UUID  `gorm:"type:uuid;not null;index" json:"challenge_id"`
	UserID      string     `gorm:"not null;index" json:"user_id"`
	Progress    float64    `gorm:"default:0" json:"progress"`
	IsCompleted bool       `gorm:"default:false" json:"is_completed"`
	CompletedAt *time.Time `json:"completed_at"`
	JoinedAt    time.Time  `gorm:"default:now()" json:"joined_at"`

	// Relationship
	Challenge ActivityChallenge `gorm:"foreignKey:ChallengeID" json:"-"`
}

// IoTDevice represents IoT devices that can report activities
type IoTDevice struct {
	ID        uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID    string     `gorm:"not null;index" json:"user_id"`
	DeviceID  string     `gorm:"uniqueIndex;not null" json:"device_id"`
	Name      string     `gorm:"not null" json:"name"`
	Type      string     `gorm:"not null" json:"type"` // bike_sensor, smart_meter, etc.
	IsActive  bool       `gorm:"default:true" json:"is_active"`
	LastSeen  *time.Time `json:"last_seen"`
	APIKey    string     `gorm:"uniqueIndex;not null" json:"api_key"`
	Settings  string     `gorm:"type:jsonb" json:"settings"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

// BeforeCreate hooks
func (e *EcoActivity) BeforeCreate(tx *gorm.DB) error {
	if e.ID == uuid.Nil {
		e.ID = uuid.New()
	}
	return nil
}

func (a *ActivityType) BeforeCreate(tx *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return nil
}

func (c *CreditRule) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}

func (a *ActivityChallenge) BeforeCreate(tx *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return nil
}

func (c *ChallengeParticipant) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}

func (i *IoTDevice) BeforeCreate(tx *gorm.DB) error {
	if i.ID == uuid.Nil {
		i.ID = uuid.New()
	}
	return nil
}

// Table names
func (EcoActivity) TableName() string          { return "eco_activities" }
func (ActivityType) TableName() string         { return "activity_types" }
func (CreditRule) TableName() string           { return "credit_rules" }
func (ActivityChallenge) TableName() string    { return "activity_challenges" }
func (ChallengeParticipant) TableName() string { return "challenge_participants" }
func (IoTDevice) TableName() string            { return "iot_devices" }

// Activity categories
const (
	CategoryTransport   = "transport"
	CategoryEnergy      = "energy"
	CategoryWaste       = "waste"
	CategoryConsumption = "consumption"
	CategoryNature      = "nature"
)

// Activity sources
const (
	SourceManual  = "manual"
	SourceIoT     = "iot"
	SourceWebhook = "webhook"
	SourceAPI     = "api"
	SourceImport  = "import"
)

// Common activity types
const (
	ActivityBiking         = "biking"
	ActivityWalking        = "walking"
	ActivityPublicTransit  = "public_transit"
	ActivityCarPooling     = "car_pooling"
	ActivityRecycling      = "recycling"
	ActivityComposting     = "composting"
	ActivitySolarEnergy    = "solar_energy"
	ActivityTreePlanting   = "tree_planting"
	ActivityLocalShopping  = "local_shopping"
	ActivityVegetarianMeal = "vegetarian_meal"
)

// Device types
const (
	DeviceTypeBikeSensor     = "bike_sensor"
	DeviceTypeSmartMeter     = "smart_meter"
	DeviceTypeFitnessTracker = "fitness_tracker"
	DeviceTypeSmartScale     = "smart_scale"
	DeviceTypeWeatherStation = "weather_station"
)

// UserActivityStats represents activity statistics for a user
type UserActivityStats struct {
	UserID             string    `json:"user_id"`
	TotalActivities    int64     `json:"total_activities"`
	TotalCreditsEarned float64   `json:"total_credits_earned"`
	TotalDuration      int       `json:"total_duration"`
	TotalDistance      float64   `json:"total_distance"`
	StartDate          time.Time `json:"start_date"`
	EndDate            time.Time `json:"end_date"`
}
