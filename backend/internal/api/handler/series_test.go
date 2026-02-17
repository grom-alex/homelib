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

func TestSeriesHandler_ListSeries_Success(t *testing.T) {
	svc := &mockCatalogService{
		listSeriesFn: func(_ context.Context, query string, page, limit int) ([]models.SeriesListItem, int, error) {
			assert.Equal(t, "ring", query)
			return []models.SeriesListItem{
				{ID: 1, Name: "Lord of the Rings", BooksCount: 3},
			}, 1, nil
		},
	}
	h := NewSeriesHandler(svc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/series?q=ring&page=1&limit=20", nil)

	h.ListSeries(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	items := resp["items"].([]interface{})
	assert.Len(t, items, 1)
}

func TestSeriesHandler_ListSeries_Error(t *testing.T) {
	svc := &mockCatalogService{
		listSeriesFn: func(_ context.Context, _ string, _, _ int) ([]models.SeriesListItem, int, error) {
			return nil, 0, fmt.Errorf("db error")
		},
	}
	h := NewSeriesHandler(svc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/series", nil)

	h.ListSeries(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestSeriesHandler_ListSeries_DefaultParams(t *testing.T) {
	svc := &mockCatalogService{
		listSeriesFn: func(_ context.Context, query string, page, limit int) ([]models.SeriesListItem, int, error) {
			assert.Equal(t, "", query)
			assert.Equal(t, 1, page)
			assert.Equal(t, 20, limit)
			return []models.SeriesListItem{}, 0, nil
		},
	}
	h := NewSeriesHandler(svc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/series", nil)

	h.ListSeries(c)

	assert.Equal(t, http.StatusOK, w.Code)
}
