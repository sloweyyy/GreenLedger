package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// Wallet represents a user's carbon credit wallet
type Wallet struct {
	ID               uuid.UUID       `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID           string          `gorm:"uniqueIndex;not null" json:"user_id"`
	AvailableCredits decimal.Decimal `gorm:"type:decimal(15,3);not null;default:0" json:"available_credits"`
	PendingCredits   decimal.Decimal `gorm:"type:decimal(15,3);not null;default:0" json:"pending_credits"`
	TotalEarned      decimal.Decimal `gorm:"type:decimal(15,3);not null;default:0" json:"total_earned"`
	TotalSpent       decimal.Decimal `gorm:"type:decimal(15,3);not null;default:0" json:"total_spent"`
	LastUpdated      time.Time       `gorm:"not null;default:now()" json:"last_updated"`
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
	
	// Relationships
	Transactions []Transaction `gorm:"foreignKey:UserID;references:UserID" json:"transactions,omitempty"`
}

// Transaction represents a credit transaction
type Transaction struct {
	ID            uuid.UUID       `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID        string          `gorm:"not null;index" json:"user_id"`
	Type          string          `gorm:"not null;index" json:"type"`
	Status        string          `gorm:"not null;index;default:'pending'" json:"status"`
	Amount        decimal.Decimal `gorm:"type:decimal(15,3);not null" json:"amount"`
	BalanceAfter  decimal.Decimal `gorm:"type:decimal(15,3);not null" json:"balance_after"`
	Source        string          `gorm:"not null" json:"source"`
	Description   string          `gorm:"not null" json:"description"`
	ReferenceID   string          `gorm:"index" json:"reference_id"`
	FromUserID    string          `gorm:"index" json:"from_user_id"`
	ToUserID      string          `gorm:"index" json:"to_user_id"`
	Metadata      string          `gorm:"type:jsonb" json:"metadata"`
	ProcessedAt   *time.Time      `json:"processed_at"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
	
	// Relationship
	Wallet Wallet `gorm:"foreignKey:UserID;references:UserID" json:"-"`
}

// TransactionBatch represents a batch of transactions for atomic processing
type TransactionBatch struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	BatchID      string    `gorm:"uniqueIndex;not null" json:"batch_id"`
	Status       string    `gorm:"not null;default:'pending'" json:"status"`
	TotalAmount  decimal.Decimal `gorm:"type:decimal(15,3);not null" json:"total_amount"`
	Description  string    `json:"description"`
	ProcessedAt  *time.Time `json:"processed_at"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	
	// Relationships
	Transactions []Transaction `gorm:"foreignKey:ReferenceID;references:BatchID" json:"transactions,omitempty"`
}

// CreditReservation represents a temporary hold on credits
type CreditReservation struct {
	ID          uuid.UUID       `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID      string          `gorm:"not null;index" json:"user_id"`
	Amount      decimal.Decimal `gorm:"type:decimal(15,3);not null" json:"amount"`
	Purpose     string          `gorm:"not null" json:"purpose"`
	ReferenceID string          `gorm:"index" json:"reference_id"`
	ExpiresAt   time.Time       `gorm:"not null" json:"expires_at"`
	IsReleased  bool            `gorm:"default:false" json:"is_released"`
	ReleasedAt  *time.Time      `json:"released_at"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

// WalletSnapshot represents a point-in-time snapshot of wallet balances
type WalletSnapshot struct {
	ID               uuid.UUID       `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID           string          `gorm:"not null;index" json:"user_id"`
	AvailableCredits decimal.Decimal `gorm:"type:decimal(15,3);not null" json:"available_credits"`
	PendingCredits   decimal.Decimal `gorm:"type:decimal(15,3);not null" json:"pending_credits"`
	TotalEarned      decimal.Decimal `gorm:"type:decimal(15,3);not null" json:"total_earned"`
	TotalSpent       decimal.Decimal `gorm:"type:decimal(15,3);not null" json:"total_spent"`
	SnapshotDate     time.Time       `gorm:"not null;index" json:"snapshot_date"`
	CreatedAt        time.Time       `json:"created_at"`
}

// BeforeCreate hooks
func (w *Wallet) BeforeCreate(tx *gorm.DB) error {
	if w.ID == uuid.Nil {
		w.ID = uuid.New()
	}
	return nil
}

func (t *Transaction) BeforeCreate(tx *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return nil
}

func (tb *TransactionBatch) BeforeCreate(tx *gorm.DB) error {
	if tb.ID == uuid.Nil {
		tb.ID = uuid.New()
	}
	return nil
}

func (cr *CreditReservation) BeforeCreate(tx *gorm.DB) error {
	if cr.ID == uuid.Nil {
		cr.ID = uuid.New()
	}
	return nil
}

func (ws *WalletSnapshot) BeforeCreate(tx *gorm.DB) error {
	if ws.ID == uuid.Nil {
		ws.ID = uuid.New()
	}
	return nil
}

// Table names
func (Wallet) TableName() string            { return "wallets" }
func (Transaction) TableName() string       { return "transactions" }
func (TransactionBatch) TableName() string  { return "transaction_batches" }
func (CreditReservation) TableName() string { return "credit_reservations" }
func (WalletSnapshot) TableName() string    { return "wallet_snapshots" }

// Transaction types
const (
	TransactionTypeCreditEarned = "credit_earned"
	TransactionTypeCreditSpent  = "credit_spent"
	TransactionTypeTransferIn   = "transfer_in"
	TransactionTypeTransferOut  = "transfer_out"
	TransactionTypeAdjustment   = "adjustment"
	TransactionTypeRefund       = "refund"
	TransactionTypePenalty      = "penalty"
	TransactionTypeBonus        = "bonus"
)

// Transaction statuses
const (
	TransactionStatusPending   = "pending"
	TransactionStatusCompleted = "completed"
	TransactionStatusFailed    = "failed"
	TransactionStatusCancelled = "cancelled"
	TransactionStatusExpired   = "expired"
)

// Credit sources
const (
	CreditSourceEcoActivity   = "eco_activity"
	CreditSourceCarbonOffset  = "carbon_offset"
	CreditSourcePurchase      = "purchase"
	CreditSourceReward        = "reward"
	CreditSourceTransfer      = "transfer"
	CreditSourceAdjustment    = "adjustment"
	CreditSourceRefund        = "refund"
	CreditSourceBonus         = "bonus"
	CreditSourceChallenge     = "challenge"
	CreditSourceReferral      = "referral"
)

// Batch statuses
const (
	BatchStatusPending   = "pending"
	BatchStatusProcessed = "processed"
	BatchStatusFailed    = "failed"
	BatchStatusCancelled = "cancelled"
)

// Helper methods for Wallet
func (w *Wallet) GetTotalBalance() decimal.Decimal {
	return w.AvailableCredits.Add(w.PendingCredits)
}

func (w *Wallet) CanSpend(amount decimal.Decimal) bool {
	return w.AvailableCredits.GreaterThanOrEqual(amount)
}

func (w *Wallet) GetNetBalance() decimal.Decimal {
	return w.TotalEarned.Sub(w.TotalSpent)
}

// Helper methods for Transaction
func (t *Transaction) IsCompleted() bool {
	return t.Status == TransactionStatusCompleted
}

func (t *Transaction) IsPending() bool {
	return t.Status == TransactionStatusPending
}

func (t *Transaction) IsCredit() bool {
	return t.Type == TransactionTypeCreditEarned || 
		   t.Type == TransactionTypeTransferIn || 
		   t.Type == TransactionTypeRefund ||
		   t.Type == TransactionTypeBonus
}

func (t *Transaction) IsDebit() bool {
	return t.Type == TransactionTypeCreditSpent || 
		   t.Type == TransactionTypeTransferOut || 
		   t.Type == TransactionTypePenalty
}

// Helper methods for CreditReservation
func (cr *CreditReservation) IsExpired() bool {
	return time.Now().After(cr.ExpiresAt) && !cr.IsReleased
}

func (cr *CreditReservation) IsActive() bool {
	return !cr.IsReleased && !cr.IsExpired()
}
