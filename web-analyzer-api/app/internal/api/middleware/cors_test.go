package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestCORSMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("OPTIONS Request", func(t *testing.T) {
		router := gin.New()
		router.Use(CORSMiddleware())
		router.GET("/test", func(c *gin.Context) {
			c.String(http.StatusOK, "ok")
		})

		req, _ := http.NewRequest(http.MethodOptions, "/test", nil)
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusNoContent, resp.Code)
		assert.Equal(t, "*", resp.Header().Get("Access-Control-Allow-Origin"))
		assert.Equal(t, "true", resp.Header().Get("Access-Control-Allow-Credentials"))
		assert.Contains(t, resp.Header().Get("Access-Control-Allow-Methods"), "OPTIONS")
		assert.Empty(t, resp.Body.String())
	})

	t.Run("GET Request", func(t *testing.T) {
		router := gin.New()
		router.Use(CORSMiddleware())
		router.GET("/test", func(c *gin.Context) {
			c.String(http.StatusOK, "ok")
		})

		req, _ := http.NewRequest(http.MethodGet, "/test", nil)
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusOK, resp.Code)
		assert.Equal(t, "*", resp.Header().Get("Access-Control-Allow-Origin"))
		assert.Equal(t, "ok", resp.Body.String())
	})
}
