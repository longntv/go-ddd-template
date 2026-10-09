package middleware

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

// Logger logs HTTP requests.
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		end := time.Now()
		latency := end.Sub(start)

		log.Printf(
			"[GIN] %s %s %s | %d | %v | %s",
			c.Request.Method,
			path,
			query,
			c.Writer.Status(),
			latency,
			c.Errors.String(),
		)
	}
}
