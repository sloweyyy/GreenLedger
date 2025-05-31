package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sloweyyy/GreenLedger/services/wallet/internal/service"
	"github.com/sloweyyy/GreenLedger/shared/logger"
	"github.com/sloweyyy/GreenLedger/shared/middleware"
	"github.com/shopspring/decimal"
)

// WalletHandler handles HTTP requests for wallet operations
type WalletHandler struct {
	walletService *service.WalletService
	logger        *logger.Logger
}

// NewWalletHandler creates a new wallet handler
func NewWalletHandler(walletService *service.WalletService, logger *logger.Logger) *WalletHandler {
	return &WalletHandler{
		walletService: walletService,
		logger:        logger,
	}
}

// RegisterRoutes registers wallet routes
func (h *WalletHandler) RegisterRoutes(router *gin.RouterGroup, authMiddleware *middleware.AuthMiddleware) {
	wallet := router.Group("/wallet")
	{
		// Protected routes
		wallet.Use(authMiddleware.RequireAuth())
		wallet.GET("/balance", h.GetBalance)
		wallet.GET("/transactions", h.GetTransactionHistory)
		wallet.GET("/transactions/:id", h.GetTransactionByID)
		wallet.POST("/transfer", h.TransferCredits)
		wallet.GET("/stats", h.GetWalletStats)

		// Admin routes
		admin := wallet.Group("/admin")
		admin.Use(authMiddleware.RequireRole("admin"))
		{
			admin.POST("/credit", h.CreditBalance)
			admin.POST("/debit", h.DebitBalance)
			admin.GET("/transactions/pending", h.GetPendingTransactions)
			admin.GET("/users/top", h.GetTopUsers)
		}
	}
}

// GetBalance godoc
// @Summary Get wallet balance
// @Description Get wallet balance for the authenticated user
// @Tags wallet
// @Produce json
// @Success 200 {object} service.WalletResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BearerAuth
// @Router /wallet/balance [get]
func (h *WalletHandler) GetBalance(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "User not authenticated"})
		return
	}

	balance, err := h.walletService.GetBalance(c.Request.Context(), userID)
	if err != nil {
		h.logger.LogError(c.Request.Context(), "failed to get wallet balance", err,
			logger.String("user_id", userID))
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to get balance",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, balance)
}

// GetTransactionHistory godoc
// @Summary Get transaction history
// @Description Get transaction history for the authenticated user
// @Tags wallet
// @Produce json
// @Param limit query int false "Limit" default(20)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} TransactionHistoryResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BearerAuth
// @Router /wallet/transactions [get]
func (h *WalletHandler) GetTransactionHistory(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "User not authenticated"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	transactions, total, err := h.walletService.GetTransactionHistory(c.Request.Context(), userID, limit, offset)
	if err != nil {
		h.logger.LogError(c.Request.Context(), "failed to get transaction history", err,
			logger.String("user_id", userID))
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to get transaction history",
			Details: err.Error(),
		})
		return
	}

	response := TransactionHistoryResponse{
		Transactions: transactions,
		Total:        total,
		Limit:        limit,
		Offset:       offset,
	}

	c.JSON(http.StatusOK, response)
}

// GetTransactionByID godoc
// @Summary Get transaction by ID
// @Description Get a specific transaction by ID
// @Tags wallet
// @Produce json
// @Param id path string true "Transaction ID"
// @Success 200 {object} service.TransactionResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BearerAuth
// @Router /wallet/transactions/{id} [get]
func (h *WalletHandler) GetTransactionByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid transaction ID",
			Details: err.Error(),
		})
		return
	}

	// This would need to be implemented in the service
	c.JSON(http.StatusOK, gin.H{
		"message":        "Get transaction by ID - to be implemented",
		"transaction_id": id.String(),
	})
}

// TransferCredits godoc
// @Summary Transfer credits
// @Description Transfer credits to another user
// @Tags wallet
// @Accept json
// @Produce json
// @Param request body service.TransferCreditsRequest true "Transfer request"
// @Success 200 {object} service.TransferResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BearerAuth
// @Router /wallet/transfer [post]
func (h *WalletHandler) TransferCredits(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "User not authenticated"})
		return
	}

	var req service.TransferCreditsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.LogError(c.Request.Context(), "invalid request body", err)
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request body",
			Details: err.Error(),
		})
		return
	}

	// Set from user ID from authenticated user
	req.FromUserID = userID

	response, err := h.walletService.TransferCredits(c.Request.Context(), &req)
	if err != nil {
		h.logger.LogError(c.Request.Context(), "failed to transfer credits", err,
			logger.String("from_user_id", userID),
			logger.String("to_user_id", req.ToUserID))
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to transfer credits",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetWalletStats godoc
// @Summary Get wallet statistics
// @Description Get wallet statistics for the authenticated user
// @Tags wallet
// @Produce json
// @Param start_date query string false "Start date (RFC3339 format)"
// @Param end_date query string false "End date (RFC3339 format)"
// @Success 200 {object} WalletStatsResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BearerAuth
// @Router /wallet/stats [get]
func (h *WalletHandler) GetWalletStats(c *gin.Context) {
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

	// This would need to be implemented in the service
	c.JSON(http.StatusOK, gin.H{
		"message":    "Get wallet stats - to be implemented",
		"user_id":    userID,
		"start_date": startDate,
		"end_date":   endDate,
	})
}

// CreditBalance godoc
// @Summary Credit wallet balance (Admin)
// @Description Credit a user's wallet balance (admin only)
// @Tags wallet
// @Accept json
// @Produce json
// @Param request body CreditBalanceRequest true "Credit request"
// @Success 200 {object} service.TransactionResponse
// @Failure 400 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BearerAuth
// @Router /wallet/admin/credit [post]
func (h *WalletHandler) CreditBalance(c *gin.Context) {
	var req CreditBalanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request body",
			Details: err.Error(),
		})
		return
	}

	// Convert to service request
	serviceReq := &service.CreditBalanceRequest{
		UserID:      req.UserID,
		Amount:      req.Amount,
		Source:      req.Source,
		Description: req.Description,
		ReferenceID: req.ReferenceID,
		Metadata:    req.Metadata,
	}

	response, err := h.walletService.CreditBalance(c.Request.Context(), serviceReq)
	if err != nil {
		h.logger.LogError(c.Request.Context(), "failed to credit balance", err,
			logger.String("user_id", req.UserID))
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to credit balance",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// DebitBalance godoc
// @Summary Debit wallet balance (Admin)
// @Description Debit a user's wallet balance (admin only)
// @Tags wallet
// @Accept json
// @Produce json
// @Param request body DebitBalanceRequest true "Debit request"
// @Success 200 {object} service.TransactionResponse
// @Failure 400 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BearerAuth
// @Router /wallet/admin/debit [post]
func (h *WalletHandler) DebitBalance(c *gin.Context) {
	var req DebitBalanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request body",
			Details: err.Error(),
		})
		return
	}

	// Convert to service request
	serviceReq := &service.DebitBalanceRequest{
		UserID:      req.UserID,
		Amount:      req.Amount,
		Description: req.Description,
		ReferenceID: req.ReferenceID,
		Metadata:    req.Metadata,
	}

	response, err := h.walletService.DebitBalance(c.Request.Context(), serviceReq)
	if err != nil {
		h.logger.LogError(c.Request.Context(), "failed to debit balance", err,
			logger.String("user_id", req.UserID))
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to debit balance",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// Placeholder implementations for remaining endpoints
func (h *WalletHandler) GetPendingTransactions(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Get pending transactions - to be implemented"})
}

func (h *WalletHandler) GetTopUsers(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Get top users - to be implemented"})
}

// Request/Response types
type CreditBalanceRequest struct {
	UserID      string                 `json:"user_id" binding:"required"`
	Amount      decimal.Decimal        `json:"amount" binding:"required"`
	Source      string                 `json:"source" binding:"required"`
	Description string                 `json:"description" binding:"required"`
	ReferenceID string                 `json:"reference_id"`
	Metadata    map[string]interface{} `json:"metadata"`
}

type DebitBalanceRequest struct {
	UserID      string                 `json:"user_id" binding:"required"`
	Amount      decimal.Decimal        `json:"amount" binding:"required"`
	Description string                 `json:"description" binding:"required"`
	ReferenceID string                 `json:"reference_id"`
	Metadata    map[string]interface{} `json:"metadata"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Details string `json:"details,omitempty"`
}

type TransactionHistoryResponse struct {
	Transactions interface{} `json:"transactions"`
	Total        int64       `json:"total"`
	Limit        int         `json:"limit"`
	Offset       int         `json:"offset"`
}

type WalletStatsResponse struct {
	Stats interface{} `json:"stats"`
}
