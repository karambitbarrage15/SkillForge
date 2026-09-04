package processor

import (
	"context"
	"errors"
	"fmt"

	"streamforge/internal/event"
	"streamforge/internal/worker"
)

var (
	ErrUnknownEventType = errors.New("unknown event type")
	ErrNilProcessor     = errors.New("processor cannot be nil")
)

// Registry manages the routing of events to their respective processors.
// It implements the worker.Processor interface so it can be used directly by the worker pool.
// Processors should be registered during application startup before workers begin processing.
// After startup, the registry should be treated as immutable/read-only for thread safety.
type Registry struct {
	processors map[event.EventType]worker.Processor
}

// NewRegistry creates a new, empty Processor Registry.
func NewRegistry() *Registry {
	return &Registry{
		processors: make(map[event.EventType]worker.Processor),
	}
}

// Register adds a processor for a specific event type.
// It returns an error if a processor is already registered for that type or if the processor is nil.
func (r *Registry) Register(eventType event.EventType, p worker.Processor) error {
	if p == nil {
		return ErrNilProcessor
	}

	if _, exists := r.processors[eventType]; exists {
		return fmt.Errorf("processor already registered for event type: %s", eventType)
	}

	r.processors[eventType] = p
	return nil
}

// Process routes the event to the registered processor for its type.
func (r *Registry) Process(ctx context.Context, e *event.Event) error {
	p, ok := r.processors[e.Type]
	if !ok {
		return fmt.Errorf("%w: %s", ErrUnknownEventType, e.Type)
	}

	return p.Process(ctx, e)
}
