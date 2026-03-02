package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/grom-alex/homelib/backend/internal/models"
)

// --- mock parental service ---

type mockParentalService struct {
	getRestrictedGenreCodesFn func(ctx context.Context) ([]string, error)
	setRestrictedGenreCodesFn func(ctx context.Context, codes []string) error
	setPinFn                  func(ctx context.Context, pin string) error
	verifyPinFn               func(ctx context.Context, pin string) (bool, error)
	removePinFn               func(ctx context.Context) error
	isPinSetFn                func(ctx context.Context) (bool, error)
	getRestrictedGenreIDsFn   func(ctx context.Context) ([]int, error)
	isAdultContentEnabledFn   func(ctx context.Context, userID string) (bool, error)
	setAdultContentEnabledFn  func(ctx context.Context, userID string, enabled bool) error
	listUsersAdultStatusFn    func(ctx context.Context) ([]models.UserAdultStatus, error)
}

func (m *mockParentalService) GetRestrictedGenreCodes(ctx context.Context) ([]string, error) {
	if m.getRestrictedGenreCodesFn != nil {
		return m.getRestrictedGenreCodesFn(ctx)
	}
	return nil, fmt.Errorf("not implemented")
}
func (m *mockParentalService) SetRestrictedGenreCodes(ctx context.Context, codes []string) error {
	if m.setRestrictedGenreCodesFn != nil {
		return m.setRestrictedGenreCodesFn(ctx, codes)
	}
	return fmt.Errorf("not implemented")
}
func (m *mockParentalService) SetPin(ctx context.Context, pin string) error {
	if m.setPinFn != nil {
		return m.setPinFn(ctx, pin)
	}
	return fmt.Errorf("not implemented")
}
func (m *mockParentalService) VerifyPin(ctx context.Context, pin string) (bool, error) {
	if m.verifyPinFn != nil {
		return m.verifyPinFn(ctx, pin)
	}
	return false, fmt.Errorf("not implemented")
}
func (m *mockParentalService) RemovePin(ctx context.Context) error {
	if m.removePinFn != nil {
		return m.removePinFn(ctx)
	}
	return fmt.Errorf("not implemented")
}
func (m *mockParentalService) IsPinSet(ctx context.Context) (bool, error) {
	if m.isPinSetFn != nil {
		return m.isPinSetFn(ctx)
	}
	return false, fmt.Errorf("not implemented")
}
func (m *mockParentalService) GetRestrictedGenreIDs(ctx context.Context) ([]int, error) {
	if m.getRestrictedGenreIDsFn != nil {
		return m.getRestrictedGenreIDsFn(ctx)
	}
	return nil, fmt.Errorf("not implemented")
}
func (m *mockParentalService) IsAdultContentEnabled(ctx context.Context, userID string) (bool, error) {
	if m.isAdultContentEnabledFn != nil {
		return m.isAdultContentEnabledFn(ctx, userID)
	}
	return false, fmt.Errorf("not implemented")
}
func (m *mockParentalService) SetAdultContentEnabled(ctx context.Context, userID string, enabled bool) error {
	if m.setAdultContentEnabledFn != nil {
		return m.setAdultContentEnabledFn(ctx, userID, enabled)
	}
	return fmt.Errorf("not implemented")
}
func (m *mockParentalService) ListUsersAdultStatus(ctx context.Context) ([]models.UserAdultStatus, error) {
	if m.listUsersAdultStatusFn != nil {
		return m.listUsersAdultStatusFn(ctx)
	}
	return nil, fmt.Errorf("not implemented")
}

// --- Admin: GetRestrictedGenres ---

func TestParentalHandler_GetRestrictedGenres_Success(t *testing.T) {
	svc := &mockParentalService{
		getRestrictedGenreCodesFn: func(_ context.Context) ([]string, error) {
			return []string{"love", "erotica"}, nil
		},
	}
	h := NewParentalHandler(svc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/admin/parental/genres", nil)

	h.GetRestrictedGenres(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string][]string
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, []string{"love", "erotica"}, resp["codes"])
}

func TestParentalHandler_GetRestrictedGenres_Empty(t *testing.T) {
	svc := &mockParentalService{
		getRestrictedGenreCodesFn: func(_ context.Context) ([]string, error) {
			return nil, nil
		},
	}
	h := NewParentalHandler(svc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/admin/parental/genres", nil)

	h.GetRestrictedGenres(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string][]string
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Empty(t, resp["codes"])
}

func TestParentalHandler_GetRestrictedGenres_Error(t *testing.T) {
	svc := &mockParentalService{
		getRestrictedGenreCodesFn: func(_ context.Context) ([]string, error) {
			return nil, fmt.Errorf("db error")
		},
	}
	h := NewParentalHandler(svc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/admin/parental/genres", nil)

	h.GetRestrictedGenres(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// --- Admin: UpdateRestrictedGenres ---

func TestParentalHandler_UpdateRestrictedGenres_Success(t *testing.T) {
	svc := &mockParentalService{
		setRestrictedGenreCodesFn: func(_ context.Context, codes []string) error {
			return nil
		},
	}
	h := NewParentalHandler(svc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPut, "/api/admin/parental/genres",
		strings.NewReader(`{"codes":["love"]}`))
	c.Request.Header.Set("Content-Type", "application/json")

	h.UpdateRestrictedGenres(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string][]string
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, []string{"love"}, resp["codes"])
}

func TestParentalHandler_UpdateRestrictedGenres_InvalidInput(t *testing.T) {
	h := NewParentalHandler(&mockParentalService{})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPut, "/api/admin/parental/genres",
		strings.NewReader(`{}`))
	c.Request.Header.Set("Content-Type", "application/json")

	h.UpdateRestrictedGenres(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestParentalHandler_UpdateRestrictedGenres_Error(t *testing.T) {
	svc := &mockParentalService{
		setRestrictedGenreCodesFn: func(_ context.Context, _ []string) error {
			return fmt.Errorf("db error")
		},
	}
	h := NewParentalHandler(svc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPut, "/api/admin/parental/genres",
		strings.NewReader(`{"codes":["love"]}`))
	c.Request.Header.Set("Content-Type", "application/json")

	h.UpdateRestrictedGenres(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// --- Admin: SetPin ---

func TestParentalHandler_SetPin_Success(t *testing.T) {
	svc := &mockParentalService{
		setPinFn: func(_ context.Context, pin string) error {
			return nil
		},
	}
	h := NewParentalHandler(svc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/admin/parental/pin",
		strings.NewReader(`{"pin":"1234"}`))
	c.Request.Header.Set("Content-Type", "application/json")

	h.SetPin(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestParentalHandler_SetPin_InvalidInput(t *testing.T) {
	h := NewParentalHandler(&mockParentalService{})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	// PIN too short
	c.Request = httptest.NewRequest(http.MethodPost, "/api/admin/parental/pin",
		strings.NewReader(`{"pin":"12"}`))
	c.Request.Header.Set("Content-Type", "application/json")

	h.SetPin(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestParentalHandler_SetPin_NonNumeric(t *testing.T) {
	h := NewParentalHandler(&mockParentalService{})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/admin/parental/pin",
		strings.NewReader(`{"pin":"abcd"}`))
	c.Request.Header.Set("Content-Type", "application/json")

	h.SetPin(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestParentalHandler_SetPin_Error(t *testing.T) {
	svc := &mockParentalService{
		setPinFn: func(_ context.Context, _ string) error {
			return fmt.Errorf("hash error")
		},
	}
	h := NewParentalHandler(svc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/admin/parental/pin",
		strings.NewReader(`{"pin":"1234"}`))
	c.Request.Header.Set("Content-Type", "application/json")

	h.SetPin(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// --- Admin: RemovePin ---

func TestParentalHandler_RemovePin_Success(t *testing.T) {
	svc := &mockParentalService{
		removePinFn: func(_ context.Context) error { return nil },
	}
	h := NewParentalHandler(svc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodDelete, "/api/admin/parental/pin", nil)

	h.RemovePin(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestParentalHandler_RemovePin_Error(t *testing.T) {
	svc := &mockParentalService{
		removePinFn: func(_ context.Context) error { return fmt.Errorf("db error") },
	}
	h := NewParentalHandler(svc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodDelete, "/api/admin/parental/pin", nil)

	h.RemovePin(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// --- Admin: GetAdminParentalStatus ---

func TestParentalHandler_GetAdminParentalStatus_Success(t *testing.T) {
	svc := &mockParentalService{
		isPinSetFn: func(_ context.Context) (bool, error) { return true, nil },
		getRestrictedGenreCodesFn: func(_ context.Context) ([]string, error) {
			return []string{"erotica"}, nil
		},
	}
	h := NewParentalHandler(svc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/admin/parental/status", nil)

	h.GetAdminParentalStatus(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp models.AdminParentalStatus
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.True(t, resp.PinSet)
	assert.Equal(t, []string{"erotica"}, resp.RestrictedGenreCodes)
}

func TestParentalHandler_GetAdminParentalStatus_PinError(t *testing.T) {
	svc := &mockParentalService{
		isPinSetFn: func(_ context.Context) (bool, error) { return false, fmt.Errorf("err") },
	}
	h := NewParentalHandler(svc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/admin/parental/status", nil)

	h.GetAdminParentalStatus(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestParentalHandler_GetAdminParentalStatus_CodesError(t *testing.T) {
	svc := &mockParentalService{
		isPinSetFn: func(_ context.Context) (bool, error) { return true, nil },
		getRestrictedGenreCodesFn: func(_ context.Context) ([]string, error) {
			return nil, fmt.Errorf("err")
		},
	}
	h := NewParentalHandler(svc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/admin/parental/status", nil)

	h.GetAdminParentalStatus(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// --- Admin: ListUsersAdultStatus ---

func TestParentalHandler_ListUsersAdultStatus_Success(t *testing.T) {
	users := []models.UserAdultStatus{
		{UserID: "u1", Username: "admin", DisplayName: "Admin", Role: "admin", AdultContentEnabled: true},
	}
	svc := &mockParentalService{
		listUsersAdultStatusFn: func(_ context.Context) ([]models.UserAdultStatus, error) {
			return users, nil
		},
	}
	h := NewParentalHandler(svc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/admin/parental/users", nil)

	h.ListUsersAdultStatus(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp []models.UserAdultStatus
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Len(t, resp, 1)
	assert.Equal(t, "u1", resp[0].UserID)
}

func TestParentalHandler_ListUsersAdultStatus_Empty(t *testing.T) {
	svc := &mockParentalService{
		listUsersAdultStatusFn: func(_ context.Context) ([]models.UserAdultStatus, error) {
			return nil, nil
		},
	}
	h := NewParentalHandler(svc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/admin/parental/users", nil)

	h.ListUsersAdultStatus(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp []models.UserAdultStatus
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Empty(t, resp)
}

func TestParentalHandler_ListUsersAdultStatus_Error(t *testing.T) {
	svc := &mockParentalService{
		listUsersAdultStatusFn: func(_ context.Context) ([]models.UserAdultStatus, error) {
			return nil, fmt.Errorf("db error")
		},
	}
	h := NewParentalHandler(svc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/admin/parental/users", nil)

	h.ListUsersAdultStatus(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// --- Admin: SetUserAdultContent ---

func TestParentalHandler_SetUserAdultContent_Success(t *testing.T) {
	svc := &mockParentalService{
		setAdultContentEnabledFn: func(_ context.Context, userID string, enabled bool) error {
			return nil
		},
	}
	h := NewParentalHandler(svc)

	w := httptest.NewRecorder()
	c, r := gin.CreateTestContext(w)
	r.PUT("/api/admin/parental/users/:userId", h.SetUserAdultContent)
	c.Request = httptest.NewRequest(http.MethodPut, "/api/admin/parental/users/user-1",
		strings.NewReader(`{"adult_content_enabled":true}`))
	c.Request.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, c.Request)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, true, resp["adult_content_enabled"])
}

func TestParentalHandler_SetUserAdultContent_InvalidInput(t *testing.T) {
	h := NewParentalHandler(&mockParentalService{})

	w := httptest.NewRecorder()
	c, r := gin.CreateTestContext(w)
	r.PUT("/api/admin/parental/users/:userId", h.SetUserAdultContent)
	c.Request = httptest.NewRequest(http.MethodPut, "/api/admin/parental/users/user-1",
		strings.NewReader(`invalid json`))
	c.Request.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, c.Request)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestParentalHandler_SetUserAdultContent_Error(t *testing.T) {
	svc := &mockParentalService{
		setAdultContentEnabledFn: func(_ context.Context, _ string, _ bool) error {
			return fmt.Errorf("db error")
		},
	}
	h := NewParentalHandler(svc)

	w := httptest.NewRecorder()
	c, r := gin.CreateTestContext(w)
	r.PUT("/api/admin/parental/users/:userId", h.SetUserAdultContent)
	c.Request = httptest.NewRequest(http.MethodPut, "/api/admin/parental/users/user-1",
		strings.NewReader(`{"adult_content_enabled":false}`))
	c.Request.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, c.Request)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// --- User: GetMyParentalStatus ---

func TestParentalHandler_GetMyParentalStatus_Success(t *testing.T) {
	svc := &mockParentalService{
		isPinSetFn:              func(_ context.Context) (bool, error) { return true, nil },
		isAdultContentEnabledFn: func(_ context.Context, _ string) (bool, error) { return true, nil },
	}
	h := NewParentalHandler(svc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("user_id", "user-1")
	c.Request = httptest.NewRequest(http.MethodGet, "/api/me/parental/status", nil)

	h.GetMyParentalStatus(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp models.ParentalStatus
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.True(t, resp.PinSet)
	assert.True(t, resp.AdultContentEnabled)
}

func TestParentalHandler_GetMyParentalStatus_NoUserID(t *testing.T) {
	h := NewParentalHandler(&mockParentalService{})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/me/parental/status", nil)

	h.GetMyParentalStatus(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestParentalHandler_GetMyParentalStatus_PinError(t *testing.T) {
	svc := &mockParentalService{
		isPinSetFn: func(_ context.Context) (bool, error) { return false, fmt.Errorf("err") },
	}
	h := NewParentalHandler(svc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("user_id", "user-1")
	c.Request = httptest.NewRequest(http.MethodGet, "/api/me/parental/status", nil)

	h.GetMyParentalStatus(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestParentalHandler_GetMyParentalStatus_AdultError(t *testing.T) {
	svc := &mockParentalService{
		isPinSetFn:              func(_ context.Context) (bool, error) { return true, nil },
		isAdultContentEnabledFn: func(_ context.Context, _ string) (bool, error) { return false, fmt.Errorf("err") },
	}
	h := NewParentalHandler(svc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("user_id", "user-1")
	c.Request = httptest.NewRequest(http.MethodGet, "/api/me/parental/status", nil)

	h.GetMyParentalStatus(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// --- User: UnlockAdultContent ---

func TestParentalHandler_UnlockAdultContent_Success(t *testing.T) {
	svc := &mockParentalService{
		verifyPinFn: func(_ context.Context, pin string) (bool, error) { return true, nil },
		setAdultContentEnabledFn: func(_ context.Context, _ string, _ bool) error { return nil },
	}
	h := NewParentalHandler(svc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("user_id", "user-1")
	c.Request = httptest.NewRequest(http.MethodPost, "/api/me/parental/unlock",
		strings.NewReader(`{"pin":"1234"}`))
	c.Request.Header.Set("Content-Type", "application/json")

	h.UnlockAdultContent(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, true, resp["adult_content_enabled"])
}

func TestParentalHandler_UnlockAdultContent_NoUserID(t *testing.T) {
	h := NewParentalHandler(&mockParentalService{})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/me/parental/unlock",
		strings.NewReader(`{"pin":"1234"}`))
	c.Request.Header.Set("Content-Type", "application/json")

	h.UnlockAdultContent(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestParentalHandler_UnlockAdultContent_InvalidInput(t *testing.T) {
	h := NewParentalHandler(&mockParentalService{})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("user_id", "user-1")
	c.Request = httptest.NewRequest(http.MethodPost, "/api/me/parental/unlock",
		strings.NewReader(`{}`))
	c.Request.Header.Set("Content-Type", "application/json")

	h.UnlockAdultContent(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestParentalHandler_UnlockAdultContent_WrongPin(t *testing.T) {
	svc := &mockParentalService{
		verifyPinFn: func(_ context.Context, _ string) (bool, error) { return false, nil },
	}
	h := NewParentalHandler(svc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("user_id", "user-1")
	c.Request = httptest.NewRequest(http.MethodPost, "/api/me/parental/unlock",
		strings.NewReader(`{"pin":"0000"}`))
	c.Request.Header.Set("Content-Type", "application/json")

	h.UnlockAdultContent(c)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestParentalHandler_UnlockAdultContent_VerifyError(t *testing.T) {
	svc := &mockParentalService{
		verifyPinFn: func(_ context.Context, _ string) (bool, error) { return false, fmt.Errorf("err") },
	}
	h := NewParentalHandler(svc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("user_id", "user-1")
	c.Request = httptest.NewRequest(http.MethodPost, "/api/me/parental/unlock",
		strings.NewReader(`{"pin":"1234"}`))
	c.Request.Header.Set("Content-Type", "application/json")

	h.UnlockAdultContent(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestParentalHandler_UnlockAdultContent_SetError(t *testing.T) {
	svc := &mockParentalService{
		verifyPinFn:              func(_ context.Context, _ string) (bool, error) { return true, nil },
		setAdultContentEnabledFn: func(_ context.Context, _ string, _ bool) error { return fmt.Errorf("err") },
	}
	h := NewParentalHandler(svc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("user_id", "user-1")
	c.Request = httptest.NewRequest(http.MethodPost, "/api/me/parental/unlock",
		strings.NewReader(`{"pin":"1234"}`))
	c.Request.Header.Set("Content-Type", "application/json")

	h.UnlockAdultContent(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// --- User: LockAdultContent ---

func TestParentalHandler_LockAdultContent_Success(t *testing.T) {
	svc := &mockParentalService{
		setAdultContentEnabledFn: func(_ context.Context, _ string, _ bool) error { return nil },
	}
	h := NewParentalHandler(svc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("user_id", "user-1")
	c.Request = httptest.NewRequest(http.MethodPost, "/api/me/parental/lock", nil)

	h.LockAdultContent(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, false, resp["adult_content_enabled"])
}

func TestParentalHandler_LockAdultContent_NoUserID(t *testing.T) {
	h := NewParentalHandler(&mockParentalService{})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/me/parental/lock", nil)

	h.LockAdultContent(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestParentalHandler_LockAdultContent_Error(t *testing.T) {
	svc := &mockParentalService{
		setAdultContentEnabledFn: func(_ context.Context, _ string, _ bool) error { return fmt.Errorf("err") },
	}
	h := NewParentalHandler(svc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("user_id", "user-1")
	c.Request = httptest.NewRequest(http.MethodPost, "/api/me/parental/lock", nil)

	h.LockAdultContent(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
