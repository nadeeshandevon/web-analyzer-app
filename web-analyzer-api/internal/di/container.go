package di

import (
	v1 "web-analyzer-api/internal/api/v1"
	"web-analyzer-api/internal/logger"
)

type HTTPHandlers struct {
	WebAnalyzerHandler v1.WebAnalyzerHandler // Handles web analyzer requests
}

type Container struct {
	HTTPHandlers HTTPHandlers // All HTTP request handlers
}

func NewContainer(logger *logger.Logger) *Container {
	webAnalyzerHandler := v1.NewWebAnalyzerHandler(logger)

	logger.Info("Dependency injection container initialized successfully")

	return &Container{
		HTTPHandlers: HTTPHandlers{
			WebAnalyzerHandler: *webAnalyzerHandler,
		},
	}
}
