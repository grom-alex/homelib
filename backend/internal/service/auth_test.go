package service

import (
	"testing"
	"time"

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

	// Generate a token manually
	token, err := svc.generateAccessToken(&mockUser)
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	// Validate it
	info, err := svc.ValidateToken(token)
	require.NoError(t, err)
	assert.Equal(t, "user-123", info.ID)
	assert.Equal(t, "test@example.com", info.Email)
	assert.Equal(t, "admin", info.Role)
}

func TestAuthService_ValidateToken_InvalidToken(t *testing.T) {
	cfg := config.AuthConfig{
		JWTSecret: "test-secret",
	}
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

func TestHashToken_Deterministic(t *testing.T) {
	h1 := hashToken("test-token")
	h2 := hashToken("test-token")
	assert.Equal(t, h1, h2)

	h3 := hashToken("different-token")
	assert.NotEqual(t, h1, h3)
}
