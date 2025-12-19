package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"web-analyzer-api/internal/logger"

	"github.com/gin-gonic/gin"
)

func main() {

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	log := logger.Get(os.Getenv("LOG_LEVEL"))
	log.Info("Starting Web Analyzer application")

	router := gin.New()

	// err = api.SetupRouter(router, app.AppConfig, app.HTTPHandlers, log)
	// if err != nil {
	// 	log.Fatal().Err(err).Msg("Failed to setup router")
	// }
	// log.Info().Msg("Router configuration completed")

	// log.Info().Str("port", appConfig.ServerPort).Msg("Starting HTTP server")
	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%s", "8081"),
		Handler: router,
	}

	go func() {
		log.Info(fmt.Sprintf("address: %s", httpServer.Addr))
		if err := httpServer.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			log.Error(fmt.Sprintf("HTTP server failed unexpectedly: %v", err))
		}
	}()

	<-ctx.Done()
	log.Info("Shutdown signal received, initiating graceful shutdown")

	ctxTimeout, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctxTimeout); err != nil {
		log.Error(fmt.Sprintf("Failed to shutdown server gracefully: %v", err))
	}

	log.Info("HTTP server shutdown completed successfully")
}
