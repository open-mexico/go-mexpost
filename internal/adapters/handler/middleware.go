package handler

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

// SlogMiddleware logs HTTP requests in structured JSON format.
func SlogMiddleware(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()
		method := c.Request.Method
		clientIP := c.ClientIP()
		errors := c.Errors.ByType(gin.ErrorTypePrivate).String()

		attributes := []any{
			slog.Int("status", status),
			slog.String("method", method),
			slog.String("path", path),
			slog.String("query", query),
			slog.String("ip", clientIP),
			slog.String("latency", latency.String()),
			slog.Int64("latency_ms", latency.Milliseconds()),
			slog.String("user_agent", c.Request.UserAgent()),
		}

		if len(errors) > 0 {
			attributes = append(attributes, slog.String("errors", errors))
		}

		if status >= 500 {
			logger.Error("Server error", attributes...)
		} else if status >= 400 {
			logger.Warn("Client error", attributes...)
		} else {
			logger.Info("HTTP request", attributes...)
		}
	}
}
