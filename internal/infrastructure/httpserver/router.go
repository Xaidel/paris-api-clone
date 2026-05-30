package httpserver

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// RouteRegistrar registers routes on a Gin engine.
type RouteRegistrar interface {
	// RegisterRoutes attaches the adapter's endpoints to the shared engine.
	RegisterRoutes(engine *gin.Engine)
}

// NewRouter builds the HTTP router for the application.
func NewRouter(logger *zap.Logger, registrars ...RouteRegistrar) (*gin.Engine, error) {
	if logger == nil {
		return nil, errors.New("logger is required")
	}

	engine := gin.New()
	engine.Use(gin.Recovery(), requestContextMiddleware(logger))
	// Keep a trivial health endpoint outside the versioned API routes so startup
	// and deployment checks do not depend on application registrars.
	engine.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	for _, registrar := range registrars {
		if registrar == nil {
			return nil, errors.New("route registrar is required")
		}

		registrar.RegisterRoutes(engine)
	}

	return engine, nil
}

func requestContextMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		startedAt := time.Now()
		// Attach the logger to the request context for handlers that need a shared
		// request-scoped logger without recreating one.
		c.Set("request_logger", logger)
		c.Next()

		// Log after the handler finishes so the final status code and duration are
		// both available in one structured event.
		logger.Info(
			"http request completed",
			zap.String("method", c.Request.Method),
			zap.String("path", c.FullPath()),
			zap.Int("status_code", c.Writer.Status()),
			zap.Duration("duration", time.Since(startedAt)),
		)
	}
}
