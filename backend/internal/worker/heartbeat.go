package worker

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Heartbeater defines the interface for worker registration and liveness checks.
type Heartbeater interface {
	Ping(ctx context.Context, workerID string, ttl time.Duration) error
}

// RedisHeartbeater implements Heartbeater using Redis.
type RedisHeartbeater struct {
	client *redis.Client
}

// NewRedisHeartbeater creates a new RedisHeartbeater.
func NewRedisHeartbeater(client *redis.Client) *RedisHeartbeater {
	return &RedisHeartbeater{
		client: client,
	}
}

// Ping writes the heartbeat key to Redis with the specified TTL.
func (h *RedisHeartbeater) Ping(ctx context.Context, workerID string, ttl time.Duration) error {
	key := fmt.Sprintf("worker:%s:heartbeat", workerID)
	// We just set a basic payload "alive", the existence and TTL is what matters.
	err := h.client.Set(ctx, key, "alive", ttl).Err()
	if err != nil {
		return fmt.Errorf("redis heartbeat failed: %w", err)
	}
	return nil
}
