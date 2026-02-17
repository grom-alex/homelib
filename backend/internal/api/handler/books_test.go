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

	"github.com/grom-alex/homelib/backend/internal/models"
	"github.com/grom-alex/homelib/backend/internal/service"
)

func TestBooksHandler_ListBooks_Success(t *testing.T) {
	svc := &mockCatalogService{
		listBooksFn: func(_ context.Context, f models.BookFilter) ([]models.BookListItem, int, error) {
			return []models.BookListItem{
				{ID: 1, Title: "Book 1", Format: "fb2"},
				{ID: 2, Title: "Book 2", Format: "epub"},
			}, 2, nil
		},
	}
	h := NewBooksHandler(svc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/books?page=1&limit=20", nil)

	h.ListBooks(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	items := resp["items"].([]interface{})
	assert.Len(t, items, 2)
	assert.Equal(t, float64(2), resp["total"])
}

func TestBooksHandler_ListBooks_BadParams(t *testing.T) {
	h := NewBooksHandler(nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/books?page=abc", nil)

	h.ListBooks(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestBooksHandler_ListBooks_ServiceError(t *testing.T) {
	svc := &mockCatalogService{
		listBooksFn: func(_ context.Context, _ models.BookFilter) ([]models.BookListItem, int, error) {
			return nil, 0, fmt.Errorf("db error")
		},
	}
	h := NewBooksHandler(svc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/books", nil)

	h.ListBooks(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestBooksHandler_GetBook_Success(t *testing.T) {
	svc := &mockCatalogService{
		getBookFn: func(_ context.Context, id int64) (*models.BookDetail, error) {
			assert.Equal(t, int64(42), id)
			return &models.BookDetail{ID: 42, Title: "Test Book"}, nil
		},
	}
	h := NewBooksHandler(svc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/books/42", nil)
	c.Params = gin.Params{{Key: "id", Value: "42"}}

	h.GetBook(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp models.BookDetail
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, int64(42), resp.ID)
	assert.Equal(t, "Test Book", resp.Title)
}

func TestBooksHandler_GetBook_InvalidID(t *testing.T) {
	h := NewBooksHandler(nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/books/abc", nil)
	c.Params = gin.Params{{Key: "id", Value: "abc"}}

	h.GetBook(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp map[string]string
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "invalid book id", resp["error"])
}

func TestBooksHandler_GetBook_NotFound(t *testing.T) {
	svc := &mockCatalogService{
		getBookFn: func(_ context.Context, _ int64) (*models.BookDetail, error) {
			return nil, fmt.Errorf("not found")
		},
	}
	h := NewBooksHandler(svc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/books/999", nil)
	c.Params = gin.Params{{Key: "id", Value: "999"}}

	h.GetBook(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestBooksHandler_GetStats_Success(t *testing.T) {
	svc := &mockCatalogService{
		getStatsFn: func(_ context.Context) (*service.Stats, error) {
			return &service.Stats{
				BooksCount:   1000,
				AuthorsCount: 500,
				GenresCount:  30,
				SeriesCount:  100,
				Languages:    []string{"en", "ru"},
				Formats:      []string{"epub", "fb2"},
			}, nil
		},
	}
	h := NewBooksHandler(svc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/stats", nil)

	h.GetStats(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp service.Stats
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, 1000, resp.BooksCount)
}

func TestBooksHandler_GetStats_Error(t *testing.T) {
	svc := &mockCatalogService{
		getStatsFn: func(_ context.Context) (*service.Stats, error) {
			return nil, fmt.Errorf("db error")
		},
	}
	h := NewBooksHandler(svc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/stats", nil)

	h.GetStats(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
