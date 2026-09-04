// Package main is the entry point for the StreamForge API server.
//
// Phase 1 implements:
//   - HTTP server with configurable port and timeouts
//   - GET /health  — liveness probe
//   - GET /ready   — readiness probe (dependencies checked in later phases)
//   - Graceful shutdown on SIGTERM / SIGINT
package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"streamforge/internal/api"
	"streamforge/internal/config"

	"streamforge/internal/queue/redis"
	"streamforge/internal/store/postgres"

	goredis "github.com/redis/go-redis/v9"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	cfg := config.Load()

	// Initialize database connection
	ctx := context.Background()
	db, err := postgres.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("failed to connect to database", "err", err)
		os.Exit(1)
	}
	defer db.Close()

	// Initialize Redis connection
	opts, err := goredis.ParseURL(cfg.RedisURL)
	if err != nil {
		logger.Error("failed to parse redis url", "err", err)
		os.Exit(1)
	}
	rdb := goredis.NewClient(opts)
	if err := rdb.Ping(ctx).Err(); err != nil {
		logger.Error("failed to connect to redis", "err", err)
		os.Exit(1)
	}
	defer rdb.Close()

	// Initialize repositories and publisher
	eventRepo := postgres.NewEventRepo(db)
	publisher := redis.NewRedisQueue(rdb)

	eventHandler := api.NewEventHandler(eventRepo, publisher)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", handleHealth)
	mux.HandleFunc("GET /ready", handleReady)
	mux.HandleFunc("POST /api/v1/events", eventHandler.HandleCreateEvent)
	mux.HandleFunc("GET /api/v1/events/{id}", eventHandler.HandleGetEventByID)
	mux.HandleFunc("GET /api/v1/events", eventHandler.HandleListEvents)

	srv := &http.Server{
		Addr:         ":" + cfg.HTTPPort,
		Handler:      mux,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	}

	// Start the HTTP server in a background goroutine.
	// The main goroutine blocks on the shutdown signal.
	serverErr := make(chan error, 1)
	go func() {
		logger.Info("API server starting", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErr <- err
		}
	}()

	// Block until SIGTERM, SIGINT, or a fatal server error.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)

	select {
	case sig := <-quit:
		logger.Info("shutdown signal received", "signal", sig)
	case err := <-serverErr:
		logger.Error("fatal server error", "err", err)
		os.Exit(1)
	}

	// Graceful shutdown: give in-flight requests time to complete.
	ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	logger.Info("shutting down API server", "timeout", cfg.ShutdownTimeout)
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("server shutdown error", "err", err)
		os.Exit(1)
	}

	logger.Info("API server stopped cleanly")
}

// handleHealth is the liveness probe.
// Returns 200 OK as long as the process is running.
func handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// handleReady is the readiness probe.
// Phase 1: always ready — no dependencies to check yet.
// Later phases will verify PostgreSQL and Redis connectivity before returning 200.
func handleReady(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
}

// writeJSON encodes v as JSON and writes it to w with the given status code.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		// At this point the header is already sent; we can only log.
		slog.Error("writeJSON encode error", "err", err)
	}
}
