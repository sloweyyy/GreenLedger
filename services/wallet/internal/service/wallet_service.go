package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/greenledger/services/wallet/internal/models"
	"github.com/greenledger/services/wallet/internal/repository"
	"github.com/greenledger/shared/logger"
	"github.com/shopspring/decimal"
)

// WalletService handles wallet operations
type WalletService struct {
	walletRepo      *repository.WalletRepository
	transactionRepo *repository.TransactionRepository
	eventPublisher  EventPublisher
	logger          *logger.Logger
}

// NewWalletService creates a new wallet service
func NewWalletService(
	walletRepo *repository.WalletRepository,
	transactionRepo *repository.TransactionRepository,
	eventPublisher EventPublisher,
	logger *logger.Logger,
) *WalletService {
	return &WalletService{
		walletRepo:      walletRepo,
		transactionRepo: transactionRepo,
		eventPublisher:  eventPublisher,
		logger:          logger,
	}
}

// CreditBalanceRequest represents a request to credit a wallet
type CreditBalanceRequest struct {
	UserID      string                 `json:"user_id" binding:"required"`
	Amount      decimal.Decimal        `json:"amount" binding:"required"`
	Source      string                 `json:"source" binding:"required"`
	Description string                 `json:"description" binding:"required"`
	ReferenceID string                 `json:"reference_id"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// DebitBalanceRequest represents a request to debit a wallet
type DebitBalanceRequest struct {
	UserID      string                 `json:"user_id" binding:"required"`
	Amount      decimal.Decimal        `json:"amount" binding:"required"`
	Description string                 `json:"description" binding:"required"`
	ReferenceID string                 `json:"reference_id"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// TransferCreditsRequest represents a request to transfer credits
type TransferCreditsRequest struct {
	FromUserID  string          `json:"from_user_id" binding:"required"`
	ToUserID    string          `json:"to_user_id" binding:"required"`
	Amount      decimal.Decimal `json:"amount" binding:"required"`
	Description string          `json:"description" binding:"required"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// WalletResponse represents a wallet in API responses
type WalletResponse struct {
	UserID           string          `json:"user_id"`
	AvailableCredits decimal.Decimal `json:"available_credits"`
	PendingCredits   decimal.Decimal `json:"pending_credits"`
	TotalEarned      decimal.Decimal `json:"total_earned"`
	TotalSpent       decimal.Decimal `json:"total_spent"`
	LastUpdated      time.Time       `json:"last_updated"`
}

// TransactionResponse represents a transaction in API responses
type TransactionResponse struct {
	ID           uuid.UUID       `json:"id"`
	UserID       string          `json:"user_id"`
	Type         string          `json:"type"`
	Status       string          `json:"status"`
	Amount       decimal.Decimal `json:"amount"`
	BalanceAfter decimal.Decimal `json:"balance_after"`
	Source       string          `json:"source"`
	Description  string          `json:"description"`
	ReferenceID  string          `json:"reference_id"`
	FromUserID   string          `json:"from_user_id,omitempty"`
	ToUserID     string          `json:"to_user_id,omitempty"`
	ProcessedAt  *time.Time      `json:"processed_at"`
	CreatedAt    time.Time       `json:"created_at"`
}

// EventPublisher interface for publishing wallet events
type EventPublisher interface {
	PublishBalanceUpdated(ctx context.Context, event *BalanceUpdatedEvent) error
	PublishTransferCompleted(ctx context.Context, event *TransferCompletedEvent) error
}

// BalanceUpdatedEvent represents a balance update event
type BalanceUpdatedEvent struct {
	UserID           string          `json:"user_id"`
	TransactionID    string          `json:"transaction_id"`
	TransactionType  string          `json:"transaction_type"`
	Amount           decimal.Decimal `json:"amount"`
	BalanceAfter     decimal.Decimal `json:"balance_after"`
	Source           string          `json:"source"`
	Timestamp        time.Time       `json:"timestamp"`
}

// TransferCompletedEvent represents a transfer completion event
type TransferCompletedEvent struct {
	TransferID   string          `json:"transfer_id"`
	FromUserID   string          `json:"from_user_id"`
	ToUserID     string          `json:"to_user_id"`
	Amount       decimal.Decimal `json:"amount"`
	Description  string          `json:"description"`
	Timestamp    time.Time       `json:"timestamp"`
}

// GetBalance retrieves a user's wallet balance
func (s *WalletService) GetBalance(ctx context.Context, userID string) (*WalletResponse, error) {
	wallet, err := s.walletRepo.GetByUserID(ctx, userID)
	if err != nil {
		// Create wallet if it doesn't exist
		if err.Error() == "record not found" {
			wallet, err = s.createWallet(ctx, userID)
			if err != nil {
				return nil, fmt.Errorf("failed to create wallet: %w", err)
			}
		} else {
			return nil, fmt.Errorf("failed to get wallet: %w", err)
		}
	}

	return s.walletToResponse(wallet), nil
}

// CreditBalance credits a user's wallet
func (s *WalletService) CreditBalance(ctx context.Context, req *CreditBalanceRequest) (*TransactionResponse, error) {
	s.logger.LogInfo(ctx, "crediting wallet balance",
		logger.String("user_id", req.UserID),
		logger.String("amount", req.Amount.String()),
		logger.String("source", req.Source))

	// Validate amount
	if req.Amount.LessThanOrEqual(decimal.Zero) {
		return nil, fmt.Errorf("amount must be positive")
	}

	// Get or create wallet
	wallet, err := s.getOrCreateWallet(ctx, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get wallet: %w", err)
	}

	// Create transaction
	transaction := &models.Transaction{
		UserID:      req.UserID,
		Type:        models.TransactionTypeCreditEarned,
		Status:      models.TransactionStatusCompleted,
		Amount:      req.Amount,
		Source:      req.Source,
		Description: req.Description,
		ReferenceID: req.ReferenceID,
	}

	// Process transaction atomically
	updatedWallet, err := s.processTransaction(ctx, wallet, transaction)
	if err != nil {
		return nil, fmt.Errorf("failed to process transaction: %w", err)
	}

	// Publish event
	event := &BalanceUpdatedEvent{
		UserID:          req.UserID,
		TransactionID:   transaction.ID.String(),
		TransactionType: transaction.Type,
		Amount:          req.Amount,
		BalanceAfter:    updatedWallet.AvailableCredits,
		Source:          req.Source,
		Timestamp:       time.Now().UTC(),
	}

	if err := s.eventPublisher.PublishBalanceUpdated(ctx, event); err != nil {
		s.logger.LogError(ctx, "failed to publish balance updated event", err)
	}

	s.logger.LogInfo(ctx, "wallet balance credited successfully",
		logger.String("user_id", req.UserID),
		logger.String("transaction_id", transaction.ID.String()))

	return s.transactionToResponse(transaction), nil
}

// DebitBalance debits a user's wallet
func (s *WalletService) DebitBalance(ctx context.Context, req *DebitBalanceRequest) (*TransactionResponse, error) {
	s.logger.LogInfo(ctx, "debiting wallet balance",
		logger.String("user_id", req.UserID),
		logger.String("amount", req.Amount.String()))

	// Validate amount
	if req.Amount.LessThanOrEqual(decimal.Zero) {
		return nil, fmt.Errorf("amount must be positive")
	}

	// Get wallet
	wallet, err := s.walletRepo.GetByUserID(ctx, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get wallet: %w", err)
	}

	// Check if user has sufficient balance
	if !wallet.CanSpend(req.Amount) {
		return nil, fmt.Errorf("insufficient balance")
	}

	// Create transaction
	transaction := &models.Transaction{
		UserID:      req.UserID,
		Type:        models.TransactionTypeCreditSpent,
		Status:      models.TransactionStatusCompleted,
		Amount:      req.Amount,
		Source:      "spending",
		Description: req.Description,
		ReferenceID: req.ReferenceID,
	}

	// Process transaction atomically
	updatedWallet, err := s.processTransaction(ctx, wallet, transaction)
	if err != nil {
		return nil, fmt.Errorf("failed to process transaction: %w", err)
	}

	// Publish event
	event := &BalanceUpdatedEvent{
		UserID:          req.UserID,
		TransactionID:   transaction.ID.String(),
		TransactionType: transaction.Type,
		Amount:          req.Amount.Neg(),
		BalanceAfter:    updatedWallet.AvailableCredits,
		Source:          "spending",
		Timestamp:       time.Now().UTC(),
	}

	if err := s.eventPublisher.PublishBalanceUpdated(ctx, event); err != nil {
		s.logger.LogError(ctx, "failed to publish balance updated event", err)
	}

	s.logger.LogInfo(ctx, "wallet balance debited successfully",
		logger.String("user_id", req.UserID),
		logger.String("transaction_id", transaction.ID.String()))

	return s.transactionToResponse(transaction), nil
}

// TransferCredits transfers credits between users
func (s *WalletService) TransferCredits(ctx context.Context, req *TransferCreditsRequest) (*TransferResponse, error) {
	s.logger.LogInfo(ctx, "transferring credits",
		logger.String("from_user_id", req.FromUserID),
		logger.String("to_user_id", req.ToUserID),
		logger.String("amount", req.Amount.String()))

	// Validate amount
	if req.Amount.LessThanOrEqual(decimal.Zero) {
		return nil, fmt.Errorf("amount must be positive")
	}

	// Validate users are different
	if req.FromUserID == req.ToUserID {
		return nil, fmt.Errorf("cannot transfer to the same user")
	}

	// Get sender wallet
	fromWallet, err := s.walletRepo.GetByUserID(ctx, req.FromUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get sender wallet: %w", err)
	}

	// Check if sender has sufficient balance
	if !fromWallet.CanSpend(req.Amount) {
		return nil, fmt.Errorf("insufficient balance")
	}

	// Get or create receiver wallet
	toWallet, err := s.getOrCreateWallet(ctx, req.ToUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get receiver wallet: %w", err)
	}

	// Generate transfer ID
	transferID := uuid.New().String()

	// Create debit transaction for sender
	debitTransaction := &models.Transaction{
		UserID:      req.FromUserID,
		Type:        models.TransactionTypeTransferOut,
		Status:      models.TransactionStatusCompleted,
		Amount:      req.Amount,
		Source:      models.CreditSourceTransfer,
		Description: req.Description,
		ReferenceID: transferID,
		ToUserID:    req.ToUserID,
	}

	// Create credit transaction for receiver
	creditTransaction := &models.Transaction{
		UserID:      req.ToUserID,
		Type:        models.TransactionTypeTransferIn,
		Status:      models.TransactionStatusCompleted,
		Amount:      req.Amount,
		Source:      models.CreditSourceTransfer,
		Description: req.Description,
		ReferenceID: transferID,
		FromUserID:  req.FromUserID,
	}

	// Process transfer atomically
	updatedFromWallet, updatedToWallet, err := s.processTransfer(ctx, fromWallet, toWallet, debitTransaction, creditTransaction)
	if err != nil {
		return nil, fmt.Errorf("failed to process transfer: %w", err)
	}

	// Publish transfer completed event
	transferEvent := &TransferCompletedEvent{
		TransferID:  transferID,
		FromUserID:  req.FromUserID,
		ToUserID:    req.ToUserID,
		Amount:      req.Amount,
		Description: req.Description,
		Timestamp:   time.Now().UTC(),
	}

	if err := s.eventPublisher.PublishTransferCompleted(ctx, transferEvent); err != nil {
		s.logger.LogError(ctx, "failed to publish transfer completed event", err)
	}

	s.logger.LogInfo(ctx, "credits transferred successfully",
		logger.String("transfer_id", transferID))

	return &TransferResponse{
		TransferID:      transferID,
		FromTransaction: s.transactionToResponse(debitTransaction),
		ToTransaction:   s.transactionToResponse(creditTransaction),
		FromBalance:     s.walletToResponse(updatedFromWallet),
		ToBalance:       s.walletToResponse(updatedToWallet),
	}, nil
}

// GetTransactionHistory retrieves transaction history for a user
func (s *WalletService) GetTransactionHistory(ctx context.Context, userID string, limit, offset int) ([]*TransactionResponse, int64, error) {
	transactions, total, err := s.transactionRepo.GetByUserID(ctx, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get transaction history: %w", err)
	}

	responses := make([]*TransactionResponse, len(transactions))
	for i, transaction := range transactions {
		responses[i] = s.transactionToResponse(transaction)
	}

	return responses, total, nil
}

// Helper methods
func (s *WalletService) createWallet(ctx context.Context, userID string) (*models.Wallet, error) {
	wallet := &models.Wallet{
		UserID:           userID,
		AvailableCredits: decimal.Zero,
		PendingCredits:   decimal.Zero,
		TotalEarned:      decimal.Zero,
		TotalSpent:       decimal.Zero,
		LastUpdated:      time.Now().UTC(),
	}

	if err := s.walletRepo.Create(ctx, wallet); err != nil {
		return nil, err
	}

	return wallet, nil
}

func (s *WalletService) getOrCreateWallet(ctx context.Context, userID string) (*models.Wallet, error) {
	wallet, err := s.walletRepo.GetByUserID(ctx, userID)
	if err != nil {
		if err.Error() == "record not found" {
			return s.createWallet(ctx, userID)
		}
		return nil, err
	}
	return wallet, nil
}

func (s *WalletService) processTransaction(ctx context.Context, wallet *models.Wallet, transaction *models.Transaction) (*models.Wallet, error) {
	// Update wallet balance based on transaction type
	if transaction.IsCredit() {
		wallet.AvailableCredits = wallet.AvailableCredits.Add(transaction.Amount)
		wallet.TotalEarned = wallet.TotalEarned.Add(transaction.Amount)
	} else if transaction.IsDebit() {
		wallet.AvailableCredits = wallet.AvailableCredits.Sub(transaction.Amount)
		wallet.TotalSpent = wallet.TotalSpent.Add(transaction.Amount)
	}

	wallet.LastUpdated = time.Now().UTC()
	transaction.BalanceAfter = wallet.AvailableCredits
	transaction.ProcessedAt = &wallet.LastUpdated

	// Save both wallet and transaction atomically
	if err := s.walletRepo.UpdateWithTransaction(ctx, wallet, transaction); err != nil {
		return nil, err
	}

	return wallet, nil
}

func (s *WalletService) processTransfer(ctx context.Context, fromWallet, toWallet *models.Wallet, debitTx, creditTx *models.Transaction) (*models.Wallet, *models.Wallet, error) {
	// Update sender wallet
	fromWallet.AvailableCredits = fromWallet.AvailableCredits.Sub(debitTx.Amount)
	fromWallet.TotalSpent = fromWallet.TotalSpent.Add(debitTx.Amount)
	fromWallet.LastUpdated = time.Now().UTC()
	debitTx.BalanceAfter = fromWallet.AvailableCredits
	debitTx.ProcessedAt = &fromWallet.LastUpdated

	// Update receiver wallet
	toWallet.AvailableCredits = toWallet.AvailableCredits.Add(creditTx.Amount)
	toWallet.TotalEarned = toWallet.TotalEarned.Add(creditTx.Amount)
	toWallet.LastUpdated = time.Now().UTC()
	creditTx.BalanceAfter = toWallet.AvailableCredits
	creditTx.ProcessedAt = &toWallet.LastUpdated

	// Save transfer atomically
	if err := s.walletRepo.ProcessTransfer(ctx, fromWallet, toWallet, debitTx, creditTx); err != nil {
		return nil, nil, err
	}

	return fromWallet, toWallet, nil
}

func (s *WalletService) walletToResponse(wallet *models.Wallet) *WalletResponse {
	return &WalletResponse{
		UserID:           wallet.UserID,
		AvailableCredits: wallet.AvailableCredits,
		PendingCredits:   wallet.PendingCredits,
		TotalEarned:      wallet.TotalEarned,
		TotalSpent:       wallet.TotalSpent,
		LastUpdated:      wallet.LastUpdated,
	}
}

func (s *WalletService) transactionToResponse(transaction *models.Transaction) *TransactionResponse {
	return &TransactionResponse{
		ID:           transaction.ID,
		UserID:       transaction.UserID,
		Type:         transaction.Type,
		Status:       transaction.Status,
		Amount:       transaction.Amount,
		BalanceAfter: transaction.BalanceAfter,
		Source:       transaction.Source,
		Description:  transaction.Description,
		ReferenceID:  transaction.ReferenceID,
		FromUserID:   transaction.FromUserID,
		ToUserID:     transaction.ToUserID,
		ProcessedAt:  transaction.ProcessedAt,
		CreatedAt:    transaction.CreatedAt,
	}
}

// TransferResponse represents a transfer response
type TransferResponse struct {
	TransferID      string               `json:"transfer_id"`
	FromTransaction *TransactionResponse `json:"from_transaction"`
	ToTransaction   *TransactionResponse `json:"to_transaction"`
	FromBalance     *WalletResponse      `json:"from_balance"`
	ToBalance       *WalletResponse      `json:"to_balance"`
}
