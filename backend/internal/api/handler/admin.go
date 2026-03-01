package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/grom-alex/homelib/backend/internal/service"
)

type AdminHandler struct {
	importSvc    ImportServicer
	genreTreeSvc GenreTreeServicer
}

func NewAdminHandler(importSvc ImportServicer, genreTreeSvc GenreTreeServicer) *AdminHandler {
	return &AdminHandler{importSvc: importSvc, genreTreeSvc: genreTreeSvc}
}

// StartImport handles POST /api/admin/import.
func (h *AdminHandler) StartImport(c *gin.Context) {
	if err := h.importSvc.StartImport(); err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusAccepted, gin.H{"message": "import started"})
}

// ImportStatus handles GET /api/admin/import/status.
func (h *AdminHandler) ImportStatus(c *gin.Context) {
	status := h.importSvc.GetStatus()
	c.JSON(http.StatusOK, status)
}

// CancelImport handles POST /api/admin/import/cancel.
func (h *AdminHandler) CancelImport(c *gin.Context) {
	h.importSvc.CancelImport()
	c.JSON(http.StatusOK, gin.H{"message": "import cancellation requested"})
}

// ReloadGenres handles POST /api/admin/genres/reload.
func (h *AdminHandler) ReloadGenres(c *gin.Context) {
	result, err := h.genreTreeSvc.ForceReload(c.Request.Context())
	if err != nil {
		if errors.Is(err, service.ErrGenreReloadAlreadyRunning) {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "genre reload failed"})
		return
	}
	c.JSON(http.StatusOK, result)
}
