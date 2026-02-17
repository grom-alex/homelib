package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// TokenValidator validates JWT tokens and returns user claims.
type TokenValidator interface {
	ValidateToken(tokenString string) (*Claims, error)
}

// Claims represents the JWT claims extracted from a token.
type Claims struct {
	UserID string
	Email  string
	Role   string
}

type AuthMiddleware struct {
	validator TokenValidator
}

func NewAuthMiddleware(validator TokenValidator) *AuthMiddleware {
	return &AuthMiddleware{validator: validator}
}

// RequireAuth returns middleware that requires a valid JWT token.
func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header"})
			c.Abort()
			return
		}

		claims, err := m.validator.ValidateToken(parts[1])
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("user_role", claims.Role)
		c.Next()
	}
}

// RequireAdmin returns middleware that requires the admin role.
func (m *AuthMiddleware) RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("user_role")
		if !exists || role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "admin only"})
			c.Abort()
			return
		}
		c.Next()
	}
}
