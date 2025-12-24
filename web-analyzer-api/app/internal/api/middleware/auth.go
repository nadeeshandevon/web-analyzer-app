package middleware

import (
	"os"
	"web-analyzer-api/app/internal/core/apperror"
	"web-analyzer-api/app/internal/util"
	"web-analyzer-api/app/internal/util/logger"

	"github.com/gin-gonic/gin"
)

const (
	ApiKeyHeader     = "x-api-key"
	DefaultApiKey    = "dev-key-123"
	ApiKeyEnvVarName = "API_KEY"
)

func AuthMiddleware(log *logger.Logger) gin.HandlerFunc {
	apiKey := os.Getenv(ApiKeyEnvVarName)
	if apiKey == "" {
		apiKey = DefaultApiKey
	}

	return func(c *gin.Context) {
		requestKey := c.GetHeader(ApiKeyHeader)

		if requestKey == "" {
			util.SetRequestError(c, apperror.Unauthorized("Missing API key"), log)
			c.Abort()
			return
		}

		if requestKey != apiKey {
			util.SetRequestError(c, apperror.Unauthorized("Invalid API key"), log)
			c.Abort()
			return
		}

		c.Next()
	}
}
