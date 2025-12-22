package middleware

import (
	"net/http"
	"web-analyzer-api/app/internal/core/apperror"
	"web-analyzer-api/app/internal/util/logger"

	"github.com/gin-gonic/gin"
)

func ErrorHandler(logger logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) > 0 {
			err := c.Errors.Last().Err
			if err != nil {
				logger.Error("Error in request", "error", err)

				switch e := err.(type) {
				case *apperror.AppError:
					if !c.Writer.Written() {
						c.JSON(e.StatusCode, gin.H{"status_code": e.StatusCode, "message": e.Message, "category": e.Category, "reason": e.Reason})
					}
				default:
					if !c.Writer.Written() {
						c.JSON(http.StatusInternalServerError, gin.H{"status_code": http.StatusInternalServerError,
							"message": "Internal server error", "category": "Internal", "reason": "Internal"})
					}
				}
			}
		}
	}
}
