package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	gin.SetMode(gin.TestMode)
}

type mockValidator struct {
	claims *Claims
	err    error
}

func (m *mockValidator) ValidateToken(token string) (*Claims, error) {
	return m.claims, m.err
}

func TestRequireAuth_NoHeader(t *testing.T) {
	mw := NewAuthMiddleware(&mockValidator{})
	w := httptest.NewRecorder()
	c, r := gin.CreateTestContext(w)

	r.GET("/test", mw.RequireAuth(), func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)
	r.ServeHTTP(w, c.Request)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestRequireAuth_InvalidHeader(t *testing.T) {
	mw := NewAuthMiddleware(&mockValidator{})
	w := httptest.NewRecorder()
	c, r := gin.CreateTestContext(w)

	r.GET("/test", mw.RequireAuth(), func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)
	c.Request.Header.Set("Authorization", "Basic abc123")
	r.ServeHTTP(w, c.Request)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestRequireAuth_InvalidToken(t *testing.T) {
	mw := NewAuthMiddleware(&mockValidator{err: fmt.Errorf("invalid")})
	w := httptest.NewRecorder()
	c, r := gin.CreateTestContext(w)

	r.GET("/test", mw.RequireAuth(), func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)
	c.Request.Header.Set("Authorization", "Bearer invalid-token")
	r.ServeHTTP(w, c.Request)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestRequireAuth_ValidToken(t *testing.T) {
	claims := &Claims{UserID: "user-1", Email: "test@test.com", Role: "user"}
	mw := NewAuthMiddleware(&mockValidator{claims: claims})
	w := httptest.NewRecorder()
	c, r := gin.CreateTestContext(w)

	r.GET("/test", mw.RequireAuth(), func(c *gin.Context) {
		userID, _ := c.Get("user_id")
		c.JSON(200, gin.H{"user_id": userID})
	})

	c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)
	c.Request.Header.Set("Authorization", "Bearer valid-token")
	r.ServeHTTP(w, c.Request)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]string
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "user-1", resp["user_id"])
}

func TestRequireAdmin_NoRole(t *testing.T) {
	mw := NewAuthMiddleware(&mockValidator{})
	w := httptest.NewRecorder()
	c, r := gin.CreateTestContext(w)

	r.GET("/test", mw.RequireAdmin(), func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)
	r.ServeHTTP(w, c.Request)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestRequireAdmin_UserRole(t *testing.T) {
	mw := NewAuthMiddleware(&mockValidator{})
	w := httptest.NewRecorder()
	c, r := gin.CreateTestContext(w)

	r.GET("/test", func(c *gin.Context) {
		c.Set("user_role", "user")
		c.Next()
	}, mw.RequireAdmin(), func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)
	r.ServeHTTP(w, c.Request)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestRequireAdmin_AdminRole(t *testing.T) {
	mw := NewAuthMiddleware(&mockValidator{})
	w := httptest.NewRecorder()
	c, r := gin.CreateTestContext(w)

	r.GET("/test", func(c *gin.Context) {
		c.Set("user_role", "admin")
		c.Next()
	}, mw.RequireAdmin(), func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)
	r.ServeHTTP(w, c.Request)

	assert.Equal(t, http.StatusOK, w.Code)
}
