package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/grom-alex/homelib/backend/internal/bookfile"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// --- GetBookContent ---

func TestReaderHandler_GetBookContent_Success(t *testing.T) {
	svc := &mockReaderService{
		getBookContentFn: func(_ context.Context, bookID int64) (*bookfile.BookContent, error) {
			assert.Equal(t, int64(42), bookID)
			return &bookfile.BookContent{
				Metadata: bookfile.BookMetadata{
					Title:    "Test Book",
					Author:   "Test Author",
					Language: "ru",
					Format:   "fb2",
				},
				TOC: []bookfile.TOCEntry{
					{ID: "ch1", Title: "Chapter 1", Level: 0},
				},
				ChapterIDs:    []string{"ch1"},
				TotalChapters: 1,
			}, nil
		},
	}
	h := NewReaderHandler(svc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/books/42/content", nil)
	c.Params = gin.Params{{Key: "id", Value: "42"}}

	h.GetBookContent(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp bookfile.BookContent
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "Test Book", resp.Metadata.Title)
	assert.Equal(t, "Test Author", resp.Metadata.Author)
	assert.Equal(t, 1, resp.TotalChapters)
	assert.Equal(t, []string{"ch1"}, resp.ChapterIDs)
}

func TestReaderHandler_GetBookContent_InvalidID(t *testing.T) {
	h := NewReaderHandler(&mockReaderService{})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/books/abc/content", nil)
	c.Params = gin.Params{{Key: "id", Value: "abc"}}

	h.GetBookContent(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestReaderHandler_GetBookContent_NotFound(t *testing.T) {
	svc := &mockReaderService{
		getBookContentFn: func(_ context.Context, _ int64) (*bookfile.BookContent, error) {
			return nil, fmt.Errorf("book not found: no rows")
		},
	}
	h := NewReaderHandler(svc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/books/999/content", nil)
	c.Params = gin.Params{{Key: "id", Value: "999"}}

	h.GetBookContent(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "not_found")
}

func TestReaderHandler_GetBookContent_UnsupportedFormat(t *testing.T) {
	svc := &mockReaderService{
		getBookContentFn: func(_ context.Context, _ int64) (*bookfile.BookContent, error) {
			return nil, fmt.Errorf("unsupported format: epub")
		},
	}
	h := NewReaderHandler(svc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/books/1/content", nil)
	c.Params = gin.Params{{Key: "id", Value: "1"}}

	h.GetBookContent(c)

	assert.Equal(t, http.StatusUnsupportedMediaType, w.Code)
	assert.Contains(t, w.Body.String(), "unsupported_format")
}

func TestReaderHandler_GetBookContent_MalformedFB2(t *testing.T) {
	svc := &mockReaderService{
		getBookContentFn: func(_ context.Context, _ int64) (*bookfile.BookContent, error) {
			return nil, fmt.Errorf("parse book: XML syntax error")
		},
	}
	h := NewReaderHandler(svc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/books/1/content", nil)
	c.Params = gin.Params{{Key: "id", Value: "1"}}

	h.GetBookContent(c)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	assert.Contains(t, w.Body.String(), "malformed_file")
}

// --- GetChapter ---

func TestReaderHandler_GetChapter_Success(t *testing.T) {
	svc := &mockReaderService{
		getChapterFn: func(_ context.Context, bookID int64, chapterID string) (*bookfile.ChapterContent, error) {
			assert.Equal(t, int64(42), bookID)
			assert.Equal(t, "ch1", chapterID)
			return &bookfile.ChapterContent{
				ID:    "ch1",
				Title: "Chapter 1",
				HTML:  `<h2 class="chapter-title">Chapter 1</h2><p>Hello</p>`,
			}, nil
		},
	}
	h := NewReaderHandler(svc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/books/42/chapter/ch1", nil)
	c.Params = gin.Params{{Key: "id", Value: "42"}, {Key: "chapterId", Value: "ch1"}}

	h.GetChapter(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp bookfile.ChapterContent
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "ch1", resp.ID)
	assert.Equal(t, "Chapter 1", resp.Title)
	assert.Contains(t, resp.HTML, "chapter-title")
}

func TestReaderHandler_GetChapter_InvalidBookID(t *testing.T) {
	h := NewReaderHandler(&mockReaderService{})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/books/abc/chapter/ch1", nil)
	c.Params = gin.Params{{Key: "id", Value: "abc"}, {Key: "chapterId", Value: "ch1"}}

	h.GetChapter(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestReaderHandler_GetChapter_ChapterNotFound(t *testing.T) {
	svc := &mockReaderService{
		getChapterFn: func(_ context.Context, _ int64, _ string) (*bookfile.ChapterContent, error) {
			return nil, fmt.Errorf("chapter \"nonexistent\": chapter \"nonexistent\" not found")
		},
	}
	h := NewReaderHandler(svc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/books/1/chapter/nonexistent", nil)
	c.Params = gin.Params{{Key: "id", Value: "1"}, {Key: "chapterId", Value: "nonexistent"}}

	h.GetChapter(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// --- GetBookImage ---

func TestReaderHandler_GetBookImage_Success(t *testing.T) {
	imgData := []byte{0x89, 'P', 'N', 'G', 0x0d, 0x0a, 0x1a, 0x0a}
	svc := &mockReaderService{
		getBookImageFn: func(_ context.Context, bookID int64, imageID string) (*bookfile.ImageData, error) {
			assert.Equal(t, int64(42), bookID)
			assert.Equal(t, "cover.jpg", imageID)
			return &bookfile.ImageData{
				ID:          "cover.jpg",
				ContentType: "image/jpeg",
				Data:        imgData,
			}, nil
		},
	}
	h := NewReaderHandler(svc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/books/42/image/cover.jpg", nil)
	c.Params = gin.Params{{Key: "id", Value: "42"}, {Key: "imageId", Value: "cover.jpg"}}

	h.GetBookImage(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "image/jpeg", w.Header().Get("Content-Type"))
	assert.Equal(t, "public, max-age=86400", w.Header().Get("Cache-Control"))
	assert.Equal(t, imgData, w.Body.Bytes())
}

func TestReaderHandler_GetBookImage_PNG(t *testing.T) {
	svc := &mockReaderService{
		getBookImageFn: func(_ context.Context, _ int64, _ string) (*bookfile.ImageData, error) {
			return &bookfile.ImageData{
				ID:          "img1.png",
				ContentType: "image/png",
				Data:        []byte{0x89, 'P', 'N', 'G'},
			}, nil
		},
	}
	h := NewReaderHandler(svc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/books/1/image/img1.png", nil)
	c.Params = gin.Params{{Key: "id", Value: "1"}, {Key: "imageId", Value: "img1.png"}}

	h.GetBookImage(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "image/png", w.Header().Get("Content-Type"))
}

func TestReaderHandler_GetBookImage_InvalidID(t *testing.T) {
	h := NewReaderHandler(&mockReaderService{})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/books/abc/image/img.png", nil)
	c.Params = gin.Params{{Key: "id", Value: "abc"}, {Key: "imageId", Value: "img.png"}}

	h.GetBookImage(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestReaderHandler_GetBookImage_NotFound(t *testing.T) {
	svc := &mockReaderService{
		getBookImageFn: func(_ context.Context, _ int64, _ string) (*bookfile.ImageData, error) {
			return nil, fmt.Errorf("image \"nonexistent\": image \"nonexistent\" not found")
		},
	}
	h := NewReaderHandler(svc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/books/1/image/nonexistent", nil)
	c.Params = gin.Params{{Key: "id", Value: "1"}, {Key: "imageId", Value: "nonexistent"}}

	h.GetBookImage(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// --- Error mapping ---

func TestReaderHandler_InternalError(t *testing.T) {
	svc := &mockReaderService{
		getBookContentFn: func(_ context.Context, _ int64) (*bookfile.BookContent, error) {
			return nil, fmt.Errorf("unexpected disk error")
		},
	}
	h := NewReaderHandler(svc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/books/1/content", nil)
	c.Params = gin.Params{{Key: "id", Value: "1"}}

	h.GetBookContent(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "internal_error")
}
