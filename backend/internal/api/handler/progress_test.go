package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/grom-alex/homelib/backend/internal/models"
)

// --- GetReadingProgress ---

func TestProgressHandler_GetReadingProgress_Found(t *testing.T) {
	now := time.Date(2026, 2, 18, 14, 30, 0, 0, time.UTC)
	repo := &mockProgressRepo{
		getFn: func(_ context.Context, userID string, bookID int64) (*models.ReadingProgress, error) {
			assert.Equal(t, "user-123", userID)
			assert.Equal(t, int64(42), bookID)
			return &models.ReadingProgress{
				ChapterID:       "ch3",
				ChapterProgress: 55,
				TotalProgress:   30,
				Device:          "desktop",
				UpdatedAt:       now,
			}, nil
		},
	}
	h := NewProgressHandler(repo)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/me/books/42/progress", nil)
	c.Params = gin.Params{{Key: "bookId", Value: "42"}}
	c.Set("user_id", "user-123")

	h.GetReadingProgress(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.ReadingProgress
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "ch3", resp.ChapterID)
	assert.Equal(t, 55, resp.ChapterProgress)
	assert.Equal(t, 30, resp.TotalProgress)
	assert.Equal(t, "desktop", resp.Device)
}

func TestProgressHandler_GetReadingProgress_NoContent(t *testing.T) {
	repo := &mockProgressRepo{
		getFn: func(_ context.Context, _ string, _ int64) (*models.ReadingProgress, error) {
			return nil, nil
		},
	}
	h := NewProgressHandler(repo)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/me/books/42/progress", nil)
	c.Params = gin.Params{{Key: "bookId", Value: "42"}}
	c.Set("user_id", "user-123")

	h.GetReadingProgress(c)

	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Empty(t, w.Body.String())
}

func TestProgressHandler_GetReadingProgress_InvalidBookID(t *testing.T) {
	h := NewProgressHandler(&mockProgressRepo{})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/me/books/abc/progress", nil)
	c.Params = gin.Params{{Key: "bookId", Value: "abc"}}
	c.Set("user_id", "user-123")

	h.GetReadingProgress(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "invalid_id")
}

func TestProgressHandler_GetReadingProgress_NoUserID(t *testing.T) {
	h := NewProgressHandler(&mockProgressRepo{})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/me/books/42/progress", nil)
	c.Params = gin.Params{{Key: "bookId", Value: "42"}}

	h.GetReadingProgress(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestProgressHandler_GetReadingProgress_DBError(t *testing.T) {
	repo := &mockProgressRepo{
		getFn: func(_ context.Context, _ string, _ int64) (*models.ReadingProgress, error) {
			return nil, fmt.Errorf("connection lost")
		},
	}
	h := NewProgressHandler(repo)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/me/books/42/progress", nil)
	c.Params = gin.Params{{Key: "bookId", Value: "42"}}
	c.Set("user_id", "user-123")

	h.GetReadingProgress(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "internal_error")
}

// --- SaveReadingProgress ---

func TestProgressHandler_SaveReadingProgress_Success(t *testing.T) {
	repo := &mockProgressRepo{
		upsertFn: func(_ context.Context, p *models.ReadingProgress) error {
			assert.Equal(t, "user-123", p.UserID)
			assert.Equal(t, int64(42), p.BookID)
			assert.Equal(t, "ch5", p.ChapterID)
			assert.Equal(t, 75, p.ChapterProgress)
			assert.Equal(t, 50, p.TotalProgress)
			assert.Equal(t, "mobile", p.Device)
			p.ID = 1
			p.UpdatedAt = time.Date(2026, 2, 18, 15, 0, 0, 0, time.UTC)
			return nil
		},
	}
	h := NewProgressHandler(repo)

	body := `{"chapterId":"ch5","chapterProgress":75,"totalProgress":50,"device":"mobile"}`
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPut, "/api/me/books/42/progress", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "bookId", Value: "42"}}
	c.Set("user_id", "user-123")

	h.SaveReadingProgress(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.ReadingProgress
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "ch5", resp.ChapterID)
	assert.Equal(t, 75, resp.ChapterProgress)
	assert.Equal(t, 50, resp.TotalProgress)
}

func TestProgressHandler_SaveReadingProgress_EmptyChapterID(t *testing.T) {
	h := NewProgressHandler(&mockProgressRepo{})

	body := `{"chapterId":"","chapterProgress":50,"totalProgress":25}`
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPut, "/api/me/books/42/progress", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "bookId", Value: "42"}}
	c.Set("user_id", "user-123")

	h.SaveReadingProgress(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "validation_error")
}

func TestProgressHandler_SaveReadingProgress_ProgressOutOfRange(t *testing.T) {
	h := NewProgressHandler(&mockProgressRepo{})

	body := `{"chapterId":"ch1","chapterProgress":150,"totalProgress":50}`
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPut, "/api/me/books/42/progress", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "bookId", Value: "42"}}
	c.Set("user_id", "user-123")

	h.SaveReadingProgress(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "validation_error")
}

func TestProgressHandler_SaveReadingProgress_NegativeProgress(t *testing.T) {
	h := NewProgressHandler(&mockProgressRepo{})

	body := `{"chapterId":"ch1","chapterProgress":-5,"totalProgress":50}`
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPut, "/api/me/books/42/progress", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "bookId", Value: "42"}}
	c.Set("user_id", "user-123")

	h.SaveReadingProgress(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestProgressHandler_SaveReadingProgress_InvalidJSON(t *testing.T) {
	h := NewProgressHandler(&mockProgressRepo{})

	body := `not json`
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPut, "/api/me/books/42/progress", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "bookId", Value: "42"}}
	c.Set("user_id", "user-123")

	h.SaveReadingProgress(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestProgressHandler_SaveReadingProgress_InvalidBookID(t *testing.T) {
	h := NewProgressHandler(&mockProgressRepo{})

	body := `{"chapterId":"ch1","chapterProgress":50,"totalProgress":25}`
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPut, "/api/me/books/42/progress", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "bookId", Value: "abc"}}
	c.Set("user_id", "user-123")

	h.SaveReadingProgress(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestProgressHandler_SaveReadingProgress_NoUserID(t *testing.T) {
	h := NewProgressHandler(&mockProgressRepo{})

	body := `{"chapterId":"ch1","chapterProgress":50,"totalProgress":25}`
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPut, "/api/me/books/42/progress", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "bookId", Value: "42"}}

	h.SaveReadingProgress(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestProgressHandler_SaveReadingProgress_DBError(t *testing.T) {
	repo := &mockProgressRepo{
		upsertFn: func(_ context.Context, _ *models.ReadingProgress) error {
			return fmt.Errorf("disk full")
		},
	}
	h := NewProgressHandler(repo)

	body := `{"chapterId":"ch1","chapterProgress":50,"totalProgress":25}`
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPut, "/api/me/books/42/progress", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "bookId", Value: "42"}}
	c.Set("user_id", "user-123")

	h.SaveReadingProgress(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "internal_error")
}
