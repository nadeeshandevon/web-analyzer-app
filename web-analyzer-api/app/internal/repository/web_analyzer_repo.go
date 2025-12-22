package repository

import "web-analyzer-api/app/internal/model"

type WebAnalyzerRepository interface {
	Save(webAnalyzer model.WebAnalyzer) (string, error)
	Update(webAnalyzer model.WebAnalyzer) (string, error)
	GetById(id string) (*model.WebAnalyzer, error)
}
