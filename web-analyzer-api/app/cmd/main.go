package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"web-analyzer-api/internal/api"
	"web-analyzer-api/internal/di"
	"web-analyzer-api/internal/util/logger"

	"github.com/gin-gonic/gin"
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

	log.Info("Starting HTTP server", "port", "8081")
	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%s", "8081"),
		Handler: router,
	}

	go func() {
		log.Info("HTTP server started", "address", httpServer.Addr)
		if err := httpServer.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			log.Error("HTTP server failed unexpectedly", "error", err)
		}
	}()

	<-ctx.Done()
	log.Info("Shutdown signal received, initiating graceful shutdown")

	ctxTimeout, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctxTimeout); err != nil {
		log.Error("Failed to shutdown server gracefully", "error", err)
	}

	log.Info("HTTP server shutdown completed successfully")
}
