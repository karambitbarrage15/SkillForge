package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"streamforge/internal/config"
	"streamforge/internal/event"
	"streamforge/internal/queue/redis"
	"streamforge/internal/worker"

	goredis "github.com/redis/go-redis/v9"
)

// NoOpProcessor is a temporary processor used only for Phase 7 infrastructure testing.
type NoOpProcessor struct{}

func (p *NoOpProcessor) Process(ctx context.Context, e *event.Event) error {
	slog.Info("Dummy processing event", "id", e.ID, "type", e.Type)
	return nil
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	logger.Info("Starting StreamForge Worker Pool...")

	cfg := config.Load()

	if cfg.RedisURL == "" {
		logger.Error("REDIS_URL is required")
		os.Exit(1)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	opts, err := goredis.ParseURL(cfg.RedisURL)
	if err != nil {
		logger.Error("failed to parse redis url", "err", err)
		os.Exit(1)
	}
	rdb := goredis.NewClient(opts)
	defer rdb.Close()

	if err := rdb.Ping(ctx).Err(); err != nil {
		logger.Error("failed to connect to redis", "err", err)
		os.Exit(1)
	}

	rQueue := redis.NewRedisQueue(rdb)
	if err := rQueue.InitConsumerGroups(ctx, cfg.WorkerConsumerGroup); err != nil {
		logger.Error("failed to initialize consumer groups", "err", err)
		os.Exit(1)
	}

	heartbeater := worker.NewRedisHeartbeater(rdb)
	proc := &NoOpProcessor{}

	pool := worker.NewPool(
		cfg.WorkerCount,
		cfg.WorkerConsumerGroup,
		rQueue,
		rQueue,
		heartbeater,
		proc,
		cfg.HeartbeatInterval,
		cfg.HeartbeatTTL,
	)

	logger.Info("Starting worker pool", "count", cfg.WorkerCount, "group", cfg.WorkerConsumerGroup)
	if err := pool.Start(ctx); err != nil {
		logger.Error("failed to start worker pool", "err", err)
		os.Exit(1)
	}

	<-ctx.Done()
	logger.Info("Received shutdown signal, stopping pool gracefully...")

	pool.Stop()
	logger.Info("Worker pool stopped cleanly.")
}
