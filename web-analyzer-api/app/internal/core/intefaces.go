package core

import (
	"context"
	"net/url"
	"web-analyzer-api/pkg/contract"
)

type WebAnalyzerService interface {
	GetAnalyzeData(ctx context.Context, analyzeId string) (*contract.WebAnalyzeResponse, error)
	AnalyzeWebsite(ctx context.Context, baseURL *url.URL) (analysisId string, err error)
	UpdateAnalysisStatus(analyzeId string, status string, errorDescription string)
}
