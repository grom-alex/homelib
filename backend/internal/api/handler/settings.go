package handler

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
)

type SettingsHandler struct {
	settingsRepo SettingsRepoer
}

func NewSettingsHandler(repo SettingsRepoer) *SettingsHandler {
	return &SettingsHandler{settingsRepo: repo}
}

// GetUserSettings handles GET /api/me/settings.
func (h *SettingsHandler) GetUserSettings(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "Пользователь не авторизован"})
		return
	}

	settings, err := h.settingsRepo.GetSettings(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Внутренняя ошибка сервера"})
		return
	}

	// Return {} if settings is empty or null
	if len(settings) == 0 || string(settings) == "{}" || string(settings) == "null" {
		c.JSON(http.StatusOK, json.RawMessage(`{}`))
		return
	}

	c.JSON(http.StatusOK, settings)
}

// UpdateUserSettings handles PUT /api/me/settings.
func (h *SettingsHandler) UpdateUserSettings(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "Пользователь не авторизован"})
		return
	}

	var patch json.RawMessage
	if err := c.ShouldBindJSON(&patch); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_json", "message": "Невалидный JSON"})
		return
	}

	result, err := h.settingsRepo.UpdateSettings(c.Request.Context(), userID, patch)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Внутренняя ошибка сервера"})
		return
	}

	c.JSON(http.StatusOK, result)
}
