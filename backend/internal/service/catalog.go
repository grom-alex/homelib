package service

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/grom-alex/homelib/backend/internal/models"
	"github.com/grom-alex/homelib/backend/internal/repository"
)

type CatalogService struct {
	pool           *pgxpool.Pool
	bookRepo       *repository.BookRepo
	authorRepo     *repository.AuthorRepo
	genreRepo      *repository.GenreRepo
	seriesRepo     *repository.SeriesRepo
	collectionRepo *repository.CollectionRepo
}

func NewCatalogService(
	pool *pgxpool.Pool,
	bookRepo *repository.BookRepo,
	authorRepo *repository.AuthorRepo,
	genreRepo *repository.GenreRepo,
	seriesRepo *repository.SeriesRepo,
	collectionRepo *repository.CollectionRepo,
) *CatalogService {
	return &CatalogService{
		pool:           pool,
		bookRepo:       bookRepo,
		authorRepo:     authorRepo,
		genreRepo:      genreRepo,
		seriesRepo:     seriesRepo,
		collectionRepo: collectionRepo,
	}
}

func (s *CatalogService) ListBooks(ctx context.Context, f models.BookFilter) ([]models.BookListItem, int, error) {
	f.SetDefaults()
	return s.bookRepo.List(ctx, f)
}

func (s *CatalogService) GetBook(ctx context.Context, id int64) (*models.BookDetail, error) {
	return s.bookRepo.GetByID(ctx, id)
}

func (s *CatalogService) ListAuthors(ctx context.Context, query string, page, limit int) ([]models.AuthorListItem, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit
	return s.authorRepo.ListWithBookCount(ctx, query, limit, offset)
}

func (s *CatalogService) GetAuthor(ctx context.Context, id int64) (*models.AuthorDetail, error) {
	author, err := s.authorRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Get books by this author
	f := models.BookFilter{AuthorID: &id, Page: 1, Limit: 100}
	f.SetDefaults()
	books, total, err := s.bookRepo.List(ctx, f)
	if err != nil {
		return nil, fmt.Errorf("list author books: %w", err)
	}

	return &models.AuthorDetail{
		ID:         author.ID,
		Name:       author.Name,
		Books:      books,
		BooksCount: total,
	}, nil
}

func (s *CatalogService) ListGenres(ctx context.Context) ([]models.GenreTreeItem, error) {
	return s.genreRepo.GetAll(ctx)
}

func (s *CatalogService) ListSeries(ctx context.Context, query string, page, limit int) ([]models.SeriesListItem, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit
	return s.seriesRepo.ListWithBookCount(ctx, query, limit, offset)
}

type Stats struct {
	BooksCount   int      `json:"books_count"`
	AuthorsCount int      `json:"authors_count"`
	GenresCount  int      `json:"genres_count"`
	SeriesCount  int      `json:"series_count"`
	Languages    []string `json:"languages"`
	Formats      []string `json:"formats"`
}

func (s *CatalogService) GetStats(ctx context.Context) (*Stats, error) {
	stats := &Stats{}

	err := s.pool.QueryRow(ctx, "SELECT COUNT(*) FROM books WHERE NOT is_deleted").Scan(&stats.BooksCount)
	if err != nil {
		return nil, fmt.Errorf("count books: %w", err)
	}

	err = s.pool.QueryRow(ctx, "SELECT COUNT(*) FROM authors").Scan(&stats.AuthorsCount)
	if err != nil {
		return nil, fmt.Errorf("count authors: %w", err)
	}

	err = s.pool.QueryRow(ctx, "SELECT COUNT(*) FROM genres").Scan(&stats.GenresCount)
	if err != nil {
		return nil, fmt.Errorf("count genres: %w", err)
	}

	err = s.pool.QueryRow(ctx, "SELECT COUNT(*) FROM series").Scan(&stats.SeriesCount)
	if err != nil {
		return nil, fmt.Errorf("count series: %w", err)
	}

	// Languages
	rows, err := s.pool.Query(ctx,
		"SELECT DISTINCT lang FROM books WHERE NOT is_deleted AND lang != '' ORDER BY lang")
	if err != nil {
		return nil, fmt.Errorf("list languages: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var lang string
		if err := rows.Scan(&lang); err != nil {
			return nil, err
		}
		stats.Languages = append(stats.Languages, lang)
	}

	// Formats
	rows2, err := s.pool.Query(ctx,
		"SELECT DISTINCT format FROM books WHERE NOT is_deleted AND format != '' ORDER BY format")
	if err != nil {
		return nil, fmt.Errorf("list formats: %w", err)
	}
	defer rows2.Close()
	for rows2.Next() {
		var format string
		if err := rows2.Scan(&format); err != nil {
			return nil, err
		}
		stats.Formats = append(stats.Formats, format)
	}

	return stats, nil
}
