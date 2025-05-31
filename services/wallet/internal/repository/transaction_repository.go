package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/greenledger/services/wallet/internal/models"
	"github.com/greenledger/shared/database"
	"github.com/greenledger/shared/logger"
	"gorm.io/gorm"
)

// TransactionRepository handles transaction data operations
type TransactionRepository struct {
	db     *database.PostgresDB
	logger *logger.Logger
}

// NewTransactionRepository creates a new transaction repository
func NewTransactionRepository(db *database.PostgresDB, logger *logger.Logger) *TransactionRepository {
	return &TransactionRepository{
		db:     db,
		logger: logger,
	}
}

// Create creates a new transaction
func (r *TransactionRepository) Create(ctx context.Context, transaction *models.Transaction) error {
	err := r.db.WithContext(ctx).Create(transaction).Error
	if err != nil {
		r.logger.LogError(ctx, "failed to create transaction", err,
			logger.String("user_id", transaction.UserID))
		return fmt.Errorf("failed to create transaction: %w", err)
	}

	return nil
}

// GetByID retrieves a transaction by ID
func (r *TransactionRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Transaction, error) {
	var transaction models.Transaction
	
	err := r.db.WithContext(ctx).First(&transaction, "id = ?", id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, database.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get transaction: %w", err)
	}

	return &transaction, nil
}

// GetByUserID retrieves transactions for a specific user
func (r *TransactionRepository) GetByUserID(ctx context.Context, userID string, limit, offset int) ([]*models.Transaction, int64, error) {
	var transactions []*models.Transaction
	var total int64

	// Get total count
	if err := r.db.WithContext(ctx).Model(&models.Transaction{}).
		Where("user_id = ?", userID).Count(&total).Error; err != nil {
		r.logger.LogError(ctx, "failed to count transactions", err,
			logger.String("user_id", userID))
		return nil, 0, fmt.Errorf("failed to count transactions: %w", err)
	}

	// Get transactions
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&transactions).Error

	if err != nil {
		r.logger.LogError(ctx, "failed to get transactions by user ID", err,
			logger.String("user_id", userID))
		return nil, 0, fmt.Errorf("failed to get transactions: %w", err)
	}

	return transactions, total, nil
}

// GetByUserIDAndDateRange retrieves transactions for a user within a date range
func (r *TransactionRepository) GetByUserIDAndDateRange(ctx context.Context, userID string, startDate, endDate time.Time, limit, offset int) ([]*models.Transaction, int64, error) {
	var transactions []*models.Transaction
	var total int64

	query := r.db.WithContext(ctx).Model(&models.Transaction{}).
		Where("user_id = ? AND created_at >= ? AND created_at <= ?", userID, startDate, endDate)

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		r.logger.LogError(ctx, "failed to count transactions in date range", err,
			logger.String("user_id", userID))
		return nil, 0, fmt.Errorf("failed to count transactions: %w", err)
	}

	// Get transactions
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND created_at >= ? AND created_at <= ?", userID, startDate, endDate).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&transactions).Error

	if err != nil {
		r.logger.LogError(ctx, "failed to get transactions by date range", err,
			logger.String("user_id", userID))
		return nil, 0, fmt.Errorf("failed to get transactions: %w", err)
	}

	return transactions, total, nil
}

// GetByReferenceID retrieves transactions by reference ID
func (r *TransactionRepository) GetByReferenceID(ctx context.Context, referenceID string) ([]*models.Transaction, error) {
	var transactions []*models.Transaction
	
	err := r.db.WithContext(ctx).
		Where("reference_id = ?", referenceID).
		Order("created_at DESC").
		Find(&transactions).Error
	
	if err != nil {
		return nil, fmt.Errorf("failed to get transactions by reference ID: %w", err)
	}

	return transactions, nil
}

// GetByType retrieves transactions by type
func (r *TransactionRepository) GetByType(ctx context.Context, transactionType string, limit, offset int) ([]*models.Transaction, int64, error) {
	var transactions []*models.Transaction
	var total int64

	// Get total count
	if err := r.db.WithContext(ctx).Model(&models.Transaction{}).
		Where("type = ?", transactionType).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count transactions by type: %w", err)
	}

	// Get transactions
	err := r.db.WithContext(ctx).
		Where("type = ?", transactionType).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&transactions).Error

	if err != nil {
		return nil, 0, fmt.Errorf("failed to get transactions by type: %w", err)
	}

	return transactions, total, nil
}

// GetPendingTransactions retrieves pending transactions
func (r *TransactionRepository) GetPendingTransactions(ctx context.Context, limit, offset int) ([]*models.Transaction, int64, error) {
	var transactions []*models.Transaction
	var total int64

	// Get total count
	if err := r.db.WithContext(ctx).Model(&models.Transaction{}).
		Where("status = ?", models.TransactionStatusPending).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count pending transactions: %w", err)
	}

	// Get transactions
	err := r.db.WithContext(ctx).
		Where("status = ?", models.TransactionStatusPending).
		Order("created_at ASC").
		Limit(limit).
		Offset(offset).
		Find(&transactions).Error

	if err != nil {
		return nil, 0, fmt.Errorf("failed to get pending transactions: %w", err)
	}

	return transactions, total, nil
}

// Update updates a transaction
func (r *TransactionRepository) Update(ctx context.Context, transaction *models.Transaction) error {
	err := r.db.WithContext(ctx).Save(transaction).Error
	if err != nil {
		r.logger.LogError(ctx, "failed to update transaction", err,
			logger.String("transaction_id", transaction.ID.String()))
		return fmt.Errorf("failed to update transaction: %w", err)
	}

	return nil
}

// GetRecentTransactions retrieves recent transactions across all users
func (r *TransactionRepository) GetRecentTransactions(ctx context.Context, limit int) ([]*models.Transaction, error) {
	var transactions []*models.Transaction

	err := r.db.WithContext(ctx).
		Where("status = ?", models.TransactionStatusCompleted).
		Order("created_at DESC").
		Limit(limit).
		Find(&transactions).Error

	if err != nil {
		r.logger.LogError(ctx, "failed to get recent transactions", err)
		return nil, fmt.Errorf("failed to get recent transactions: %w", err)
	}

	return transactions, nil
}

// GetTransactionSummary retrieves transaction summary for a user
func (r *TransactionRepository) GetTransactionSummary(ctx context.Context, userID string, startDate, endDate time.Time) (*TransactionSummary, error) {
	var summary TransactionSummary

	err := r.db.WithContext(ctx).
		Model(&models.Transaction{}).
		Select(`
			COUNT(*) as total_transactions,
			COUNT(CASE WHEN type IN ('credit_earned', 'transfer_in', 'refund', 'bonus') THEN 1 END) as credit_transactions,
			COUNT(CASE WHEN type IN ('credit_spent', 'transfer_out', 'penalty') THEN 1 END) as debit_transactions,
			COALESCE(SUM(CASE WHEN type IN ('credit_earned', 'transfer_in', 'refund', 'bonus') THEN amount END), 0) as total_credits,
			COALESCE(SUM(CASE WHEN type IN ('credit_spent', 'transfer_out', 'penalty') THEN amount END), 0) as total_debits
		`).
		Where("user_id = ? AND created_at >= ? AND created_at <= ? AND status = 'completed'", userID, startDate, endDate).
		Scan(&summary).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get transaction summary: %w", err)
	}

	summary.UserID = userID
	summary.StartDate = startDate
	summary.EndDate = endDate

	return &summary, nil
}

// TransactionSummary represents a summary of transactions
type TransactionSummary struct {
	UserID              string    `json:"user_id"`
	TotalTransactions   int64     `json:"total_transactions"`
	CreditTransactions  int64     `json:"credit_transactions"`
	DebitTransactions   int64     `json:"debit_transactions"`
	TotalCredits        float64   `json:"total_credits"`
	TotalDebits         float64   `json:"total_debits"`
	StartDate           time.Time `json:"start_date"`
	EndDate             time.Time `json:"end_date"`
}
