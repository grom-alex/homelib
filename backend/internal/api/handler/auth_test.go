package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/grom-alex/homelib/backend/internal/models"
	"github.com/grom-alex/homelib/backend/internal/service"
)

func TestAuthHandler_Register_Success(t *testing.T) {
	svc := &mockAuthService{
		registerFn: func(_ context.Context, input models.CreateUserInput) (*service.AuthResult, error) {
			assert.Equal(t, "test@example.com", input.Email)
			return mockAuthResult(), nil
		},
	}
	h := NewAuthHandler(svc, 30*24*time.Hour, false)

	body := `{"email":"test@example.com","username":"testuser","display_name":"Test","password":"password123"}`
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBufferString(body))
	c.Request.Header.Set("Content-Type", "application/json")

	h.Register(c)

	assert.Equal(t, http.StatusCreated, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "mock-access-token", resp["access_token"])
	assert.NotNil(t, resp["user"])
}

func TestAuthHandler_Register_InvalidInput(t *testing.T) {
	h := NewAuthHandler(&mockAuthService{}, 30*24*time.Hour, false)

	body := `{"email":"not-an-email"}` // missing required fields
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBufferString(body))
	c.Request.Header.Set("Content-Type", "application/json")

	h.Register(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthHandler_Register_UserAlreadyExists(t *testing.T) {
	svc := &mockAuthService{
		registerFn: func(_ context.Context, _ models.CreateUserInput) (*service.AuthResult, error) {
			return nil, service.ErrUserAlreadyExists
		},
	}
	h := NewAuthHandler(svc, 30*24*time.Hour, false)

	body := `{"email":"test@example.com","username":"testuser","display_name":"Test","password":"password123"}`
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBufferString(body))
	c.Request.Header.Set("Content-Type", "application/json")

	h.Register(c)

	assert.Equal(t, http.StatusConflict, w.Code)
}

func TestAuthHandler_Register_RegistrationDisabled(t *testing.T) {
	svc := &mockAuthService{
		registerFn: func(_ context.Context, _ models.CreateUserInput) (*service.AuthResult, error) {
			return nil, service.ErrRegistrationDisabled
		},
	}
	h := NewAuthHandler(svc, 30*24*time.Hour, false)

	body := `{"email":"test@example.com","username":"testuser","display_name":"Test","password":"password123"}`
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBufferString(body))
	c.Request.Header.Set("Content-Type", "application/json")

	h.Register(c)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestAuthHandler_Register_PasswordTooLong(t *testing.T) {
	svc := &mockAuthService{
		registerFn: func(_ context.Context, _ models.CreateUserInput) (*service.AuthResult, error) {
			return nil, service.ErrPasswordTooLong
		},
	}
	h := NewAuthHandler(svc, 30*24*time.Hour, false)

	body := `{"email":"test@example.com","username":"testuser","display_name":"Test","password":"password123"}`
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBufferString(body))
	c.Request.Header.Set("Content-Type", "application/json")

	h.Register(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthHandler_Register_InternalError(t *testing.T) {
	svc := &mockAuthService{
		registerFn: func(_ context.Context, _ models.CreateUserInput) (*service.AuthResult, error) {
			return nil, fmt.Errorf("db connection failed")
		},
	}
	h := NewAuthHandler(svc, 30*24*time.Hour, false)

	body := `{"email":"test@example.com","username":"testuser","display_name":"Test","password":"password123"}`
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBufferString(body))
	c.Request.Header.Set("Content-Type", "application/json")

	h.Register(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestAuthHandler_Login_Success(t *testing.T) {
	svc := &mockAuthService{
		loginFn: func(_ context.Context, input models.LoginInput) (*service.AuthResult, error) {
			assert.Equal(t, "test@example.com", input.Email)
			return mockAuthResult(), nil
		},
	}
	h := NewAuthHandler(svc, 30*24*time.Hour, false)

	body := `{"email":"test@example.com","password":"password123"}`
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBufferString(body))
	c.Request.Header.Set("Content-Type", "application/json")

	h.Login(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "mock-access-token", resp["access_token"])
}

func TestAuthHandler_Login_InvalidInput(t *testing.T) {
	h := NewAuthHandler(&mockAuthService{}, 30*24*time.Hour, false)

	body := `{"email":"bad"}` // missing password
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBufferString(body))
	c.Request.Header.Set("Content-Type", "application/json")

	h.Login(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthHandler_Login_InvalidCredentials(t *testing.T) {
	svc := &mockAuthService{
		loginFn: func(_ context.Context, _ models.LoginInput) (*service.AuthResult, error) {
			return nil, service.ErrInvalidCredentials
		},
	}
	h := NewAuthHandler(svc, 30*24*time.Hour, false)

	body := `{"email":"test@example.com","password":"wrongpassword"}`
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBufferString(body))
	c.Request.Header.Set("Content-Type", "application/json")

	h.Login(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthHandler_Refresh_Success(t *testing.T) {
	svc := &mockAuthService{
		refreshTokenFn: func(_ context.Context, tokenStr string) (*service.AuthResult, error) {
			assert.Equal(t, "old-refresh-token", tokenStr)
			return mockAuthResult(), nil
		},
	}
	h := NewAuthHandler(svc, 30*24*time.Hour, false)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/auth/refresh", nil)
	c.Request.AddCookie(&http.Cookie{Name: "refresh_token", Value: "old-refresh-token"})

	h.Refresh(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "mock-access-token", resp["access_token"])
}

func TestAuthHandler_Refresh_NoCookie(t *testing.T) {
	h := NewAuthHandler(&mockAuthService{}, 30*24*time.Hour, false)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/auth/refresh", nil)

	h.Refresh(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthHandler_Refresh_InvalidToken(t *testing.T) {
	svc := &mockAuthService{
		refreshTokenFn: func(_ context.Context, _ string) (*service.AuthResult, error) {
			return nil, service.ErrInvalidRefreshToken
		},
	}
	h := NewAuthHandler(svc, 30*24*time.Hour, false)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/auth/refresh", nil)
	c.Request.AddCookie(&http.Cookie{Name: "refresh_token", Value: "bad-token"})

	h.Refresh(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthHandler_Logout_Success(t *testing.T) {
	logoutCalled := false
	svc := &mockAuthService{
		logoutFn: func(_ context.Context, tokenStr string) error {
			logoutCalled = true
			assert.Equal(t, "some-refresh-token", tokenStr)
			return nil
		},
	}
	h := NewAuthHandler(svc, 30*24*time.Hour, false)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
	c.Request.AddCookie(&http.Cookie{Name: "refresh_token", Value: "some-refresh-token"})

	h.Logout(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.True(t, logoutCalled)
}

func TestAuthHandler_Logout_NoCookie(t *testing.T) {
	h := NewAuthHandler(&mockAuthService{}, 30*24*time.Hour, false)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)

	h.Logout(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAuthHandler_SetRefreshCookie(t *testing.T) {
	svc := &mockAuthService{
		loginFn: func(_ context.Context, _ models.LoginInput) (*service.AuthResult, error) {
			return mockAuthResult(), nil
		},
	}
	h := NewAuthHandler(svc, 30*24*time.Hour, true)

	body := `{"email":"test@example.com","password":"password123"}`
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBufferString(body))
	c.Request.Header.Set("Content-Type", "application/json")

	h.Login(c)

	cookies := w.Result().Cookies()
	require.Len(t, cookies, 1)
	assert.Equal(t, "refresh_token", cookies[0].Name)
	assert.Equal(t, "mock-refresh-token", cookies[0].Value)
	assert.True(t, cookies[0].HttpOnly)
	assert.True(t, cookies[0].Secure)
}
