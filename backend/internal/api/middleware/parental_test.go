package middleware

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
)

// --- mock parental service for middleware ---

type mockParentalSvc struct {
	isAdultContentEnabledFn func(ctx context.Context, userID string) (bool, error)
	getRestrictedGenreIDsFn func(ctx context.Context) ([]int, error)
}

func (m *mockParentalSvc) IsAdultContentEnabled(ctx context.Context, userID string) (bool, error) {
	if m.isAdultContentEnabledFn != nil {
		return m.isAdultContentEnabledFn(ctx, userID)
	}
	return false, nil
}

func (m *mockParentalSvc) GetRestrictedGenreIDs(ctx context.Context) ([]int, error) {
	if m.getRestrictedGenreIDsFn != nil {
		return m.getRestrictedGenreIDsFn(ctx)
	}
	return nil, nil
}

// --- Tests ---

func TestParentalFilter_NoUserID(t *testing.T) {
	svc := &mockParentalSvc{}
	mw := ParentalFilter(svc)

	w := httptest.NewRecorder()
	c, r := gin.CreateTestContext(w)
	handlerCalled := false
	r.GET("/test", mw, func(c *gin.Context) {
		handlerCalled = true
		c.JSON(200, gin.H{"ok": true})
	})

	c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)
	r.ServeHTTP(w, c.Request)

	assert.True(t, handlerCalled)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestParentalFilter_AdultEnabled(t *testing.T) {
	svc := &mockParentalSvc{
		isAdultContentEnabledFn: func(_ context.Context, _ string) (bool, error) {
			return true, nil
		},
	}
	mw := ParentalFilter(svc)

	w := httptest.NewRecorder()
	c, r := gin.CreateTestContext(w)
	r.GET("/test", func(c *gin.Context) {
		c.Set("user_id", "user-1")
		c.Next()
	}, mw, func(c *gin.Context) {
		_, exists := c.Get("restricted_genre_ids")
		c.JSON(200, gin.H{"restricted": exists})
	})

	c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)
	r.ServeHTTP(w, c.Request)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]bool
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.False(t, resp["restricted"])
}

func TestParentalFilter_AdultDisabled_WithRestrictions(t *testing.T) {
	svc := &mockParentalSvc{
		isAdultContentEnabledFn: func(_ context.Context, _ string) (bool, error) {
			return false, nil
		},
		getRestrictedGenreIDsFn: func(_ context.Context) ([]int, error) {
			return []int{10, 20, 30}, nil
		},
	}
	mw := ParentalFilter(svc)

	w := httptest.NewRecorder()
	c, r := gin.CreateTestContext(w)
	r.GET("/test", func(c *gin.Context) {
		c.Set("user_id", "user-1")
		c.Next()
	}, mw, func(c *gin.Context) {
		ids, _ := c.Get("restricted_genre_ids")
		c.JSON(200, gin.H{"ids": ids})
	})

	c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)
	r.ServeHTTP(w, c.Request)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string][]int
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, []int{10, 20, 30}, resp["ids"])
}

func TestParentalFilter_AdultDisabled_NoRestrictions(t *testing.T) {
	svc := &mockParentalSvc{
		isAdultContentEnabledFn: func(_ context.Context, _ string) (bool, error) {
			return false, nil
		},
		getRestrictedGenreIDsFn: func(_ context.Context) ([]int, error) {
			return nil, nil
		},
	}
	mw := ParentalFilter(svc)

	w := httptest.NewRecorder()
	c, r := gin.CreateTestContext(w)
	r.GET("/test", func(c *gin.Context) {
		c.Set("user_id", "user-1")
		c.Next()
	}, mw, func(c *gin.Context) {
		_, exists := c.Get("restricted_genre_ids")
		c.JSON(200, gin.H{"restricted": exists})
	})

	c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)
	r.ServeHTTP(w, c.Request)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]bool
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.False(t, resp["restricted"])
}

func TestParentalFilter_AdultCheckError_FailClosed(t *testing.T) {
	svc := &mockParentalSvc{
		isAdultContentEnabledFn: func(_ context.Context, _ string) (bool, error) {
			return false, fmt.Errorf("db error")
		},
		getRestrictedGenreIDsFn: func(_ context.Context) ([]int, error) {
			return []int{10}, nil
		},
	}
	mw := ParentalFilter(svc)

	w := httptest.NewRecorder()
	c, r := gin.CreateTestContext(w)
	r.GET("/test", func(c *gin.Context) {
		c.Set("user_id", "user-1")
		c.Next()
	}, mw, func(c *gin.Context) {
		ids, _ := c.Get("restricted_genre_ids")
		c.JSON(200, gin.H{"ids": ids})
	})

	c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)
	r.ServeHTTP(w, c.Request)

	// Should still proceed with restrictions (fail-closed: assume not enabled)
	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string][]int
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, []int{10}, resp["ids"])
}

func TestParentalFilter_GetRestrictedIDsError_Aborts(t *testing.T) {
	svc := &mockParentalSvc{
		isAdultContentEnabledFn: func(_ context.Context, _ string) (bool, error) {
			return false, nil
		},
		getRestrictedGenreIDsFn: func(_ context.Context) ([]int, error) {
			return nil, fmt.Errorf("db error")
		},
	}
	mw := ParentalFilter(svc)

	w := httptest.NewRecorder()
	c, r := gin.CreateTestContext(w)
	handlerCalled := false
	r.GET("/test", func(c *gin.Context) {
		c.Set("user_id", "user-1")
		c.Next()
	}, mw, func(c *gin.Context) {
		handlerCalled = true
		c.JSON(200, gin.H{"ok": true})
	})

	c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)
	r.ServeHTTP(w, c.Request)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	assert.False(t, handlerCalled, "downstream handler should not be called on restriction error")
}
