package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/greenledger/services/user-auth/internal/service"
	"github.com/greenledger/shared/logger"
	"github.com/greenledger/shared/middleware"
)

// AuthHandler handles HTTP requests for authentication
type AuthHandler struct {
	authService *service.AuthService
	userService *service.UserService
	logger      *logger.Logger
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(authService *service.AuthService, userService *service.UserService, logger *logger.Logger) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		userService: userService,
		logger:      logger,
	}
}

// RegisterRoutes registers authentication routes
func (h *AuthHandler) RegisterRoutes(router *gin.RouterGroup, authMiddleware *middleware.AuthMiddleware) {
	auth := router.Group("/auth")
	{
		// Public routes
		auth.POST("/register", h.Register)
		auth.POST("/login", h.Login)
		auth.POST("/refresh", h.RefreshToken)
		auth.POST("/logout", h.Logout)

		// Protected routes
		protected := auth.Group("")
		protected.Use(authMiddleware.RequireAuth())
		{
			protected.GET("/profile", h.GetProfile)
			protected.PUT("/profile", h.UpdateProfile)
			protected.POST("/change-password", h.ChangePassword)
			protected.GET("/sessions", h.GetSessions)
			protected.DELETE("/sessions/:id", h.DeleteSession)
		}

		// Admin routes
		admin := auth.Group("/admin")
		admin.Use(authMiddleware.RequireAuth())
		admin.Use(authMiddleware.RequireRole("admin"))
		{
			admin.GET("/users", h.ListUsers)
			admin.GET("/users/:id", h.GetUser)
			admin.PUT("/users/:id", h.UpdateUser)
			admin.DELETE("/users/:id", h.DeleteUser)
			admin.POST("/users/:id/roles", h.AssignRole)
			admin.DELETE("/users/:id/roles/:role_id", h.RemoveRole)
		}
	}
}

// Register godoc
// @Summary Register a new user
// @Description Register a new user account
// @Tags auth
// @Accept json
// @Produce json
// @Param request body service.RegisterRequest true "Registration request"
// @Success 201 {object} service.AuthResponse
// @Failure 400 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req service.RegisterRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.LogError(c.Request.Context(), "invalid request body", err)
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request body",
			Details: err.Error(),
		})
		return
	}

	response, err := h.authService.Register(c.Request.Context(), &req)
	if err != nil {
		h.logger.LogError(c.Request.Context(), "registration failed", err,
			logger.String("email", req.Email))

		statusCode := http.StatusInternalServerError
		if strings.Contains(err.Error(), "already exists") {
			statusCode = http.StatusConflict
		}

		c.JSON(statusCode, ErrorResponse{
			Error:   "Registration failed",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, response)
}

// Login godoc
// @Summary Login user
// @Description Authenticate user and return tokens
// @Tags auth
// @Accept json
// @Produce json
// @Param request body service.LoginRequest true "Login request"
// @Success 200 {object} service.AuthResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req service.LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.LogError(c.Request.Context(), "invalid request body", err)
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request body",
			Details: err.Error(),
		})
		return
	}

	response, err := h.authService.Login(c.Request.Context(), &req)
	if err != nil {
		h.logger.LogError(c.Request.Context(), "login failed", err,
			logger.String("email", req.Email))

		statusCode := http.StatusInternalServerError
		if strings.Contains(err.Error(), "invalid credentials") ||
			strings.Contains(err.Error(), "deactivated") {
			statusCode = http.StatusUnauthorized
		}

		c.JSON(statusCode, ErrorResponse{
			Error:   "Login failed",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// RefreshToken godoc
// @Summary Refresh access token
// @Description Refresh access token using refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RefreshTokenRequest true "Refresh token request"
// @Success 200 {object} service.AuthResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req RefreshTokenRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request body",
			Details: err.Error(),
		})
		return
	}

	response, err := h.authService.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		h.logger.LogError(c.Request.Context(), "token refresh failed", err)
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "Token refresh failed",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// Logout godoc
// @Summary Logout user
// @Description Logout user and invalidate session
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RefreshTokenRequest true "Logout request"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	var req RefreshTokenRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request body",
			Details: err.Error(),
		})
		return
	}

	if err := h.authService.Logout(c.Request.Context(), req.RefreshToken); err != nil {
		h.logger.LogError(c.Request.Context(), "logout failed", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Logout failed",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "Logged out successfully",
	})
}

// GetProfile godoc
// @Summary Get user profile
// @Description Get authenticated user's profile
// @Tags auth
// @Produce json
// @Success 200 {object} service.UserResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BearerAuth
// @Router /auth/profile [get]
func (h *AuthHandler) GetProfile(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "User not authenticated"})
		return
	}

	user, err := h.userService.GetByID(c.Request.Context(), userID)
	if err != nil {
		h.logger.LogError(c.Request.Context(), "failed to get user profile", err,
			logger.String("user_id", userID))
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to get profile",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, user)
}

// UpdateProfile godoc
// @Summary Update user profile
// @Description Update authenticated user's profile
// @Tags auth
// @Accept json
// @Produce json
// @Param request body service.UpdateProfileRequest true "Update profile request"
// @Success 200 {object} service.UserResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BearerAuth
// @Router /auth/profile [put]
func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "User not authenticated"})
		return
	}

	var req service.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request body",
			Details: err.Error(),
		})
		return
	}

	user, err := h.userService.UpdateProfile(c.Request.Context(), userID, &req)
	if err != nil {
		h.logger.LogError(c.Request.Context(), "failed to update profile", err,
			logger.String("user_id", userID))
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to update profile",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, user)
}

// ChangePassword godoc
// @Summary Change user password
// @Description Change authenticated user's password
// @Tags auth
// @Accept json
// @Produce json
// @Param request body ChangePasswordRequest true "Change password request"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BearerAuth
// @Router /auth/change-password [post]
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "User not authenticated"})
		return
	}

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request body",
			Details: err.Error(),
		})
		return
	}

	if err := h.userService.ChangePassword(c.Request.Context(), userID, req.CurrentPassword, req.NewPassword); err != nil {
		h.logger.LogError(c.Request.Context(), "failed to change password", err,
			logger.String("user_id", userID))

		statusCode := http.StatusInternalServerError
		if strings.Contains(err.Error(), "invalid password") {
			statusCode = http.StatusBadRequest
		}

		c.JSON(statusCode, ErrorResponse{
			Error:   "Failed to change password",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "Password changed successfully",
	})
}

// Placeholder implementations for remaining endpoints
func (h *AuthHandler) GetSessions(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Get sessions - to be implemented"})
}

func (h *AuthHandler) DeleteSession(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Delete session - to be implemented"})
}

func (h *AuthHandler) ListUsers(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "List users - to be implemented"})
}

func (h *AuthHandler) GetUser(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Get user - to be implemented"})
}

func (h *AuthHandler) UpdateUser(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Update user - to be implemented"})
}

func (h *AuthHandler) DeleteUser(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Delete user - to be implemented"})
}

func (h *AuthHandler) AssignRole(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Assign role - to be implemented"})
}

func (h *AuthHandler) RemoveRole(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Remove role - to be implemented"})
}

// Request/Response types
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=8"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Details string `json:"details,omitempty"`
}

type SuccessResponse struct {
	Message string `json:"message"`
}
