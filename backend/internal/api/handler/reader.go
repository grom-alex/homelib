package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/grom-alex/homelib/backend/internal/service"
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
	if err := service.ValidateResourceID(chapterID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_chapter", "message": "Некорректный ID главы"})
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
	if err := service.ValidateResourceID(imageID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_image", "message": "Некорректный ID изображения"})
		return
	}

	img, err := h.readerSvc.GetBookImage(c.Request.Context(), id, imageID)
	if err != nil {
		h.handleReaderError(c, err)
		return
	}

	// Only serve image/* content types to prevent Content-Type spoofing
	contentType := img.ContentType
	if !strings.HasPrefix(contentType, "image/") {
		contentType = "application/octet-stream"
	}

	c.Header("Cache-Control", "public, max-age=86400")
	c.Data(http.StatusOK, contentType, img.Data)
}

// handleReaderError maps service errors to HTTP responses per contract.
func (h *ReaderHandler) handleReaderError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrUnsupportedFormat):
		c.JSON(http.StatusUnsupportedMediaType, gin.H{
			"error":   "unsupported_format",
			"message": "Формат книги не поддерживается для чтения в браузере",
		})
	case errors.Is(err, service.ErrBookNotFound):
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "not_found",
			"message": "Книга или запрашиваемый ресурс не найдены",
		})
	case errors.Is(err, service.ErrInvalidResourceID):
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_id",
			"message": "Некорректный идентификатор ресурса",
		})
	case errors.Is(err, service.ErrMalformedFile):
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
