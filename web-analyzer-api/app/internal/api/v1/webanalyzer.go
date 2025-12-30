package v1

import (
	"encoding/json"
	"net/http"
	"net/url"
	"web-analyzer-api/app/internal/contract"
	"web-analyzer-api/app/internal/core"
	"web-analyzer-api/app/internal/core/apperror"
	"web-analyzer-api/app/internal/util"
	"web-analyzer-api/app/internal/util/logger"

	"github.com/gin-gonic/gin"
)

type WebAnalyzerHandler struct {
	log                *logger.Logger
	webAnalyzerService core.WebAnalyzerService
}

func NewWebAnalyzerHandler(logger *logger.Logger, webAnalyzerService core.WebAnalyzerService) *WebAnalyzerHandler {
	return &WebAnalyzerHandler{
		log:                logger,
		webAnalyzerService: webAnalyzerService,
	}
}

func (h WebAnalyzerHandler) RegisterRoutes(v1 *gin.RouterGroup) {

	v1.POST("/web-analyzer/analyze",
		h.analyzeWebsite)

	v1.GET("/web-analyzer/:analyze_id/analyze",
		h.getAnalyzeData)
}

func (h WebAnalyzerHandler) analyzeWebsite(c *gin.Context) {
	var req contract.WebAnalyzeRequest
	parsedURL, ok := h.validateRequest(c, &req)
	if !ok {
		return
	}

	result, err := h.webAnalyzerService.AnalyzeWebsite(c.Request.Context(), parsedURL)

	if err != nil {
		util.SetRequestError(c, err, h.log)
		return
	}

	c.JSON(http.StatusOK, gin.H{"analyze_id": result})
}

func (h WebAnalyzerHandler) getAnalyzeData(c *gin.Context) {
	analyzeId := c.Param("analyze_id")
	if analyzeId == "" {
		util.SetRequestError(c, apperror.BadRequest("Analyze id cannot be empty"), h.log)
		return
	}

	result, err := h.webAnalyzerService.GetAnalyzeData(c.Request.Context(), analyzeId)

	if err != nil {
		util.SetRequestError(c, err, h.log)
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *WebAnalyzerHandler) validateRequest(c *gin.Context, req *contract.WebAnalyzeRequest) (*url.URL, bool) {
	if err := json.NewDecoder(c.Request.Body).Decode(req); err != nil {
		util.SetRequestError(c, apperror.BadRequest("Invalid request body: "+err.Error()), h.log)
		return nil, false
	}

	// Validate URL format
	parsedURL, err := url.Parse(req.URL)
	if err != nil {
		util.SetRequestError(c, apperror.BadRequest("Invalid URL format: "+err.Error()), h.log)
		return nil, false
	}
	if parsedURL.Scheme == "" || parsedURL.Host == "" {
		util.SetRequestError(c, apperror.BadRequest("Invalid URL format. Please provide a valid URL with scheme (http:// or https://)"), h.log)
		return nil, false
	}
	return parsedURL, true
}
