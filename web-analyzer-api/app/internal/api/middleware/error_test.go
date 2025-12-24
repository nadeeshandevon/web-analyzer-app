package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"web-analyzer-api/app/internal/core/apperror"
	"web-analyzer-api/app/internal/util/logger"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestErrorHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	log := logger.Get("debug")

	t.Run("No Error", func(t *testing.T) {
		router := gin.New()
		router.Use(ErrorHandler(*log))
		router.GET("/test", func(c *gin.Context) {
			c.String(http.StatusOK, "ok")
		})

		req, _ := http.NewRequest(http.MethodGet, "/test", nil)
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusOK, resp.Code)
		assert.Equal(t, "ok", resp.Body.String())
	})

	t.Run("AppError handling", func(t *testing.T) {
		router := gin.New()
		router.Use(ErrorHandler(*log))
		router.GET("/test", func(c *gin.Context) {
			c.Error(apperror.BadRequest("bad request test"))
		})

		req, _ := http.NewRequest(http.MethodGet, "/test", nil)
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusBadRequest, resp.Code)
		assert.Contains(t, resp.Body.String(), "bad request test")
		assert.Contains(t, resp.Body.String(), "validation")
	})

	t.Run("Generic error handling", func(t *testing.T) {
		router := gin.New()
		router.Use(ErrorHandler(*log))
		router.GET("/test", func(c *gin.Context) {
			c.Error(errors.New("generic error"))
		})

		req, _ := http.NewRequest(http.MethodGet, "/test", nil)
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusInternalServerError, resp.Code)
		assert.Contains(t, resp.Body.String(), "Internal server error")
	})

	t.Run("Response already written", func(t *testing.T) {
		router := gin.New()
		router.Use(ErrorHandler(*log))
		router.GET("/test", func(c *gin.Context) {
			c.String(http.StatusCreated, "already written")
			c.Error(errors.New("should not be written"))
		})

		req, _ := http.NewRequest(http.MethodGet, "/test", nil)
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusCreated, resp.Code)
		assert.Equal(t, "already written", resp.Body.String())
	})
}
