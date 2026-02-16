package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/grom-alex/homelib/backend/internal/models"
)

type GenreRepo struct {
	pool *pgxpool.Pool
}

func NewGenreRepo(pool *pgxpool.Pool) *GenreRepo {
	return &GenreRepo{pool: pool}
}

// UpsertGenres inserts genre codes, ignoring duplicates.
// Returns a map of code → id.
func (r *GenreRepo) UpsertGenres(ctx context.Context, tx pgx.Tx, codes []string) (map[string]int, error) {
	if len(codes) == 0 {
		return make(map[string]int), nil
	}

	result := make(map[string]int, len(codes))

	for _, code := range codes {
		var id int
		err := tx.QueryRow(ctx,
			`INSERT INTO genres (code, name)
			 VALUES ($1, $1)
			 ON CONFLICT (code) DO UPDATE SET code = EXCLUDED.code
			 RETURNING id`,
			code,
		).Scan(&id)
		if err != nil {
			return nil, fmt.Errorf("upsert genre %q: %w", code, err)
		}
		result[code] = id
	}

	return result, nil
}

func (r *GenreRepo) GetAll(ctx context.Context) ([]models.GenreTreeItem, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT g.id, g.code, g.name, g.parent_id, g.meta_group,
				COUNT(bg.book_id) as books_count
		 FROM genres g
		 LEFT JOIN book_genres bg ON bg.genre_id = g.id
		 GROUP BY g.id, g.code, g.name, g.parent_id, g.meta_group
		 ORDER BY g.meta_group, g.name`)
	if err != nil {
		return nil, fmt.Errorf("query genres: %w", err)
	}
	defer rows.Close()

	type flatGenre struct {
		models.GenreTreeItem
		ParentID *int
	}

	var flat []flatGenre

	for rows.Next() {
		var g flatGenre
		if err := rows.Scan(&g.ID, &g.Code, &g.Name, &g.ParentID, &g.MetaGroup, &g.BooksCount); err != nil {
			return nil, fmt.Errorf("scan genre: %w", err)
		}
		flat = append(flat, g)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	genreMap := make(map[int]*flatGenre, len(flat))
	for i := range flat {
		genreMap[flat[i].ID] = &flat[i]
	}

	// Build tree
	var roots []models.GenreTreeItem
	for i := range flat {
		g := &flat[i]
		if g.ParentID != nil {
			if parent, ok := genreMap[*g.ParentID]; ok {
				parent.Children = append(parent.Children, g.GenreTreeItem)
				continue
			}
		}
		roots = append(roots, g.GenreTreeItem)
	}

	return roots, nil
}

func (r *GenreRepo) GetByID(ctx context.Context, id int) (*models.Genre, error) {
	var g models.Genre
	err := r.pool.QueryRow(ctx,
		`SELECT id, code, name, parent_id, meta_group FROM genres WHERE id = $1`, id,
	).Scan(&g.ID, &g.Code, &g.Name, &g.ParentID, &g.MetaGroup)
	if err != nil {
		return nil, fmt.Errorf("get genre %d: %w", id, err)
	}
	return &g, nil
}
