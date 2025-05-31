package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/greenledger/services/user-auth/internal/models"
	"github.com/greenledger/shared/database"
	"github.com/greenledger/shared/logger"
	"gorm.io/gorm"
)

// SessionRepository handles session data operations
type SessionRepository struct {
	db     *database.PostgresDB
	logger *logger.Logger
}

// NewSessionRepository creates a new session repository
func NewSessionRepository(db *database.PostgresDB, logger *logger.Logger) *SessionRepository {
	return &SessionRepository{
		db:     db,
		logger: logger,
	}
}

// Create creates a new session
func (r *SessionRepository) Create(ctx context.Context, session *models.Session) error {
	err := r.db.WithContext(ctx).Create(session).Error
	if err != nil {
		r.logger.LogError(ctx, "failed to create session", err,
			logger.String("user_id", session.UserID.String()))
		return fmt.Errorf("failed to create session: %w", err)
	}

	r.logger.LogInfo(ctx, "session created successfully",
		logger.String("session_id", session.ID.String()),
		logger.String("user_id", session.UserID.String()))

	return nil
}

// GetByToken retrieves a session by token
func (r *SessionRepository) GetByToken(ctx context.Context, token string) (*models.Session, error) {
	var session models.Session
	
	err := r.db.WithContext(ctx).
		Preload("User").
		First(&session, "token = ?", token).Error
	
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, database.ErrNotFound
		}
		r.logger.LogError(ctx, "failed to get session by token", err)
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	return &session, nil
}

// GetByUserID retrieves active sessions for a user
func (r *SessionRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*models.Session, error) {
	var sessions []*models.Session
	
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND is_active = true AND expires_at > ?", userID, time.Now()).
		Order("created_at DESC").
		Find(&sessions).Error
	
	if err != nil {
		r.logger.LogError(ctx, "failed to get sessions by user ID", err,
			logger.String("user_id", userID.String()))
		return nil, fmt.Errorf("failed to get sessions: %w", err)
	}

	return sessions, nil
}

// Update updates a session
func (r *SessionRepository) Update(ctx context.Context, session *models.Session) error {
	err := r.db.WithContext(ctx).Save(session).Error
	if err != nil {
		r.logger.LogError(ctx, "failed to update session", err,
			logger.String("session_id", session.ID.String()))
		return fmt.Errorf("failed to update session: %w", err)
	}

	return nil
}

// Delete deletes a session
func (r *SessionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	err := r.db.WithContext(ctx).Delete(&models.Session{}, "id = ?", id).Error
	if err != nil {
		r.logger.LogError(ctx, "failed to delete session", err,
			logger.String("session_id", id.String()))
		return fmt.Errorf("failed to delete session: %w", err)
	}

	return nil
}

// InvalidateUserSessions invalidates all sessions for a user
func (r *SessionRepository) InvalidateUserSessions(ctx context.Context, userID uuid.UUID) error {
	err := r.db.WithContext(ctx).
		Model(&models.Session{}).
		Where("user_id = ?", userID).
		Update("is_active", false).Error

	if err != nil {
		r.logger.LogError(ctx, "failed to invalidate user sessions", err,
			logger.String("user_id", userID.String()))
		return fmt.Errorf("failed to invalidate sessions: %w", err)
	}

	r.logger.LogInfo(ctx, "user sessions invalidated",
		logger.String("user_id", userID.String()))

	return nil
}

// CleanupExpiredSessions removes expired sessions
func (r *SessionRepository) CleanupExpiredSessions(ctx context.Context) error {
	result := r.db.WithContext(ctx).
		Where("expires_at < ?", time.Now()).
		Delete(&models.Session{})

	if result.Error != nil {
		r.logger.LogError(ctx, "failed to cleanup expired sessions", result.Error)
		return fmt.Errorf("failed to cleanup expired sessions: %w", result.Error)
	}

	if result.RowsAffected > 0 {
		r.logger.LogInfo(ctx, "expired sessions cleaned up",
			logger.Int("count", int(result.RowsAffected)))
	}

	return nil
}
