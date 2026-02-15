package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type RefreshTokenRepo struct {
	pool *pgxpool.Pool
}

func NewRefreshTokenRepo(pool *pgxpool.Pool) *RefreshTokenRepo {
	return &RefreshTokenRepo{pool: pool}
}

func (r *RefreshTokenRepo) Create(ctx context.Context, userID, tokenHash string, expiresAt time.Time) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO refresh_tokens (user_id, token_hash, expires_at)
		 VALUES ($1, $2, $3)`,
		userID, tokenHash, expiresAt,
	)
	if err != nil {
		return fmt.Errorf("create refresh token: %w", err)
	}
	return nil
}

// GetUserIDByTokenHash returns the user_id for an unexpired refresh token.
func (r *RefreshTokenRepo) GetUserIDByTokenHash(ctx context.Context, tokenHash string) (string, error) {
	var userID string
	err := r.pool.QueryRow(ctx,
		`SELECT user_id FROM refresh_tokens
		 WHERE token_hash = $1 AND expires_at > NOW()`,
		tokenHash,
	).Scan(&userID)
	if err != nil {
		return "", fmt.Errorf("get refresh token: %w", err)
	}
	return userID, nil
}

func (r *RefreshTokenRepo) Delete(ctx context.Context, tokenHash string) error {
	_, err := r.pool.Exec(ctx,
		"DELETE FROM refresh_tokens WHERE token_hash = $1", tokenHash)
	if err != nil {
		return fmt.Errorf("delete refresh token: %w", err)
	}
	return nil
}

func (r *RefreshTokenRepo) DeleteAllForUser(ctx context.Context, userID string) error {
	_, err := r.pool.Exec(ctx,
		"DELETE FROM refresh_tokens WHERE user_id = $1", userID)
	if err != nil {
		return fmt.Errorf("delete all refresh tokens for user: %w", err)
	}
	return nil
}

func (r *RefreshTokenRepo) CleanupExpired(ctx context.Context) error {
	_, err := r.pool.Exec(ctx,
		"DELETE FROM refresh_tokens WHERE expires_at < NOW()")
	if err != nil {
		return fmt.Errorf("cleanup expired tokens: %w", err)
	}
	return nil
}
