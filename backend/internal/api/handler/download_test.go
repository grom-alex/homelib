package handler

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/grom-alex/homelib/backend/internal/service"
)

func TestDownloadHandler_DownloadBook_Success(t *testing.T) {
	content := "fake book content"
	svc := &mockDownloadService{
		downloadBookFn: func(_ context.Context, id int64) (*service.DownloadResult, error) {
			assert.Equal(t, int64(42), id)
			return &service.DownloadResult{
				Reader:      nopReadCloser{strings.NewReader(content)},
				Filename:    "book.fb2",
				ContentType: "application/xml",
				Size:        int64(len(content)),
			}, nil
		},
	}
	h := NewDownloadHandler(svc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/books/42/download", nil)
	c.Params = gin.Params{{Key: "id", Value: "42"}}

	h.DownloadBook(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Header().Get("Content-Disposition"), "book.fb2")
	assert.Equal(t, "application/xml", w.Header().Get("Content-Type"))
	assert.Equal(t, content, w.Body.String())
}

func TestDownloadHandler_DownloadBook_InvalidID(t *testing.T) {
	h := NewDownloadHandler(&mockDownloadService{})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/books/abc/download", nil)
	c.Params = gin.Params{{Key: "id", Value: "abc"}}

	h.DownloadBook(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDownloadHandler_DownloadBook_NotFound(t *testing.T) {
	svc := &mockDownloadService{
		downloadBookFn: func(_ context.Context, _ int64) (*service.DownloadResult, error) {
			return nil, fmt.Errorf("book not found")
		},
	}
	h := NewDownloadHandler(svc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/books/999/download", nil)
	c.Params = gin.Params{{Key: "id", Value: "999"}}

	h.DownloadBook(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestDownloadHandler_DownloadBook_ZeroSize(t *testing.T) {
	svc := &mockDownloadService{
		downloadBookFn: func(_ context.Context, _ int64) (*service.DownloadResult, error) {
			return &service.DownloadResult{
				Reader:      nopReadCloser{strings.NewReader("")},
				Filename:    "book.epub",
				ContentType: "application/epub+zip",
				Size:        0,
			}, nil
		},
	}
	h := NewDownloadHandler(svc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/books/1/download", nil)
	c.Params = gin.Params{{Key: "id", Value: "1"}}

	h.DownloadBook(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Empty(t, w.Header().Get("Content-Length"))
}
