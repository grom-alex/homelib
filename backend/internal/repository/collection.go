package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/grom-alex/homelib/backend/internal/models"
)

type CollectionRepo struct {
	pool *pgxpool.Pool
}

func NewCollectionRepo(pool *pgxpool.Pool) *CollectionRepo {
	return &CollectionRepo{pool: pool}
}

// Upsert creates or updates a collection by code.
func (r *CollectionRepo) Upsert(ctx context.Context, tx pgx.Tx, coll *models.Collection) error {
	err := tx.QueryRow(ctx,
		`INSERT INTO collections (name, code, collection_type, description, source_url, version, version_date, books_count, last_import_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())
		 ON CONFLICT (code) DO UPDATE SET
			name = EXCLUDED.name,
			collection_type = EXCLUDED.collection_type,
			description = EXCLUDED.description,
			source_url = EXCLUDED.source_url,
			version = EXCLUDED.version,
			version_date = EXCLUDED.version_date,
			books_count = EXCLUDED.books_count,
			last_import_at = NOW(),
			updated_at = NOW()
		 RETURNING id`,
		coll.Name, coll.Code, coll.CollectionType, coll.Description,
		coll.SourceURL, coll.Version, coll.VersionDate, coll.BooksCount,
	).Scan(&coll.ID)
	if err != nil {
		return fmt.Errorf("upsert collection %q: %w", coll.Code, err)
	}
	return nil
}

func (r *CollectionRepo) GetByCode(ctx context.Context, code string) (*models.Collection, error) {
	var c models.Collection
	err := r.pool.QueryRow(ctx,
		`SELECT id, name, code, collection_type, description, source_url, version,
				version_date, books_count, last_import_at, created_at, updated_at
		 FROM collections WHERE code = $1`, code,
	).Scan(&c.ID, &c.Name, &c.Code, &c.CollectionType, &c.Description,
		&c.SourceURL, &c.Version, &c.VersionDate, &c.BooksCount,
		&c.LastImportAt, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get collection %q: %w", code, err)
	}
	return &c, nil
}

func (r *CollectionRepo) GetAll(ctx context.Context) ([]models.Collection, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, name, code, collection_type, description, source_url, version,
				version_date, books_count, last_import_at, created_at, updated_at
		 FROM collections ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("list collections: %w", err)
	}
	defer rows.Close()

	var items []models.Collection
	for rows.Next() {
		var c models.Collection
		if err := rows.Scan(&c.ID, &c.Name, &c.Code, &c.CollectionType, &c.Description,
			&c.SourceURL, &c.Version, &c.VersionDate, &c.BooksCount,
			&c.LastImportAt, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan collection: %w", err)
		}
		items = append(items, c)
	}

	return items, rows.Err()
}
