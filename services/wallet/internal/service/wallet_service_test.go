package service

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/sloweyyy/GreenLedger/services/wallet/internal/models"
)

func TestWalletModel_Creation(t *testing.T) {
	wallet := &models.Wallet{
		ID:               uuid.New(),
		UserID:           "test-user-123",
		AvailableCredits: decimal.NewFromFloat(150.75),
		PendingCredits:   decimal.NewFromFloat(25.0),
		TotalEarned:      decimal.NewFromFloat(200.0),
		TotalSpent:       decimal.NewFromFloat(24.25),
	}

	if wallet.UserID != "test-user-123" {
		t.Errorf("Expected UserID 'test-user-123', got %s", wallet.UserID)
	}

	if !wallet.AvailableCredits.Equal(decimal.NewFromFloat(150.75)) {
		t.Errorf("Expected AvailableCredits 150.75, got %s", wallet.AvailableCredits)
	}

	if !wallet.PendingCredits.Equal(decimal.NewFromFloat(25.0)) {
		t.Errorf("Expected PendingCredits 25.0, got %s", wallet.PendingCredits)
	}

	if !wallet.TotalEarned.Equal(decimal.NewFromFloat(200.0)) {
		t.Errorf("Expected TotalEarned 200.0, got %s", wallet.TotalEarned)
	}
}

func TestWalletModel_CanSpend(t *testing.T) {
	wallet := &models.Wallet{
		ID:               uuid.New(),
		AvailableCredits: decimal.NewFromFloat(100.0),
	}

	// Test sufficient balance
	if !wallet.CanSpend(decimal.NewFromFloat(50.0)) {
		t.Error("Expected wallet to have sufficient balance for 50.0")
	}

	// Test insufficient balance
	if wallet.CanSpend(decimal.NewFromFloat(150.0)) {
		t.Error("Expected wallet to not have sufficient balance for 150.0")
	}

	// Test exact balance
	if !wallet.CanSpend(decimal.NewFromFloat(100.0)) {
		t.Error("Expected wallet to have sufficient balance for exact amount")
	}
}

func TestTransactionModel_Creation(t *testing.T) {
	now := time.Now()
	transaction := &models.Transaction{
		ID:          uuid.New(),
		UserID:      "test-user-123",
		Type:        models.TransactionTypeCreditEarned,
		Amount:      decimal.NewFromFloat(25.50),
		Description: "Carbon credit earned",
		Status:      models.TransactionStatusCompleted,
		ReferenceID: "REF-12345",
		ProcessedAt: &now,
	}

	if transaction.Type != models.TransactionTypeCreditEarned {
		t.Errorf("Expected Type %s, got %s", models.TransactionTypeCreditEarned, transaction.Type)
	}

	if !transaction.Amount.Equal(decimal.NewFromFloat(25.50)) {
		t.Errorf("Expected Amount 25.50, got %s", transaction.Amount)
	}

	if transaction.Description != "Carbon credit earned" {
		t.Errorf("Expected Description 'Carbon credit earned', got %s", transaction.Description)
	}

	if transaction.Status != models.TransactionStatusCompleted {
		t.Errorf("Expected Status %s, got %s", models.TransactionStatusCompleted, transaction.Status)
	}

	if transaction.ReferenceID != "REF-12345" {
		t.Errorf("Expected ReferenceID 'REF-12345', got %s", transaction.ReferenceID)
	}
}

func TestTransactionModel_IsCompleted(t *testing.T) {
	completedTransaction := &models.Transaction{
		ID:     uuid.New(),
		Status: models.TransactionStatusCompleted,
	}

	pendingTransaction := &models.Transaction{
		ID:     uuid.New(),
		Status: models.TransactionStatusPending,
	}

	if !completedTransaction.IsCompleted() {
		t.Error("Expected completed transaction to be completed")
	}

	if pendingTransaction.IsCompleted() {
		t.Error("Expected pending transaction to not be completed")
	}
}

func TestTransactionModel_IsPending(t *testing.T) {
	pendingTransaction := &models.Transaction{
		ID:     uuid.New(),
		Status: models.TransactionStatusPending,
	}

	completedTransaction := &models.Transaction{
		ID:     uuid.New(),
		Status: models.TransactionStatusCompleted,
	}

	if !pendingTransaction.IsPending() {
		t.Error("Expected pending transaction to be pending")
	}

	if completedTransaction.IsPending() {
		t.Error("Expected completed transaction to not be pending")
	}
}

func TestCreditReservationModel_Creation(t *testing.T) {
	reservation := &models.CreditReservation{
		ID:          uuid.New(),
		UserID:      "test-user-123",
		Amount:      decimal.NewFromFloat(50.0),
		Purpose:     "Certificate purchase",
		ReferenceID: "CERT-123",
		ExpiresAt:   time.Now().Add(24 * time.Hour),
		IsReleased:  false,
	}

	if reservation.UserID != "test-user-123" {
		t.Errorf("Expected UserID 'test-user-123', got %s", reservation.UserID)
	}

	if !reservation.Amount.Equal(decimal.NewFromFloat(50.0)) {
		t.Errorf("Expected Amount 50.0, got %s", reservation.Amount)
	}

	if reservation.Purpose != "Certificate purchase" {
		t.Errorf("Expected Purpose 'Certificate purchase', got %s", reservation.Purpose)
	}

	if reservation.IsReleased {
		t.Error("Expected reservation to not be released")
	}
}

func TestCreditReservationModel_IsActive(t *testing.T) {
	activeReservation := &models.CreditReservation{
		ID:         uuid.New(),
		ExpiresAt:  time.Now().Add(1 * time.Hour), // 1 hour from now
		IsReleased: false,
	}

	releasedReservation := &models.CreditReservation{
		ID:         uuid.New(),
		ExpiresAt:  time.Now().Add(1 * time.Hour),
		IsReleased: true,
	}

	if !activeReservation.IsActive() {
		t.Error("Expected active reservation to be active")
	}

	if releasedReservation.IsActive() {
		t.Error("Expected released reservation to not be active")
	}
}
