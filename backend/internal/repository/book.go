package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/grom-alex/homelib/backend/internal/models"
)

type BookRepo struct {
	pool *pgxpool.Pool
}

func NewBookRepo(pool *pgxpool.Pool) *BookRepo {
	return &BookRepo{pool: pool}
}

// BatchUpsert inserts or updates books using ON CONFLICT (collection_id, lib_id).
// Returns the number of inserted and updated books.
func (r *BookRepo) BatchUpsert(ctx context.Context, tx pgx.Tx, books []models.Book) (inserted, updated int, err error) {
	for i := range books {
		var id int64
		var isNew bool

		err := tx.QueryRow(ctx,
			`INSERT INTO books (collection_id, title, lang, year, format, file_size,
				archive_name, file_in_archive, series_id, series_num, series_type,
				lib_id, lib_rate, is_deleted, description, keywords, date_added)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
			 ON CONFLICT (collection_id, lib_id) DO UPDATE SET
				title = EXCLUDED.title,
				lang = EXCLUDED.lang,
				year = EXCLUDED.year,
				format = EXCLUDED.format,
				file_size = EXCLUDED.file_size,
				archive_name = EXCLUDED.archive_name,
				file_in_archive = EXCLUDED.file_in_archive,
				series_id = EXCLUDED.series_id,
				series_num = EXCLUDED.series_num,
				series_type = EXCLUDED.series_type,
				lib_rate = EXCLUDED.lib_rate,
				is_deleted = EXCLUDED.is_deleted,
				description = EXCLUDED.description,
				keywords = EXCLUDED.keywords,
				date_added = EXCLUDED.date_added,
				updated_at = NOW()
			 RETURNING id, (xmax = 0) as is_new`,
			books[i].CollectionID, books[i].Title, books[i].Lang, books[i].Year, books[i].Format, books[i].FileSize,
			books[i].ArchiveName, books[i].FileInArchive, books[i].SeriesID, books[i].SeriesNum, books[i].SeriesType,
			books[i].LibID, books[i].LibRate, books[i].IsDeleted, books[i].Description, books[i].Keywords, books[i].DateAdded,
		).Scan(&id, &isNew)
		if err != nil {
			return inserted, updated, fmt.Errorf("upsert book lib_id=%s: %w", books[i].LibID, err)
		}

		books[i].ID = id
		if isNew {
			inserted++
		} else {
			updated++
		}
	}

	return inserted, updated, nil
}

// SetBookAuthors replaces all author associations for a book.
func (r *BookRepo) SetBookAuthors(ctx context.Context, tx pgx.Tx, bookID int64, authorIDs []int64) error {
	_, err := tx.Exec(ctx, `DELETE FROM book_authors WHERE book_id = $1`, bookID)
	if err != nil {
		return fmt.Errorf("delete book_authors: %w", err)
	}

	for _, authorID := range authorIDs {
		_, err := tx.Exec(ctx,
			`INSERT INTO book_authors (book_id, author_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
			bookID, authorID,
		)
		if err != nil {
			return fmt.Errorf("insert book_author: %w", err)
		}
	}

	return nil
}

// SetBookGenres replaces all genre associations for a book.
func (r *BookRepo) SetBookGenres(ctx context.Context, tx pgx.Tx, bookID int64, genreIDs []int) error {
	_, err := tx.Exec(ctx, `DELETE FROM book_genres WHERE book_id = $1`, bookID)
	if err != nil {
		return fmt.Errorf("delete book_genres: %w", err)
	}

	for _, genreID := range genreIDs {
		_, err := tx.Exec(ctx,
			`INSERT INTO book_genres (book_id, genre_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
			bookID, genreID,
		)
		if err != nil {
			return fmt.Errorf("insert book_genre: %w", err)
		}
	}

	return nil
}

func (r *BookRepo) GetByID(ctx context.Context, id int64) (*models.BookDetail, error) {
	var b models.BookDetail
	err := r.pool.QueryRow(ctx,
		`SELECT b.id, b.title, b.lang, b.year, b.format, b.file_size,
				b.lib_rate, b.is_deleted, b.description, b.keywords, b.date_added
		 FROM books b WHERE b.id = $1`, id,
	).Scan(&b.ID, &b.Title, &b.Lang, &b.Year, &b.Format, &b.FileSize,
		&b.LibRate, &b.IsDeleted, &b.Description, &b.Keywords, &b.DateAdded)
	if err != nil {
		return nil, fmt.Errorf("get book %d: %w", id, err)
	}

	// Load authors
	rows, err := r.pool.Query(ctx,
		`SELECT a.id, a.name FROM authors a
		 JOIN book_authors ba ON ba.author_id = a.id
		 WHERE ba.book_id = $1 ORDER BY a.name_sort`, id)
	if err != nil {
		return nil, fmt.Errorf("get book authors: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var ref models.BookAuthorRef
		if err := rows.Scan(&ref.ID, &ref.Name); err != nil {
			return nil, err
		}
		b.Authors = append(b.Authors, ref)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Load genres
	rows2, err := r.pool.Query(ctx,
		`SELECT g.id, g.code, g.name FROM genres g
		 JOIN book_genres bg ON bg.genre_id = g.id
		 WHERE bg.book_id = $1 ORDER BY g.name`, id)
	if err != nil {
		return nil, fmt.Errorf("get book genres: %w", err)
	}
	defer rows2.Close()
	for rows2.Next() {
		var ref models.BookGenreDetailRef
		if err := rows2.Scan(&ref.ID, &ref.Code, &ref.Name); err != nil {
			return nil, err
		}
		b.Genres = append(b.Genres, ref)
	}
	if err := rows2.Err(); err != nil {
		return nil, err
	}

	// Load series
	var seriesID *int64
	var seriesNum *int
	var seriesType *string
	err = r.pool.QueryRow(ctx,
		`SELECT b.series_id, b.series_num, b.series_type FROM books b WHERE b.id = $1`, id,
	).Scan(&seriesID, &seriesNum, &seriesType)
	if err == nil && seriesID != nil {
		var ref models.BookSeriesDetailRef
		ref.Num = seriesNum
		ref.Type = seriesType
		err = r.pool.QueryRow(ctx,
			`SELECT id, name FROM series WHERE id = $1`, *seriesID,
		).Scan(&ref.ID, &ref.Name)
		if err == nil {
			b.Series = &ref
		}
	}

	// Load collection
	var collID *int
	_ = r.pool.QueryRow(ctx,
		`SELECT collection_id FROM books WHERE id = $1`, id,
	).Scan(&collID)
	if collID != nil {
		var ref models.BookCollectionRef
		err = r.pool.QueryRow(ctx,
			`SELECT id, name FROM collections WHERE id = $1`, *collID,
		).Scan(&ref.ID, &ref.Name)
		if err == nil {
			b.Collection = &ref
		}
	}

	return &b, nil
}

// List returns books matching the filter with pagination.
func (r *BookRepo) List(ctx context.Context, f models.BookFilter) ([]models.BookListItem, int, error) {
	var conditions []string
	var args []any
	argIdx := 1

	if f.Query != "" {
		conditions = append(conditions, fmt.Sprintf(
			"b.search_vector @@ plainto_tsquery('russian', $%d)", argIdx))
		args = append(args, f.Query)
		argIdx++
	}
	if f.AuthorID != nil {
		conditions = append(conditions, fmt.Sprintf(
			"EXISTS (SELECT 1 FROM book_authors ba WHERE ba.book_id = b.id AND ba.author_id = $%d)", argIdx))
		args = append(args, *f.AuthorID)
		argIdx++
	}
	if f.GenreID != nil {
		conditions = append(conditions, fmt.Sprintf(
			"EXISTS (SELECT 1 FROM book_genres bg WHERE bg.book_id = b.id AND bg.genre_id = $%d)", argIdx))
		args = append(args, *f.GenreID)
		argIdx++
	}
	if f.SeriesID != nil {
		conditions = append(conditions, fmt.Sprintf("b.series_id = $%d", argIdx))
		args = append(args, *f.SeriesID)
		argIdx++
	}
	if f.Lang != "" {
		conditions = append(conditions, fmt.Sprintf("b.lang = $%d", argIdx))
		args = append(args, f.Lang)
		argIdx++
	}
	if f.Format != "" {
		conditions = append(conditions, fmt.Sprintf("b.format = $%d", argIdx))
		args = append(args, f.Format)
		argIdx++
	}

	where := ""
	if len(conditions) > 0 {
		where = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Count
	var total int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM books b %s", where)
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count books: %w", err)
	}

	// Sort
	orderCol := "b.title"
	switch f.Sort {
	case "year":
		orderCol = "b.year"
	case "added_at":
		orderCol = "b.added_at"
	case "lib_rate":
		orderCol = "b.lib_rate"
	}
	orderDir := "ASC"
	if strings.EqualFold(f.Order, "desc") {
		orderDir = "DESC"
	}

	listQuery := fmt.Sprintf(
		`SELECT b.id, b.title, b.lang, b.year, b.format, b.file_size, b.lib_rate, b.is_deleted
		 FROM books b %s
		 ORDER BY %s %s NULLS LAST
		 LIMIT $%d OFFSET $%d`,
		where, orderCol, orderDir, argIdx, argIdx+1,
	)
	args = append(args, f.Limit, f.Offset())

	rows, err := r.pool.Query(ctx, listQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list books: %w", err)
	}
	defer rows.Close()

	var items []models.BookListItem
	for rows.Next() {
		var item models.BookListItem
		if err := rows.Scan(&item.ID, &item.Title, &item.Lang, &item.Year,
			&item.Format, &item.FileSize, &item.LibRate, &item.IsDeleted); err != nil {
			return nil, 0, fmt.Errorf("scan book: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	// Batch load authors, genres, series for all books (3 queries instead of 3*N)
	if len(items) > 0 {
		bookIDs := make([]int64, len(items))
		for i := range items {
			bookIDs[i] = items[i].ID
		}

		authorsMap, err := r.getBookAuthorRefsBatch(ctx, bookIDs)
		if err != nil {
			return nil, 0, fmt.Errorf("batch load authors: %w", err)
		}
		genresMap, err := r.getBookGenreRefsBatch(ctx, bookIDs)
		if err != nil {
			return nil, 0, fmt.Errorf("batch load genres: %w", err)
		}
		seriesMap, err := r.getBookSeriesRefsBatch(ctx, bookIDs)
		if err != nil {
			return nil, 0, fmt.Errorf("batch load series: %w", err)
		}

		for i := range items {
			items[i].Authors = authorsMap[items[i].ID]
			items[i].Genres = genresMap[items[i].ID]
			items[i].Series = seriesMap[items[i].ID]
		}
	}

	return items, total, nil
}

// GetBookForDownload returns archive and file info for downloading.
func (r *BookRepo) GetBookForDownload(ctx context.Context, id int64) (archiveName, fileInArchive, format string, err error) {
	err = r.pool.QueryRow(ctx,
		`SELECT archive_name, file_in_archive, format FROM books WHERE id = $1 AND NOT is_deleted`, id,
	).Scan(&archiveName, &fileInArchive, &format)
	if err != nil {
		err = fmt.Errorf("get book download info %d: %w", id, err)
	}
	return
}

func (r *BookRepo) getBookAuthorRefsBatch(ctx context.Context, bookIDs []int64) (map[int64][]models.BookAuthorRef, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT ba.book_id, a.id, a.name FROM authors a
		 JOIN book_authors ba ON ba.author_id = a.id
		 WHERE ba.book_id = ANY($1) ORDER BY a.name_sort`, bookIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[int64][]models.BookAuthorRef)
	for rows.Next() {
		var bookID int64
		var ref models.BookAuthorRef
		if err := rows.Scan(&bookID, &ref.ID, &ref.Name); err != nil {
			return nil, err
		}
		result[bookID] = append(result[bookID], ref)
	}
	return result, rows.Err()
}

func (r *BookRepo) getBookGenreRefsBatch(ctx context.Context, bookIDs []int64) (map[int64][]models.BookGenreRef, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT bg.book_id, g.id, g.name FROM genres g
		 JOIN book_genres bg ON bg.genre_id = g.id
		 WHERE bg.book_id = ANY($1) ORDER BY g.name`, bookIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[int64][]models.BookGenreRef)
	for rows.Next() {
		var bookID int64
		var ref models.BookGenreRef
		if err := rows.Scan(&bookID, &ref.ID, &ref.Name); err != nil {
			return nil, err
		}
		result[bookID] = append(result[bookID], ref)
	}
	return result, rows.Err()
}

func (r *BookRepo) getBookSeriesRefsBatch(ctx context.Context, bookIDs []int64) (map[int64]*models.BookSeriesRef, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT b.id, s.id, s.name, b.series_num FROM books b
		 JOIN series s ON s.id = b.series_id
		 WHERE b.id = ANY($1) AND b.series_id IS NOT NULL`, bookIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[int64]*models.BookSeriesRef)
	for rows.Next() {
		var bookID int64
		var ref models.BookSeriesRef
		if err := rows.Scan(&bookID, &ref.ID, &ref.Name, &ref.Num); err != nil {
			return nil, err
		}
		result[bookID] = &ref
	}
	return result, rows.Err()
}
