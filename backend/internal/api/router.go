package api

import (
	"github.com/gin-gonic/gin"

	"github.com/grom-alex/homelib/backend/internal/api/handler"
	"github.com/grom-alex/homelib/backend/internal/api/middleware"
)

type Handlers struct {
	Books    *handler.BooksHandler
	Authors  *handler.AuthorsHandler
	Genres   *handler.GenresHandler
	Series   *handler.SeriesHandler
	Admin    *handler.AdminHandler
	Auth     *handler.AuthHandler
	Download *handler.DownloadHandler
}

func SetupRouter(h Handlers, authMw *middleware.AuthMiddleware) *gin.Engine {
	r := gin.Default()

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	api := r.Group("/api")
	{
		// Public endpoints
		api.GET("/stats", h.Books.GetStats)

		// Auth endpoints
		auth := api.Group("/auth")
		{
			if h.Auth != nil {
				auth.POST("/register", h.Auth.Register)
				auth.POST("/login", h.Auth.Login)
				auth.POST("/refresh", h.Auth.Refresh)
				auth.POST("/logout", h.Auth.Logout)
			}
		}

		// Authenticated endpoints
		authorized := api.Group("")
		if authMw != nil {
			authorized.Use(authMw.RequireAuth())
		}
		{
			authorized.GET("/books", h.Books.ListBooks)
			authorized.GET("/books/:id", h.Books.GetBook)
			if h.Download != nil {
				authorized.GET("/books/:id/download", h.Download.DownloadBook)
			}
			authorized.GET("/authors", h.Authors.ListAuthors)
			authorized.GET("/authors/:id", h.Authors.GetAuthor)
			authorized.GET("/genres", h.Genres.ListGenres)
			authorized.GET("/series", h.Series.ListSeries)
		}

		// Admin endpoints
		admin := api.Group("/admin")
		if authMw != nil {
			admin.Use(authMw.RequireAuth(), authMw.RequireAdmin())
		}
		{
			admin.POST("/import", h.Admin.StartImport)
			admin.GET("/import/status", h.Admin.ImportStatus)
			admin.POST("/import/cancel", h.Admin.CancelImport)
		}
	}

	return r
}
