package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- GetUserSettings ---

func TestSettingsHandler_GetUserSettings_Empty(t *testing.T) {
	repo := &mockSettingsRepo{
		getSettingsFn: func(_ context.Context, userID string) (json.RawMessage, error) {
			assert.Equal(t, "user-123", userID)
			return json.RawMessage(`{}`), nil
		},
	}
	h := NewSettingsHandler(repo)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/me/settings", nil)
	c.Set("user_id", "user-123")

	h.GetUserSettings(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{}`, w.Body.String())
}

func TestSettingsHandler_GetUserSettings_WithReader(t *testing.T) {
	settings := `{"reader":{"fontSize":20,"theme":"dark"}}`
	repo := &mockSettingsRepo{
		getSettingsFn: func(_ context.Context, _ string) (json.RawMessage, error) {
			return json.RawMessage(settings), nil
		},
	}
	h := NewSettingsHandler(repo)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/me/settings", nil)
	c.Set("user_id", "user-123")

	h.GetUserSettings(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	reader := resp["reader"].(map[string]interface{})
	assert.Equal(t, float64(20), reader["fontSize"])
	assert.Equal(t, "dark", reader["theme"])
}

func TestSettingsHandler_GetUserSettings_NoUserID(t *testing.T) {
	h := NewSettingsHandler(&mockSettingsRepo{})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/me/settings", nil)

	h.GetUserSettings(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestSettingsHandler_GetUserSettings_DBError(t *testing.T) {
	repo := &mockSettingsRepo{
		getSettingsFn: func(_ context.Context, _ string) (json.RawMessage, error) {
			return nil, fmt.Errorf("connection lost")
		},
	}
	h := NewSettingsHandler(repo)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/me/settings", nil)
	c.Set("user_id", "user-123")

	h.GetUserSettings(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// --- UpdateUserSettings ---

func TestSettingsHandler_UpdateUserSettings_PartialMerge(t *testing.T) {
	merged := `{"reader":{"fontSize":20,"fontFamily":"Georgia","theme":"dark"}}`
	repo := &mockSettingsRepo{
		updateSettingsFn: func(_ context.Context, userID string, patch json.RawMessage) (json.RawMessage, error) {
			assert.Equal(t, "user-123", userID)
			var p map[string]interface{}
			require.NoError(t, json.Unmarshal(patch, &p))
			reader := p["reader"].(map[string]interface{})
			assert.Equal(t, float64(20), reader["fontSize"])
			assert.Equal(t, "dark", reader["theme"])
			return json.RawMessage(merged), nil
		},
	}
	h := NewSettingsHandler(repo)

	body := `{"reader":{"fontSize":20,"theme":"dark"}}`
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPut, "/api/me/settings", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_id", "user-123")

	h.UpdateUserSettings(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	reader := resp["reader"].(map[string]interface{})
	assert.Equal(t, float64(20), reader["fontSize"])
	assert.Equal(t, "Georgia", reader["fontFamily"])
	assert.Equal(t, "dark", reader["theme"])
}

func TestSettingsHandler_UpdateUserSettings_InvalidJSON(t *testing.T) {
	h := NewSettingsHandler(&mockSettingsRepo{})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPut, "/api/me/settings", strings.NewReader("not json"))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_id", "user-123")

	h.UpdateUserSettings(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "invalid_json")
}

func TestSettingsHandler_UpdateUserSettings_NoUserID(t *testing.T) {
	h := NewSettingsHandler(&mockSettingsRepo{})

	body := `{"reader":{"fontSize":20}}`
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPut, "/api/me/settings", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	h.UpdateUserSettings(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestSettingsHandler_UpdateUserSettings_DBError(t *testing.T) {
	repo := &mockSettingsRepo{
		updateSettingsFn: func(_ context.Context, _ string, _ json.RawMessage) (json.RawMessage, error) {
			return nil, fmt.Errorf("disk full")
		},
	}
	h := NewSettingsHandler(repo)

	body := `{"reader":{"fontSize":20}}`
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPut, "/api/me/settings", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_id", "user-123")

	h.UpdateUserSettings(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
