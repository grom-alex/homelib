package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/grom-alex/homelib/backend/internal/service"
)

type SeriesHandler struct {
	catalogSvc *service.CatalogService
}

func NewSeriesHandler(catalogSvc *service.CatalogService) *SeriesHandler {
	return &SeriesHandler{catalogSvc: catalogSvc}
}

// ListSeries handles GET /api/series.
func (h *SeriesHandler) ListSeries(c *gin.Context) {
	query := c.Query("q")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	series, total, err := h.catalogSvc.ListSeries(c.Request.Context(), query, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list series"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items": series,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}
