package main

import (
	"errors"
	"log/slog"
	"net/http"
	"os"
	"time"

	"allai/backend/internal/api"
	"allai/backend/internal/config"
	"allai/backend/internal/provider"
)

func main() {
	cfg := config.Load()
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	openRouter := provider.NewOpenRouter(cfg.OpenRouterKey, cfg.FrontendURL, cfg.AppName)
	server := api.NewServer(openRouter, cfg.FrontendURL, logger)

	httpServer := &http.Server{
		Addr:              cfg.Host + ":" + cfg.Port,
		Handler:           server.Handler(),
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	logger.Info("Allai API listening", "address", "http://"+cfg.Host+":"+cfg.Port, "openrouter_configured", openRouter.Configured())
	if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Error("server stopped", "error", err)
		os.Exit(1)
	}
}
