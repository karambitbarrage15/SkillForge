// Package main is the entry point for the StreamForge load simulator.
//
// The simulator generates synthetic events at configurable rates to
// drive load testing scenarios (Phase 18).
//
// Phase 1: binary compiles and exits cleanly. No simulation logic yet.
package main

import (
	"log/slog"
	"os"

	"streamforge/internal/config"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	cfg := config.Load()

	// cfg will be used by the simulator in Phase 18.
	_ = cfg

	logger.Info("simulator not yet implemented — see Phase 18")
	os.Exit(0)
}
