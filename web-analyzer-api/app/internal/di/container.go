package di

import (
	v1 "web-analyzer-api/app/internal/api/v1"
	webanalyzer "web-analyzer-api/app/internal/core/web_analyzer"
	"web-analyzer-api/app/internal/repositorysql"
	"web-analyzer-api/app/internal/util/logger"
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
