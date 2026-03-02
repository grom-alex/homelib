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
		listGenresFn: func(_ context.Context, _ []int) ([]models.GenreTreeItem, error) {
			return []models.GenreTreeItem{
				{
					ID: 1, Code: "sf_all", Name: "Фантастика", Position: "0.1", BooksCount: 350,
					Children: []models.GenreTreeItem{
						{ID: 3, Code: "sf_history", Name: "Альтернативная история", Position: "0.1.1", BooksCount: 150},
						{ID: 4, Code: "sf_action", Name: "Боевая фантастика", Position: "0.1.2", BooksCount: 200},
					},
				},
				{ID: 2, Code: "det_all", Name: "Детективы", Position: "0.2", BooksCount: 200},
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
	assert.Equal(t, "Фантастика", resp[0].Name)
	assert.Equal(t, "0.1", resp[0].Position)
	assert.Equal(t, 350, resp[0].BooksCount)
	assert.Len(t, resp[0].Children, 2)
}

func TestGenresHandler_ListGenres_Error(t *testing.T) {
	svc := &mockCatalogService{
		listGenresFn: func(_ context.Context, _ []int) ([]models.GenreTreeItem, error) {
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
