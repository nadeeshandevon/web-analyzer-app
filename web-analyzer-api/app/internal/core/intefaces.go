package core

import (
	"context"
	"net/http"
	"net/url"
	"sync"
	"web-analyzer-api/app/internal/model"
	"web-analyzer-api/pkg/contract"
)

type LinkChecker interface {
	CheckLink(ctx context.Context, client *http.Client, link string, baseURL *url.URL) *model.LinkCheckResult
	RunWorker(ctx context.Context, linksChan <-chan string, resultsChan chan<- model.LinkCheckResult, baseURL *url.URL, wg *sync.WaitGroup)
}

type WebAnalyzerService interface {
	GetAnalyzeData(ctx context.Context, analyzeId string) (*contract.WebAnalyzeResponse, error)
	AnalyzeWebsite(ctx context.Context, baseURL *url.URL) (analysisId string, err error)
	UpdateAnalysisStatus(analyzeId string, status string, errorDescription string)
}
