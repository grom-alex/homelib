package handler

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/grom-alex/homelib/backend/internal/models"
	"github.com/grom-alex/homelib/backend/internal/service"
)

type AuthHandler struct {
	authSvc    *service.AuthService
	refreshTTL time.Duration
}

func NewAuthHandler(authSvc *service.AuthService, refreshTTL time.Duration) *AuthHandler {
	return &AuthHandler{authSvc: authSvc, refreshTTL: refreshTTL}
}

// Register handles POST /api/auth/register.
func (h *AuthHandler) Register(c *gin.Context) {
	var input models.CreateUserInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input: " + err.Error()})
		return
	}

	result, err := h.authSvc.Register(c.Request.Context(), input)
	if err != nil {
		msg := err.Error()
		switch {
		case strings.Contains(msg, "already exists"):
			c.JSON(http.StatusConflict, gin.H{"error": msg})
		case strings.Contains(msg, "registration is disabled"):
			c.JSON(http.StatusForbidden, gin.H{"error": msg})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "registration failed"})
		}
		return
	}

	h.setRefreshCookie(c, result.RefreshToken)
	c.JSON(http.StatusCreated, gin.H{
		"user":         result.User,
		"access_token": result.AccessToken,
	})
}

// Login handles POST /api/auth/login.
func (h *AuthHandler) Login(c *gin.Context) {
	var input models.LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		return
	}

	result, err := h.authSvc.Login(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	h.setRefreshCookie(c, result.RefreshToken)
	c.JSON(http.StatusOK, gin.H{
		"user":         result.User,
		"access_token": result.AccessToken,
	})
}

// Refresh handles POST /api/auth/refresh.
func (h *AuthHandler) Refresh(c *gin.Context) {
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil || refreshToken == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "no refresh token"})
		return
	}

	result, err := h.authSvc.RefreshToken(c.Request.Context(), refreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid refresh token"})
		return
	}

	h.setRefreshCookie(c, result.RefreshToken)
	c.JSON(http.StatusOK, gin.H{
		"user":         result.User,
		"access_token": result.AccessToken,
	})
}

// Logout handles POST /api/auth/logout.
func (h *AuthHandler) Logout(c *gin.Context) {
	refreshToken, _ := c.Cookie("refresh_token")
	if refreshToken != "" {
		_ = h.authSvc.Logout(c.Request.Context(), refreshToken)
	}

	c.SetCookie("refresh_token", "", -1, "/", "", false, true)
	c.JSON(http.StatusOK, gin.H{"message": "logged out"})
}

func (h *AuthHandler) setRefreshCookie(c *gin.Context, token string) {
	maxAge := int(h.refreshTTL.Seconds())
	c.SetCookie("refresh_token", token, maxAge, "/", "", false, true)
}
