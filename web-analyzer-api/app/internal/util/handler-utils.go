package util

import (
	"web-analyzer-api/app/internal/util/logger"

	"github.com/gin-gonic/gin"
)

const errDefaultMessage = "Error while setting error in context"

func SetRequestError(c *gin.Context, err error, log *logger.Logger) {
	err1 := c.Error(err)
	if err1 == nil {
		log.Error(errDefaultMessage, "error", errDefaultMessage)
	}
}
