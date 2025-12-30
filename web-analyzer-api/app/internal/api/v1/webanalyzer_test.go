package v1

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"web-analyzer-api/app/internal/api/middleware"
	"web-analyzer-api/app/internal/contract"
	"web-analyzer-api/app/internal/core/apperror"
	"web-analyzer-api/app/internal/util/logger"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock WebAnalyzerService
type MockWebAnalyzerService struct {
	mock.Mock
}

func (m *MockWebAnalyzerService) GetAnalyzeData(ctx context.Context, analyzeId string) (*contract.WebAnalyzeResponse, error) {
	args := m.Called(ctx, analyzeId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*contract.WebAnalyzeResponse), args.Error(1)
}

func (m *MockWebAnalyzerService) AnalyzeWebsite(ctx context.Context, baseURL *url.URL) (string, error) {
	args := m.Called(ctx, baseURL)
	return args.String(0), args.Error(1)
}

func (m *MockWebAnalyzerService) UpdateAnalysisStatus(analyzeId string, status string, errorDescription string) {
	m.Called(analyzeId, status, errorDescription)
}

// Setup Test
func setupTest() (service *MockWebAnalyzerService, handler *WebAnalyzerHandler, router *gin.Engine) {
	log := logger.Get("debug")
	mockService := new(MockWebAnalyzerService)
	handler = NewWebAnalyzerHandler(log, mockService)
	router = gin.New()
	router.Use(middleware.ErrorHandler(*log))
	router.Use(middleware.AuthMiddleware(log))
	return mockService, handler, router
}

func TestWebAnalyzerHandler_AnalyzeWebsite(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Success", func(t *testing.T) {
		mockService, handler, router := setupTest()
		router.POST("/analyze", handler.analyzeWebsite)

		reqBody := contract.WebAnalyzeRequest{URL: "http://my-app.com"}
		body, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPost, "/analyze", bytes.NewBuffer(body))
		req.Header.Set("x-api-key", "dev-key-123")
		resp := httptest.NewRecorder()

		parsedURL, _ := url.Parse("http://my-app.com")
		mockService.On("AnalyzeWebsite", mock.Anything, parsedURL).Return("test-id", nil)

		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusOK, resp.Code)
		var result map[string]string
		json.Unmarshal(resp.Body.Bytes(), &result)
		assert.Equal(t, "test-id", result["analyze_id"])
		mockService.AssertExpectations(t)
	})

	t.Run("Invalid JSON", func(t *testing.T) {
		_, handler, router := setupTest()

		router.POST("/analyze", handler.analyzeWebsite)

		req, _ := http.NewRequest(http.MethodPost, "/analyze", bytes.NewBufferString("invalid-json"))
		req.Header.Set("x-api-key", "dev-key-123")
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusBadRequest, resp.Code)
	})

	t.Run("Invalid URL", func(t *testing.T) {
		_, handler, router := setupTest()
		router.POST("/analyze", handler.analyzeWebsite)

		reqBody := contract.WebAnalyzeRequest{URL: "not-a-url"}
		body, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPost, "/analyze", bytes.NewBuffer(body))
		req.Header.Set("x-api-key", "dev-key-123")
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusBadRequest, resp.Code)
	})

	t.Run("Service Error", func(t *testing.T) {
		mockService, handler, router := setupTest()
		router.POST("/analyze", handler.analyzeWebsite)

		reqBody := contract.WebAnalyzeRequest{URL: "http://example.com"}
		body, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPost, "/analyze", bytes.NewBuffer(body))
		req.Header.Set("x-api-key", "dev-key-123")
		resp := httptest.NewRecorder()

		mockService.On("AnalyzeWebsite", mock.Anything, mock.Anything).Return("", apperror.InternalServerError("service error"))

		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusInternalServerError, resp.Code)
	})
}

func TestWebAnalyzerHandler_GetAnalyzeData(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Success", func(t *testing.T) {
		mockService, handler, router := setupTest()
		router.GET("/analyze/:analyze_id", handler.getAnalyzeData)

		expectedResponse := &contract.WebAnalyzeResponse{
			URL: "http://my-app.com",
		}
		mockService.On("GetAnalyzeData", mock.Anything, "test-id").Return(expectedResponse, nil)

		req, _ := http.NewRequest(http.MethodGet, "/analyze/test-id", nil)
		req.Header.Set("x-api-key", "dev-key-123")
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusOK, resp.Code)
		var actualResponse contract.WebAnalyzeResponse
		json.Unmarshal(resp.Body.Bytes(), &actualResponse)
		assert.Equal(t, expectedResponse.URL, actualResponse.URL)
		mockService.AssertExpectations(t)
	})

	t.Run("Service Error", func(t *testing.T) {
		mockService, handler, router := setupTest()
		router.GET("/analyze/:analyze_id", handler.getAnalyzeData)

		mockService.On("GetAnalyzeData", mock.Anything, "test-id").Return(nil, apperror.InternalServerError("service error"))

		req, _ := http.NewRequest(http.MethodGet, "/analyze/test-id", nil)
		req.Header.Set("x-api-key", "dev-key-123")
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusInternalServerError, resp.Code)
	})

	t.Run("Not found Error", func(t *testing.T) {
		mockService, handler, router := setupTest()
		router.GET("/analyze/:analyze_id", handler.getAnalyzeData)

		mockService.On("GetAnalyzeData", mock.Anything, "test-id").Return(nil, apperror.NotFound("not found"))

		req, _ := http.NewRequest(http.MethodGet, "/analyze/test-id", nil)
		req.Header.Set("x-api-key", "dev-key-123")
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusNotFound, resp.Code)
	})

	t.Run("Unauthorized", func(t *testing.T) {
		_, handler, router := setupTest()
		router.POST("/analyze", handler.analyzeWebsite)

		reqBody := contract.WebAnalyzeRequest{URL: "http://my-app.com"}
		body, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPost, "/analyze", bytes.NewBuffer(body))
		req.Header.Set("x-api-key", "321456")
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusUnauthorized, resp.Code)
	})
}
