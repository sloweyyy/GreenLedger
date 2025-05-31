package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sloweyyy/GreenLedger/services/user-auth/internal/models"
	"github.com/sloweyyy/GreenLedger/shared/database"
	"github.com/sloweyyy/GreenLedger/shared/logger"
	"gorm.io/gorm"
)

// UserRepository handles user data operations
type UserRepository struct {
	db     *database.PostgresDB
	logger *logger.Logger
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *database.PostgresDB, logger *logger.Logger) *UserRepository {
	return &UserRepository{
		db:     db,
		logger: logger,
	}
}

// Create creates a new user
func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	err := r.db.WithContext(ctx).Create(user).Error
	if err != nil {
		r.logger.LogError(ctx, "failed to create user", err,
			logger.String("email", user.Email))
		return fmt.Errorf("failed to create user: %w", err)
	}

	r.logger.LogInfo(ctx, "user created successfully",
		logger.String("user_id", user.ID.String()),
		logger.String("email", user.Email))

	return nil
}

// GetByID retrieves a user by ID
func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	var user models.User
	
	err := r.db.WithContext(ctx).
		Preload("Roles").
		Preload("Roles.Permissions").
		Preload("Profile").
		First(&user, "id = ?", id).Error
	
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, database.ErrNotFound
		}
		r.logger.LogError(ctx, "failed to get user by ID", err,
			logger.String("user_id", id.String()))
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

// GetByEmail retrieves a user by email
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	
	err := r.db.WithContext(ctx).
		Preload("Roles").
		Preload("Roles.Permissions").
		Preload("Profile").
		First(&user, "email = ?", email).Error
	
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, database.ErrNotFound
		}
		r.logger.LogError(ctx, "failed to get user by email", err,
			logger.String("email", email))
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

// GetByUsername retrieves a user by username
func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	var user models.User
	
	err := r.db.WithContext(ctx).
		Preload("Roles").
		Preload("Roles.Permissions").
		Preload("Profile").
		First(&user, "username = ?", username).Error
	
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, database.ErrNotFound
		}
		r.logger.LogError(ctx, "failed to get user by username", err,
			logger.String("username", username))
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

// Update updates a user
func (r *UserRepository) Update(ctx context.Context, user *models.User) error {
	err := r.db.WithContext(ctx).Save(user).Error
	if err != nil {
		r.logger.LogError(ctx, "failed to update user", err,
			logger.String("user_id", user.ID.String()))
		return fmt.Errorf("failed to update user: %w", err)
	}

	r.logger.LogInfo(ctx, "user updated successfully",
		logger.String("user_id", user.ID.String()))

	return nil
}

// Delete deletes a user (soft delete)
func (r *UserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	err := r.db.WithContext(ctx).Delete(&models.User{}, "id = ?", id).Error
	if err != nil {
		r.logger.LogError(ctx, "failed to delete user", err,
			logger.String("user_id", id.String()))
		return fmt.Errorf("failed to delete user: %w", err)
	}

	r.logger.LogInfo(ctx, "user deleted successfully",
		logger.String("user_id", id.String()))

	return nil
}

// UpdateLastLogin updates the user's last login time
func (r *UserRepository) UpdateLastLogin(ctx context.Context, userID uuid.UUID) error {
	now := time.Now().UTC()
	err := r.db.WithContext(ctx).
		Model(&models.User{}).
		Where("id = ?", userID).
		Update("last_login_at", now).Error

	if err != nil {
		r.logger.LogError(ctx, "failed to update last login", err,
			logger.String("user_id", userID.String()))
		return fmt.Errorf("failed to update last login: %w", err)
	}

	return nil
}

// AssignRole assigns a role to a user
func (r *UserRepository) AssignRole(ctx context.Context, userID, roleID uuid.UUID) error {
	return r.db.WithTransaction(ctx, func(tx *gorm.DB) error {
		// Check if user exists
		var user models.User
		if err := tx.First(&user, "id = ?", userID).Error; err != nil {
			return fmt.Errorf("user not found: %w", err)
		}

		// Check if role exists
		var role models.Role
		if err := tx.First(&role, "id = ?", roleID).Error; err != nil {
			return fmt.Errorf("role not found: %w", err)
		}

		// Assign role
		if err := tx.Model(&user).Association("Roles").Append(&role); err != nil {
			return fmt.Errorf("failed to assign role: %w", err)
		}

		r.logger.LogInfo(ctx, "role assigned to user",
			logger.String("user_id", userID.String()),
			logger.String("role_id", roleID.String()))

		return nil
	})
}

// RemoveRole removes a role from a user
func (r *UserRepository) RemoveRole(ctx context.Context, userID, roleID uuid.UUID) error {
	return r.db.WithTransaction(ctx, func(tx *gorm.DB) error {
		// Check if user exists
		var user models.User
		if err := tx.First(&user, "id = ?", userID).Error; err != nil {
			return fmt.Errorf("user not found: %w", err)
		}

		// Check if role exists
		var role models.Role
		if err := tx.First(&role, "id = ?", roleID).Error; err != nil {
			return fmt.Errorf("role not found: %w", err)
		}

		// Remove role
		if err := tx.Model(&user).Association("Roles").Delete(&role); err != nil {
			return fmt.Errorf("failed to remove role: %w", err)
		}

		r.logger.LogInfo(ctx, "role removed from user",
			logger.String("user_id", userID.String()),
			logger.String("role_id", roleID.String()))

		return nil
	})
}

// List retrieves users with pagination
func (r *UserRepository) List(ctx context.Context, limit, offset int) ([]*models.User, int64, error) {
	var users []*models.User
	var total int64

	// Get total count
	if err := r.db.WithContext(ctx).Model(&models.User{}).Count(&total).Error; err != nil {
		r.logger.LogError(ctx, "failed to count users", err)
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}

	// Get users
	err := r.db.WithContext(ctx).
		Preload("Roles").
		Preload("Profile").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&users).Error

	if err != nil {
		r.logger.LogError(ctx, "failed to list users", err)
		return nil, 0, fmt.Errorf("failed to list users: %w", err)
	}

	return users, total, nil
}

// Search searches users by email, username, or name
func (r *UserRepository) Search(ctx context.Context, query string, limit, offset int) ([]*models.User, int64, error) {
	var users []*models.User
	var total int64

	searchPattern := "%" + query + "%"
	
	dbQuery := r.db.WithContext(ctx).
		Where("email ILIKE ? OR username ILIKE ? OR first_name ILIKE ? OR last_name ILIKE ?",
			searchPattern, searchPattern, searchPattern, searchPattern)

	// Get total count
	if err := dbQuery.Model(&models.User{}).Count(&total).Error; err != nil {
		r.logger.LogError(ctx, "failed to count search results", err)
		return nil, 0, fmt.Errorf("failed to count search results: %w", err)
	}

	// Get users
	err := dbQuery.
		Preload("Roles").
		Preload("Profile").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&users).Error

	if err != nil {
		r.logger.LogError(ctx, "failed to search users", err)
		return nil, 0, fmt.Errorf("failed to search users: %w", err)
	}

	return users, total, nil
}

// EmailExists checks if an email already exists
func (r *UserRepository) EmailExists(ctx context.Context, email string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.User{}).
		Where("email = ?", email).
		Count(&count).Error

	if err != nil {
		r.logger.LogError(ctx, "failed to check email existence", err,
			logger.String("email", email))
		return false, fmt.Errorf("failed to check email existence: %w", err)
	}

	return count > 0, nil
}

// UsernameExists checks if a username already exists
func (r *UserRepository) UsernameExists(ctx context.Context, username string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.User{}).
		Where("username = ?", username).
		Count(&count).Error

	if err != nil {
		r.logger.LogError(ctx, "failed to check username existence", err,
			logger.String("username", username))
		return false, fmt.Errorf("failed to check username existence: %w", err)
	}

	return count > 0, nil
}
