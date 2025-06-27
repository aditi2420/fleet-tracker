package middleware

import (
	"github.com/gin-gonic/gin"
	"log/slog"
	"time"
)

func HTTPLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// let the request proceed
		c.Next()

		// after handler finishes
		slog.Info("http request",
			"method", c.Request.Method,
			"path", c.FullPath(),
			"status", c.Writer.Status(),
			"duration", time.Since(start).Milliseconds(),
			"clientIP", c.ClientIP(),
		)
	}
}
