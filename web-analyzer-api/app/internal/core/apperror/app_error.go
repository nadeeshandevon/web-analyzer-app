package apperror

import (
	"fmt"
	"net/http"
)

const (
	CategoryValidation = "validation"
	CategoryAuth       = "auth"
	CategoryInternal   = "internal"
	CategoryDownstream = "downstream"
	CategoryUnknown    = "unknown"
)

type AppError struct {
	StatusCode   int
	Message      string
	ChainedError error
	Category     string
	Reason       string
}

func BadRequest(message string) *AppError {
	return categorizedError(message, http.StatusBadRequest, CategoryValidation)
}

func InternalServerError(message string) *AppError {
	return categorizedError(message, http.StatusInternalServerError, CategoryInternal)
}

func Unauthorized(message string) *AppError {
	return categorizedError(message, http.StatusUnauthorized, CategoryAuth)
}

func NotFound(message string) *AppError {
	return categorizedError(message, http.StatusNotFound, CategoryValidation)
}

func (e *AppError) Error() string {
	result := fmt.Sprintf("status code: %d, message: %s", e.StatusCode, e.Message)
	if e.ChainedError != nil {
		result += fmt.Sprintf("\nchained error: %s", e.ChainedError.Error())
	}
	return result
}

func categorizedError(message string, status int, category string) *AppError {
	return &AppError{StatusCode: status, Message: message, Category: category}
}
