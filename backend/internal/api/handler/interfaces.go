package handler

import (
	"context"
	"encoding/json"

	"github.com/grom-alex/homelib/backend/internal/bookfile"
	"github.com/grom-alex/homelib/backend/internal/models"
	"github.com/grom-alex/homelib/backend/internal/service"
)

// AuthServicer is the interface that auth handlers need from the auth service.
type AuthServicer interface {
	Register(ctx context.Context, input models.CreateUserInput) (*service.AuthResult, error)
	Login(ctx context.Context, input models.LoginInput) (*service.AuthResult, error)
	RefreshToken(ctx context.Context, tokenStr string) (*service.AuthResult, error)
	Logout(ctx context.Context, tokenStr string) error
}

// CatalogServicer is the interface that catalog handlers need from the catalog service.
type CatalogServicer interface {
	ListBooks(ctx context.Context, f models.BookFilter) ([]models.BookListItem, int, error)
	GetBook(ctx context.Context, id int64) (*models.BookDetail, error)
	ListAuthors(ctx context.Context, query string, page, limit int) ([]models.AuthorListItem, int, error)
	GetAuthor(ctx context.Context, id int64) (*models.AuthorDetail, error)
	ListGenres(ctx context.Context) ([]models.GenreTreeItem, error)
	ListSeries(ctx context.Context, query string, page, limit int) ([]models.SeriesListItem, int, error)
	GetStats(ctx context.Context) (*service.Stats, error)
}

// DownloadServicer is the interface that the download handler needs.
type DownloadServicer interface {
	DownloadBook(ctx context.Context, id int64) (*service.DownloadResult, error)
}

// ImportServicer is the interface that admin handlers need from the import service.
type ImportServicer interface {
	StartImport(parentCtx ...context.Context) error
	GetStatus() models.ImportStatus
	CancelImport()
}

// ReaderServicer is the interface that reader handlers need from the reader service.
type ReaderServicer interface {
	GetBookContent(ctx context.Context, bookID int64) (*bookfile.BookContent, error)
	GetChapter(ctx context.Context, bookID int64, chapterID string) (*bookfile.ChapterContent, error)
	GetBookImage(ctx context.Context, bookID int64, imageID string) (*bookfile.ImageData, error)
}

// ProgressRepoer is the interface that progress handlers need from the reading progress repo.
type ProgressRepoer interface {
	Get(ctx context.Context, userID string, bookID int64) (*models.ReadingProgress, error)
	Upsert(ctx context.Context, p *models.ReadingProgress) error
}

// SettingsRepoer is the interface that settings handlers need from the user repo.
type SettingsRepoer interface {
	GetSettings(ctx context.Context, userID string) (json.RawMessage, error)
	UpdateSettings(ctx context.Context, userID string, patch json.RawMessage) (json.RawMessage, error)
}
