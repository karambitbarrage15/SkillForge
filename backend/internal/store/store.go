package store

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"streamforge/internal/event"
)

var (
	ErrEventNotFound          = errors.New("event not found")
	ErrInvalidStateTransition = errors.New("invalid state transition")
)

// Worker represents a node in the worker pool.
type Worker struct {
	ID             uuid.UUID
	Status         string
	LastHeartbeat  time.Time
	CurrentEventID *uuid.UUID
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// Execution represents an attempt to process an event.
type Execution struct {
	ID         uuid.UUID
	EventID    uuid.UUID
	WorkerID   uuid.UUID
	Attempt    int
	Status     string
	Error      *string
	StartedAt  time.Time
	FinishedAt *time.Time
}

// IdempotencyRecord prevents duplicate processing results.
type IdempotencyRecord struct {
	EventID    uuid.UUID
	ResultHash string
	CreatedAt  time.Time
}

// EventRepository handles persistence for Events.
type EventRepository interface {
	Create(ctx context.Context, e *event.Event) error
	GetByID(ctx context.Context, id uuid.UUID) (*event.Event, error)
	// UpdateStatus uses optimistic locking (WHERE status = old).
	// It returns ErrEventNotFound if the ID doesn't exist.
	// It returns ErrInvalidStateTransition if the ID exists but the status does not match old.
	UpdateStatus(ctx context.Context, id uuid.UUID, old, new event.EventStatus) error
	List(ctx context.Context, limit, offset int) ([]*event.Event, error)
}

// WorkerRepository tracks active workers.
type WorkerRepository interface {
	Upsert(ctx context.Context, w *Worker) error
	MarkOffline(ctx context.Context, threshold time.Time) (int, error)
}

// ExecutionRepository records processing attempts.
type ExecutionRepository interface {
	Create(ctx context.Context, ex *Execution) error
	Complete(ctx context.Context, id uuid.UUID, status string, errStr *string) error
}

// IdempotencyRepository guarantees at-most-once delivery of results.
type IdempotencyRepository interface {
	// CheckOrInsert attempts to insert a record.
	// Returns true if successfully inserted.
	// Returns false if a record for the EventID already exists.
	CheckOrInsert(ctx context.Context, record *IdempotencyRecord) (bool, error)
}
