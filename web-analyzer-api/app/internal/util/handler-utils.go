package util

import (
	"web-analyzer-api/app/internal/util/logger"

	"github.com/gin-gonic/gin"
)

const errDefaultMessage = "Error while setting error in context"

func SetRequestError(c *gin.Context, err error, log *logger.Logger) {
	if err == nil {
		log.Error(errDefaultMessage, "error", "nil error provided")
		return
	}
	c.Error(err)
}
