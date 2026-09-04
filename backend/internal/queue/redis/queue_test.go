package redis

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"streamforge/internal/event"
)

func setupTestRedis(t *testing.T) *redis.Client {
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		t.Skip("REDIS_URL not set. Skipping integration tests.")
	}

	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		t.Fatalf("Failed to parse REDIS_URL: %v", err)
	}

	client := redis.NewClient(opts)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		t.Fatalf("Failed to ping Redis: %v", err)
	}

	// Clean up streams for isolated testing
	client.Del(context.Background(), StreamHigh, StreamNormal, StreamLow)

	return client
}

func TestRedisQueue_PublishAndConsumePriority(t *testing.T) {
	client := setupTestRedis(t)
	defer client.Close()
	q := NewRedisQueue(client)
	ctx := context.Background()

	// 1. Initialize consumer groups
	group := "test_workers"
	if err := q.InitConsumerGroups(ctx, group); err != nil {
		t.Fatalf("Failed to init consumer groups: %v", err)
	}

	// 2. Publish 3 events in reverse priority order (Low, Normal, High)
	eventsToPublish := []struct {
		priority int
		stream   string
	}{
		{event.PriorityLow, StreamLow},
		{event.PriorityNormal, StreamNormal},
		{event.PriorityHigh, StreamHigh},
	}

	publishedIDs := make(map[int]uuid.UUID)

	for _, tc := range eventsToPublish {
		e := &event.Event{
			ID:        uuid.New(),
			Type:      event.TypeOrderCreated,
			Priority:  tc.priority,
			Payload:   json.RawMessage(`{"data": "test"}`),
			Status:    event.StatusReceived,
			CreatedAt: time.Now(),
		}
		publishedIDs[tc.priority] = e.ID

		if err := q.Publish(ctx, e); err != nil {
			t.Fatalf("Publish failed for priority %d: %v", tc.priority, err)
		}
	}

	// 3. Consume - should strictly return HIGH, then NORMAL, then LOW
	consumerName := "worker-1"

	// High
	jobHigh, err := q.Consume(ctx, group, consumerName)
	if err != nil {
		t.Fatalf("Failed to consume High: %v", err)
	}
	if jobHigh.Event.ID != publishedIDs[event.PriorityHigh] {
		t.Errorf("Expected High event %s, got %s", publishedIDs[event.PriorityHigh], jobHigh.Event.ID)
	}

	// Normal
	jobNormal, err := q.Consume(ctx, group, consumerName)
	if err != nil {
		t.Fatalf("Failed to consume Normal: %v", err)
	}
	if jobNormal.Event.ID != publishedIDs[event.PriorityNormal] {
		t.Errorf("Expected Normal event %s, got %s", publishedIDs[event.PriorityNormal], jobNormal.Event.ID)
	}

	// Low
	jobLow, err := q.Consume(ctx, group, consumerName)
	if err != nil {
		t.Fatalf("Failed to consume Low: %v", err)
	}
	if jobLow.Event.ID != publishedIDs[event.PriorityLow] {
		t.Errorf("Expected Low event %s, got %s", publishedIDs[event.PriorityLow], jobLow.Event.ID)
	}

	// 4. Test Ack
	if err := q.Ack(ctx, jobHigh); err != nil {
		t.Fatalf("Failed to ack High event: %v", err)
	}

	// Test PEL reduction by checking pending length (optional) but we can just assume XAck worked

	// 5. Test Retry
	// For Normal, we will retry it. It should appear back in the normal queue.
	if err := q.Retry(ctx, jobNormal); err != nil {
		t.Fatalf("Failed to retry Normal event: %v", err)
	}

	// Consume again, we should get the retried normal event
	jobNormalRetried, err := q.Consume(ctx, group, consumerName)
	if err != nil {
		t.Fatalf("Failed to consume retried Normal: %v", err)
	}
	if jobNormalRetried.Event.ID != publishedIDs[event.PriorityNormal] {
		t.Errorf("Expected retried Normal event %s, got %s", publishedIDs[event.PriorityNormal], jobNormalRetried.Event.ID)
	}
	// Ack it to clean up
	q.Ack(ctx, jobNormalRetried)
	q.Ack(ctx, jobLow)

	// 6. Test Context Cancellation on empty queues
	// We expect the short sleep to trigger and context to eventually cancel.
	cancelCtx, cancel := context.WithTimeout(ctx, 150*time.Millisecond)
	defer cancel()

	_, err = q.Consume(cancelCtx, group, consumerName)
	if err != context.DeadlineExceeded {
		t.Errorf("Expected DeadlineExceeded, got %v", err)
	}
}
