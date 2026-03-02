package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type MetadataRepo struct {
	pool *pgxpool.Pool
}

func NewMetadataRepo(pool *pgxpool.Pool) *MetadataRepo {
	return &MetadataRepo{pool: pool}
}

// Get returns the value for the given key, or empty string if not found.
func (r *MetadataRepo) Get(ctx context.Context, key string) (string, error) {
	var value string
	err := r.pool.QueryRow(ctx,
		`SELECT value FROM app_metadata WHERE key = $1`, key,
	).Scan(&value)
	if err == pgx.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("get metadata %q: %w", key, err)
	}
	return value, nil
}

// Set upserts a key-value pair in app_metadata.
func (r *MetadataRepo) Set(ctx context.Context, key, value string) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO app_metadata (key, value, updated_at)
		 VALUES ($1, $2, NOW())
		 ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value, updated_at = NOW()`,
		key, value,
	)
	if err != nil {
		return fmt.Errorf("set metadata %q: %w", key, err)
	}
	return nil
}
