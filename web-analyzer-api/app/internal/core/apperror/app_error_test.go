package apperror

import (
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBadRequest(t *testing.T) {
	msg := "invalid input"
	err := BadRequest(msg)
	assert.Equal(t, http.StatusBadRequest, err.StatusCode)
	assert.Equal(t, msg, err.Message)
	assert.Equal(t, CategoryValidation, err.Category)
}

func TestInternalServerError(t *testing.T) {
	msg := "something went wrong"
	err := InternalServerError(msg)
	assert.Equal(t, http.StatusInternalServerError, err.StatusCode)
	assert.Equal(t, msg, err.Message)
	assert.Equal(t, CategoryInternal, err.Category)
}

func TestUnauthorized(t *testing.T) {
	msg := "no api key"
	err := Unauthorized(msg)
	assert.Equal(t, http.StatusUnauthorized, err.StatusCode)
	assert.Equal(t, msg, err.Message)
	assert.Equal(t, CategoryAuth, err.Category)
}

func TestNotFound(t *testing.T) {
	msg := "resource missing"
	err := NotFound(msg)
	assert.Equal(t, http.StatusNotFound, err.StatusCode)
	assert.Equal(t, msg, err.Message)
	assert.Equal(t, CategoryNotFound, err.Category)
}

func TestErrorMethod(t *testing.T) {
	t.Run("Basic error", func(t *testing.T) {
		err := &AppError{StatusCode: 400, Message: "test"}
		assert.Equal(t, "status code: 400, message: test", err.Error())
	})

	t.Run("Chained error", func(t *testing.T) {
		cause := errors.New("underlying cause")
		err := &AppError{
			StatusCode:   500,
			Message:      "outer",
			ChainedError: cause,
		}
		expected := "status code: 500, message: outer\nchained error: underlying cause"
		assert.Equal(t, expected, err.Error())
	})
}
