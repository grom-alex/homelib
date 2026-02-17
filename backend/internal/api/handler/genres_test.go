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

func TestGenresHandler_ListGenres_Success(t *testing.T) {
	svc := &mockCatalogService{
		listGenresFn: func(_ context.Context) ([]models.GenreTreeItem, error) {
			return []models.GenreTreeItem{
				{ID: 1, Code: "sf", Name: "Science Fiction", BooksCount: 100},
				{ID: 2, Code: "det", Name: "Detective", BooksCount: 200},
			}, nil
		},
	}
	h := NewGenresHandler(svc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/genres", nil)

	h.ListGenres(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp []models.GenreTreeItem
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Len(t, resp, 2)
	assert.Equal(t, "Science Fiction", resp[0].Name)
}

func TestGenresHandler_ListGenres_Error(t *testing.T) {
	svc := &mockCatalogService{
		listGenresFn: func(_ context.Context) ([]models.GenreTreeItem, error) {
			return nil, fmt.Errorf("db error")
		},
	}
	h := NewGenresHandler(svc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/genres", nil)

	h.ListGenres(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
