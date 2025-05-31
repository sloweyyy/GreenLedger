package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// Certificate represents a carbon offset certificate
type Certificate struct {
	ID                uuid.UUID       `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID            string          `gorm:"not null;index" json:"user_id"`
	CertificateNumber string          `gorm:"uniqueIndex;not null" json:"certificate_number"`
	Type              string          `gorm:"not null;index" json:"type"`
	Status            string          `gorm:"not null;index;default:'pending'" json:"status"`
	CarbonOffset      decimal.Decimal `gorm:"type:decimal(15,3);not null" json:"carbon_offset"`
	CreditsUsed       decimal.Decimal `gorm:"type:decimal(15,3);not null" json:"credits_used"`
	ProjectName       string          `gorm:"not null" json:"project_name"`
	ProjectType       string          `gorm:"not null" json:"project_type"`
	ProjectLocation   string          `json:"project_location"`
	VerificationBody  string          `json:"verification_body"`
	Standard          string          `json:"standard"`
	VintageYear       int             `json:"vintage_year"`
	SerialNumber      string          `gorm:"uniqueIndex" json:"serial_number"`
	BlockchainTxHash  string          `gorm:"index" json:"blockchain_tx_hash"`
	BlockchainNetwork string          `json:"blockchain_network"`
	TokenID           string          `gorm:"index" json:"token_id"`
	MetadataURI       string          `json:"metadata_uri"`
	IssuedAt          *time.Time      `json:"issued_at"`
	ExpiresAt         *time.Time      `json:"expires_at"`
	RetiredAt         *time.Time      `json:"retired_at"`
	CreatedAt         time.Time       `json:"created_at"`
	UpdatedAt         time.Time       `json:"updated_at"`
	
	// Relationships
	Verifications []CertificateVerification `gorm:"foreignKey:CertificateID" json:"verifications,omitempty"`
	Transfers     []CertificateTransfer     `gorm:"foreignKey:CertificateID" json:"transfers,omitempty"`
}

// CertificateVerification represents a verification record for a certificate
type CertificateVerification struct {
	ID            uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CertificateID uuid.UUID `gorm:"type:uuid;not null;index" json:"certificate_id"`
	VerifierID    string    `gorm:"not null" json:"verifier_id"`
	VerifierName  string    `gorm:"not null" json:"verifier_name"`
	Status        string    `gorm:"not null" json:"status"`
	Comments      string    `json:"comments"`
	Evidence      string    `gorm:"type:jsonb" json:"evidence"`
	VerifiedAt    time.Time `gorm:"not null" json:"verified_at"`
	CreatedAt     time.Time `json:"created_at"`
	
	// Relationship
	Certificate Certificate `gorm:"foreignKey:CertificateID" json:"-"`
}

// CertificateTransfer represents a transfer of certificate ownership
type CertificateTransfer struct {
	ID            uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CertificateID uuid.UUID `gorm:"type:uuid;not null;index" json:"certificate_id"`
	FromUserID    string    `gorm:"not null;index" json:"from_user_id"`
	ToUserID      string    `gorm:"not null;index" json:"to_user_id"`
	TransferType  string    `gorm:"not null" json:"transfer_type"`
	Price         decimal.Decimal `gorm:"type:decimal(15,3)" json:"price"`
	Currency      string    `json:"currency"`
	Status        string    `gorm:"not null;default:'pending'" json:"status"`
	TxHash        string    `gorm:"index" json:"tx_hash"`
	TransferredAt *time.Time `json:"transferred_at"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	
	// Relationship
	Certificate Certificate `gorm:"foreignKey:CertificateID" json:"-"`
}

// CertificateTemplate represents a template for certificate generation
type CertificateTemplate struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Name        string    `gorm:"uniqueIndex;not null" json:"name"`
	Type        string    `gorm:"not null;index" json:"type"`
	Title       string    `gorm:"not null" json:"title"`
	Description string    `json:"description"`
	Template    string    `gorm:"type:text;not null" json:"template"`
	Metadata    string    `gorm:"type:jsonb" json:"metadata"`
	IsActive    bool      `gorm:"default:true" json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CertificateProject represents a carbon offset project
type CertificateProject struct {
	ID               uuid.UUID       `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Name             string          `gorm:"not null" json:"name"`
	Type             string          `gorm:"not null;index" json:"type"`
	Description      string          `json:"description"`
	Location         string          `json:"location"`
	Country          string          `json:"country"`
	Developer        string          `json:"developer"`
	VerificationBody string          `json:"verification_body"`
	Standard         string          `json:"standard"`
	Methodology      string          `json:"methodology"`
	VintageYear      int             `json:"vintage_year"`
	TotalCredits     decimal.Decimal `gorm:"type:decimal(15,3)" json:"total_credits"`
	AvailableCredits decimal.Decimal `gorm:"type:decimal(15,3)" json:"available_credits"`
	PricePerCredit   decimal.Decimal `gorm:"type:decimal(10,2)" json:"price_per_credit"`
	Currency         string          `gorm:"default:'USD'" json:"currency"`
	IsActive         bool            `gorm:"default:true" json:"is_active"`
	StartDate        time.Time       `json:"start_date"`
	EndDate          time.Time       `json:"end_date"`
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
	
	// Relationships
	Certificates []Certificate `gorm:"foreignKey:ProjectName;references:Name" json:"certificates,omitempty"`
}

// BeforeCreate hooks
func (c *Certificate) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}

func (cv *CertificateVerification) BeforeCreate(tx *gorm.DB) error {
	if cv.ID == uuid.Nil {
		cv.ID = uuid.New()
	}
	return nil
}

func (ct *CertificateTransfer) BeforeCreate(tx *gorm.DB) error {
	if ct.ID == uuid.Nil {
		ct.ID = uuid.New()
	}
	return nil
}

func (ct *CertificateTemplate) BeforeCreate(tx *gorm.DB) error {
	if ct.ID == uuid.Nil {
		ct.ID = uuid.New()
	}
	return nil
}

func (cp *CertificateProject) BeforeCreate(tx *gorm.DB) error {
	if cp.ID == uuid.Nil {
		cp.ID = uuid.New()
	}
	return nil
}

// Table names
func (Certificate) TableName() string             { return "certificates" }
func (CertificateVerification) TableName() string { return "certificate_verifications" }
func (CertificateTransfer) TableName() string     { return "certificate_transfers" }
func (CertificateTemplate) TableName() string     { return "certificate_templates" }
func (CertificateProject) TableName() string      { return "certificate_projects" }

// Certificate types
const (
	CertificateTypeOffset     = "offset"
	CertificateTypeReduction  = "reduction"
	CertificateTypeRemoval    = "removal"
	CertificateTypeAvoidance  = "avoidance"
)

// Certificate statuses
const (
	CertificateStatusPending   = "pending"
	CertificateStatusIssued    = "issued"
	CertificateStatusVerified  = "verified"
	CertificateStatusRetired   = "retired"
	CertificateStatusCancelled = "cancelled"
	CertificateStatusExpired   = "expired"
)

// Transfer types
const (
	TransferTypeSale     = "sale"
	TransferTypeGift     = "gift"
	TransferTypeRetire   = "retire"
	TransferTypeExchange = "exchange"
)

// Transfer statuses
const (
	TransferStatusPending   = "pending"
	TransferStatusCompleted = "completed"
	TransferStatusFailed    = "failed"
	TransferStatusCancelled = "cancelled"
)

// Verification statuses
const (
	VerificationStatusPending  = "pending"
	VerificationStatusApproved = "approved"
	VerificationStatusRejected = "rejected"
)

// Project types
const (
	ProjectTypeForestry      = "forestry"
	ProjectTypeRenewable     = "renewable_energy"
	ProjectTypeMethane       = "methane_capture"
	ProjectTypeSoil          = "soil_carbon"
	ProjectTypeDAC           = "direct_air_capture"
	ProjectTypeBiomass       = "biomass"
	ProjectTypeTransport     = "transport"
	ProjectTypeEfficiency    = "energy_efficiency"
)

// Standards
const (
	StandardVCS  = "VCS"
	StandardGold = "Gold_Standard"
	StandardCAR  = "CAR"
	StandardACR  = "ACR"
	StandardCDM  = "CDM"
)

// Helper methods for Certificate
func (c *Certificate) IsIssued() bool {
	return c.Status == CertificateStatusIssued || c.Status == CertificateStatusVerified
}

func (c *Certificate) IsRetired() bool {
	return c.Status == CertificateStatusRetired
}

func (c *Certificate) IsExpired() bool {
	return c.ExpiresAt != nil && time.Now().After(*c.ExpiresAt)
}

func (c *Certificate) CanTransfer() bool {
	return c.IsIssued() && !c.IsRetired() && !c.IsExpired()
}

// Helper methods for CertificateTransfer
func (ct *CertificateTransfer) IsCompleted() bool {
	return ct.Status == TransferStatusCompleted
}

func (ct *CertificateTransfer) IsPending() bool {
	return ct.Status == TransferStatusPending
}

// Helper methods for CertificateProject
func (cp *CertificateProject) HasAvailableCredits() bool {
	return cp.AvailableCredits.GreaterThan(decimal.Zero)
}

func (cp *CertificateProject) CanIssueCredits(amount decimal.Decimal) bool {
	return cp.IsActive && cp.AvailableCredits.GreaterThanOrEqual(amount)
}

// CertificateMetadata represents metadata for blockchain certificates
type CertificateMetadata struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Image       string                 `json:"image"`
	Attributes  []MetadataAttribute    `json:"attributes"`
	Properties  map[string]interface{} `json:"properties"`
}

// MetadataAttribute represents an attribute in certificate metadata
type MetadataAttribute struct {
	TraitType string      `json:"trait_type"`
	Value     interface{} `json:"value"`
}
