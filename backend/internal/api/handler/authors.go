package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type AuthorsHandler struct {
	catalogSvc CatalogServicer
}

func NewAuthorsHandler(catalogSvc CatalogServicer) *AuthorsHandler {
	return &AuthorsHandler{catalogSvc: catalogSvc}
}

// ListAuthors handles GET /api/authors.
func (h *AuthorsHandler) ListAuthors(c *gin.Context) {
	query := c.Query("q")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	authors, total, err := h.catalogSvc.ListAuthors(c.Request.Context(), query, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list authors"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items": authors,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

// GetAuthor handles GET /api/authors/:id.
func (h *AuthorsHandler) GetAuthor(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid author id"})
		return
	}

	author, err := h.catalogSvc.GetAuthor(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "author not found"})
		return
	}

	c.JSON(http.StatusOK, author)
}
