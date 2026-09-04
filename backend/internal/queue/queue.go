package queue

import (
	"context"

	"streamforge/internal/event"
)

// Publisher defines the contract for putting events onto the distributed queue.
type Publisher interface {
	Publish(ctx context.Context, e *event.Event) error
}

// Job wraps an Event with queue-specific metadata (like message ID and origin queue)
// so that the queue implementation can safely acknowledge or retry it later
// without leaking infrastructure details into the domain model.
type Job struct {
	Event   *event.Event
	Receipt string // Infrastructure-specific message ID (e.g., Redis Message ID)
	Queue   string // Infrastructure-specific queue name (e.g., events:high)
	Group   string // Consumer group that claimed the message
}

// Consumer defines the contract for taking events off the distributed queue.
type Consumer interface {
	Consume(ctx context.Context, consumerGroup, consumerName string) (*Job, error)
}

// Acker defines the contract for acknowledging successful processing.
type Acker interface {
	Ack(ctx context.Context, j *Job) error
}

// Retrier defines the contract for retrying a failed event.
type Retrier interface {
	Retry(ctx context.Context, j *Job) error
}

// NoopPublisher is a stub used in Phase 4 to satisfy the dependency flow
// before Redis is implemented in Phase 5. It does NOT actually queue the event.
type NoopPublisher struct{}

// Publish implements the Publisher interface as a no-op.
func (p *NoopPublisher) Publish(ctx context.Context, e *event.Event) error {
	// Do nothing in Phase 4
	return nil
}
