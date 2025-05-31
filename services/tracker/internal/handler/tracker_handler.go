package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sloweyyy/GreenLedger/services/tracker/internal/service"
	"github.com/sloweyyy/GreenLedger/shared/logger"
	"github.com/sloweyyy/GreenLedger/shared/middleware"
)

// TrackerHandler handles HTTP requests for activity tracking
type TrackerHandler struct {
	trackerService *service.TrackerService
	logger         *logger.Logger
}

// NewTrackerHandler creates a new tracker handler
func NewTrackerHandler(trackerService *service.TrackerService, logger *logger.Logger) *TrackerHandler {
	return &TrackerHandler{
		trackerService: trackerService,
		logger:         logger,
	}
}

// RegisterRoutes registers tracker routes
func (h *TrackerHandler) RegisterRoutes(router *gin.RouterGroup, authMiddleware *middleware.AuthMiddleware) {
	tracker := router.Group("/tracker")
	{
		// Public routes (for webhooks and IoT devices)
		tracker.POST("/webhook", h.HandleWebhook)
		tracker.POST("/iot", h.HandleIoTData)

		// Protected routes
		tracker.Use(authMiddleware.RequireAuth())
		tracker.POST("/activities", h.LogActivity)
		tracker.GET("/activities", h.GetUserActivities)
		tracker.GET("/activities/:id", h.GetActivityByID)
		tracker.GET("/stats", h.GetUserStats)
		tracker.GET("/activity-types", h.GetActivityTypes)
		tracker.GET("/activity-types/:category", h.GetActivityTypesByCategory)

		// Admin/Moderator routes
		admin := tracker.Group("/admin")
		admin.Use(authMiddleware.RequireRole("admin"))
		{
			admin.GET("/activities/unverified", h.GetUnverifiedActivities)
			admin.PUT("/activities/:id/verify", h.VerifyActivity)
			admin.GET("/activities/recent", h.GetRecentActivities)
		}
	}
}

// LogActivity godoc
// @Summary Log eco-friendly activity
// @Description Log a new eco-friendly activity for the authenticated user
// @Tags tracker
// @Accept json
// @Produce json
// @Param request body service.LogActivityRequest true "Activity request"
// @Success 201 {object} service.ActivityResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BearerAuth
// @Router /tracker/activities [post]
func (h *TrackerHandler) LogActivity(c *gin.Context) {
	var req service.LogActivityRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.LogError(c.Request.Context(), "invalid request body", err)
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request body",
			Details: err.Error(),
		})
		return
	}

	// Get user ID from context
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "User not authenticated"})
		return
	}
	req.UserID = userID

	// Set source if not provided
	if req.Source == "" {
		req.Source = "manual"
	}

	response, err := h.trackerService.LogActivity(c.Request.Context(), &req)
	if err != nil {
		h.logger.LogError(c.Request.Context(), "failed to log activity", err,
			logger.String("user_id", userID))
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to log activity",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, response)
}

// GetUserActivities godoc
// @Summary Get user activities
// @Description Get activities for the authenticated user
// @Tags tracker
// @Produce json
// @Param limit query int false "Limit" default(20)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} ActivityListResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BearerAuth
// @Router /tracker/activities [get]
func (h *TrackerHandler) GetUserActivities(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "User not authenticated"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	activities, total, err := h.trackerService.GetUserActivities(c.Request.Context(), userID, limit, offset)
	if err != nil {
		h.logger.LogError(c.Request.Context(), "failed to get user activities", err,
			logger.String("user_id", userID))
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to get activities",
			Details: err.Error(),
		})
		return
	}

	response := ActivityListResponse{
		Activities: activities,
		Total:      total,
		Limit:      limit,
		Offset:     offset,
	}

	c.JSON(http.StatusOK, response)
}

// GetActivityByID godoc
// @Summary Get activity by ID
// @Description Get a specific activity by ID
// @Tags tracker
// @Produce json
// @Param id path string true "Activity ID"
// @Success 200 {object} service.ActivityResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BearerAuth
// @Router /tracker/activities/{id} [get]
func (h *TrackerHandler) GetActivityByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid activity ID",
			Details: err.Error(),
		})
		return
	}

	activity, err := h.trackerService.GetActivityByID(c.Request.Context(), id)
	if err != nil {
		h.logger.LogError(c.Request.Context(), "failed to get activity", err,
			logger.String("activity_id", id.String()))
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to get activity",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, activity)
}

// GetUserStats godoc
// @Summary Get user activity statistics
// @Description Get activity statistics for the authenticated user
// @Tags tracker
// @Produce json
// @Param start_date query string false "Start date (RFC3339 format)"
// @Param end_date query string false "End date (RFC3339 format)"
// @Success 200 {object} models.UserActivityStats
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BearerAuth
// @Router /tracker/stats [get]
func (h *TrackerHandler) GetUserStats(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "User not authenticated"})
		return
	}

	// Parse date range (default to last 30 days)
	endDate := time.Now().UTC()
	startDate := endDate.AddDate(0, 0, -30)

	if startDateStr := c.Query("start_date"); startDateStr != "" {
		if parsed, err := time.Parse(time.RFC3339, startDateStr); err == nil {
			startDate = parsed
		}
	}
	if endDateStr := c.Query("end_date"); endDateStr != "" {
		if parsed, err := time.Parse(time.RFC3339, endDateStr); err == nil {
			endDate = parsed
		}
	}

	stats, err := h.trackerService.GetUserStats(c.Request.Context(), userID, startDate, endDate)
	if err != nil {
		h.logger.LogError(c.Request.Context(), "failed to get user stats", err,
			logger.String("user_id", userID))
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to get stats",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetActivityTypes godoc
// @Summary Get activity types
// @Description Get all available activity types
// @Tags tracker
// @Produce json
// @Success 200 {object} ActivityTypesResponse
// @Failure 500 {object} ErrorResponse
// @Security BearerAuth
// @Router /tracker/activity-types [get]
func (h *TrackerHandler) GetActivityTypes(c *gin.Context) {
	// This would be implemented with ActivityTypeService
	c.JSON(http.StatusOK, gin.H{
		"message": "Get activity types - to be implemented",
	})
}

// GetActivityTypesByCategory godoc
// @Summary Get activity types by category
// @Description Get activity types for a specific category
// @Tags tracker
// @Produce json
// @Param category path string true "Activity category"
// @Success 200 {object} ActivityTypesResponse
// @Failure 500 {object} ErrorResponse
// @Security BearerAuth
// @Router /tracker/activity-types/{category} [get]
func (h *TrackerHandler) GetActivityTypesByCategory(c *gin.Context) {
	category := c.Param("category")

	// This would be implemented with ActivityTypeService
	c.JSON(http.StatusOK, gin.H{
		"message":  "Get activity types by category - to be implemented",
		"category": category,
	})
}

// VerifyActivity godoc
// @Summary Verify activity
// @Description Verify an activity (admin only)
// @Tags tracker
// @Produce json
// @Param id path string true "Activity ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BearerAuth
// @Router /tracker/admin/activities/{id}/verify [put]
func (h *TrackerHandler) VerifyActivity(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid activity ID",
			Details: err.Error(),
		})
		return
	}

	verifiedBy, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "User not authenticated"})
		return
	}

	if err := h.trackerService.VerifyActivity(c.Request.Context(), id, verifiedBy); err != nil {
		h.logger.LogError(c.Request.Context(), "failed to verify activity", err,
			logger.String("activity_id", id.String()))
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to verify activity",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "Activity verified successfully",
	})
}

// Placeholder implementations for remaining endpoints
func (h *TrackerHandler) HandleWebhook(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Webhook handler - to be implemented"})
}

func (h *TrackerHandler) HandleIoTData(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "IoT data handler - to be implemented"})
}

func (h *TrackerHandler) GetUnverifiedActivities(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Get unverified activities - to be implemented"})
}

func (h *TrackerHandler) GetRecentActivities(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Get recent activities - to be implemented"})
}

// Response types
type ErrorResponse struct {
	Error   string `json:"error"`
	Details string `json:"details,omitempty"`
}

type SuccessResponse struct {
	Message string `json:"message"`
}

type ActivityListResponse struct {
	Activities interface{} `json:"activities"`
	Total      int64       `json:"total"`
	Limit      int         `json:"limit"`
	Offset     int         `json:"offset"`
}

type ActivityTypesResponse struct {
	ActivityTypes interface{} `json:"activity_types"`
	Total         int64       `json:"total"`
}
