package v1

import (
	"web-analyzer-api/internal/logger"

	"github.com/gin-gonic/gin"
)

type WebAnalyzerHandler struct {
	log *logger.Logger // Logger for logging messages
}

func NewWebAnalyzerHandler(logger *logger.Logger) *WebAnalyzerHandler {
	return &WebAnalyzerHandler{
		log: logger,
	}
}

func (h WebAnalyzerHandler) RegisterRoutes(v1 *gin.RouterGroup) {

	v1.POST("/web-analyzer/analyze",
		h.analyzeWebPage)
}

func (h WebAnalyzerHandler) analyzeWebPage(c *gin.Context) {
}
