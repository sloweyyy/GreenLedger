package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/sloweyyy/GreenLedger/services/user-auth/internal/models"
	"github.com/sloweyyy/GreenLedger/services/user-auth/internal/repository"
	"github.com/sloweyyy/GreenLedger/shared/logger"
)

// AuthService handles authentication operations
type AuthService struct {
	userRepo    *repository.UserRepository
	sessionRepo *repository.SessionRepository
	roleRepo    *repository.RoleRepository
	jwtSecret   []byte
	logger      *logger.Logger
}

// NewAuthService creates a new auth service
func NewAuthService(
	userRepo *repository.UserRepository,
	sessionRepo *repository.SessionRepository,
	roleRepo *repository.RoleRepository,
	jwtSecret string,
	logger *logger.Logger,
) *AuthService {
	return &AuthService{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		roleRepo:    roleRepo,
		jwtSecret:   []byte(jwtSecret),
		logger:      logger,
	}
}

// RegisterRequest represents a user registration request
type RegisterRequest struct {
	Email     string `json:"email" binding:"required,email"`
	Username  string `json:"username" binding:"required,min=3,max=50"`
	FirstName string `json:"first_name" binding:"required,min=1,max=100"`
	LastName  string `json:"last_name" binding:"required,min=1,max=100"`
	Password  string `json:"password" binding:"required,min=8"`
}

// LoginRequest represents a user login request
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// AuthResponse represents an authentication response
type AuthResponse struct {
	User         *UserResponse `json:"user"`
	AccessToken  string        `json:"access_token"`
	RefreshToken string        `json:"refresh_token"`
	ExpiresAt    time.Time     `json:"expires_at"`
}

// UserResponse represents a user in API responses
type UserResponse struct {
	ID         uuid.UUID            `json:"id"`
	Email      string               `json:"email"`
	Username   string               `json:"username"`
	FirstName  string               `json:"first_name"`
	LastName   string               `json:"last_name"`
	IsActive   bool                 `json:"is_active"`
	IsVerified bool                 `json:"is_verified"`
	Roles      []string             `json:"roles"`
	CreatedAt  time.Time            `json:"created_at"`
	Profile    *UserProfileResponse `json:"profile,omitempty"`
}

// UserProfileResponse represents user profile information in API responses
type UserProfileResponse struct {
	Bio         string `json:"bio"`
	Location    string `json:"location"`
	Website     string `json:"website"`
	PhoneNumber string `json:"phone_number"`
}

// JWTClaims represents JWT token claims
type JWTClaims struct {
	UserID string   `json:"user_id"`
	Email  string   `json:"email"`
	Roles  []string `json:"roles"`
	jwt.RegisteredClaims
}

// Register registers a new user
func (s *AuthService) Register(ctx context.Context, req *RegisterRequest) (*AuthResponse, error) {
	s.logger.LogInfo(ctx, "starting user registration",
		logger.String("email", req.Email),
		logger.String("username", req.Username))

	// Check if email already exists
	emailExists, err := s.userRepo.EmailExists(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to check email existence: %w", err)
	}
	if emailExists {
		return nil, fmt.Errorf("email already exists")
	}

	// Check if username already exists
	usernameExists, err := s.userRepo.UsernameExists(ctx, req.Username)
	if err != nil {
		return nil, fmt.Errorf("failed to check username existence: %w", err)
	}
	if usernameExists {
		return nil, fmt.Errorf("username already exists")
	}

	// Create user
	user := &models.User{
		Email:     req.Email,
		Username:  req.Username,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		IsActive:  true,
	}

	// Hash password
	if err := user.HashPassword(req.Password); err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Save user
	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Assign default user role
	defaultRole, err := s.roleRepo.GetByName(ctx, models.RoleUser)
	if err != nil {
		s.logger.LogError(ctx, "failed to get default role", err)
		// Continue without role assignment for now
	} else {
		if err := s.userRepo.AssignRole(ctx, user.ID, defaultRole.ID); err != nil {
			s.logger.LogError(ctx, "failed to assign default role", err)
		}
	}

	// Reload user with roles
	user, err = s.userRepo.GetByID(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to reload user: %w", err)
	}

	// Generate tokens
	accessToken, refreshToken, expiresAt, err := s.generateTokens(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	// Create session
	session := &models.Session{
		UserID:    user.ID,
		Token:     refreshToken,
		ExpiresAt: expiresAt.Add(24 * time.Hour), // Refresh token expires in 24 hours
		IsActive:  true,
	}

	if err := s.sessionRepo.Create(ctx, session); err != nil {
		s.logger.LogError(ctx, "failed to create session", err)
		// Continue without session for now
	}

	s.logger.LogInfo(ctx, "user registered successfully",
		logger.String("user_id", user.ID.String()),
		logger.String("email", user.Email))

	return &AuthResponse{
		User:         s.userToResponse(user),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
	}, nil
}

// Login authenticates a user
func (s *AuthService) Login(ctx context.Context, req *LoginRequest) (*AuthResponse, error) {
	s.logger.LogInfo(ctx, "starting user login",
		logger.String("email", req.Email))

	// Get user by email
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		if err == fmt.Errorf("record not found") {
			return nil, fmt.Errorf("invalid credentials")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Check if user is active
	if !user.IsActive {
		return nil, fmt.Errorf("user account is deactivated")
	}

	// Check password
	if !user.CheckPassword(req.Password) {
		s.logger.LogWarn(ctx, "invalid password attempt",
			logger.String("user_id", user.ID.String()),
			logger.String("email", user.Email))
		return nil, fmt.Errorf("invalid credentials")
	}

	// Update last login
	if err := s.userRepo.UpdateLastLogin(ctx, user.ID); err != nil {
		s.logger.LogError(ctx, "failed to update last login", err)
		// Continue without updating last login
	}

	// Generate tokens
	accessToken, refreshToken, expiresAt, err := s.generateTokens(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	// Create session
	session := &models.Session{
		UserID:    user.ID,
		Token:     refreshToken,
		ExpiresAt: expiresAt.Add(24 * time.Hour), // Refresh token expires in 24 hours
		IsActive:  true,
	}

	if err := s.sessionRepo.Create(ctx, session); err != nil {
		s.logger.LogError(ctx, "failed to create session", err)
		// Continue without session for now
	}

	s.logger.LogInfo(ctx, "user logged in successfully",
		logger.String("user_id", user.ID.String()),
		logger.String("email", user.Email))

	return &AuthResponse{
		User:         s.userToResponse(user),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
	}, nil
}

// RefreshToken refreshes an access token using a refresh token
func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (*AuthResponse, error) {
	// Get session by token
	session, err := s.sessionRepo.GetByToken(ctx, refreshToken)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token")
	}

	// Check if session is valid
	if !session.IsSessionValid() {
		return nil, fmt.Errorf("refresh token expired")
	}

	// Get user
	user, err := s.userRepo.GetByID(ctx, session.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Check if user is active
	if !user.IsActive {
		return nil, fmt.Errorf("user account is deactivated")
	}

	// Generate new tokens
	accessToken, newRefreshToken, expiresAt, err := s.generateTokens(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	// Update session with new refresh token
	session.Token = newRefreshToken
	session.ExpiresAt = expiresAt.Add(24 * time.Hour)
	if err := s.sessionRepo.Update(ctx, session); err != nil {
		s.logger.LogError(ctx, "failed to update session", err)
	}

	return &AuthResponse{
		User:         s.userToResponse(user),
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		ExpiresAt:    expiresAt,
	}, nil
}

// Logout logs out a user by invalidating their session
func (s *AuthService) Logout(ctx context.Context, refreshToken string) error {
	session, err := s.sessionRepo.GetByToken(ctx, refreshToken)
	if err != nil {
		return nil // Already logged out
	}

	session.IsActive = false
	if err := s.sessionRepo.Update(ctx, session); err != nil {
		s.logger.LogError(ctx, "failed to invalidate session", err)
		return fmt.Errorf("failed to logout: %w", err)
	}

	s.logger.LogInfo(ctx, "user logged out successfully",
		logger.String("user_id", session.UserID.String()))

	return nil
}

// ValidateToken validates a JWT token and returns the user
func (s *AuthService) ValidateToken(ctx context.Context, tokenString string) (*models.User, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return s.jwtSecret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID in token")
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	if !user.IsActive {
		return nil, fmt.Errorf("user account is deactivated")
	}

	return user, nil
}

// generateTokens generates access and refresh tokens
func (s *AuthService) generateTokens(user *models.User) (string, string, time.Time, error) {
	expiresAt := time.Now().Add(1 * time.Hour) // Access token expires in 1 hour

	// Create access token claims
	claims := &JWTClaims{
		UserID: user.ID.String(),
		Email:  user.Email,
		Roles:  user.GetRoleNames(),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "greenledger-auth",
			Subject:   user.ID.String(),
		},
	}

	// Generate access token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	accessToken, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return "", "", time.Time{}, fmt.Errorf("failed to sign access token: %w", err)
	}

	// Generate refresh token (random string)
	refreshToken, err := s.generateRandomToken()
	if err != nil {
		return "", "", time.Time{}, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return accessToken, refreshToken, expiresAt, nil
}

// generateRandomToken generates a random token
func (s *AuthService) generateRandomToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// userToResponse converts a user model to response format
func (s *AuthService) userToResponse(user *models.User) *UserResponse {
	return &UserResponse{
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
}
