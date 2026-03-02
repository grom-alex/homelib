package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/grom-alex/homelib/backend/internal/models"
)

type BooksHandler struct {
	catalogSvc         CatalogServicer
	restrictionChecker BookRestrictionChecker
}

func NewBooksHandler(catalogSvc CatalogServicer, restrictionChecker BookRestrictionChecker) *BooksHandler {
	return &BooksHandler{catalogSvc: catalogSvc, restrictionChecker: restrictionChecker}
}

// ListBooks handles GET /api/books.
func (h *BooksHandler) ListBooks(c *gin.Context) {
	var f models.BookFilter
	if err := c.ShouldBindQuery(&f); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid query parameters"})
		return
	}

	// Apply parental content filter
	f.ExcludeGenreIDs = getRestrictedGenreIDs(c)

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

	// Check parental restriction (fail-closed: block on error)
	if restrictedIDs := getRestrictedGenreIDs(c); len(restrictedIDs) > 0 {
		restricted, err := h.restrictionChecker.IsBookRestricted(c.Request.Context(), id, restrictedIDs)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
			return
		}
		if restricted {
			c.JSON(http.StatusForbidden, gin.H{"error": "content_restricted"})
			return
		}
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
