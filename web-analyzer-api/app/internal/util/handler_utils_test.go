package util

import (
	"errors"
	"net/http/httptest"
	"testing"
	"web-analyzer-api/app/internal/util/logger"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestSetRequestError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	log := logger.Get("debug")

	t.Run("Valid Error", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		testErr := errors.New("test error")
		SetRequestError(c, testErr, log)

		assert.Len(t, c.Errors, 1)
		assert.Equal(t, testErr, c.Errors[0].Err)
	})

	t.Run("Nil Error", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		SetRequestError(c, nil, log)
		assert.Len(t, c.Errors, 0)
	})
}
