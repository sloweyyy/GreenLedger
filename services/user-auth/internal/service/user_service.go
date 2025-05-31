package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/sloweyyy/GreenLedger/services/user-auth/internal/models"
	"github.com/sloweyyy/GreenLedger/services/user-auth/internal/repository"
	"github.com/sloweyyy/GreenLedger/shared/logger"
)

// UserService handles user management operations
type UserService struct {
	userRepo *repository.UserRepository
	logger   *logger.Logger
}

// NewUserService creates a new user service
func NewUserService(userRepo *repository.UserRepository, logger *logger.Logger) *UserService {
	return &UserService{
		userRepo: userRepo,
		logger:   logger,
	}
}

// UpdateProfileRequest represents a profile update request
type UpdateProfileRequest struct {
	FirstName   string `json:"first_name" binding:"required,min=1,max=100"`
	LastName    string `json:"last_name" binding:"required,min=1,max=100"`
	Bio         string `json:"bio"`
	Location    string `json:"location"`
	Website     string `json:"website"`
	PhoneNumber string `json:"phone_number"`
}

// GetByID retrieves a user by ID
func (s *UserService) GetByID(ctx context.Context, userID string) (*UserResponse, error) {
	id, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return s.userToResponse(user), nil
}

// UpdateProfile updates a user's profile
func (s *UserService) UpdateProfile(ctx context.Context, userID string, req *UpdateProfileRequest) (*UserResponse, error) {
	id, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Update user fields
	user.FirstName = req.FirstName
	user.LastName = req.LastName

	// Update or create profile
	if user.Profile == nil {
		user.Profile = &models.UserProfile{
			UserID: user.ID,
		}
	}

	user.Profile.Bio = req.Bio
	user.Profile.Location = req.Location
	user.Profile.Website = req.Website
	user.Profile.PhoneNumber = req.PhoneNumber

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	s.logger.LogInfo(ctx, "user profile updated",
		logger.String("user_id", userID))

	return s.userToResponse(user), nil
}

// ChangePassword changes a user's password
func (s *UserService) ChangePassword(ctx context.Context, userID, currentPassword, newPassword string) error {
	id, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("invalid user ID: %w", err)
	}

	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	// Verify current password
	if !user.CheckPassword(currentPassword) {
		return fmt.Errorf("invalid current password")
	}

	// Hash new password
	if err := user.HashPassword(newPassword); err != nil {
		return fmt.Errorf("failed to hash new password: %w", err)
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	s.logger.LogInfo(ctx, "user password changed",
		logger.String("user_id", userID))

	return nil
}

// List retrieves users with pagination
func (s *UserService) List(ctx context.Context, limit, offset int) ([]*UserResponse, int64, error) {
	users, total, err := s.userRepo.List(ctx, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list users: %w", err)
	}

	responses := make([]*UserResponse, len(users))
	for i, user := range users {
		responses[i] = s.userToResponse(user)
	}

	return responses, total, nil
}

// Search searches users
func (s *UserService) Search(ctx context.Context, query string, limit, offset int) ([]*UserResponse, int64, error) {
	users, total, err := s.userRepo.Search(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to search users: %w", err)
	}

	responses := make([]*UserResponse, len(users))
	for i, user := range users {
		responses[i] = s.userToResponse(user)
	}

	return responses, total, nil
}

// UpdateUser updates a user (admin operation)
func (s *UserService) UpdateUser(ctx context.Context, userID string, req *UpdateUserRequest) (*UserResponse, error) {
	id, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Update fields if provided
	if req.FirstName != "" {
		user.FirstName = req.FirstName
	}
	if req.LastName != "" {
		user.LastName = req.LastName
	}
	if req.Email != "" {
		// Check if email already exists
		emailExists, err := s.userRepo.EmailExists(ctx, req.Email)
		if err != nil {
			return nil, fmt.Errorf("failed to check email existence: %w", err)
		}
		if emailExists && req.Email != user.Email {
			return nil, fmt.Errorf("email already exists")
		}
		user.Email = req.Email
	}
	if req.IsActive != nil {
		user.IsActive = *req.IsActive
	}
	if req.IsVerified != nil {
		user.IsVerified = *req.IsVerified
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	s.logger.LogInfo(ctx, "user updated by admin",
		logger.String("user_id", userID))

	return s.userToResponse(user), nil
}

// DeleteUser deletes a user (admin operation)
func (s *UserService) DeleteUser(ctx context.Context, userID string) error {
	id, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("invalid user ID: %w", err)
	}

	if err := s.userRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	s.logger.LogInfo(ctx, "user deleted by admin",
		logger.String("user_id", userID))

	return nil
}

// AssignRole assigns a role to a user (admin operation)
func (s *UserService) AssignRole(ctx context.Context, userID, roleID string) error {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("invalid user ID: %w", err)
	}

	rid, err := uuid.Parse(roleID)
	if err != nil {
		return fmt.Errorf("invalid role ID: %w", err)
	}

	if err := s.userRepo.AssignRole(ctx, uid, rid); err != nil {
		return fmt.Errorf("failed to assign role: %w", err)
	}

	s.logger.LogInfo(ctx, "role assigned to user",
		logger.String("user_id", userID),
		logger.String("role_id", roleID))

	return nil
}

// RemoveRole removes a role from a user (admin operation)
func (s *UserService) RemoveRole(ctx context.Context, userID, roleID string) error {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("invalid user ID: %w", err)
	}

	rid, err := uuid.Parse(roleID)
	if err != nil {
		return fmt.Errorf("invalid role ID: %w", err)
	}

	if err := s.userRepo.RemoveRole(ctx, uid, rid); err != nil {
		return fmt.Errorf("failed to remove role: %w", err)
	}

	s.logger.LogInfo(ctx, "role removed from user",
		logger.String("user_id", userID),
		logger.String("role_id", roleID))

	return nil
}

// userToResponse converts a user model to response format
func (s *UserService) userToResponse(user *models.User) *UserResponse {
	response := &UserResponse{
		ID:         user.ID,
		Email:      user.Email,
		Username:   user.Username,
		FirstName:  user.FirstName,
		LastName:   user.LastName,
		IsActive:   user.IsActive,
		IsVerified: user.IsVerified,
		Roles:      user.GetRoleNames(),
		CreatedAt:  user.CreatedAt,
	}

	// Add profile information if available
	if user.Profile != nil {
		response.Profile = &UserProfileResponse{
			Bio:         user.Profile.Bio,
			Location:    user.Profile.Location,
			Website:     user.Profile.Website,
			PhoneNumber: user.Profile.PhoneNumber,
		}
	}

	return response
}

// Additional request/response types
type UpdateUserRequest struct {
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	Email      string `json:"email"`
	IsActive   *bool  `json:"is_active"`
	IsVerified *bool  `json:"is_verified"`
}

type UserProfileResponse struct {
	Bio         string `json:"bio"`
	Location    string `json:"location"`
	Website     string `json:"website"`
	PhoneNumber string `json:"phone_number"`
}

// Update UserResponse to include profile
type UserResponseWithProfile struct {
	UserResponse
	Profile *UserProfileResponse `json:"profile,omitempty"`
}
