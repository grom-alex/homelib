package middleware

import (
	"context"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ParentalServicer abstracts the parental service methods needed by the middleware.
type ParentalServicer interface {
	IsAdultContentEnabled(ctx context.Context, userID string) (bool, error)
	GetRestrictedGenreIDs(ctx context.Context) ([]int, error)
}

// ParentalFilter returns middleware that resolves restricted genre IDs
// and stores them in the Gin context for downstream handlers.
// All users (including admins) are subject to parental filtering.
func ParentalFilter(svc ParentalServicer) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("user_id")
		if userID == "" {
			c.Next()
			return
		}

		// Fail-closed: on error checking status, assume content is restricted.
		enabled, err := svc.IsAdultContentEnabled(c.Request.Context(), userID)
		if err != nil {
			log.Printf("parental middleware: check adult content: %v", err)
			enabled = false
		}

		if !enabled {
			ids, err := svc.GetRestrictedGenreIDs(c.Request.Context())
			if err != nil {
				// Fail-closed: cannot determine restricted genres, block access.
				log.Printf("parental middleware: get restricted IDs: %v", err)
				c.JSON(http.StatusServiceUnavailable, gin.H{"error": "service_unavailable", "message": "Не удалось проверить ограничения контента"})
				c.Abort()
				return
			}
			if len(ids) > 0 {
				c.Set("restricted_genre_ids", ids)
			}
		}

		c.Next()
	}
}
