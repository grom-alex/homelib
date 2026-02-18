package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type ReaderHandler struct {
	readerSvc ReaderServicer
}

func NewReaderHandler(readerSvc ReaderServicer) *ReaderHandler {
	return &ReaderHandler{readerSvc: readerSvc}
}

// GetBookContent handles GET /api/books/:id/content.
func (h *ReaderHandler) GetBookContent(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_id", "message": "Некорректный ID книги"})
		return
	}

	content, err := h.readerSvc.GetBookContent(c.Request.Context(), id)
	if err != nil {
		h.handleReaderError(c, err)
		return
	}

	c.JSON(http.StatusOK, content)
}

// GetChapter handles GET /api/books/:id/chapter/:chapterId.
func (h *ReaderHandler) GetChapter(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_id", "message": "Некорректный ID книги"})
		return
	}

	chapterID := c.Param("chapterId")
	if chapterID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_chapter", "message": "ID главы не указан"})
		return
	}

	ch, err := h.readerSvc.GetChapter(c.Request.Context(), id, chapterID)
	if err != nil {
		h.handleReaderError(c, err)
		return
	}

	c.JSON(http.StatusOK, ch)
}

// GetBookImage handles GET /api/books/:id/image/:imageId.
func (h *ReaderHandler) GetBookImage(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_id", "message": "Некорректный ID книги"})
		return
	}

	imageID := c.Param("imageId")
	if imageID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_image", "message": "ID изображения не указан"})
		return
	}

	img, err := h.readerSvc.GetBookImage(c.Request.Context(), id, imageID)
	if err != nil {
		h.handleReaderError(c, err)
		return
	}

	c.Header("Cache-Control", "public, max-age=86400")
	c.Data(http.StatusOK, img.ContentType, img.Data)
}

// handleReaderError maps service errors to HTTP responses per contract.
func (h *ReaderHandler) handleReaderError(c *gin.Context, err error) {
	msg := err.Error()

	switch {
	case strings.Contains(msg, "unsupported format"):
		c.JSON(http.StatusUnsupportedMediaType, gin.H{
			"error":   "unsupported_format",
			"message": "Формат книги не поддерживается для чтения в браузере",
		})
	case strings.Contains(msg, "not found"):
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "not_found",
			"message": "Книга или запрашиваемый ресурс не найдены",
		})
	case strings.Contains(msg, "parse book"):
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"error":   "malformed_file",
			"message": "Файл книги повреждён или имеет некорректный формат",
		})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "Внутренняя ошибка сервера",
		})
	}
}
