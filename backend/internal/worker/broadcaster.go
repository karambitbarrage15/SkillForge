package worker

import (
	"context"
	"log"

	"github.com/redis/go-redis/v9"
)

type RedisBroadcaster struct {
	client *redis.Client
}

func NewRedisBroadcaster(client *redis.Client) *RedisBroadcaster {
	return &RedisBroadcaster{client: client}
}

func (b *RedisBroadcaster) Broadcast(ctx context.Context, payload []byte) {
	// Execute the broadcast in a background goroutine so we never block critical path
	// A failure here does not break event ingestion or processing
	go func(data string) {
		err := b.client.Publish(context.Background(), "ws:events", data).Err()
		if err != nil {
			log.Printf("[RedisBroadcaster] Failed to broadcast message: %v", err)
		}
	}(string(payload))
}

type MockBroadcaster struct {
	Payloads [][]byte
}

func (m *MockBroadcaster) Broadcast(ctx context.Context, payload []byte) {
	m.Payloads = append(m.Payloads, append([]byte(nil), payload...))
}
