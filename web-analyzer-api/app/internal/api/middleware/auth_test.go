package middleware

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"web-analyzer-api/app/internal/util/logger"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestAuthMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	log := logger.Get("debug")

	t.Run("Valid Key from Env", func(t *testing.T) {
		const testKey = "dev-key-xyz"
		os.Setenv(ApiKeyEnvVarName, testKey)
		defer os.Unsetenv(ApiKeyEnvVarName)

		router := gin.New()
		router.Use(ErrorHandler(*log))
		router.Use(AuthMiddleware(log))
		router.GET("/test", func(c *gin.Context) {
			c.String(http.StatusOK, "ok")
		})

		req, _ := http.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set(ApiKeyHeader, testKey)
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusOK, resp.Code)
		assert.Equal(t, "ok", resp.Body.String())
	})

	t.Run("Invalid Key", func(t *testing.T) {
		os.Setenv(ApiKeyEnvVarName, "x-api-key-test")
		defer os.Unsetenv(ApiKeyEnvVarName)

		router := gin.New()
		router.Use(ErrorHandler(*log))
		router.Use(AuthMiddleware(log))
		router.GET("/test", func(c *gin.Context) {
			c.String(http.StatusOK, "ok")
		})

		req, _ := http.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set(ApiKeyHeader, "x-api-key-test1")
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusUnauthorized, resp.Code)
		assert.Contains(t, resp.Body.String(), "Invalid API key")
	})

	t.Run("Missing Key", func(t *testing.T) {
		router := gin.New()
		router.Use(ErrorHandler(*log))
		router.Use(AuthMiddleware(log))
		router.GET("/test", func(c *gin.Context) {
			c.String(http.StatusOK, "ok")
		})

		req, _ := http.NewRequest(http.MethodGet, "/test", nil)
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusUnauthorized, resp.Code)
		assert.Contains(t, resp.Body.String(), "Missing API key")
	})
}
