package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/grom-alex/homelib/backend/internal/bookfile"
	"github.com/grom-alex/homelib/backend/internal/models"
	"github.com/grom-alex/homelib/backend/internal/service"
)

// --- Auth service mock ---

type mockAuthService struct {
	registerFn     func(ctx context.Context, input models.CreateUserInput) (*service.AuthResult, error)
	loginFn        func(ctx context.Context, input models.LoginInput) (*service.AuthResult, error)
	refreshTokenFn func(ctx context.Context, tokenStr string) (*service.AuthResult, error)
	logoutFn       func(ctx context.Context, tokenStr string) error
}

func (m *mockAuthService) Register(ctx context.Context, input models.CreateUserInput) (*service.AuthResult, error) {
	if m.registerFn != nil {
		return m.registerFn(ctx, input)
	}
	return nil, fmt.Errorf("not implemented")
}

func (m *mockAuthService) Login(ctx context.Context, input models.LoginInput) (*service.AuthResult, error) {
	if m.loginFn != nil {
		return m.loginFn(ctx, input)
	}
	return nil, fmt.Errorf("not implemented")
}

func (m *mockAuthService) RefreshToken(ctx context.Context, tokenStr string) (*service.AuthResult, error) {
	if m.refreshTokenFn != nil {
		return m.refreshTokenFn(ctx, tokenStr)
	}
	return nil, fmt.Errorf("not implemented")
}

func (m *mockAuthService) Logout(ctx context.Context, tokenStr string) error {
	if m.logoutFn != nil {
		return m.logoutFn(ctx, tokenStr)
	}
	return nil
}

// --- Catalog service mock ---

type mockCatalogService struct {
	listBooksFn   func(ctx context.Context, f models.BookFilter) ([]models.BookListItem, int, error)
	getBookFn     func(ctx context.Context, id int64) (*models.BookDetail, error)
	listAuthorsFn func(ctx context.Context, query string, page, limit int) ([]models.AuthorListItem, int, error)
	getAuthorFn   func(ctx context.Context, id int64) (*models.AuthorDetail, error)
	listGenresFn  func(ctx context.Context) ([]models.GenreTreeItem, error)
	listSeriesFn  func(ctx context.Context, query string, page, limit int) ([]models.SeriesListItem, int, error)
	getStatsFn    func(ctx context.Context) (*service.Stats, error)
}

func (m *mockCatalogService) ListBooks(ctx context.Context, f models.BookFilter) ([]models.BookListItem, int, error) {
	if m.listBooksFn != nil {
		return m.listBooksFn(ctx, f)
	}
	return nil, 0, fmt.Errorf("not implemented")
}

func (m *mockCatalogService) GetBook(ctx context.Context, id int64) (*models.BookDetail, error) {
	if m.getBookFn != nil {
		return m.getBookFn(ctx, id)
	}
	return nil, fmt.Errorf("not implemented")
}

func (m *mockCatalogService) ListAuthors(ctx context.Context, query string, page, limit int) ([]models.AuthorListItem, int, error) {
	if m.listAuthorsFn != nil {
		return m.listAuthorsFn(ctx, query, page, limit)
	}
	return nil, 0, fmt.Errorf("not implemented")
}

func (m *mockCatalogService) GetAuthor(ctx context.Context, id int64) (*models.AuthorDetail, error) {
	if m.getAuthorFn != nil {
		return m.getAuthorFn(ctx, id)
	}
	return nil, fmt.Errorf("not implemented")
}

func (m *mockCatalogService) ListGenres(ctx context.Context) ([]models.GenreTreeItem, error) {
	if m.listGenresFn != nil {
		return m.listGenresFn(ctx)
	}
	return nil, fmt.Errorf("not implemented")
}

func (m *mockCatalogService) ListSeries(ctx context.Context, query string, page, limit int) ([]models.SeriesListItem, int, error) {
	if m.listSeriesFn != nil {
		return m.listSeriesFn(ctx, query, page, limit)
	}
	return nil, 0, fmt.Errorf("not implemented")
}

func (m *mockCatalogService) GetStats(ctx context.Context) (*service.Stats, error) {
	if m.getStatsFn != nil {
		return m.getStatsFn(ctx)
	}
	return nil, fmt.Errorf("not implemented")
}

// --- Download service mock ---

type mockDownloadService struct {
	downloadBookFn func(ctx context.Context, id int64) (*service.DownloadResult, error)
}

func (m *mockDownloadService) DownloadBook(ctx context.Context, id int64) (*service.DownloadResult, error) {
	if m.downloadBookFn != nil {
		return m.downloadBookFn(ctx, id)
	}
	return nil, fmt.Errorf("not implemented")
}

// --- Import service mock ---

type mockImportService struct {
	startImportFn func(parentCtx ...context.Context) error
	getStatusFn   func() models.ImportStatus
	cancelFn      func()
}

func (m *mockImportService) StartImport(parentCtx ...context.Context) error {
	if m.startImportFn != nil {
		return m.startImportFn(parentCtx...)
	}
	return nil
}

func (m *mockImportService) GetStatus() models.ImportStatus {
	if m.getStatusFn != nil {
		return m.getStatusFn()
	}
	return models.ImportStatus{Status: "idle"}
}

func (m *mockImportService) CancelImport() {
	if m.cancelFn != nil {
		m.cancelFn()
	}
}

// --- Reader service mock ---

type mockReaderService struct {
	getBookContentFn func(ctx context.Context, bookID int64) (*bookfile.BookContent, error)
	getChapterFn     func(ctx context.Context, bookID int64, chapterID string) (*bookfile.ChapterContent, error)
	getBookImageFn   func(ctx context.Context, bookID int64, imageID string) (*bookfile.ImageData, error)
}

func (m *mockReaderService) GetBookContent(ctx context.Context, bookID int64) (*bookfile.BookContent, error) {
	if m.getBookContentFn != nil {
		return m.getBookContentFn(ctx, bookID)
	}
	return nil, fmt.Errorf("not implemented")
}

func (m *mockReaderService) GetChapter(ctx context.Context, bookID int64, chapterID string) (*bookfile.ChapterContent, error) {
	if m.getChapterFn != nil {
		return m.getChapterFn(ctx, bookID, chapterID)
	}
	return nil, fmt.Errorf("not implemented")
}

func (m *mockReaderService) GetBookImage(ctx context.Context, bookID int64, imageID string) (*bookfile.ImageData, error) {
	if m.getBookImageFn != nil {
		return m.getBookImageFn(ctx, bookID, imageID)
	}
	return nil, fmt.Errorf("not implemented")
}

// --- Progress repo mock ---

type mockProgressRepo struct {
	getFn    func(ctx context.Context, userID string, bookID int64) (*models.ReadingProgress, error)
	upsertFn func(ctx context.Context, p *models.ReadingProgress) error
}

func (m *mockProgressRepo) Get(ctx context.Context, userID string, bookID int64) (*models.ReadingProgress, error) {
	if m.getFn != nil {
		return m.getFn(ctx, userID, bookID)
	}
	return nil, fmt.Errorf("not implemented")
}

func (m *mockProgressRepo) Upsert(ctx context.Context, p *models.ReadingProgress) error {
	if m.upsertFn != nil {
		return m.upsertFn(ctx, p)
	}
	return fmt.Errorf("not implemented")
}

// --- Settings repo mock ---

type mockSettingsRepo struct {
	getSettingsFn    func(ctx context.Context, userID string) (json.RawMessage, error)
	updateSettingsFn func(ctx context.Context, userID string, patch json.RawMessage) (json.RawMessage, error)
}

func (m *mockSettingsRepo) GetSettings(ctx context.Context, userID string) (json.RawMessage, error) {
	if m.getSettingsFn != nil {
		return m.getSettingsFn(ctx, userID)
	}
	return nil, fmt.Errorf("not implemented")
}

func (m *mockSettingsRepo) UpdateSettings(ctx context.Context, userID string, patch json.RawMessage) (json.RawMessage, error) {
	if m.updateSettingsFn != nil {
		return m.updateSettingsFn(ctx, userID, patch)
	}
	return nil, fmt.Errorf("not implemented")
}

// --- Helper: nopCloser wraps an io.Reader to satisfy io.ReadCloser ---

type nopReadCloser struct {
	io.Reader
}

func (nopReadCloser) Close() error { return nil }

// --- Helper: mock AuthResult ---

func mockAuthResult() *service.AuthResult {
	return &service.AuthResult{
		User: models.UserInfo{
			ID:          "user-123",
			Email:       "test@example.com",
			Username:    "testuser",
			DisplayName: "Test User",
			Role:        "user",
		},
		AccessToken:  "mock-access-token",
		RefreshToken: "mock-refresh-token",
	}
}
