package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sloweyyy/GreenLedger/services/reporting/internal/service"
	"github.com/sloweyyy/GreenLedger/shared/logger"
	"github.com/sloweyyy/GreenLedger/shared/middleware"
)

// ReportingHandler handles HTTP requests for reporting operations
type ReportingHandler struct {
	reportingService *service.ReportingService
	logger           *logger.Logger
}

// NewReportingHandler creates a new reporting handler
func NewReportingHandler(reportingService *service.ReportingService, logger *logger.Logger) *ReportingHandler {
	return &ReportingHandler{
		reportingService: reportingService,
		logger:           logger,
	}
}

// RegisterRoutes registers reporting routes
func (h *ReportingHandler) RegisterRoutes(router *gin.RouterGroup, authMiddleware *middleware.AuthMiddleware) {
	reports := router.Group("/reports")
	{
		// Protected routes
		reports.Use(authMiddleware.RequireAuth())
		reports.POST("/", h.GenerateReport)
		reports.GET("/", h.GetUserReports)
		reports.GET("/:id", h.GetReport)
		reports.DELETE("/:id", h.DeleteReport)

		// Admin routes
		admin := reports.Group("/admin")
		admin.Use(authMiddleware.RequireRole("admin"))
		{
			admin.GET("/all", h.GetAllReports)
			admin.GET("/templates", h.GetReportTemplates)
		}
	}
}

// GenerateReport godoc
// @Summary Generate a new report
// @Description Generate a new report for the authenticated user
// @Tags reports
// @Accept json
// @Produce json
// @Param request body service.GenerateReportRequest true "Report generation request"
// @Success 201 {object} service.ReportResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BearerAuth
// @Router /reports [post]
func (h *ReportingHandler) GenerateReport(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "User not authenticated"})
		return
	}

	var req service.GenerateReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.LogError(c.Request.Context(), "invalid request body", err)
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request body",
			Details: err.Error(),
		})
		return
	}

	// Set user ID from authenticated user
	req.UserID = userID

	response, err := h.reportingService.GenerateReport(c.Request.Context(), &req)
	if err != nil {
		h.logger.LogError(c.Request.Context(), "failed to generate report", err,
			logger.String("user_id", userID),
			logger.String("report_type", req.Type))
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to generate report",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, response)
}

// GetReport godoc
// @Summary Get report by ID
// @Description Get a specific report by ID
// @Tags reports
// @Produce json
// @Param id path string true "Report ID"
// @Success 200 {object} service.ReportResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BearerAuth
// @Router /reports/{id} [get]
func (h *ReportingHandler) GetReport(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "User not authenticated"})
		return
	}

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid report ID",
			Details: err.Error(),
		})
		return
	}

	response, err := h.reportingService.GetReport(c.Request.Context(), id, userID)
	if err != nil {
		h.logger.LogError(c.Request.Context(), "failed to get report", err,
			logger.String("report_id", id.String()),
			logger.String("user_id", userID))
		c.JSON(http.StatusNotFound, ErrorResponse{
			Error:   "Report not found",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetUserReports godoc
// @Summary Get user reports
// @Description Get reports for the authenticated user
// @Tags reports
// @Produce json
// @Param limit query int false "Limit" default(20)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} ReportListResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BearerAuth
// @Router /reports [get]
func (h *ReportingHandler) GetUserReports(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "User not authenticated"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	reports, total, err := h.reportingService.GetUserReports(c.Request.Context(), userID, limit, offset)
	if err != nil {
		h.logger.LogError(c.Request.Context(), "failed to get user reports", err,
			logger.String("user_id", userID))
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to get reports",
			Details: err.Error(),
		})
		return
	}

	response := ReportListResponse{
		Reports: reports,
		Total:   total,
		Limit:   limit,
		Offset:  offset,
	}

	c.JSON(http.StatusOK, response)
}

// DeleteReport godoc
// @Summary Delete report
// @Description Delete a report by ID
// @Tags reports
// @Produce json
// @Param id path string true "Report ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BearerAuth
// @Router /reports/{id} [delete]
func (h *ReportingHandler) DeleteReport(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "User not authenticated"})
		return
	}

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid report ID",
			Details: err.Error(),
		})
		return
	}

	err = h.reportingService.DeleteReport(c.Request.Context(), id, userID)
	if err != nil {
		h.logger.LogError(c.Request.Context(), "failed to delete report", err,
			logger.String("report_id", id.String()),
			logger.String("user_id", userID))
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to delete report",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "Report deleted successfully",
	})
}

// Placeholder implementations for admin endpoints
func (h *ReportingHandler) GetAllReports(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Get all reports - to be implemented"})
}

func (h *ReportingHandler) GetReportTemplates(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Get report templates - to be implemented"})
}

// Response types
type ErrorResponse struct {
	Error   string `json:"error"`
	Details string `json:"details,omitempty"`
}

type SuccessResponse struct {
	Message string `json:"message"`
}

type ReportListResponse struct {
	Reports []*service.ReportResponse `json:"reports"`
	Total   int64                     `json:"total"`
	Limit   int                       `json:"limit"`
	Offset  int                       `json:"offset"`
}
