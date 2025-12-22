package core

import (
	"context"
	"net/url"
	"web-analyzer-api/pkg/contract"
)

type WebAnalyzerService interface {
	AnalyzeWebsite(ctx context.Context, baseURL *url.URL) (analysisId string, err error)
	GetAnalyzeData(ctx context.Context, analyzeId string) (*contract.WebAnalyzeResponse, error)
}
