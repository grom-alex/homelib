package handler

import (
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/grom-alex/homelib/backend/internal/service"
)

type DownloadHandler struct {
	downloadSvc *service.DownloadService
}

func NewDownloadHandler(downloadSvc *service.DownloadService) *DownloadHandler {
	return &DownloadHandler{downloadSvc: downloadSvc}
}

// DownloadBook handles GET /api/books/:id/download.
func (h *DownloadHandler) DownloadBook(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid book id"})
		return
	}

	result, err := h.downloadSvc.DownloadBook(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "book not found or file unavailable"})
		return
	}
	defer func() { _ = result.Reader.Close() }()

	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, result.Filename))
	c.Header("Content-Type", result.ContentType)
	if result.Size > 0 {
		c.Header("Content-Length", strconv.FormatInt(result.Size, 10))
	}

	c.Status(http.StatusOK)
	c.Stream(func(w io.Writer) bool {
		_, err := io.Copy(w, result.Reader)
		return err == nil
	})
}
