// Tengeneza CV API entrypoint. This process only serves HTTP; background work is
// intentionally delegated to cmd/worker so request latency stays predictable.
package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/troubleman96/tengeneza-cv-desktop/internal/config"
	"github.com/troubleman96/tengeneza-cv-desktop/internal/httpapi"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("configuration validation failed", "error", err)
		os.Exit(1)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: cfg.LogLevel}))
	app, err := httpapi.NewWithError(cfg, logger)
	if err != nil {
		logger.Error("API bootstrap failed", "error", err)
		os.Exit(1)
	}

	go func() {
		logger.Info("tengenezacv API listening", "port", cfg.Port, "environment", cfg.Environment)
		if err := app.Listen(":" + cfg.Port); err != nil {
			logger.Error("http server stopped", "error", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := app.ShutdownWithContext(ctx); err != nil {
		logger.Error("graceful shutdown failed", "error", err)
	}
}
