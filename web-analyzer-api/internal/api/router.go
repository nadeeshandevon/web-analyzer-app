package api

import (
	"web-analyzer-api/internal/api/middleware"
	"web-analyzer-api/internal/di"
	"web-analyzer-api/internal/logger"

	"github.com/gin-gonic/gin"
)

func SetupRouter(router *gin.Engine, handlers di.HTTPHandlers, logger *logger.Logger) error {
	router.Use(middleware.ErrorHandler(*logger))
	router.Use(gin.Recovery())

	v1 := router.Group("/api/v1")
	handlers.WebAnalyzerHandler.RegisterRoutes(v1)

	logger.Info("Router setup completed successfully")
	return nil
}
