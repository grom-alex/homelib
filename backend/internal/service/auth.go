package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/grom-alex/homelib/backend/internal/config"
	"github.com/grom-alex/homelib/backend/internal/models"
	"github.com/grom-alex/homelib/backend/internal/repository"
)

type AuthService struct {
	cfg            config.AuthConfig
	userRepo       *repository.UserRepo
	refreshRepo    *repository.RefreshTokenRepo
}

func NewAuthService(cfg config.AuthConfig, userRepo *repository.UserRepo, refreshRepo *repository.RefreshTokenRepo) *AuthService {
	return &AuthService{
		cfg:         cfg,
		userRepo:    userRepo,
		refreshRepo: refreshRepo,
	}
}

type AuthResult struct {
	User         models.UserInfo `json:"user"`
	AccessToken  string          `json:"access_token"`
	RefreshToken string          `json:"-"` // sent via httpOnly cookie
}

// Register creates a new user. First user becomes admin.
// Uses a transaction to prevent race conditions on first registration.
func (s *AuthService) Register(ctx context.Context, input models.CreateUserInput) (*AuthResult, error) {
	// Validate password length in bytes (bcrypt silently truncates at 72 bytes)
	if len([]byte(input.Password)) > 72 {
		return nil, ErrPasswordTooLong
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	user := &models.User{
		Email:        strings.ToLower(strings.TrimSpace(input.Email)),
		Username:     strings.TrimSpace(input.Username),
		DisplayName:  strings.TrimSpace(input.DisplayName),
		PasswordHash: string(hash),
		Role:         "user",
		IsActive:     true,
	}

	if err := s.userRepo.RegisterUser(ctx, user, s.cfg.RegistrationEnabled); err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			return nil, ErrUserAlreadyExists
		}
		if strings.Contains(err.Error(), "registration is disabled") {
			return nil, ErrRegistrationDisabled
		}
		return nil, err
	}

	return s.createTokenPair(ctx, user)
}

// Login authenticates a user by email and password.
func (s *AuthService) Login(ctx context.Context, input models.LoginInput) (*AuthResult, error) {
	user, err := s.userRepo.GetByEmail(ctx, strings.ToLower(strings.TrimSpace(input.Email)))
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	if !user.IsActive {
		return nil, ErrAccountDeactivated
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	_ = s.userRepo.UpdateLastLogin(ctx, user.ID)

	return s.createTokenPair(ctx, user)
}

// RefreshToken atomically rotates the refresh token and returns new tokens.
// Uses DELETE ... RETURNING to prevent race conditions with concurrent requests.
func (s *AuthService) RefreshToken(ctx context.Context, refreshTokenStr string) (*AuthResult, error) {
	tokenHash := hashToken(refreshTokenStr)

	// Atomic: delete old token and get user_id in one query
	userID, err := s.refreshRepo.DeleteAndReturnUserID(ctx, tokenHash)
	if err != nil {
		return nil, ErrInvalidRefreshToken
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	if !user.IsActive {
		return nil, ErrAccountDeactivated
	}

	return s.createTokenPair(ctx, user)
}

// Logout invalidates the refresh token.
func (s *AuthService) Logout(ctx context.Context, refreshTokenStr string) error {
	tokenHash := hashToken(refreshTokenStr)
	return s.refreshRepo.Delete(ctx, tokenHash)
}

// ValidateToken validates a JWT access token and returns claims.
func (s *AuthService) ValidateToken(tokenString string) (*models.UserInfo, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.cfg.JWTSecret), nil
	})
	if err != nil {
		return nil, fmt.Errorf("parse token: %w", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	return &models.UserInfo{
		ID:    claimString(claims, "sub"),
		Email: claimString(claims, "email"),
		Role:  claimString(claims, "role"),
	}, nil
}

func (s *AuthService) createTokenPair(ctx context.Context, user *models.User) (*AuthResult, error) {
	accessToken, err := s.generateAccessToken(user)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.generateRefreshToken(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	return &AuthResult{
		User:         user.ToUserInfo(),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *AuthService) generateAccessToken(user *models.User) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"sub":   user.ID,
		"email": user.Email,
		"role":  user.Role,
		"name":  user.DisplayName,
		"iat":   now.Unix(),
		"exp":   now.Add(s.cfg.AccessTokenTTL).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.cfg.JWTSecret))
}

func (s *AuthService) generateRefreshToken(ctx context.Context, userID string) (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate random bytes: %w", err)
	}
	tokenStr := hex.EncodeToString(b)
	tokenHash := hashToken(tokenStr)

	expiresAt := time.Now().Add(s.cfg.RefreshTokenTTL)
	if err := s.refreshRepo.Create(ctx, userID, tokenHash, expiresAt); err != nil {
		return "", err
	}

	return tokenStr, nil
}

func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

func claimString(claims jwt.MapClaims, key string) string {
	if v, ok := claims[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}
