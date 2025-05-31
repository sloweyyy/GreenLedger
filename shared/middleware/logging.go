package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/greenledger/shared/logger"
)

// RequestLogger creates a logging middleware for Gin
func RequestLogger(log *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Generate request ID
		requestID := uuid.New().String()
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)

		// Start timer
		start := time.Now()

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(start)

		// Log request
		log.LogInfo(c.Request.Context(), "HTTP request",
			logger.String("method", c.Request.Method),
			logger.String("path", c.Request.URL.Path),
			logger.String("query", c.Request.URL.RawQuery),
			logger.Int("status", c.Writer.Status()),
			logger.String("latency", latency.String()),
			logger.String("client_ip", c.ClientIP()),
			logger.String("user_agent", c.Request.UserAgent()),
			logger.String("request_id", requestID),
		)
	}
}

// ErrorLogger logs errors that occur during request processing
func ErrorLogger(log *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Log errors if any
		if len(c.Errors) > 0 {
			for _, err := range c.Errors {
				log.LogError(c.Request.Context(), "request error", err.Err,
					logger.String("type", err.Type.String()),
					logger.String("meta", err.Meta.(string)))
			}
		}
	}
}
