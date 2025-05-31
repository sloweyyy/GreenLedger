package service

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sloweyyy/GreenLedger/services/user-auth/internal/models"
)

func TestUserModel_Creation(t *testing.T) {
	user := &models.User{
		ID:       uuid.New(),
		Email:    "test@example.com",
		Username: "testuser",
		IsActive: true,
		Role:     models.UserRoleUser,
	}

	if user.Email != "test@example.com" {
		t.Errorf("Expected Email 'test@example.com', got %s", user.Email)
	}

	if user.Username != "testuser" {
		t.Errorf("Expected Username 'testuser', got %s", user.Username)
	}

	if !user.IsActive {
		t.Error("Expected user to be active")
	}

	if user.Role != models.UserRoleUser {
		t.Errorf("Expected Role %s, got %s", models.UserRoleUser, user.Role)
	}
}

func TestUserModel_IsAdmin(t *testing.T) {
	adminUser := &models.User{
		ID:   uuid.New(),
		Role: models.UserRoleAdmin,
	}

	regularUser := &models.User{
		ID:   uuid.New(),
		Role: models.UserRoleUser,
	}

	if !adminUser.IsAdmin() {
		t.Error("Expected admin user to be admin")
	}

	if regularUser.IsAdmin() {
		t.Error("Expected regular user to not be admin")
	}
}

func TestUserModel_IsVerified(t *testing.T) {
	verifiedUser := &models.User{
		ID:         uuid.New(),
		VerifiedAt: &time.Time{},
	}
	*verifiedUser.VerifiedAt = time.Now()

	unverifiedUser := &models.User{
		ID:         uuid.New(),
		VerifiedAt: nil,
	}

	if !verifiedUser.IsVerified() {
		t.Error("Expected verified user to be verified")
	}

	if unverifiedUser.IsVerified() {
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

func TestSessionModel_IsExpired(t *testing.T) {
	expiredSession := &models.Session{
		ID:        uuid.New(),
		ExpiresAt: time.Now().Add(-1 * time.Hour), // 1 hour ago
	}

	validSession := &models.Session{
		ID:        uuid.New(),
		ExpiresAt: time.Now().Add(1 * time.Hour), // 1 hour from now
	}

	if !expiredSession.IsExpired() {
		t.Error("Expected expired session to be expired")
	}

	if validSession.IsExpired() {
		t.Error("Expected valid session to not be expired")
	}
}

func TestSessionModel_IsValid(t *testing.T) {
	validSession := &models.Session{
		ID:        uuid.New(),
		IsActive:  true,
		ExpiresAt: time.Now().Add(1 * time.Hour), // 1 hour from now
	}

	inactiveSession := &models.Session{
		ID:        uuid.New(),
		IsActive:  false,
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}

	expiredSession := &models.Session{
		ID:        uuid.New(),
		IsActive:  true,
		ExpiresAt: time.Now().Add(-1 * time.Hour), // 1 hour ago
	}

	if !validSession.IsValid() {
		t.Error("Expected valid session to be valid")
	}

	if inactiveSession.IsValid() {
		t.Error("Expected inactive session to not be valid")
	}

	if expiredSession.IsValid() {
		t.Error("Expected expired session to not be valid")
	}
}

func TestRefreshTokenModel_Creation(t *testing.T) {
	refreshToken := &models.RefreshToken{
		ID:        uuid.New(),
		UserID:    uuid.New(),
		Token:     "refresh-token-456",
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour), // 7 days
		IsActive:  true,
	}

	if refreshToken.Token != "refresh-token-456" {
		t.Errorf("Expected Token 'refresh-token-456', got %s", refreshToken.Token)
	}

	if !refreshToken.IsActive {
		t.Error("Expected refresh token to be active")
	}

	if refreshToken.ExpiresAt.Before(time.Now()) {
		t.Error("Expected refresh token to not be expired")
	}
}

func TestRefreshTokenModel_IsExpired(t *testing.T) {
	expiredToken := &models.RefreshToken{
		ID:        uuid.New(),
		ExpiresAt: time.Now().Add(-1 * time.Hour), // 1 hour ago
	}

	validToken := &models.RefreshToken{
		ID:        uuid.New(),
		ExpiresAt: time.Now().Add(1 * time.Hour), // 1 hour from now
	}

	if !expiredToken.IsExpired() {
		t.Error("Expected expired token to be expired")
	}

	if validToken.IsExpired() {
		t.Error("Expected valid token to not be expired")
	}
}

func TestRefreshTokenModel_IsValid(t *testing.T) {
	validToken := &models.RefreshToken{
		ID:        uuid.New(),
		IsActive:  true,
		ExpiresAt: time.Now().Add(1 * time.Hour), // 1 hour from now
	}

	inactiveToken := &models.RefreshToken{
		ID:        uuid.New(),
		IsActive:  false,
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}

	expiredToken := &models.RefreshToken{
		ID:        uuid.New(),
		IsActive:  true,
		ExpiresAt: time.Now().Add(-1 * time.Hour), // 1 hour ago
	}

	if !validToken.IsValid() {
		t.Error("Expected valid token to be valid")
	}

	if inactiveToken.IsValid() {
		t.Error("Expected inactive token to not be valid")
	}

	if expiredToken.IsValid() {
		t.Error("Expected expired token to not be valid")
	}
}

func TestPasswordResetModel_Creation(t *testing.T) {
	passwordReset := &models.PasswordReset{
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

func TestPasswordResetModel_IsExpired(t *testing.T) {
	expiredReset := &models.PasswordReset{
		ID:        uuid.New(),
		ExpiresAt: time.Now().Add(-1 * time.Hour), // 1 hour ago
	}

	validReset := &models.PasswordReset{
		ID:        uuid.New(),
		ExpiresAt: time.Now().Add(1 * time.Hour), // 1 hour from now
	}

	if !expiredReset.IsExpired() {
		t.Error("Expected expired reset to be expired")
	}

	if validReset.IsExpired() {
		t.Error("Expected valid reset to not be expired")
	}
}

func TestPasswordResetModel_IsValid(t *testing.T) {
	validReset := &models.PasswordReset{
		ID:        uuid.New(),
		IsUsed:    false,
		ExpiresAt: time.Now().Add(1 * time.Hour), // 1 hour from now
	}

	usedReset := &models.PasswordReset{
		ID:        uuid.New(),
		IsUsed:    true,
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}

	expiredReset := &models.PasswordReset{
		ID:        uuid.New(),
		IsUsed:    false,
		ExpiresAt: time.Now().Add(-1 * time.Hour), // 1 hour ago
	}

	if !validReset.IsValid() {
		t.Error("Expected valid reset to be valid")
	}

	if usedReset.IsValid() {
		t.Error("Expected used reset to not be valid")
	}

	if expiredReset.IsValid() {
		t.Error("Expected expired reset to not be valid")
	}
}
