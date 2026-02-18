package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/grom-alex/homelib/backend/internal/models"
)

type ReadingProgressRepo struct {
	pool Pool
}

func NewReadingProgressRepo(pool Pool) *ReadingProgressRepo {
	return &ReadingProgressRepo{pool: pool}
}

// Get returns the reading progress for a user and book, or nil if not found.
func (r *ReadingProgressRepo) Get(ctx context.Context, userID string, bookID int64) (*models.ReadingProgress, error) {
	var p models.ReadingProgress
	err := r.pool.QueryRow(ctx,
		`SELECT id, user_id, book_id, chapter_id, chapter_progress, total_progress, device, updated_at
		 FROM reading_progress WHERE user_id = $1 AND book_id = $2`,
		userID, bookID,
	).Scan(&p.ID, &p.UserID, &p.BookID, &p.ChapterID, &p.ChapterProgress, &p.TotalProgress, &p.Device, &p.UpdatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get reading progress: %w", err)
	}
	return &p, nil
}

// Upsert inserts or updates reading progress using ON CONFLICT (user_id, book_id).
func (r *ReadingProgressRepo) Upsert(ctx context.Context, p *models.ReadingProgress) error {
	err := r.pool.QueryRow(ctx,
		`INSERT INTO reading_progress (user_id, book_id, chapter_id, chapter_progress, total_progress, device, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, NOW())
		 ON CONFLICT (user_id, book_id) DO UPDATE SET
			chapter_id = EXCLUDED.chapter_id,
			chapter_progress = EXCLUDED.chapter_progress,
			total_progress = EXCLUDED.total_progress,
			device = EXCLUDED.device,
			updated_at = NOW()
		 RETURNING id, updated_at`,
		p.UserID, p.BookID, p.ChapterID, p.ChapterProgress, p.TotalProgress, p.Device,
	).Scan(&p.ID, &p.UpdatedAt)

	if err != nil {
		return fmt.Errorf("upsert reading progress: %w", err)
	}
	return nil
}

// GetByUser returns all reading progress entries for a user.
func (r *ReadingProgressRepo) GetByUser(ctx context.Context, userID string) ([]models.ReadingProgress, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, user_id, book_id, chapter_id, chapter_progress, total_progress, device, updated_at
		 FROM reading_progress WHERE user_id = $1 ORDER BY updated_at DESC`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("get reading progress by user: %w", err)
	}
	defer rows.Close()

	var result []models.ReadingProgress
	for rows.Next() {
		var p models.ReadingProgress
		if err := rows.Scan(&p.ID, &p.UserID, &p.BookID, &p.ChapterID, &p.ChapterProgress, &p.TotalProgress, &p.Device, &p.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan reading progress: %w", err)
		}
		result = append(result, p)
	}
	return result, rows.Err()
}
