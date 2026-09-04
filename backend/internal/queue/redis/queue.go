package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"streamforge/internal/event"
	"streamforge/internal/queue"
)

const (
	StreamHigh   = "events:high"
	StreamNormal = "events:normal"
	StreamLow    = "events:low"
)

var AllStreams = []string{StreamHigh, StreamNormal, StreamLow}

type RedisQueue struct {
	client *redis.Client
}

func NewRedisQueue(client *redis.Client) *RedisQueue {
	return &RedisQueue{
		client: client,
	}
}

// InitConsumerGroups establishes the consumer group across all streams.
// This is typically called by the Worker on startup.
func (q *RedisQueue) InitConsumerGroups(ctx context.Context, group string) error {
	for _, stream := range AllStreams {
		// MKSTREAM creates the stream if it doesn't exist
		err := q.client.XGroupCreateMkStream(ctx, stream, group, "$").Err()
		if err != nil {
			// Ignore BUSYGROUP error (group already exists)
			if err.Error() != "BUSYGROUP Consumer Group name already exists" &&
				err.Error() != "ERR BUSYGROUP Consumer Group name already exists" {
				return fmt.Errorf("failed to create consumer group for stream %s: %w", stream, err)
			}
		}
	}
	return nil
}

// Publish implements queue.Publisher using XADD.
func (q *RedisQueue) Publish(ctx context.Context, e *event.Event) error {
	b, err := json.Marshal(e)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	stream := StreamLow
	switch e.Priority {
	case event.PriorityHigh:
		stream = StreamHigh
	case event.PriorityNormal:
		stream = StreamNormal
	}

	args := &redis.XAddArgs{
		Stream: stream,
		ID:     "*", // Auto-generated timestamp-based ID
		Values: map[string]interface{}{
			"data": string(b),
		},
	}

	return q.client.XAdd(ctx, args).Err()
}

func (q *RedisQueue) Consume(ctx context.Context, group, consumer string) (*queue.Job, error) {
	// Polling loop to block slightly without risking pulling low priority into PEL
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			// Check sequentially to guarantee Priority DESC
			for _, stream := range AllStreams {
				args := &redis.XReadGroupArgs{
					Group:    group,
					Consumer: consumer,
					Streams:  []string{stream, ">"},
					Count:    1,
					Block:    -1, // Non-blocking!
				}

				streams, err := q.client.XReadGroup(ctx, args).Result()
				if err != nil && err != redis.Nil {
					return nil, fmt.Errorf("failed to read from stream %s: %w", stream, err)
				}

				if len(streams) > 0 && len(streams[0].Messages) > 0 {
					msg := streams[0].Messages[0]
					dataStr, ok := msg.Values["data"].(string)
					if !ok {
						return nil, errors.New("invalid message format: missing 'data' field")
					}

					var e event.Event
					if err := json.Unmarshal([]byte(dataStr), &e); err != nil {
						return nil, fmt.Errorf("failed to unmarshal event: %w", err)
					}

					return &queue.Job{
						Event:   &e,
						Receipt: msg.ID,
						Queue:   stream,
						Group:   group,
					}, nil
				}
			}

			// All streams were empty. Sleep briefly to prevent tight loop,
			// then check again.
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(100 * time.Millisecond):
				// Continue loop
			}
		}
	}
}

// Ack implements queue.Acker
func (q *RedisQueue) Ack(ctx context.Context, j *queue.Job) error {
	if j.Receipt == "" || j.Queue == "" || j.Group == "" {
		return errors.New("invalid job: missing receipt, queue, or group metadata")
	}
	return q.client.XAck(ctx, j.Queue, j.Group, j.Receipt).Err()
}

// Retry implements queue.Retrier
func (q *RedisQueue) Retry(ctx context.Context, j *queue.Job) error {
	if j.Receipt == "" || j.Queue == "" || j.Group == "" {
		return errors.New("invalid job: missing receipt, queue, or group metadata")
	}

	// 1. Publish back to the queue (this generates a new Message ID and puts it at the back)
	if err := q.Publish(ctx, j.Event); err != nil {
		return fmt.Errorf("failed to republish event during retry: %w", err)
	}

	// 2. Acknowledge the old message so it's removed from PEL
	if err := q.Ack(ctx, j); err != nil {
		return fmt.Errorf("failed to ack original message during retry: %w", err)
	}

	return nil
}
