package handler

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/grom-alex/homelib/backend/internal/models"
)

// ParentalServicer is the interface that parental handlers need from the parental service.
type ParentalServicer interface {
	GetRestrictedGenreCodes(ctx context.Context) ([]string, error)
	SetRestrictedGenreCodes(ctx context.Context, codes []string) error
	SetPin(ctx context.Context, pin string) error
	VerifyPin(ctx context.Context, pin string) (bool, error)
	RemovePin(ctx context.Context) error
	IsPinSet(ctx context.Context) (bool, error)
	GetRestrictedGenreIDs(ctx context.Context) ([]int, error)
	IsAdultContentEnabled(ctx context.Context, userID string) (bool, error)
	SetAdultContentEnabled(ctx context.Context, userID string, enabled bool) error
	ListUsersAdultStatus(ctx context.Context) ([]models.UserAdultStatus, error)
}

type ParentalHandler struct {
	parentalSvc ParentalServicer
}

func NewParentalHandler(svc ParentalServicer) *ParentalHandler {
	return &ParentalHandler{parentalSvc: svc}
}

// --- Admin endpoints ---

// GetRestrictedGenres handles GET /api/admin/parental/genres.
func (h *ParentalHandler) GetRestrictedGenres(c *gin.Context) {
	codes, err := h.parentalSvc.GetRestrictedGenreCodes(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Не удалось получить список ограниченных жанров"})
		return
	}
	if codes == nil {
		codes = []string{}
	}
	c.JSON(http.StatusOK, gin.H{"codes": codes})
}

// UpdateRestrictedGenres handles PUT /api/admin/parental/genres.
func (h *ParentalHandler) UpdateRestrictedGenres(c *gin.Context) {
	var input models.UpdateRestrictedGenresInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_input", "message": "Невалидные данные"})
		return
	}
	if err := h.parentalSvc.SetRestrictedGenreCodes(c.Request.Context(), input.Codes); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Не удалось обновить список жанров"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"codes": input.Codes})
}

// SetPin handles POST /api/admin/parental/pin.
func (h *ParentalHandler) SetPin(c *gin.Context) {
	var input models.SetPinInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_input", "message": "PIN должен быть от 4 до 6 символов"})
		return
	}
	if err := h.parentalSvc.SetPin(c.Request.Context(), input.Pin); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Не удалось установить PIN"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// RemovePin handles DELETE /api/admin/parental/pin.
func (h *ParentalHandler) RemovePin(c *gin.Context) {
	if err := h.parentalSvc.RemovePin(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Не удалось удалить PIN"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// GetAdminParentalStatus handles GET /api/admin/parental/status.
func (h *ParentalHandler) GetAdminParentalStatus(c *gin.Context) {
	pinSet, err := h.parentalSvc.IsPinSet(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}
	codes, err := h.parentalSvc.GetRestrictedGenreCodes(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}
	if codes == nil {
		codes = []string{}
	}
	c.JSON(http.StatusOK, models.AdminParentalStatus{
		PinSet:               pinSet,
		RestrictedGenreCodes: codes,
	})
}

// ListUsersAdultStatus handles GET /api/admin/parental/users.
func (h *ParentalHandler) ListUsersAdultStatus(c *gin.Context) {
	users, err := h.parentalSvc.ListUsersAdultStatus(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Не удалось получить список пользователей"})
		return
	}
	if users == nil {
		users = []models.UserAdultStatus{}
	}
	c.JSON(http.StatusOK, users)
}

// SetUserAdultContent handles PUT /api/admin/parental/users/:userId.
func (h *ParentalHandler) SetUserAdultContent(c *gin.Context) {
	userID := c.Param("userId")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_input", "message": "ID пользователя не указан"})
		return
	}

	var input models.SetUserAdultContentInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_input", "message": "Невалидные данные"})
		return
	}

	if err := h.parentalSvc.SetAdultContentEnabled(c.Request.Context(), userID, input.AdultContentEnabled); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Не удалось обновить настройки пользователя"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok", "adult_content_enabled": input.AdultContentEnabled})
}

// --- User endpoints ---

// GetMyParentalStatus handles GET /api/me/parental/status.
func (h *ParentalHandler) GetMyParentalStatus(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	pinSet, err := h.parentalSvc.IsPinSet(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}

	enabled, err := h.parentalSvc.IsAdultContentEnabled(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}

	c.JSON(http.StatusOK, models.ParentalStatus{
		PinSet:              pinSet,
		AdultContentEnabled: enabled,
	})
}

// UnlockAdultContent handles POST /api/me/parental/unlock.
func (h *ParentalHandler) UnlockAdultContent(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var input models.VerifyPinInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_input", "message": "PIN не указан"})
		return
	}

	valid, err := h.parentalSvc.VerifyPin(c.Request.Context(), input.Pin)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}
	if !valid {
		c.JSON(http.StatusForbidden, gin.H{"error": "invalid_pin", "message": "Неверный PIN"})
		return
	}

	if err := h.parentalSvc.SetAdultContentEnabled(c.Request.Context(), userID, true); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok", "adult_content_enabled": true})
}

// LockAdultContent handles POST /api/me/parental/lock.
func (h *ParentalHandler) LockAdultContent(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	if err := h.parentalSvc.SetAdultContentEnabled(c.Request.Context(), userID, false); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok", "adult_content_enabled": false})
}
