package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/grom-alex/homelib/backend/internal/models"
	"github.com/grom-alex/homelib/backend/internal/service"
)

type ProgressHandler struct {
	progressRepo ProgressRepository
}

func NewProgressHandler(repo ProgressRepository) *ProgressHandler {
	return &ProgressHandler{progressRepo: repo}
}

// GetReadingProgress handles GET /api/me/books/:bookId/progress.
func (h *ProgressHandler) GetReadingProgress(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "Пользователь не авторизован"})
		return
	}

	bookID, err := strconv.ParseInt(c.Param("bookId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_id", "message": "Некорректный ID книги"})
		return
	}

	progress, err := h.progressRepo.Get(c.Request.Context(), userID, bookID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Внутренняя ошибка сервера"})
		return
	}

	if progress == nil {
		c.Status(http.StatusNoContent)
		c.Writer.WriteHeaderNow()
		return
	}

	c.JSON(http.StatusOK, progress)
}

// SaveReadingProgress handles PUT /api/me/books/:bookId/progress.
func (h *ProgressHandler) SaveReadingProgress(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "Пользователь не авторизован"})
		return
	}

	bookID, err := strconv.ParseInt(c.Param("bookId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_id", "message": "Некорректный ID книги"})
		return
	}

	var input models.SaveProgressInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "validation_error", "message": "Невалидные данные прогресса чтения"})
		return
	}

	// Validate chapter ID format (safe for filesystem use in caching)
	if err := service.ValidateResourceID(input.ChapterID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_chapter", "message": "Некорректный ID главы"})
		return
	}

	progress := &models.ReadingProgress{
		UserID:          userID,
		BookID:          bookID,
		ChapterID:       input.ChapterID,
		ChapterProgress: input.ChapterProgress,
		TotalProgress:   input.TotalProgress,
		Device:          input.Device,
	}

	if err := h.progressRepo.Upsert(c.Request.Context(), progress); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Внутренняя ошибка сервера"})
		return
	}

	c.JSON(http.StatusOK, progress)
}

// GetAllProgress handles GET /api/me/progress.
// Returns a compact map of bookID → totalProgress for the authenticated user.
func (h *ProgressHandler) GetAllProgress(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "Пользователь не авторизован"})
		return
	}

	items, err := h.progressRepo.GetByUser(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Внутренняя ошибка сервера"})
		return
	}

	result := make(map[int64]int, len(items))
	for _, p := range items {
		result[p.BookID] = p.TotalProgress
	}

	c.JSON(http.StatusOK, result)
}
