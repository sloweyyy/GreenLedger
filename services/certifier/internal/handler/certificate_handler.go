package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sloweyyy/GreenLedger/services/certifier/internal/service"
	"github.com/sloweyyy/GreenLedger/shared/logger"
	"github.com/sloweyyy/GreenLedger/shared/middleware"
)

// CertificateHandler handles HTTP requests for certificate operations
type CertificateHandler struct {
	certificateService *service.CertificateService
	logger             *logger.Logger
}

// NewCertificateHandler creates a new certificate handler
func NewCertificateHandler(certificateService *service.CertificateService, logger *logger.Logger) *CertificateHandler {
	return &CertificateHandler{
		certificateService: certificateService,
		logger:             logger,
	}
}

// RegisterRoutes registers certificate routes
func (h *CertificateHandler) RegisterRoutes(router *gin.RouterGroup, authMiddleware *middleware.AuthMiddleware) {
	certificates := router.Group("/certificates")
	{
		// Public routes
		certificates.GET("/verify/:certificate_number", h.VerifyCertificate)

		// Protected routes
		certificates.Use(authMiddleware.RequireAuth())
		certificates.POST("/", h.IssueCertificate)
		certificates.GET("/", h.GetUserCertificates)
		certificates.GET("/:id", h.GetCertificate)
		certificates.POST("/:id/retire", h.RetireCertificate)

		// Admin routes
		admin := certificates.Group("/admin")
		admin.Use(authMiddleware.RequireRole("admin"))
		{
			admin.GET("/all", h.GetAllCertificates)
			admin.GET("/pending", h.GetPendingCertificates)
		}
	}
}

// IssueCertificate godoc
// @Summary Issue a new certificate
// @Description Issue a new carbon offset certificate
// @Tags certificates
// @Accept json
// @Produce json
// @Param request body service.IssueCertificateRequest true "Certificate issue request"
// @Success 201 {object} service.CertificateResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BearerAuth
// @Router /certificates [post]
func (h *CertificateHandler) IssueCertificate(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "User not authenticated"})
		return
	}

	var req service.IssueCertificateRequest
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

	response, err := h.certificateService.IssueCertificate(c.Request.Context(), &req)
	if err != nil {
		h.logger.LogError(c.Request.Context(), "failed to issue certificate", err,
			logger.String("user_id", userID))
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to issue certificate",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, response)
}

// GetCertificate godoc
// @Summary Get certificate by ID
// @Description Get a specific certificate by ID
// @Tags certificates
// @Produce json
// @Param id path string true "Certificate ID"
// @Success 200 {object} service.CertificateResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BearerAuth
// @Router /certificates/{id} [get]
func (h *CertificateHandler) GetCertificate(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "User not authenticated"})
		return
	}

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid certificate ID",
			Details: err.Error(),
		})
		return
	}

	response, err := h.certificateService.GetCertificate(c.Request.Context(), id, userID)
	if err != nil {
		h.logger.LogError(c.Request.Context(), "failed to get certificate", err,
			logger.String("certificate_id", id.String()),
			logger.String("user_id", userID))
		c.JSON(http.StatusNotFound, ErrorResponse{
			Error:   "Certificate not found",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetUserCertificates godoc
// @Summary Get user certificates
// @Description Get certificates for the authenticated user
// @Tags certificates
// @Produce json
// @Param limit query int false "Limit" default(20)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} CertificateListResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BearerAuth
// @Router /certificates [get]
func (h *CertificateHandler) GetUserCertificates(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "User not authenticated"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	certificates, total, err := h.certificateService.GetUserCertificates(c.Request.Context(), userID, limit, offset)
	if err != nil {
		h.logger.LogError(c.Request.Context(), "failed to get user certificates", err,
			logger.String("user_id", userID))
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to get certificates",
			Details: err.Error(),
		})
		return
	}

	response := CertificateListResponse{
		Certificates: certificates,
		Total:        total,
		Limit:        limit,
		Offset:       offset,
	}

	c.JSON(http.StatusOK, response)
}

// VerifyCertificate godoc
// @Summary Verify certificate
// @Description Verify a certificate by certificate number
// @Tags certificates
// @Produce json
// @Param certificate_number path string true "Certificate Number"
// @Success 200 {object} service.CertificateResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /certificates/verify/{certificate_number} [get]
func (h *CertificateHandler) VerifyCertificate(c *gin.Context) {
	certificateNumber := c.Param("certificate_number")

	response, err := h.certificateService.VerifyCertificate(c.Request.Context(), certificateNumber)
	if err != nil {
		h.logger.LogError(c.Request.Context(), "failed to verify certificate", err,
			logger.String("certificate_number", certificateNumber))
		c.JSON(http.StatusNotFound, ErrorResponse{
			Error:   "Certificate verification failed",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// RetireCertificate godoc
// @Summary Retire certificate
// @Description Retire a certificate (permanent action)
// @Tags certificates
// @Produce json
// @Param id path string true "Certificate ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BearerAuth
// @Router /certificates/{id}/retire [post]
func (h *CertificateHandler) RetireCertificate(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "User not authenticated"})
		return
	}

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid certificate ID",
			Details: err.Error(),
		})
		return
	}

	err = h.certificateService.RetireCertificate(c.Request.Context(), id, userID)
	if err != nil {
		h.logger.LogError(c.Request.Context(), "failed to retire certificate", err,
			logger.String("certificate_id", id.String()),
			logger.String("user_id", userID))
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to retire certificate",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "Certificate retired successfully",
	})
}

// Placeholder implementations for admin endpoints
func (h *CertificateHandler) GetAllCertificates(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Get all certificates - to be implemented"})
}

func (h *CertificateHandler) GetPendingCertificates(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Get pending certificates - to be implemented"})
}

// Response types
type ErrorResponse struct {
	Error   string `json:"error"`
	Details string `json:"details,omitempty"`
}

type SuccessResponse struct {
	Message string `json:"message"`
}

type CertificateListResponse struct {
	Certificates []*service.CertificateResponse `json:"certificates"`
	Total        int64                          `json:"total"`
	Limit        int                            `json:"limit"`
	Offset       int                            `json:"offset"`
}
