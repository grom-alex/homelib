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
)

func TestAuthorsHandler_ListAuthors_Success(t *testing.T) {
	svc := &mockCatalogService{
		listAuthorsFn: func(_ context.Context, query string, page, limit int) ([]models.AuthorListItem, int, error) {
			assert.Equal(t, "tolkien", query)
			return []models.AuthorListItem{
				{ID: 1, Name: "J.R.R. Tolkien", BooksCount: 12},
			}, 1, nil
		},
	}
	h := NewAuthorsHandler(svc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/authors?q=tolkien&page=1&limit=20", nil)

	h.ListAuthors(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	items := resp["items"].([]interface{})
	assert.Len(t, items, 1)
	assert.Equal(t, float64(1), resp["total"])
}

func TestAuthorsHandler_ListAuthors_Error(t *testing.T) {
	svc := &mockCatalogService{
		listAuthorsFn: func(_ context.Context, _ string, _, _ int) ([]models.AuthorListItem, int, error) {
			return nil, 0, fmt.Errorf("db error")
		},
	}
	h := NewAuthorsHandler(svc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/authors", nil)

	h.ListAuthors(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestAuthorsHandler_GetAuthor_Success(t *testing.T) {
	svc := &mockCatalogService{
		getAuthorFn: func(_ context.Context, id int64) (*models.AuthorDetail, error) {
			assert.Equal(t, int64(1), id)
			return &models.AuthorDetail{
				ID:         1,
				Name:       "J.R.R. Tolkien",
				BooksCount: 12,
				Books:      []models.BookListItem{},
			}, nil
		},
	}
	h := NewAuthorsHandler(svc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/authors/1", nil)
	c.Params = gin.Params{{Key: "id", Value: "1"}}

	h.GetAuthor(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp models.AuthorDetail
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "J.R.R. Tolkien", resp.Name)
}

func TestAuthorsHandler_GetAuthor_InvalidID(t *testing.T) {
	h := NewAuthorsHandler(nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/authors/xyz", nil)
	c.Params = gin.Params{{Key: "id", Value: "xyz"}}

	h.GetAuthor(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp map[string]string
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "invalid author id", resp["error"])
}

func TestAuthorsHandler_GetAuthor_NotFound(t *testing.T) {
	svc := &mockCatalogService{
		getAuthorFn: func(_ context.Context, _ int64) (*models.AuthorDetail, error) {
			return nil, fmt.Errorf("not found")
		},
	}
	h := NewAuthorsHandler(svc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/authors/999", nil)
	c.Params = gin.Params{{Key: "id", Value: "999"}}

	h.GetAuthor(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
}
