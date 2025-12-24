package webanalyzer

import (
	"context"
	"net/http"
	"net/url"
	"sync"
	"testing"

	core "web-analyzer-api/app/internal/core"
	"web-analyzer-api/app/internal/core/apperror"
	"web-analyzer-api/app/internal/model"
	"web-analyzer-api/app/internal/util/logger"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock WebAnalyzerRepository
type MockWebAnalyzerRepository struct {
	mock.Mock
}

func (m *MockWebAnalyzerRepository) Save(analysis model.WebAnalyzer) (string, error) {
	args := m.Called(analysis)
	return args.String(0), args.Error(1)
}

func (m *MockWebAnalyzerRepository) GetById(id string) (*model.WebAnalyzer, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.WebAnalyzer), args.Error(1)
}

func (m *MockWebAnalyzerRepository) Update(analysis model.WebAnalyzer) (string, error) {
	args := m.Called(analysis)
	return args.String(0), args.Error(1)
}

// Mock LinkChecker
type MockLinkChecker struct {
	mock.Mock
}

func (m *MockLinkChecker) CheckLink(ctx context.Context, client *http.Client, link string, baseURL *url.URL) *model.LinkCheckResult {
	args := m.Called(ctx, client, link, baseURL)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*model.LinkCheckResult)
}

func (m *MockLinkChecker) RunWorker(ctx context.Context, linksChan <-chan string, resultsChan chan<- model.LinkCheckResult, baseURL *url.URL, wg *sync.WaitGroup) {
	m.Called(ctx, linksChan, resultsChan, baseURL, wg)
	wg.Done()
}

func setupTest() (service core.WebAnalyzerService, repo *MockWebAnalyzerRepository, linkChecker *MockLinkChecker) {
	log := logger.Get("info")
	mockRepo := new(MockWebAnalyzerRepository)
	mockLinkChecker := new(MockLinkChecker)
	service = NewWebAnalyzerService(log, mockRepo, mockLinkChecker)
	return service, mockRepo, mockLinkChecker
}

func TestNewWebAnalyzerService(t *testing.T) {
	service, _, _ := setupTest()
	assert.NotNil(t, service)
}

func TestGetAnalyzeData(t *testing.T) {
	service, mockRepo, _ := setupTest()
	analysisId := "id-123456789"

	t.Run("Success", func(t *testing.T) {
		mockResult := &model.WebAnalyzer{
			URL:         "http://test.com",
			HTMLVersion: "HTML5",
			Title:       "Test Title",
			Status:      StatusSuccess,
		}

		mockRepo.On("GetById", analysisId).Return(mockResult, nil).Once()

		resp, err := service.GetAnalyzeData(context.Background(), analysisId)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, mockResult.URL, resp.URL)
		assert.Equal(t, mockResult.Title, resp.Title)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Repository Error", func(t *testing.T) {
		mockRepo.On("GetById", analysisId).Return(nil, apperror.InternalServerError("Analyse data not found")).Once()

		resp, err := service.GetAnalyzeData(context.Background(), analysisId)

		assert.Error(t, err)
		assert.Nil(t, resp)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Not Found", func(t *testing.T) {
		mockRepo.On("GetById", analysisId).Return(nil, nil).Once()

		resp, err := service.GetAnalyzeData(context.Background(), analysisId)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "not found")
		mockRepo.AssertExpectations(t)
	})
}

func TestUpdateAnalysisStatus(t *testing.T) {
	service, mockRepo, _ := setupTest()
	analysisId := "id-123456789"

	t.Run("Update Status and Error", func(t *testing.T) {
		existing := &model.WebAnalyzer{
			ID:     analysisId,
			Status: StatusPending,
		}

		mockRepo.On("GetById", analysisId).Return(existing, nil).Once()
		mockRepo.On("Update", mock.MatchedBy(func(a model.WebAnalyzer) bool {
			return a.Status == StatusFailed && *a.ErrorDescription == "error description"
		})).Return(analysisId, nil).Once()

		service.UpdateAnalysisStatus(analysisId, StatusFailed, "error description")

		mockRepo.AssertExpectations(t)
	})
}

func TestAnalyzeWebsite(t *testing.T) {
	service, mockRepo, _ := setupTest()
	baseURL, _ := url.Parse("http://test.com")

	t.Run("Success Path", func(t *testing.T) {
		mockRepo.On("Save", mock.MatchedBy(func(a model.WebAnalyzer) bool {
			return a.URL == "http://test.com" && a.Status == StatusPending
		})).Return("new-id", nil).Once()

		mockRepo.On("GetById", "new-id").Return(nil, apperror.InternalServerError("stop background job")).Maybe()

		id, err := service.AnalyzeWebsite(context.Background(), baseURL)

		assert.NoError(t, err)
		assert.Equal(t, "new-id", id)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Save Error", func(t *testing.T) {
		mockRepo.On("Save", mock.Anything).Return("", apperror.BadRequest("save error")).Once()

		id, err := service.AnalyzeWebsite(context.Background(), baseURL)

		assert.Error(t, err)
		assert.Empty(t, id)
		mockRepo.AssertExpectations(t)
	})
}
