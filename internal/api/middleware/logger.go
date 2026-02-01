package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

// Logger returns a middleware that logs HTTP requests.
func Logger(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		statusCode := c.Writer.Status()

		logger.Info("HTTP request",
			"method", c.Request.Method,
			"path", path,
			"query", query,
			"status", statusCode,
			"latency_ms", latency.Milliseconds(),
			"client_ip", c.ClientIP(),
		)
	}
}
