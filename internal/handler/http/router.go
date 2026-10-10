package http

import (
	"github.com/gin-gonic/gin"

	"github.com/longntv/go-ddd-template/internal/handler/health"
	"github.com/longntv/go-ddd-template/internal/handler/http/middleware"
	"github.com/longntv/go-ddd-template/internal/handler/http/server"
)

// Router sets up the HTTP routes.
func Router(
	userHandler *server.Handler,
	healthHandler *health.HealthHandler,
	corsConfig middleware.CORSConfig,
) *gin.Engine {
	r := gin.New()

	// Middleware
	r.Use(gin.Recovery())
	r.Use(middleware.Logger())
	r.Use(middleware.CORS(corsConfig))

	// Health check
	r.GET("/health", healthHandler.Handle)

	// API v1
	v1 := r.Group("/api/v1")
	{
		// Users
		users := v1.Group("/users")
		{
			users.POST("", userHandler.Create)
			users.GET("", userHandler.List)
			users.GET("/:id", userHandler.Get)
			users.PUT("/:id", userHandler.Update)
			users.DELETE("/:id", userHandler.Delete)
		}
	}

	return r
}
