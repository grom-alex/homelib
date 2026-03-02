package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type GenresHandler struct {
	catalogSvc CatalogServicer
}

func NewGenresHandler(catalogSvc CatalogServicer) *GenresHandler {
	return &GenresHandler{catalogSvc: catalogSvc}
}

// ListGenres handles GET /api/genres.
func (h *GenresHandler) ListGenres(c *gin.Context) {
	excludeIDs := getRestrictedGenreIDs(c)

	genres, err := h.catalogSvc.ListGenres(c.Request.Context(), excludeIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list genres"})
		return
	}

	c.JSON(http.StatusOK, genres)
}
