package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/greenledger/services/calculator/internal/service"
	"github.com/greenledger/shared/logger"
	"github.com/greenledger/shared/middleware"
)

// CalculatorHandler handles HTTP requests for carbon footprint calculations
type CalculatorHandler struct {
	calculatorService *service.CalculatorService
	logger            *logger.Logger
}

// NewCalculatorHandler creates a new calculator handler
func NewCalculatorHandler(calculatorService *service.CalculatorService, logger *logger.Logger) *CalculatorHandler {
	return &CalculatorHandler{
		calculatorService: calculatorService,
		logger:            logger,
	}
}

// RegisterRoutes registers calculator routes
func (h *CalculatorHandler) RegisterRoutes(router *gin.RouterGroup, authMiddleware *middleware.AuthMiddleware) {
	calculator := router.Group("/calculator")
	{
		// Public routes
		calculator.GET("/emission-factors", h.GetEmissionFactors)
		calculator.GET("/emission-factors/:activity_type", h.GetEmissionFactorsByType)

		// Protected routes
		calculator.Use(authMiddleware.RequireAuth())
		calculator.POST("/calculate", h.CalculateFootprint)
		calculator.GET("/calculations", h.GetCalculationHistory)
		calculator.GET("/calculations/:id", h.GetCalculationByID)
		calculator.GET("/stats", h.GetUserStats)
	}
}

// CalculateFootprint godoc
// @Summary Calculate carbon footprint
// @Description Calculate carbon footprint for given activities
// @Tags calculator
// @Accept json
// @Produce json
// @Param request body service.CalculateFootprintRequest true "Calculation request"
// @Success 200 {object} service.CalculateFootprintResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BearerAuth
// @Router /calculator/calculate [post]
func (h *CalculatorHandler) CalculateFootprint(c *gin.Context) {
	var req service.CalculateFootprintRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.LogError(c.Request.Context(), "invalid request body", err)
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request body",
			Details: err.Error(),
		})
		return
	}

	// Get user ID from context (set by auth middleware)
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "User not authenticated"})
		return
	}
	req.UserID = userID

	response, err := h.calculatorService.CalculateFootprint(c.Request.Context(), &req)
	if err != nil {
		h.logger.LogError(c.Request.Context(), "failed to calculate footprint", err,
			logger.String("user_id", userID))
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to calculate footprint",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetCalculationHistory godoc
// @Summary Get calculation history
// @Description Get calculation history for the authenticated user
// @Tags calculator
// @Produce json
// @Param start_date query string false "Start date (RFC3339 format)"
// @Param end_date query string false "End date (RFC3339 format)"
// @Param limit query int false "Limit" default(20)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} CalculationHistoryResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BearerAuth
// @Router /calculator/calculations [get]
func (h *CalculatorHandler) GetCalculationHistory(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "User not authenticated"})
		return
	}

	// Parse query parameters
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	var startDate, endDate *time.Time
	if startDateStr := c.Query("start_date"); startDateStr != "" {
		if parsed, err := time.Parse(time.RFC3339, startDateStr); err == nil {
			startDate = &parsed
		}
	}
	if endDateStr := c.Query("end_date"); endDateStr != "" {
		if parsed, err := time.Parse(time.RFC3339, endDateStr); err == nil {
			endDate = &parsed
		}
	}

	calculations, total, err := h.calculatorService.GetCalculationHistory(
		c.Request.Context(), userID, startDate, endDate, limit, offset)
	if err != nil {
		h.logger.LogError(c.Request.Context(), "failed to get calculation history", err,
			logger.String("user_id", userID))
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to get calculation history",
			Details: err.Error(),
		})
		return
	}

	response := CalculationHistoryResponse{
		Calculations: calculations,
		Total:        total,
		Limit:        limit,
		Offset:       offset,
	}

	c.JSON(http.StatusOK, response)
}

// GetCalculationByID godoc
// @Summary Get calculation by ID
// @Description Get a specific calculation by ID
// @Tags calculator
// @Produce json
// @Param id path string true "Calculation ID"
// @Success 200 {object} models.Calculation
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BearerAuth
// @Router /calculator/calculations/{id} [get]
func (h *CalculatorHandler) GetCalculationByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid calculation ID",
			Details: err.Error(),
		})
		return
	}

	calculation, err := h.calculatorService.GetCalculationByID(c.Request.Context(), id)
	if err != nil {
		h.logger.LogError(c.Request.Context(), "failed to get calculation", err,
			logger.String("calculation_id", id.String()))
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to get calculation",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, calculation)
}

// GetUserStats godoc
// @Summary Get user calculation statistics
// @Description Get calculation statistics for the authenticated user
// @Tags calculator
// @Produce json
// @Param start_date query string false "Start date (RFC3339 format)"
// @Param end_date query string false "End date (RFC3339 format)"
// @Success 200 {object} repository.UserCalculationStats
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BearerAuth
// @Router /calculator/stats [get]
func (h *CalculatorHandler) GetUserStats(c *gin.Context) {
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

	// This would need to be implemented in the service layer
	// stats, err := h.calculatorService.GetUserStats(c.Request.Context(), userID, startDate, endDate)
	// For now, return a placeholder
	c.JSON(http.StatusOK, gin.H{
		"message": "User stats endpoint - to be implemented",
		"user_id": userID,
		"start_date": startDate,
		"end_date": endDate,
	})
}

// GetEmissionFactors godoc
// @Summary Get emission factors
// @Description Get all emission factors with optional filtering
// @Tags calculator
// @Produce json
// @Param activity_type query string false "Activity type filter"
// @Param location query string false "Location filter"
// @Param limit query int false "Limit" default(50)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} EmissionFactorsResponse
// @Failure 500 {object} ErrorResponse
// @Router /calculator/emission-factors [get]
func (h *CalculatorHandler) GetEmissionFactors(c *gin.Context) {
	// This would need to be implemented in the service layer
	c.JSON(http.StatusOK, gin.H{
		"message": "Emission factors endpoint - to be implemented",
	})
}

// GetEmissionFactorsByType godoc
// @Summary Get emission factors by activity type
// @Description Get emission factors for a specific activity type
// @Tags calculator
// @Produce json
// @Param activity_type path string true "Activity type"
// @Success 200 {object} EmissionFactorsResponse
// @Failure 500 {object} ErrorResponse
// @Router /calculator/emission-factors/{activity_type} [get]
func (h *CalculatorHandler) GetEmissionFactorsByType(c *gin.Context) {
	activityType := c.Param("activity_type")
	
	// This would need to be implemented in the service layer
	c.JSON(http.StatusOK, gin.H{
		"message": "Emission factors by type endpoint - to be implemented",
		"activity_type": activityType,
	})
}

// Response types
type ErrorResponse struct {
	Error   string `json:"error"`
	Details string `json:"details,omitempty"`
}

type CalculationHistoryResponse struct {
	Calculations interface{} `json:"calculations"`
	Total        int64       `json:"total"`
	Limit        int         `json:"limit"`
	Offset       int         `json:"offset"`
}

type EmissionFactorsResponse struct {
	Factors interface{} `json:"factors"`
	Total   int64       `json:"total"`
	Limit   int         `json:"limit"`
	Offset  int         `json:"offset"`
}
