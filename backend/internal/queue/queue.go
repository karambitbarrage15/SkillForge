package queue

import (
	"context"
	"log/slog"

	"streamforge/internal/event"
)

// Publisher defines the contract for putting events onto the distributed queue.
type Publisher interface {
	Publish(ctx context.Context, e *event.Event) error
}

// NoopPublisher is a stub used in Phase 4 to satisfy the dependency flow
// before Redis is implemented in Phase 5. It does NOT actually queue the event.
type NoopPublisher struct{}

// Publish mocks a successful publish operation.
func (n *NoopPublisher) Publish(ctx context.Context, e *event.Event) error {
	slog.Debug("NoopPublisher: simulating publish", "event_id", e.ID)
	return nil
}
