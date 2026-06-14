// Package middleware provides Gin HTTP middleware for authentication, logging and metrics.
package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"github.com/sloweyyy/GreenLedger/shared/logger"
)

// contextKey is a private type for request-context keys, avoiding collisions
// with keys defined in other packages (see context.WithValue documentation).
type contextKey string

const (
	userIDKey    contextKey = "user_id"
	userEmailKey contextKey = "user_email"
	userRolesKey contextKey = "user_roles"
)

// UserIDFromContext returns the authenticated user ID stored in the request
// context by AuthMiddleware, or ("", false) if it is absent.
func UserIDFromContext(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(userIDKey).(string)
	return v, ok
}

// UserEmailFromContext returns the authenticated user email stored in the
// request context by AuthMiddleware, or ("", false) if it is absent.
func UserEmailFromContext(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(userEmailKey).(string)
	return v, ok
}

// UserRolesFromContext returns the authenticated user roles stored in the
// request context by AuthMiddleware, or (nil, false) if they are absent.
func UserRolesFromContext(ctx context.Context) ([]string, bool) {
	v, ok := ctx.Value(userRolesKey).([]string)
	return v, ok
}

// AuthMiddleware provides JWT authentication middleware
type AuthMiddleware struct {
	jwtSecret []byte
	logger    *logger.Logger
}

// NewAuthMiddleware creates a new auth middleware instance
func NewAuthMiddleware(jwtSecret string, logger *logger.Logger) *AuthMiddleware {
	return &AuthMiddleware{
		jwtSecret: []byte(jwtSecret),
		logger:    logger,
	}
}

// Claims represents JWT claims
type Claims struct {
	UserID string   `json:"user_id"`
	Email  string   `json:"email"`
	Roles  []string `json:"roles"`
	jwt.RegisteredClaims
}

// RequireAuth middleware that requires valid JWT token
func (a *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := a.extractToken(c)
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing authorization token"})
			c.Abort()
			return
		}

		claims, err := a.validateToken(token)
		if err != nil {
			a.logger.LogError(c.Request.Context(), "invalid token", err,
				logger.String("token", token[:10]+"..."))
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			c.Abort()
			return
		}

		// Add claims to context
		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("user_roles", claims.Roles)

		// Add to request context for downstream services
		ctx := context.WithValue(c.Request.Context(), userIDKey, claims.UserID)
		ctx = context.WithValue(ctx, userEmailKey, claims.Email)
		ctx = context.WithValue(ctx, userRolesKey, claims.Roles)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

// RequireRole middleware that requires specific role
func (a *AuthMiddleware) RequireRole(requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		roles, exists := c.Get("user_roles")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{"error": "no roles found"})
			c.Abort()
			return
		}

		userRoles, ok := roles.([]string)
		if !ok {
			c.JSON(http.StatusForbidden, gin.H{"error": "invalid roles format"})
			c.Abort()
			return
		}

		hasRole := false
		for _, role := range userRoles {
			if role == requiredRole || role == "admin" {
				hasRole = true
				break
			}
		}

		if !hasRole {
			userIDVal, _ := c.Get("user_id")
			userID, _ := userIDVal.(string)
			a.logger.LogWarn(c.Request.Context(), "insufficient permissions",
				logger.String("user_id", userID),
				logger.String("required_role", requiredRole),
				logger.Any("user_roles", userRoles))
			c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// OptionalAuth middleware that extracts user info if token is present
func (a *AuthMiddleware) OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := a.extractToken(c)
		if token == "" {
			c.Next()
			return
		}

		claims, err := a.validateToken(token)
		if err != nil {
			// Log but don't fail the request
			a.logger.LogDebug(c.Request.Context(), "optional auth failed",
				logger.String("error", err.Error()))
			c.Next()
			return
		}

		// Add claims to context
		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("user_roles", claims.Roles)

		c.Next()
	}
}

// extractToken extracts JWT token from Authorization header
func (a *AuthMiddleware) extractToken(c *gin.Context) string {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return ""
	}

	// Check for Bearer token
	if strings.HasPrefix(authHeader, "Bearer ") {
		return strings.TrimPrefix(authHeader, "Bearer ")
	}

	return ""
}

// validateToken validates JWT token and returns claims
func (a *AuthMiddleware) validateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return a.jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, jwt.ErrInvalidKey
}

// GetUserID extracts user ID from gin context
func GetUserID(c *gin.Context) (string, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return "", false
	}

	id, ok := userID.(string)
	return id, ok
}

// GetUserEmail extracts user email from gin context
func GetUserEmail(c *gin.Context) (string, bool) {
	email, exists := c.Get("user_email")
	if !exists {
		return "", false
	}

	userEmail, ok := email.(string)
	return userEmail, ok
}

// GetUserRoles extracts user roles from gin context
func GetUserRoles(c *gin.Context) ([]string, bool) {
	roles, exists := c.Get("user_roles")
	if !exists {
		return nil, false
	}

	userRoles, ok := roles.([]string)
	return userRoles, ok
}

// HasRole checks if user has specific role
func HasRole(c *gin.Context, role string) bool {
	roles, exists := GetUserRoles(c)
	if !exists {
		return false
	}

	for _, r := range roles {
		if r == role || r == "admin" {
			return true
		}
	}
	return false
}
