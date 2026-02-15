package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/grom-alex/homelib/backend/internal/models"
)

type AuthorRepo struct {
	pool *pgxpool.Pool
}

func NewAuthorRepo(pool *pgxpool.Pool) *AuthorRepo {
	return &AuthorRepo{pool: pool}
}

// UpsertAuthors inserts authors, ignoring duplicates on (name, name_sort).
// Returns a map of name_sort → id for the upserted authors.
func (r *AuthorRepo) UpsertAuthors(ctx context.Context, tx pgx.Tx, authors []models.Author) (map[string]int64, error) {
	if len(authors) == 0 {
		return make(map[string]int64), nil
	}

	result := make(map[string]int64, len(authors))

	for _, a := range authors {
		var id int64
		err := tx.QueryRow(ctx,
			`INSERT INTO authors (name, name_sort)
			 VALUES ($1, $2)
			 ON CONFLICT (name_sort) DO UPDATE SET name = EXCLUDED.name
			 RETURNING id`,
			a.Name, a.NameSort,
		).Scan(&id)
		if err != nil {
			return nil, fmt.Errorf("upsert author %q: %w", a.Name, err)
		}
		result[a.NameSort] = id
	}

	return result, nil
}

func (r *AuthorRepo) GetByID(ctx context.Context, id int64) (*models.Author, error) {
	var a models.Author
	err := r.pool.QueryRow(ctx,
		`SELECT id, name, name_sort, created_at FROM authors WHERE id = $1`, id,
	).Scan(&a.ID, &a.Name, &a.NameSort, &a.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("get author %d: %w", id, err)
	}
	return &a, nil
}

func (r *AuthorRepo) ListWithBookCount(ctx context.Context, query string, limit, offset int) ([]models.AuthorListItem, int, error) {
	var args []any
	where := ""
	argIdx := 1

	if query != "" {
		where = fmt.Sprintf(" WHERE a.name ILIKE '%%' || $%d || '%%'", argIdx)
		args = append(args, query)
		argIdx++
	}

	var total int
	countQuery := "SELECT COUNT(*) FROM authors a" + where
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count authors: %w", err)
	}

	listQuery := fmt.Sprintf(
		`SELECT a.id, a.name, COUNT(ba.book_id) as books_count
		 FROM authors a
		 LEFT JOIN book_authors ba ON ba.author_id = a.id
		 %s
		 GROUP BY a.id, a.name
		 ORDER BY a.name_sort
		 LIMIT $%d OFFSET $%d`,
		where, argIdx, argIdx+1,
	)
	args = append(args, limit, offset)

	rows, err := r.pool.Query(ctx, listQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list authors: %w", err)
	}
	defer rows.Close()

	var items []models.AuthorListItem
	for rows.Next() {
		var item models.AuthorListItem
		if err := rows.Scan(&item.ID, &item.Name, &item.BooksCount); err != nil {
			return nil, 0, fmt.Errorf("scan author: %w", err)
		}
		items = append(items, item)
	}

	return items, total, rows.Err()
}
