package repository

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/grom-alex/homelib/backend/internal/models"
)

func TestReadingProgressRepo_Get_Found(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := NewReadingProgressRepo(mock)
	now := time.Now()

	mock.ExpectQuery("SELECT .+ FROM reading_progress WHERE user_id = \\$1 AND book_id = \\$2").
		WithArgs("user-1", int64(42)).
		WillReturnRows(pgxmock.NewRows([]string{
			"id", "user_id", "book_id", "chapter_id",
			"chapter_progress", "total_progress", "device", "updated_at",
		}).AddRow(int64(1), "user-1", int64(42), "ch3", 55, 30, "desktop", now))

	p, err := repo.Get(context.Background(), "user-1", 42)
	require.NoError(t, err)
	require.NotNil(t, p)
	assert.Equal(t, int64(1), p.ID)
	assert.Equal(t, "user-1", p.UserID)
	assert.Equal(t, int64(42), p.BookID)
	assert.Equal(t, "ch3", p.ChapterID)
	assert.Equal(t, 55, p.ChapterProgress)
	assert.Equal(t, 30, p.TotalProgress)
	assert.Equal(t, "desktop", p.Device)
	assert.Equal(t, now, p.UpdatedAt)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestReadingProgressRepo_Get_NotFound(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := NewReadingProgressRepo(mock)

	mock.ExpectQuery("SELECT .+ FROM reading_progress WHERE user_id = \\$1 AND book_id = \\$2").
		WithArgs("user-1", int64(999)).
		WillReturnError(pgx.ErrNoRows)

	p, err := repo.Get(context.Background(), "user-1", 999)
	require.NoError(t, err)
	assert.Nil(t, p)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestReadingProgressRepo_Get_DBError(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := NewReadingProgressRepo(mock)

	mock.ExpectQuery("SELECT .+ FROM reading_progress WHERE user_id = \\$1 AND book_id = \\$2").
		WithArgs("user-1", int64(1)).
		WillReturnError(fmt.Errorf("connection refused"))

	p, err := repo.Get(context.Background(), "user-1", 1)
	assert.Error(t, err)
	assert.Nil(t, p)
	assert.Contains(t, err.Error(), "get reading progress")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestReadingProgressRepo_Upsert_Insert(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := NewReadingProgressRepo(mock)
	now := time.Now()

	p := &models.ReadingProgress{
		UserID:          "user-1",
		BookID:          42,
		ChapterID:       "ch2",
		ChapterProgress: 75,
		TotalProgress:   50,
		Device:          "mobile",
	}

	mock.ExpectQuery("INSERT INTO reading_progress .+ ON CONFLICT .+ RETURNING id, updated_at").
		WithArgs("user-1", int64(42), "ch2", 75, 50, "mobile").
		WillReturnRows(pgxmock.NewRows([]string{"id", "updated_at"}).AddRow(int64(1), now))

	err = repo.Upsert(context.Background(), p)
	require.NoError(t, err)
	assert.Equal(t, int64(1), p.ID)
	assert.Equal(t, now, p.UpdatedAt)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestReadingProgressRepo_Upsert_Update(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := NewReadingProgressRepo(mock)
	now := time.Now()

	p := &models.ReadingProgress{
		UserID:          "user-1",
		BookID:          42,
		ChapterID:       "ch5",
		ChapterProgress: 90,
		TotalProgress:   80,
		Device:          "desktop",
	}

	mock.ExpectQuery("INSERT INTO reading_progress .+ ON CONFLICT .+ RETURNING id, updated_at").
		WithArgs("user-1", int64(42), "ch5", 90, 80, "desktop").
		WillReturnRows(pgxmock.NewRows([]string{"id", "updated_at"}).AddRow(int64(1), now))

	err = repo.Upsert(context.Background(), p)
	require.NoError(t, err)
	assert.Equal(t, now, p.UpdatedAt)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestReadingProgressRepo_Upsert_DBError(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := NewReadingProgressRepo(mock)

	p := &models.ReadingProgress{
		UserID:    "user-1",
		BookID:    42,
		ChapterID: "ch1",
	}

	mock.ExpectQuery("INSERT INTO reading_progress .+ ON CONFLICT .+ RETURNING id, updated_at").
		WithArgs("user-1", int64(42), "ch1", 0, 0, "").
		WillReturnError(fmt.Errorf("foreign key violation"))

	err = repo.Upsert(context.Background(), p)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "upsert reading progress")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestReadingProgressRepo_GetByUser_MultipleEntries(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := NewReadingProgressRepo(mock)
	now := time.Now()
	earlier := now.Add(-time.Hour)

	mock.ExpectQuery("SELECT .+ FROM reading_progress WHERE user_id = \\$1 ORDER BY updated_at DESC").
		WithArgs("user-1").
		WillReturnRows(pgxmock.NewRows([]string{
			"id", "user_id", "book_id", "chapter_id",
			"chapter_progress", "total_progress", "device", "updated_at",
		}).
			AddRow(int64(2), "user-1", int64(100), "ch3", 50, 25, "desktop", now).
			AddRow(int64(1), "user-1", int64(42), "ch1", 10, 5, "mobile", earlier))

	result, err := repo.GetByUser(context.Background(), "user-1")
	require.NoError(t, err)
	require.Len(t, result, 2)
	assert.Equal(t, int64(100), result[0].BookID)
	assert.Equal(t, int64(42), result[1].BookID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestReadingProgressRepo_GetByUser_Empty(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := NewReadingProgressRepo(mock)

	mock.ExpectQuery("SELECT .+ FROM reading_progress WHERE user_id = \\$1 ORDER BY updated_at DESC").
		WithArgs("user-new").
		WillReturnRows(pgxmock.NewRows([]string{
			"id", "user_id", "book_id", "chapter_id",
			"chapter_progress", "total_progress", "device", "updated_at",
		}))

	result, err := repo.GetByUser(context.Background(), "user-new")
	require.NoError(t, err)
	assert.Empty(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestReadingProgressRepo_GetByUser_DBError(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := NewReadingProgressRepo(mock)

	mock.ExpectQuery("SELECT .+ FROM reading_progress WHERE user_id = \\$1 ORDER BY updated_at DESC").
		WithArgs("user-1").
		WillReturnError(fmt.Errorf("timeout"))

	result, err := repo.GetByUser(context.Background(), "user-1")
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "get reading progress by user")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestNewReadingProgressRepo(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := NewReadingProgressRepo(mock)
	assert.NotNil(t, repo)
}
