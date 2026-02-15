package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/grom-alex/homelib/backend/internal/models"
	"github.com/grom-alex/homelib/backend/internal/service"
)

type BooksHandler struct {
	catalogSvc *service.CatalogService
}

func NewBooksHandler(catalogSvc *service.CatalogService) *BooksHandler {
	return &BooksHandler{catalogSvc: catalogSvc}
}

// ListBooks handles GET /api/books.
func (h *BooksHandler) ListBooks(c *gin.Context) {
	var f models.BookFilter
	if err := c.ShouldBindQuery(&f); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid query parameters"})
		return
	}

	books, total, err := h.catalogSvc.ListBooks(c.Request.Context(), f)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list books"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items": books,
		"total": total,
		"page":  f.Page,
		"limit": f.Limit,
	})
}

// GetBook handles GET /api/books/:id.
func (h *BooksHandler) GetBook(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid book id"})
		return
	}

	book, err := h.catalogSvc.GetBook(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "book not found"})
		return
	}

	c.JSON(http.StatusOK, book)
}

// GetStats handles GET /api/stats.
func (h *BooksHandler) GetStats(c *gin.Context) {
	stats, err := h.catalogSvc.GetStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get stats"})
		return
	}

	c.JSON(http.StatusOK, stats)
}
