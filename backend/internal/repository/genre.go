package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/grom-alex/homelib/backend/internal/glst"
	"github.com/grom-alex/homelib/backend/internal/models"
)

type GenreRepo struct {
	pool *pgxpool.Pool
}

func NewGenreRepo(pool *pgxpool.Pool) *GenreRepo {
	return &GenreRepo{pool: pool}
}

// UpsertGenres inserts genre codes, ignoring duplicates.
// Uses pgx.Batch for a single network round-trip.
// Returns a map of code → id.
func (r *GenreRepo) UpsertGenres(ctx context.Context, tx pgx.Tx, codes []string) (map[string]int, error) {
	if len(codes) == 0 {
		return make(map[string]int), nil
	}

	const upsertSQL = `INSERT INTO genres (code, name)
		 VALUES ($1, $1)
		 ON CONFLICT (code) DO UPDATE SET code = EXCLUDED.code
		 RETURNING id`

	batch := &pgx.Batch{}
	for _, code := range codes {
		batch.Queue(upsertSQL, code)
	}

	br := tx.SendBatch(ctx, batch)
	defer func() { _ = br.Close() }()

	result := make(map[string]int, len(codes))
	for _, code := range codes {
		var id int
		if err := br.QueryRow().Scan(&id); err != nil {
			return nil, fmt.Errorf("upsert genre %q: %w", code, err)
		}
		result[code] = id
	}

	return result, nil
}

func (r *GenreRepo) GetAll(ctx context.Context) ([]models.GenreTreeItem, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT g.id, g.code, g.name, g.position, g.parent_id,
				COUNT(bg.book_id) as books_count
		 FROM genres g
		 LEFT JOIN book_genres bg ON bg.genre_id = g.id
		 WHERE g.is_active = TRUE
		 GROUP BY g.id, g.code, g.name, g.position, g.parent_id
		 ORDER BY g.sort_order`)
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
		if err := rows.Scan(&g.ID, &g.Code, &g.Name, &g.Position, &g.ParentID, &g.BooksCount); err != nil {
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

	// Build pointer-based tree, then convert to values via recursive copy.
	// This avoids the copy-before-mutation bug when children are value types.
	type pNode struct {
		ID         int
		Code       string
		Name       string
		Position   string
		BooksCount int
		ParentID   *int
		children   []*pNode
	}

	nodeMap := make(map[int]*pNode, len(flat))
	nodes := make([]pNode, len(flat))
	for i := range flat {
		nodes[i] = pNode{
			ID: flat[i].ID, Code: flat[i].Code, Name: flat[i].Name,
			Position: flat[i].Position, BooksCount: flat[i].BooksCount,
			ParentID: flat[i].ParentID,
		}
		nodeMap[flat[i].ID] = &nodes[i]
	}

	// Assign children + accumulate book counts bottom-up
	var rootNodes []*pNode
	for i := range nodes {
		n := &nodes[i]
		if n.ParentID != nil {
			if parent, ok := nodeMap[*n.ParentID]; ok {
				parent.children = append(parent.children, n)
			} else {
				rootNodes = append(rootNodes, n)
			}
		} else {
			rootNodes = append(rootNodes, n)
		}
	}

	// Recursive conversion: children are fully resolved via pointers before copying
	var convert func(n *pNode) models.GenreTreeItem
	convert = func(n *pNode) models.GenreTreeItem {
		item := models.GenreTreeItem{
			ID: n.ID, Code: n.Code, Name: n.Name,
			Position: n.Position, BooksCount: n.BooksCount,
		}
		for _, c := range n.children {
			child := convert(c)
			item.BooksCount += child.BooksCount
			item.Children = append(item.Children, child)
		}
		return item
	}

	roots := make([]models.GenreTreeItem, 0, len(rootNodes))
	for _, rn := range rootNodes {
		roots = append(roots, convert(rn))
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

// LoadTree loads genre entries from a parsed GLST file into the database.
// It deactivates all existing genres, then upserts entries from the file,
// resolving parent_id by position lookup. All operations run in a single transaction.
func (r *GenreRepo) LoadTree(ctx context.Context, entries []glst.GenreEntry) (int, error) {
	if len(entries) == 0 {
		return 0, nil
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// Deactivate all genres before loading new tree
	if _, err := tx.Exec(ctx, `UPDATE genres SET is_active = FALSE`); err != nil {
		return 0, fmt.Errorf("deactivate genres: %w", err)
	}

	// Build position → parent_id mapping incrementally.
	// We process entries in order, so parent positions are resolved before children.
	positionToID := make(map[string]int, len(entries))
	loaded := 0

	for i, e := range entries {
		var parentID *int
		if e.ParentPosition != "" {
			if pid, ok := positionToID[e.ParentPosition]; ok {
				parentID = &pid
			} else {
				// Parent not yet loaded — skip (should not happen
				// if parser validates orphans, but be defensive)
				continue
			}
		}

		var id int
		err := tx.QueryRow(ctx,
			`INSERT INTO genres (code, name, position, parent_id, sort_order, is_active)
			 VALUES ($1, $2, $3, $4, $5, TRUE)
			 ON CONFLICT (position) WHERE position IS NOT NULL
			 DO UPDATE SET code = EXCLUDED.code,
			              name = EXCLUDED.name,
			              parent_id = EXCLUDED.parent_id,
			              sort_order = EXCLUDED.sort_order,
			              is_active = TRUE
			 RETURNING id`,
			e.Code, e.Name, e.Position, parentID, i,
		).Scan(&id)
		if err != nil {
			return loaded, fmt.Errorf("upsert genre %q (position %s): %w", e.Code, e.Position, err)
		}

		positionToID[e.Position] = id
		loaded++
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("commit genre tree: %w", err)
	}

	return loaded, nil
}

// GetUnsortedGenreID returns the ID of the «Неотсортированное» genre (position='0.0').
func (r *GenreRepo) GetUnsortedGenreID(ctx context.Context) (int, error) {
	var id int
	err := r.pool.QueryRow(ctx,
		`SELECT id FROM genres WHERE position = '0.0' LIMIT 1`,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("get unsorted genre: %w", err)
	}
	return id, nil
}

// GetIDsByCodes returns a mapping of genre code to all matching genre IDs.
// One code can map to multiple IDs when duplicates exist across tree levels.
func (r *GenreRepo) GetIDsByCodes(ctx context.Context, codes []string) (map[string][]int, error) {
	if len(codes) == 0 {
		return make(map[string][]int), nil
	}

	rows, err := r.pool.Query(ctx,
		`SELECT code, id FROM genres WHERE code = ANY($1) AND is_active = TRUE`,
		codes,
	)
	if err != nil {
		return nil, fmt.Errorf("get genre IDs by codes: %w", err)
	}
	defer rows.Close()

	result := make(map[string][]int, len(codes))
	for rows.Next() {
		var code string
		var id int
		if err := rows.Scan(&code, &id); err != nil {
			return nil, fmt.Errorf("scan genre code/id: %w", err)
		}
		result[code] = append(result[code], id)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}
