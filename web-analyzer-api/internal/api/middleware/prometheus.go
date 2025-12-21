package middleware

import (
	"strconv"
	"time"
	"web-analyzer-api/internal/util/metrics"

	"github.com/gin-gonic/gin"
)

func init() {
	metrics.RegisterAPIMetrics()
}

func PrometheusMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.FullPath()

		c.Next()

		if path == "/metrics" || path == "/health" {
			return
		}

		duration := time.Since(start).Seconds()
		status := strconv.Itoa(c.Writer.Status())
		method := c.Request.Method

		if path == "" {
			path = "unknown"
		}

		metrics.RecordHttpRequestTotal(method, path, status)
		metrics.RecordHttpRequestDuration(method, path, duration)
	}
}
