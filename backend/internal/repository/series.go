package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/grom-alex/homelib/backend/internal/models"
)

type SeriesRepo struct {
	pool *pgxpool.Pool
}

func NewSeriesRepo(pool *pgxpool.Pool) *SeriesRepo {
	return &SeriesRepo{pool: pool}
}

// UpsertSeries inserts series names, ignoring duplicates.
// Returns a map of name → id.
func (r *SeriesRepo) UpsertSeries(ctx context.Context, tx pgx.Tx, names []string) (map[string]int64, error) {
	if len(names) == 0 {
		return make(map[string]int64), nil
	}

	result := make(map[string]int64, len(names))

	for _, name := range names {
		var id int64
		err := tx.QueryRow(ctx,
			`INSERT INTO series (name)
			 VALUES ($1)
			 ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name
			 RETURNING id`,
			name,
		).Scan(&id)
		if err != nil {
			return nil, fmt.Errorf("upsert series %q: %w", name, err)
		}
		result[name] = id
	}

	return result, nil
}

func (r *SeriesRepo) ListWithBookCount(ctx context.Context, query string, limit, offset int) ([]models.SeriesListItem, int, error) {
	var args []any
	where := ""
	argIdx := 1

	if query != "" {
		where = fmt.Sprintf(" WHERE s.name ILIKE '%%' || $%d || '%%'", argIdx)
		args = append(args, query)
		argIdx++
	}

	var total int
	countQuery := "SELECT COUNT(*) FROM series s" + where
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count series: %w", err)
	}

	listQuery := fmt.Sprintf(
		`SELECT s.id, s.name, COUNT(b.id) as books_count
		 FROM series s
		 LEFT JOIN books b ON b.series_id = s.id
		 %s
		 GROUP BY s.id, s.name
		 ORDER BY s.name
		 LIMIT $%d OFFSET $%d`,
		where, argIdx, argIdx+1,
	)
	args = append(args, limit, offset)

	rows, err := r.pool.Query(ctx, listQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list series: %w", err)
	}
	defer rows.Close()

	var items []models.SeriesListItem
	for rows.Next() {
		var item models.SeriesListItem
		if err := rows.Scan(&item.ID, &item.Name, &item.BooksCount); err != nil {
			return nil, 0, fmt.Errorf("scan series: %w", err)
		}
		items = append(items, item)
	}

	return items, total, rows.Err()
}
