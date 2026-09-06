package scheduler

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"

	"streamforge/internal/event"
	"streamforge/internal/metrics"
	"streamforge/internal/queue"
	redisQueue "streamforge/internal/queue/redis"
	"streamforge/internal/store"
)

type Scheduler struct {
	db        store.EventRepository
	redis     *redis.Client
	publisher queue.Publisher
	interval  time.Duration
}

func NewScheduler(db store.EventRepository, rdb *redis.Client, pub queue.Publisher, interval time.Duration) *Scheduler {
	return &Scheduler{
		db:        db,
		redis:     rdb,
		publisher: pub,
		interval:  interval,
	}
}

func (s *Scheduler) Start(ctx context.Context) {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			slog.Info("Scheduler Stopping")
			return
		case <-ticker.C:
			s.runSweep(ctx)
			s.updateMetrics(ctx)
		}
	}
}

func (s *Scheduler) runSweep(ctx context.Context) {
	s.recoverEvents(ctx, event.StatusProcessing, 30*time.Second)
	s.recoverEvents(ctx, event.StatusRetrying, 5*time.Minute)
}

func (s *Scheduler) recoverEvents(ctx context.Context, status event.EventStatus, age time.Duration) {
	cutoff := time.Now().Add(-age)

	events, err := s.db.GetStaleEvents(ctx, status, cutoff)
	if err != nil {
		slog.Error("Failed to get stale events", "status", status, "err", err)
		return
	}

	if len(events) == 0 {
		return
	}

	// For PROCESSING events, we must check if the worker is actually dead.
	deadWorkers := make(map[string]bool)

	for _, e := range events {
		if status == event.StatusProcessing {
			if e.WorkerID == nil {
				continue
			}
			workerID := e.WorkerID.String()

			// Check liveness if not already known
			if _, checked := deadWorkers[workerID]; !checked {
				key := fmt.Sprintf("worker:%s:heartbeat", workerID)
				res, err := s.redis.Exists(ctx, key).Result()
				if err != nil {
					slog.Error("Redis error checking heartbeat", "worker_id", workerID, "err", err)
					continue
				}

				deadWorkers[workerID] = (res == 0)
			}

			if !deadWorkers[workerID] {
				// Worker is alive, don't recover
				continue
			}
		}

		// Proceed with recovery
		slog.Info("Recovering stranded event", "status", status, "event_id", e.ID)

		// 1. Publish to Redis AT-LEAST-ONCE (payload status = QUEUED)
		e.Status = event.StatusQueued
		if err := s.publisher.Publish(ctx, e); err != nil {
			slog.Error("Failed to republish event to Redis", "event_id", e.ID, "err", err)
			continue // Skip DB update, we will try again next sweep
		}

		// 2. Update PostgreSQL (PROCESSING/RETRYING -> QUEUED)
		// We pass empty workerID because we are resetting ownership
		if err := s.db.AssignEvent(ctx, e.ID, "", status, event.StatusQueued); err != nil {
			slog.Error("Failed to reset DB status for event", "event_id", e.ID, "err", err)
			// We DO NOT rollback the Redis publish. Idempotency guarantees safe duplicate handling.
			continue
		}

		slog.Info("Successfully recovered event to QUEUED", "event_id", e.ID)
	}
}

func (s *Scheduler) updateMetrics(ctx context.Context) {
	// Update active workers
	keys, err := s.redis.Keys(ctx, "worker:*:heartbeat").Result()
	if err != nil {
		slog.Error("Failed to get active workers from Redis", "err", err)
	} else {
		metrics.ActiveWorkers.Set(float64(len(keys)))
	}

	// Update queue depth
	for _, stream := range redisQueue.AllStreams {
		l, err := s.redis.XLen(ctx, stream).Result()
		if err != nil {
			slog.Error("Failed to get stream length", "stream", stream, "err", err)
			continue
		}
		metrics.QueueDepth.WithLabelValues(stream).Set(float64(l))
	}
}
