package api

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	v1 "web-analyzer-api/app/internal/api/v1"
	"web-analyzer-api/app/internal/contract"
	"web-analyzer-api/app/internal/di"
	"web-analyzer-api/app/internal/util/logger"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

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

func TestSetupRouter(t *testing.T) {
	gin.SetMode(gin.TestMode)
	log := logger.Get("debug")
	mockService := new(MockWebAnalyzerService)
	handler := v1.NewWebAnalyzerHandler(log, mockService)

	handlers := di.HTTPHandlers{
		WebAnalyzerHandler: *handler,
	}

	t.Run("Router Setup Success", func(t *testing.T) {
		router := gin.New()
		err := SetupRouter(router, handlers, log)

		assert.NoError(t, err)

		req, _ := http.NewRequest(http.MethodPost, "/api/v1/web-analyzer/analyze", nil)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusUnauthorized, resp.Code)

		req, _ = http.NewRequest(http.MethodPost, "/api/v1/web-analyzer/analyze", bytes.NewBufferString("{}"))
		req.Header.Set("x-api-key", "dev-key-123")
		resp = httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusBadRequest, resp.Code)
	})

	t.Run("CORS Middleware Check", func(t *testing.T) {
		router := gin.New()
		_ = SetupRouter(router, handlers, log)

		req, _ := http.NewRequest(http.MethodOptions, "/api/v1/web-analyzer/analyze", nil)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusNoContent, resp.Code)
		assert.Equal(t, "*", resp.Header().Get("Access-Control-Allow-Origin"))
	})
}
