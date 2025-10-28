package logger

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)


func GinLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now().UTC()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)

		fields := []zap.Field{
			zap.Int("status", c.Writer.Status()),
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("query", query),
			zap.String("ip", c.ClientIP()),
			zap.String("user-agent", c.Request.UserAgent()),
			zap.Duration("latency", latency),
		}

		if len(c.Errors) > 0 {
			for _, e := range c.Errors.Errors() {
				Get().Error("Request error", append(fields, zap.String("error", e))...)
			}
		} else {
			Get().Info("Request completed", fields...)
		}
	}
}
