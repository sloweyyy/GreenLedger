package service

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sloweyyy/GreenLedger/services/user-auth/internal/models"
)

func TestUserModel_Creation(t *testing.T) {
	user := &models.User{
		ID:         uuid.New(),
		Email:      "test@example.com",
		Username:   "testuser",
		FirstName:  "Test",
		LastName:   "User",
		IsActive:   true,
		IsVerified: false,
	}

	if user.Email != "test@example.com" {
		t.Errorf("Expected Email 'test@example.com', got %s", user.Email)
	}

	if user.Username != "testuser" {
		t.Errorf("Expected Username 'testuser', got %s", user.Username)
	}

	if user.FirstName != "Test" {
		t.Errorf("Expected FirstName 'Test', got %s", user.FirstName)
	}

	if user.LastName != "User" {
		t.Errorf("Expected LastName 'User', got %s", user.LastName)
	}

	if !user.IsActive {
		t.Error("Expected user to be active")
	}

	if user.IsVerified {
		t.Error("Expected user to not be verified initially")
	}
}

func TestUserModel_HasRole(t *testing.T) {
	// Create roles
	adminRole := models.Role{
		ID:   uuid.New(),
		Name: models.RoleAdmin,
	}

	userRole := models.Role{
		ID:   uuid.New(),
		Name: models.RoleUser,
	}

	adminUser := &models.User{
		ID:    uuid.New(),
		Roles: []models.Role{adminRole},
	}

	regularUser := &models.User{
		ID:    uuid.New(),
		Roles: []models.Role{userRole},
	}

	if !adminUser.HasRole(models.RoleAdmin) {
		t.Error("Expected admin user to have admin role")
	}

	if regularUser.HasRole(models.RoleAdmin) {
		t.Error("Expected regular user to not have admin role")
	}

	if !regularUser.HasRole(models.RoleUser) {
		t.Error("Expected regular user to have user role")
	}
}

func TestUserModel_IsVerified(t *testing.T) {
	verifiedUser := &models.User{
		ID:         uuid.New(),
		IsVerified: true,
	}

	unverifiedUser := &models.User{
		ID:         uuid.New(),
		IsVerified: false,
	}

	if !verifiedUser.IsVerified {
		t.Error("Expected verified user to be verified")
	}

	if unverifiedUser.IsVerified {
		t.Error("Expected unverified user to not be verified")
	}
}

func TestSessionModel_Creation(t *testing.T) {
	session := &models.Session{
		ID:        uuid.New(),
		UserID:    uuid.New(),
		Token:     "session-token-123",
		ExpiresAt: time.Now().Add(24 * time.Hour),
		IsActive:  true,
	}

	if session.Token != "session-token-123" {
		t.Errorf("Expected Token 'session-token-123', got %s", session.Token)
	}

	if !session.IsActive {
		t.Error("Expected session to be active")
	}

	if session.ExpiresAt.Before(time.Now()) {
		t.Error("Expected session to not be expired")
	}
}

func TestSessionModel_IsSessionValid(t *testing.T) {
	expiredSession := &models.Session{
		ID:        uuid.New(),
		ExpiresAt: time.Now().Add(-1 * time.Hour), // 1 hour ago
		IsActive:  true,
	}

	validSession := &models.Session{
		ID:        uuid.New(),
		ExpiresAt: time.Now().Add(1 * time.Hour), // 1 hour from now
		IsActive:  true,
	}

	if expiredSession.IsSessionValid() {
		t.Error("Expected expired session to be invalid")
	}

	if !validSession.IsSessionValid() {
		t.Error("Expected valid session to be valid")
	}
}

func TestPasswordResetTokenModel_Creation(t *testing.T) {
	passwordReset := &models.PasswordResetToken{
		ID:        uuid.New(),
		UserID:    uuid.New(),
		Token:     "reset-token-789",
		ExpiresAt: time.Now().Add(1 * time.Hour), // 1 hour
		IsUsed:    false,
	}

	if passwordReset.Token != "reset-token-789" {
		t.Errorf("Expected Token 'reset-token-789', got %s", passwordReset.Token)
	}

	if passwordReset.IsUsed {
		t.Error("Expected password reset to not be used")
	}

	if passwordReset.ExpiresAt.Before(time.Now()) {
		t.Error("Expected password reset to not be expired")
	}
}

func TestEmailVerificationTokenModel_Creation(t *testing.T) {
	emailToken := &models.EmailVerificationToken{
		ID:        uuid.New(),
		UserID:    uuid.New(),
		Token:     "email-token-123",
		ExpiresAt: time.Now().Add(24 * time.Hour), // 24 hours
		IsUsed:    false,
	}

	if emailToken.Token != "email-token-123" {
		t.Errorf("Expected Token 'email-token-123', got %s", emailToken.Token)
	}

	if emailToken.IsUsed {
		t.Error("Expected email verification token to not be used")
	}

	if emailToken.ExpiresAt.Before(time.Now()) {
		t.Error("Expected email verification token to not be expired")
	}
}
