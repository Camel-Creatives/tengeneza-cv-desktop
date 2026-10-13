// The worker process will consume render, email, reconciliation, and usage tasks.
// It is deliberately a distinct executable to support independent autoscaling.
package main

import (
	"log/slog"
	"os"

	"github.com/troubleman96/tengeneza-cv-desktop/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("configuration validation failed", "error", err)
		os.Exit(1)
	}
	slog.Info("worker bootstrap complete", "environment", cfg.Environment, "redis_configured", cfg.RedisURL != "")
}
