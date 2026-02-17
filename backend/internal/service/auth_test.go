package service

import (
	"context"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/grom-alex/homelib/backend/internal/config"
	"github.com/grom-alex/homelib/backend/internal/models"
)

var mockUser = models.User{
	ID:          "user-123",
	Email:       "test@example.com",
	Username:    "testuser",
	DisplayName: "Test User",
	Role:        "admin",
}

func TestAuthService_ValidateToken_RoundTrip(t *testing.T) {
	cfg := config.AuthConfig{
		JWTSecret:      "test-secret-key-for-testing",
		AccessTokenTTL: 15 * time.Minute,
	}
	svc := NewAuthService(cfg, nil, nil)

	token, err := svc.generateAccessToken(&mockUser)
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	info, err := svc.ValidateToken(token)
	require.NoError(t, err)
	assert.Equal(t, "user-123", info.ID)
	assert.Equal(t, "test@example.com", info.Email)
	assert.Equal(t, "admin", info.Role)
}

func TestAuthService_ValidateToken_InvalidToken(t *testing.T) {
	cfg := config.AuthConfig{JWTSecret: "test-secret"}
	svc := NewAuthService(cfg, nil, nil)

	_, err := svc.ValidateToken("invalid-token")
	assert.Error(t, err)
}

func TestAuthService_ValidateToken_WrongSecret(t *testing.T) {
	cfg1 := config.AuthConfig{
		JWTSecret:      "secret-1",
		AccessTokenTTL: 15 * time.Minute,
	}
	svc1 := NewAuthService(cfg1, nil, nil)

	token, err := svc1.generateAccessToken(&mockUser)
	require.NoError(t, err)

	cfg2 := config.AuthConfig{JWTSecret: "secret-2"}
	svc2 := NewAuthService(cfg2, nil, nil)

	_, err = svc2.ValidateToken(token)
	assert.Error(t, err)
}

func TestAuthService_ValidateToken_ExpiredToken(t *testing.T) {
	cfg := config.AuthConfig{
		JWTSecret:      "test-secret",
		AccessTokenTTL: -1 * time.Hour, // expired
	}
	svc := NewAuthService(cfg, nil, nil)

	token, err := svc.generateAccessToken(&mockUser)
	require.NoError(t, err)

	_, err = svc.ValidateToken(token)
	assert.Error(t, err)
}

func TestAuthService_ValidateToken_WrongSigningMethod(t *testing.T) {
	// Create a token signed with RS256 (different method)
	token := jwt.NewWithClaims(jwt.SigningMethodNone, jwt.MapClaims{
		"sub":   "user-123",
		"email": "test@example.com",
		"role":  "admin",
		"exp":   time.Now().Add(time.Hour).Unix(),
	})
	tokenStr, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
	require.NoError(t, err)

	cfg := config.AuthConfig{JWTSecret: "test-secret"}
	svc := NewAuthService(cfg, nil, nil)

	_, err = svc.ValidateToken(tokenStr)
	assert.Error(t, err)
}

func TestHashToken_Deterministic(t *testing.T) {
	h1 := hashToken("test-token")
	h2 := hashToken("test-token")
	assert.Equal(t, h1, h2)

	h3 := hashToken("different-token")
	assert.NotEqual(t, h1, h3)
}

func TestHashToken_NonEmpty(t *testing.T) {
	h := hashToken("")
	assert.NotEmpty(t, h)
	assert.Len(t, h, 64) // sha256 hex = 64 chars
}

func TestClaimString_Found(t *testing.T) {
	claims := jwt.MapClaims{"key": "value"}
	assert.Equal(t, "value", claimString(claims, "key"))
}

func TestClaimString_NotFound(t *testing.T) {
	claims := jwt.MapClaims{"key": "value"}
	assert.Equal(t, "", claimString(claims, "missing"))
}

func TestClaimString_NonString(t *testing.T) {
	claims := jwt.MapClaims{"key": 42}
	assert.Equal(t, "", claimString(claims, "key"))
}

func TestAuthService_GenerateAccessToken_Claims(t *testing.T) {
	cfg := config.AuthConfig{
		JWTSecret:      "test-secret",
		AccessTokenTTL: 15 * time.Minute,
	}
	svc := NewAuthService(cfg, nil, nil)

	tokenStr, err := svc.generateAccessToken(&mockUser)
	require.NoError(t, err)

	// Parse token without validation to inspect claims
	parser := jwt.NewParser(jwt.WithoutClaimsValidation())
	token, _, err := parser.ParseUnverified(tokenStr, jwt.MapClaims{})
	require.NoError(t, err)

	claims := token.Claims.(jwt.MapClaims)
	assert.Equal(t, "user-123", claims["sub"])
	assert.Equal(t, "test@example.com", claims["email"])
	assert.Equal(t, "admin", claims["role"])
	assert.Equal(t, "Test User", claims["name"])
	assert.NotNil(t, claims["iat"])
	assert.NotNil(t, claims["exp"])
}

func TestNewAuthService(t *testing.T) {
	cfg := config.AuthConfig{JWTSecret: "secret"}
	svc := NewAuthService(cfg, nil, nil)
	assert.NotNil(t, svc)
	assert.Equal(t, "secret", svc.cfg.JWTSecret)
}

func TestAuthService_Register_PasswordTooLong(t *testing.T) {
	cfg := config.AuthConfig{JWTSecret: "secret"}
	svc := NewAuthService(cfg, nil, nil)

	// Create password longer than 72 bytes
	longPassword := string(make([]byte, 73))
	input := models.CreateUserInput{
		Email:       "test@example.com",
		Username:    "test",
		DisplayName: "Test",
		Password:    longPassword,
	}

	_, err := svc.Register(context.TODO(), input)
	assert.ErrorIs(t, err, ErrPasswordTooLong)
}
