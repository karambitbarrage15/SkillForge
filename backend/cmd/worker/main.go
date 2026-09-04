// Package main is the entry point for the StreamForge worker process.
//
// The worker process hosts:
//   - The bounded worker pool (Phase 7)
//   - Per-worker heartbeat goroutines (Phase 6)
//   - The processor registry (Phase 9)
//   - The scheduler goroutine for failure recovery (Phase 14)
//
// Phase 1: process starts, logs readiness, and waits for a shutdown signal.
// Subsequent phases will wire in the actual worker pool and dependencies.
package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"streamforge/internal/config"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	cfg := config.Load()

	// cfg will be used by the worker pool in Phase 7.
	_ = cfg

	logger.Info("worker process starting")

	// Block until shutdown signal.
	// Phase 6 will replace this with the worker pool lifecycle.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	sig := <-quit

	logger.Info("worker process stopping", "signal", sig)
	// Phase 7 will drain in-flight jobs here before exiting.
	logger.Info("worker process stopped")
}
