package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"streamforge/internal/config"
	"streamforge/internal/event"
	"streamforge/internal/processor"
	"streamforge/internal/queue/redis"
	"streamforge/internal/store/postgres"
	"streamforge/internal/worker"

	goredis "github.com/redis/go-redis/v9"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	logger.Info("Starting StreamForge Worker Pool...")

	cfg := config.Load()

	if cfg.RedisURL == "" {
		logger.Error("REDIS_URL is required")
		os.Exit(1)
	}

	if cfg.DatabaseURL == "" {
		logger.Error("DATABASE_URL is required")
		os.Exit(1)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	db, err := postgres.NewDB(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("failed to connect to postgres", "err", err)
		os.Exit(1)
	}
	defer db.Close()

	repo := postgres.NewEventRepo(db)

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

	registry := processor.NewRegistry()

	// Register dummy processors
	if err := registry.Register(event.TypeOrderCreated, processor.NewOrderProcessor()); err != nil {
		logger.Error("failed to register processor", "err", err)
		os.Exit(1)
	}
	if err := registry.Register(event.TypePaymentSuccess, processor.NewPaymentSuccessProcessor()); err != nil {
		logger.Error("failed to register processor", "err", err)
		os.Exit(1)
	}
	if err := registry.Register(event.TypePaymentFailed, processor.NewPaymentFailedProcessor()); err != nil {
		logger.Error("failed to register processor", "err", err)
		os.Exit(1)
	}
	if err := registry.Register(event.TypeInventoryUpdated, processor.NewInventoryProcessor()); err != nil {
		logger.Error("failed to register processor", "err", err)
		os.Exit(1)
	}
	if err := registry.Register(event.TypeOrderCancelled, processor.NewOrderCancelledProcessor()); err != nil {
		logger.Error("failed to register processor", "err", err)
		os.Exit(1)
	}
	if err := registry.Register(event.TypeUserRegistered, processor.NewUserRegisteredProcessor()); err != nil {
		logger.Error("failed to register processor", "err", err)
		os.Exit(1)
	}

	pool := worker.NewPool(
		cfg.WorkerCount,
		cfg.WorkerConsumerGroup,
		rQueue,
		rQueue,
		rQueue,
		heartbeater,
		registry,
		repo,
		repo,
		cfg.HeartbeatInterval,
		cfg.HeartbeatTTL,
		3,             // maxAttempts
		1*time.Second, // baseDelay
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
