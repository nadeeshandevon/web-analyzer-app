package di

import (
	v1 "web-analyzer-api/internal/api/v1"
	webanalyzer "web-analyzer-api/internal/core/web_analyzer"
	"web-analyzer-api/internal/repositorysql"
	"web-analyzer-api/internal/util/logger"
)

type HTTPHandlers struct {
	WebAnalyzerHandler v1.WebAnalyzerHandler // Handles web analyzer requests
}

type Container struct {
	HTTPHandlers HTTPHandlers
}

func NewContainer(logger *logger.Logger) *Container {
	webAnalyzerRepo := repositorysql.NewWebAnalyzerRepo(logger)
	webAnalyzerService := webanalyzer.NewWebAnalyzerService(logger, webAnalyzerRepo)
	webAnalyzerHandler := v1.NewWebAnalyzerHandler(logger, webAnalyzerService)

	logger.Info("Dependency injection container initialized successfully")

	return &Container{
		HTTPHandlers: HTTPHandlers{
			WebAnalyzerHandler: *webAnalyzerHandler,
		},
	}
}
