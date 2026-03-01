package api

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/grom-alex/homelib/backend/internal/api/handler"
	"github.com/grom-alex/homelib/backend/internal/api/middleware"
	"github.com/grom-alex/homelib/backend/internal/config"
	"github.com/grom-alex/homelib/backend/internal/glst"
	"github.com/grom-alex/homelib/backend/internal/repository"
	"github.com/grom-alex/homelib/backend/internal/service"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Server struct {
	httpServer   *http.Server
	pool         *pgxpool.Pool
	importSvc    *service.ImportService
	readerSvc    *service.ReaderService
	genreTreeSvc *service.GenreTreeService
}

// loadGenreData returns genre file data: from config path if set, otherwise embedded default.
func loadGenreData(cfg config.GenreTreeConfig) []byte {
	if cfg.FilePath != "" {
		data, err := os.ReadFile(cfg.FilePath)
		if err != nil {
			log.Printf("WARNING: failed to read genre file %q, using embedded default: %v", cfg.FilePath, err)
			return glst.DefaultGenreFile
		}
		return data
	}
	return glst.DefaultGenreFile
}

func NewServer(cfg *config.Config, pool *pgxpool.Pool) *Server {
	// Repositories
	bookRepo := repository.NewBookRepo(pool)
	authorRepo := repository.NewAuthorRepo(pool)
	genreRepo := repository.NewGenreRepo(pool)
	seriesRepo := repository.NewSeriesRepo(pool)
	collectionRepo := repository.NewCollectionRepo(pool)
	userRepo := repository.NewUserRepo(pool)
	refreshRepo := repository.NewRefreshTokenRepo(pool)
	metadataRepo := repository.NewMetadataRepo(pool)

	// Genre tree service
	genreData := loadGenreData(cfg.GenreTree)
	genreTreeSvc := service.NewGenreTreeService(genreData, metadataRepo, genreRepo, bookRepo)

	// Services
	catalogSvc := service.NewCatalogService(pool, bookRepo, authorRepo, genreRepo, seriesRepo, collectionRepo)
	importSvc := service.NewImportService(pool, cfg.Import, cfg.Library, bookRepo, authorRepo, genreRepo, seriesRepo, collectionRepo)
	authSvc := service.NewAuthService(cfg.Auth, userRepo, refreshRepo)
	downloadSvc := service.NewDownloadService(bookRepo, cfg.Library)
	readerSvc := service.NewReaderService(bookRepo, cfg.Library, cfg.Reader)

	// Reading progress repository
	progressRepo := repository.NewReadingProgressRepo(pool)

	// Auth middleware using AuthService as validator
	authValidator := &authServiceValidator{authSvc: authSvc}
	authMw := middleware.NewAuthMiddleware(authValidator)

	// Handlers
	h := Handlers{
		Books:    handler.NewBooksHandler(catalogSvc),
		Authors:  handler.NewAuthorsHandler(catalogSvc),
		Genres:   handler.NewGenresHandler(catalogSvc),
		Series:   handler.NewSeriesHandler(catalogSvc),
		Admin:    handler.NewAdminHandler(importSvc, genreTreeSvc),
		Auth:     handler.NewAuthHandler(authSvc, cfg.Auth.RefreshTokenTTL, cfg.Auth.CookieSecure),
		Download: handler.NewDownloadHandler(downloadSvc),
		Reader:   handler.NewReaderHandler(readerSvc),
		Progress: handler.NewProgressHandler(progressRepo),
		Settings: handler.NewSettingsHandler(userRepo),
	}

	router := SetupRouter(h, authMw)

	return &Server{
		httpServer: &http.Server{
			Addr:              fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
			Handler:           router,
			ReadTimeout:       30 * time.Second,
			WriteTimeout:      30 * time.Second,
			ReadHeaderTimeout: 10 * time.Second,
			IdleTimeout:       120 * time.Second,
		},
		pool:         pool,
		importSvc:    importSvc,
		readerSvc:    readerSvc,
		genreTreeSvc: genreTreeSvc,
	}
}

// authServiceValidator adapts AuthService to the middleware.TokenValidator interface.
type authServiceValidator struct {
	authSvc *service.AuthService
}

func (v *authServiceValidator) ValidateToken(tokenString string) (*middleware.Claims, error) {
	info, err := v.authSvc.ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}
	return &middleware.Claims{
		UserID: info.ID,
		Email:  info.Email,
		Role:   info.Role,
	}, nil
}

func (s *Server) Start(ctx context.Context) error {
	// Set app context so imports started via API are cancelled on shutdown
	s.importSvc.SetAppContext(ctx)

	// Load genre tree if needed (idempotent — skips if hash unchanged)
	if s.genreTreeSvc != nil {
		result, err := s.genreTreeSvc.LoadIfNeeded(ctx)
		if err != nil {
			log.Printf("WARNING: genre tree load failed: %v", err)
		} else if result.Skipped {
			log.Println("Genre tree: up to date, skipped")
		} else {
			log.Printf("Genre tree loaded: %d genres, %d books remapped, %d warnings",
				result.GenresLoaded, result.BooksRemapped, len(result.Warnings))
		}
	}

	// Start periodic cache cleanup (stops on ctx cancellation)
	s.readerSvc.StartCacheCleanup(ctx)

	errCh := make(chan error, 1)

	go func() {
		log.Printf("Starting server on %s", s.httpServer.Addr)
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		return fmt.Errorf("server error: %w", err)
	case <-ctx.Done():
		log.Println("Shutting down server...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return s.httpServer.Shutdown(shutdownCtx)
	}
}
