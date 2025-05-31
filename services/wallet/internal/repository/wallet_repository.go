package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/greenledger/services/wallet/internal/models"
	"github.com/greenledger/shared/database"
	"github.com/greenledger/shared/logger"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// WalletRepository handles wallet data operations
type WalletRepository struct {
	db     *database.PostgresDB
	logger *logger.Logger
}

// NewWalletRepository creates a new wallet repository
func NewWalletRepository(db *database.PostgresDB, logger *logger.Logger) *WalletRepository {
	return &WalletRepository{
		db:     db,
		logger: logger,
	}
}

// Create creates a new wallet
func (r *WalletRepository) Create(ctx context.Context, wallet *models.Wallet) error {
	err := r.db.WithContext(ctx).Create(wallet).Error
	if err != nil {
		r.logger.LogError(ctx, "failed to create wallet", err,
			logger.String("user_id", wallet.UserID))
		return fmt.Errorf("failed to create wallet: %w", err)
	}

	r.logger.LogInfo(ctx, "wallet created successfully",
		logger.String("wallet_id", wallet.ID.String()),
		logger.String("user_id", wallet.UserID))

	return nil
}

// GetByUserID retrieves a wallet by user ID
func (r *WalletRepository) GetByUserID(ctx context.Context, userID string) (*models.Wallet, error) {
	var wallet models.Wallet
	
	err := r.db.WithContext(ctx).First(&wallet, "user_id = ?", userID).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, database.ErrNotFound
		}
		r.logger.LogError(ctx, "failed to get wallet by user ID", err,
			logger.String("user_id", userID))
		return nil, fmt.Errorf("failed to get wallet: %w", err)
	}

	return &wallet, nil
}

// Update updates a wallet
func (r *WalletRepository) Update(ctx context.Context, wallet *models.Wallet) error {
	err := r.db.WithContext(ctx).Save(wallet).Error
	if err != nil {
		r.logger.LogError(ctx, "failed to update wallet", err,
			logger.String("wallet_id", wallet.ID.String()))
		return fmt.Errorf("failed to update wallet: %w", err)
	}

	return nil
}

// UpdateWithTransaction updates wallet and creates transaction atomically
func (r *WalletRepository) UpdateWithTransaction(ctx context.Context, wallet *models.Wallet, transaction *models.Transaction) error {
	return r.db.WithTransaction(ctx, func(tx *gorm.DB) error {
		// Update wallet
		if err := tx.Save(wallet).Error; err != nil {
			return fmt.Errorf("failed to update wallet: %w", err)
		}

		// Create transaction
		if err := tx.Create(transaction).Error; err != nil {
			return fmt.Errorf("failed to create transaction: %w", err)
		}

		r.logger.LogInfo(ctx, "wallet updated with transaction",
			logger.String("wallet_id", wallet.ID.String()),
			logger.String("transaction_id", transaction.ID.String()))

		return nil
	})
}

// ProcessTransfer processes a credit transfer between two wallets atomically
func (r *WalletRepository) ProcessTransfer(ctx context.Context, fromWallet, toWallet *models.Wallet, debitTx, creditTx *models.Transaction) error {
	return r.db.WithTransaction(ctx, func(tx *gorm.DB) error {
		// Update sender wallet
		if err := tx.Save(fromWallet).Error; err != nil {
			return fmt.Errorf("failed to update sender wallet: %w", err)
		}

		// Update receiver wallet
		if err := tx.Save(toWallet).Error; err != nil {
			return fmt.Errorf("failed to update receiver wallet: %w", err)
		}

		// Create debit transaction
		if err := tx.Create(debitTx).Error; err != nil {
			return fmt.Errorf("failed to create debit transaction: %w", err)
		}

		// Create credit transaction
		if err := tx.Create(creditTx).Error; err != nil {
			return fmt.Errorf("failed to create credit transaction: %w", err)
		}

		r.logger.LogInfo(ctx, "transfer processed successfully",
			logger.String("from_user_id", fromWallet.UserID),
			logger.String("to_user_id", toWallet.UserID),
			logger.String("amount", debitTx.Amount.String()))

		return nil
	})
}

// GetWalletStats retrieves wallet statistics
func (r *WalletRepository) GetWalletStats(ctx context.Context, userID string, startDate, endDate time.Time) (*WalletStats, error) {
	var stats WalletStats

	// Get current wallet
	wallet, err := r.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get wallet: %w", err)
	}

	stats.UserID = userID
	stats.CurrentBalance = wallet.AvailableCredits
	stats.TotalEarned = wallet.TotalEarned
	stats.TotalSpent = wallet.TotalSpent

	// Get transaction counts and amounts for the period
	var result struct {
		CreditCount  int64           `gorm:"column:credit_count"`
		CreditAmount decimal.Decimal `gorm:"column:credit_amount"`
		DebitCount   int64           `gorm:"column:debit_count"`
		DebitAmount  decimal.Decimal `gorm:"column:debit_amount"`
	}

	err = r.db.WithContext(ctx).
		Model(&models.Transaction{}).
		Select(`
			COUNT(CASE WHEN type IN ('credit_earned', 'transfer_in', 'refund', 'bonus') THEN 1 END) as credit_count,
			COALESCE(SUM(CASE WHEN type IN ('credit_earned', 'transfer_in', 'refund', 'bonus') THEN amount END), 0) as credit_amount,
			COUNT(CASE WHEN type IN ('credit_spent', 'transfer_out', 'penalty') THEN 1 END) as debit_count,
			COALESCE(SUM(CASE WHEN type IN ('credit_spent', 'transfer_out', 'penalty') THEN amount END), 0) as debit_amount
		`).
		Where("user_id = ? AND created_at >= ? AND created_at <= ? AND status = 'completed'", userID, startDate, endDate).
		Scan(&result).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get transaction stats: %w", err)
	}

	stats.PeriodCredits = result.CreditAmount
	stats.PeriodDebits = result.DebitAmount
	stats.PeriodTransactions = result.CreditCount + result.DebitCount
	stats.StartDate = startDate
	stats.EndDate = endDate

	return &stats, nil
}

// GetTopUsers retrieves users with highest balances
func (r *WalletRepository) GetTopUsers(ctx context.Context, limit int) ([]*models.Wallet, error) {
	var wallets []*models.Wallet

	err := r.db.WithContext(ctx).
		Order("available_credits DESC").
		Limit(limit).
		Find(&wallets).Error

	if err != nil {
		r.logger.LogError(ctx, "failed to get top users", err)
		return nil, fmt.Errorf("failed to get top users: %w", err)
	}

	return wallets, nil
}

// CreateSnapshot creates a wallet snapshot
func (r *WalletRepository) CreateSnapshot(ctx context.Context, snapshot *models.WalletSnapshot) error {
	err := r.db.WithContext(ctx).Create(snapshot).Error
	if err != nil {
		r.logger.LogError(ctx, "failed to create wallet snapshot", err,
			logger.String("user_id", snapshot.UserID))
		return fmt.Errorf("failed to create snapshot: %w", err)
	}

	return nil
}

// GetSnapshots retrieves wallet snapshots for a user
func (r *WalletRepository) GetSnapshots(ctx context.Context, userID string, limit int) ([]*models.WalletSnapshot, error) {
	var snapshots []*models.WalletSnapshot

	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("snapshot_date DESC").
		Limit(limit).
		Find(&snapshots).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get snapshots: %w", err)
	}

	return snapshots, nil
}

// WalletStats represents wallet statistics
type WalletStats struct {
	UserID             string          `json:"user_id"`
	CurrentBalance     decimal.Decimal `json:"current_balance"`
	TotalEarned        decimal.Decimal `json:"total_earned"`
	TotalSpent         decimal.Decimal `json:"total_spent"`
	PeriodCredits      decimal.Decimal `json:"period_credits"`
	PeriodDebits       decimal.Decimal `json:"period_debits"`
	PeriodTransactions int64           `json:"period_transactions"`
	StartDate          time.Time       `json:"start_date"`
	EndDate            time.Time       `json:"end_date"`
}
