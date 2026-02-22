package handler

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
)

// maxSettingsSize is the upper bound (bytes) for the accumulated user settings JSON.
const maxSettingsSize = 64 * 1024

// UserSettingsPatch whitelists the top-level keys a client may store.
// Unknown keys are silently dropped during JSON unmarshalling.
type UserSettingsPatch struct {
	Reader json.RawMessage `json:"reader,omitempty"`
	UI     json.RawMessage `json:"ui,omitempty"`
}

type SettingsHandler struct {
	settingsRepo SettingsRepository
}

func NewSettingsHandler(repo SettingsRepository) *SettingsHandler {
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

	// Limit request body to prevent abuse
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxSettingsSize)

	// Unmarshal into a whitelist struct so unknown top-level keys are dropped.
	var input UserSettingsPatch
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_json", "message": "Невалидный JSON"})
		return
	}

	patch, err := json.Marshal(input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Внутренняя ошибка сервера"})
		return
	}

	result, err := h.settingsRepo.UpdateSettings(c.Request.Context(), userID, patch)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Внутренняя ошибка сервера"})
		return
	}

	if len(result) > maxSettingsSize {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "settings_too_large", "message": "Настройки превышают допустимый размер"})
		return
	}

	c.JSON(http.StatusOK, result)
}
