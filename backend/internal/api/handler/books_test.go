package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBooksHandler_ListBooks_BadParams(t *testing.T) {
	h := NewBooksHandler(nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/books?page=abc", nil)

	h.ListBooks(c)

	// Should return 400 for invalid page param
	assert.Equal(t, http.StatusBadRequest, w.Code)
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
