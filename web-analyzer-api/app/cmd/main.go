package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"web-analyzer-api/app/internal/api"
	"web-analyzer-api/app/internal/di"
	"web-analyzer-api/app/internal/util/logger"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	log := logger.Get(os.Getenv("LOG_LEVEL"))
	log.Info("Starting Web Analyzer application")

	app := di.NewContainer(log)
	log.Info("Dependency injection container initialized")

	router := gin.New()

	err := api.SetupRouter(router, app.HTTPHandlers, log)
	if err != nil {
		log.Error("Failed to setup router", "error", err)
	}
	log.Info("Router configuration completed")

	serverPort := os.Getenv("SERVER_PORT")
	if serverPort == "" {
		serverPort = "8081"
	}

	metricsPort := os.Getenv("METRICS_PORT")
	if metricsPort == "" {
		metricsPort = "9090"
	}

	log.Info("Starting HTTP server", "port", serverPort)
	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%s", serverPort),
		Handler: router,
	}

	go func() {
		log.Info("HTTP server started", "address", httpServer.Addr)
		if err := httpServer.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			log.Error("HTTP server failed unexpectedly", "error", err)
		}
	}()

	log.Info("Starting Metrics server", "port", metricsPort)
	metricsRouter := gin.New()
	metricsRouter.Use(gin.Recovery())
	metricsRouter.GET("/metrics", gin.WrapH(promhttp.Handler()))

	metricsServer := &http.Server{
		Addr:    fmt.Sprintf(":%s", metricsPort),
		Handler: metricsRouter,
	}

	go func() {
		log.Info("Metrics server started", "address", metricsServer.Addr)
		if err := metricsServer.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			log.Error("Metrics server failed unexpectedly", "error", err)
		}
	}()

	<-ctx.Done()
	log.Info("Shutdown signal received, initiating graceful shutdown")

	ctxTimeout, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var errs []error
	if err := httpServer.Shutdown(ctxTimeout); err != nil {
		log.Error("Failed to shutdown main server gracefully", "error", err)
		errs = append(errs, err)
	}

	if err := metricsServer.Shutdown(ctxTimeout); err != nil {
		log.Error("Failed to shutdown metrics server gracefully", "error", err)
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		log.Error("Shutdown completed with errors")
	} else {
		log.Info("All servers shutdown completed successfully")
	}
}
