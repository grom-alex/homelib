package api

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/grom-alex/homelib/backend/internal/api/handler"
	"github.com/grom-alex/homelib/backend/internal/api/middleware"
	"github.com/grom-alex/homelib/backend/internal/config"
	"github.com/grom-alex/homelib/backend/internal/repository"
	"github.com/grom-alex/homelib/backend/internal/service"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Server struct {
	httpServer *http.Server
	pool       *pgxpool.Pool
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

	// Services
	catalogSvc := service.NewCatalogService(pool, bookRepo, authorRepo, genreRepo, seriesRepo, collectionRepo)
	importSvc := service.NewImportService(pool, cfg.Import, cfg.Library, bookRepo, authorRepo, genreRepo, seriesRepo, collectionRepo)
	authSvc := service.NewAuthService(cfg.Auth, userRepo, refreshRepo)
	downloadSvc := service.NewDownloadService(bookRepo, cfg.Library)

	// Auth middleware using AuthService as validator
	authValidator := &authServiceValidator{authSvc: authSvc}
	authMw := middleware.NewAuthMiddleware(authValidator)

	// Handlers
	h := Handlers{
		Books:    handler.NewBooksHandler(catalogSvc),
		Authors:  handler.NewAuthorsHandler(catalogSvc),
		Genres:   handler.NewGenresHandler(catalogSvc),
		Series:   handler.NewSeriesHandler(catalogSvc),
		Admin:    handler.NewAdminHandler(importSvc),
		Auth:     handler.NewAuthHandler(authSvc, cfg.Auth.RefreshTokenTTL, cfg.Auth.CookieSecure),
		Download: handler.NewDownloadHandler(downloadSvc),
	}

	router := SetupRouter(h, authMw)

	return &Server{
		httpServer: &http.Server{
			Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
			Handler:      router,
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
		},
		pool: pool,
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
