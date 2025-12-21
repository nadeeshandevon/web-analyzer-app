package core

import (
	"context"
	"web-analyzer-api/pkg/contract"
)

type WebAnalyzerService interface {
	AnalyzeWebsite(ctx context.Context, url string) (analysisId string, err error)
	GetAnalyzeData(ctx context.Context, analyzeId string) (*contract.WebAnalyzeResponse, error)
}
