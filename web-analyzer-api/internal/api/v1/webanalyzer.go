package v1

import (
	"encoding/json"
	"net/url"
	"web-analyzer-api/internal/core/apperror"
	"web-analyzer-api/internal/logger"
	"web-analyzer-api/internal/util"
	"web-analyzer-api/pkg/contract"

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
	var req contract.WebAnalyzeRequest
	if !h.validateRequest(c, &req) {
		return
	}
}

func (h *WebAnalyzerHandler) validateRequest(c *gin.Context, req *contract.WebAnalyzeRequest) bool {
	if err := json.NewDecoder(c.Request.Body).Decode(req); err != nil {
		util.SetRequestError(c, apperror.BadRequest("Invalid request body: "+err.Error()), h.log)
		return false
	}

	// Validate URL format
	parsedURL, err := url.Parse(req.URL)
	if err != nil {
		util.SetRequestError(c, apperror.BadRequest("Invalid URL format: "+err.Error()), h.log)
		return false
	}
	if parsedURL.Scheme == "" || parsedURL.Host == "" {
		util.SetRequestError(c, apperror.BadRequest("Invalid URL format. Please provide a valid URL with scheme (http:// or https://)"), h.log)
		return false
	}
	return true
}
