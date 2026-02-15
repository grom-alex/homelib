package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/grom-alex/homelib/backend/internal/config"
	"github.com/grom-alex/homelib/backend/internal/models"
	"github.com/grom-alex/homelib/backend/internal/service"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestAdminHandler_StartImport_Accepted(t *testing.T) {
	svc := service.NewImportService(nil, config.ImportConfig{}, config.LibraryConfig{INPXPath: "/dev/null"}, nil, nil, nil, nil, nil)
	h := NewAdminHandler(svc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/admin/import", nil)

	h.StartImport(c)

	assert.Equal(t, http.StatusAccepted, w.Code)

	var resp map[string]string
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "import started", resp["message"])
}

func TestAdminHandler_StartImport_Conflict(t *testing.T) {
	svc := service.NewImportService(nil, config.ImportConfig{}, config.LibraryConfig{}, nil, nil, nil, nil, nil)
	h := NewAdminHandler(svc)

	// Set running state
	svc.SetStatusForTest(models.ImportStatus{Status: "running"})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/admin/import", nil)

	h.StartImport(c)

	assert.Equal(t, http.StatusConflict, w.Code)

	var resp map[string]string
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Contains(t, resp["error"], "already running")
}

func TestAdminHandler_ImportStatus(t *testing.T) {
	svc := service.NewImportService(nil, config.ImportConfig{}, config.LibraryConfig{}, nil, nil, nil, nil, nil)
	h := NewAdminHandler(svc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/admin/import/status", nil)

	h.ImportStatus(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.ImportStatus
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "idle", resp.Status)
}
