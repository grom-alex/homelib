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
	Reader   *handler.ReaderHandler
	Progress *handler.ProgressHandler
	Settings *handler.SettingsHandler
	Parental *handler.ParentalHandler
}

func SetupRouter(h Handlers, authMw *middleware.AuthMiddleware, parentalMw gin.HandlerFunc) *gin.Engine {
	r := gin.Default()

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	api := r.Group("/api")
	{
		// Public endpoints
		api.GET("/stats", h.Books.GetStats)
		if h.Reader != nil {
			// Public: <img src> tags do not send Authorization headers.
			// Images are embedded binaries from book archives, not user data.
			api.GET("/books/:id/image/:imageId", h.Reader.GetBookImage)
		}

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

		// Authenticated endpoints (with parental filter)
		authorized := api.Group("")
		if authMw != nil {
			authorized.Use(authMw.RequireAuth())
		}
		if parentalMw != nil {
			authorized.Use(parentalMw)
		}
		{
			authorized.GET("/books", h.Books.ListBooks)
			authorized.GET("/books/:id", h.Books.GetBook)
			if h.Download != nil {
				authorized.GET("/books/:id/download", h.Download.DownloadBook)
			}
			if h.Reader != nil {
				authorized.GET("/books/:id/content", h.Reader.GetBookContent)
				authorized.GET("/books/:id/chapter/:chapterId", h.Reader.GetChapter)
			}
			if h.Progress != nil {
				authorized.GET("/me/progress", h.Progress.GetAllProgress)
				authorized.GET("/me/books/:bookId/progress", h.Progress.GetReadingProgress)
				authorized.PUT("/me/books/:bookId/progress", h.Progress.SaveReadingProgress)
				// POST duplicates PUT as a fallback for navigator.sendBeacon(),
				// which only supports POST requests.
				authorized.POST("/me/books/:bookId/progress", h.Progress.SaveReadingProgress)
			}
			if h.Settings != nil {
				authorized.GET("/me/settings", h.Settings.GetUserSettings)
				authorized.PUT("/me/settings", h.Settings.UpdateUserSettings)
			}
			if h.Parental != nil {
				authorized.GET("/me/parental/status", h.Parental.GetMyParentalStatus)
				authorized.POST("/me/parental/unlock", h.Parental.UnlockAdultContent)
				authorized.POST("/me/parental/lock", h.Parental.LockAdultContent)
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
			admin.POST("/genres/reload", h.Admin.ReloadGenres)
			if h.Parental != nil {
				admin.GET("/parental/status", h.Parental.GetAdminParentalStatus)
				admin.GET("/parental/genres", h.Parental.GetRestrictedGenres)
				admin.PUT("/parental/genres", h.Parental.UpdateRestrictedGenres)
				admin.POST("/parental/pin", h.Parental.SetPin)
				admin.DELETE("/parental/pin", h.Parental.RemovePin)
				admin.GET("/parental/users", h.Parental.ListUsersAdultStatus)
				admin.PUT("/parental/users/:userId", h.Parental.SetUserAdultContent)
			}
		}
	}

	return r
}
