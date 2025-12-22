package api

import (
	"web-analyzer-api/app/internal/api/middleware"
	"web-analyzer-api/app/internal/di"
	"web-analyzer-api/app/internal/util/logger"

	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func SetupRouter(router *gin.Engine, handlers di.HTTPHandlers, logger *logger.Logger) error {
	pprof.Register(router)
	router.Use(middleware.CORSMiddleware())
	router.Use(middleware.ErrorHandler(*logger))
	router.Use(middleware.PrometheusMiddleware())
	router.Use(gin.Recovery())

	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	v1 := router.Group("/api/v1")
	handlers.WebAnalyzerHandler.RegisterRoutes(v1)

	logger.Info("Router setup completed successfully")
	return nil
}
