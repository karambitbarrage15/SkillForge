package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"streamforge/internal/event"
	"streamforge/internal/store"
)

type EventRepo struct {
	pool *pgxpool.Pool
}

func NewEventRepo(db *DB) *EventRepo {
	return &EventRepo{pool: db.pool}
}

func (r *EventRepo) Create(ctx context.Context, e *event.Event) error {
	q := `INSERT INTO events (id, type, payload, priority, status, attempt, worker_id, created_at, updated_at, completed_at)
		  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`
	_, err := r.pool.Exec(ctx, q,
		e.ID, e.Type, e.Payload, e.Priority, e.Status, e.Attempt, e.WorkerID, e.CreatedAt, e.UpdatedAt, e.CompletedAt)
	return err
}

func (r *EventRepo) GetByID(ctx context.Context, id uuid.UUID) (*event.Event, error) {
	q := `SELECT id, type, payload, priority, status, attempt, worker_id, created_at, updated_at, completed_at
		  FROM events WHERE id = $1`

	row := r.pool.QueryRow(ctx, q, id)
	var e event.Event
	err := row.Scan(
		&e.ID, &e.Type, &e.Payload, &e.Priority, &e.Status, &e.Attempt, &e.WorkerID, &e.CreatedAt, &e.UpdatedAt, &e.CompletedAt)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, store.ErrEventNotFound
		}
		return nil, err
	}
	return &e, nil
}

func (r *EventRepo) UpdateStatus(ctx context.Context, id uuid.UUID, old, new event.EventStatus) error {
	// First enforce the domain state machine rules before hitting DB
	if !event.IsValidTransition(old, new) {
		return store.ErrInvalidStateTransition
	}

	// First check if it exists at all to distinguish between NotFound and InvalidState
	// Using a transaction to ensure atomic check-and-update behavior
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var currentStatus event.EventStatus
	err = tx.QueryRow(ctx, "SELECT status FROM events WHERE id = $1 FOR UPDATE", id).Scan(&currentStatus)
	if err != nil {
		if err == pgx.ErrNoRows {
			return store.ErrEventNotFound
		}
		return err
	}

	if currentStatus != old {
		return store.ErrInvalidStateTransition
	}

	var q string
	var args []any
	now := time.Now()

	if new == event.StatusCompleted {
		q = `UPDATE events SET status = $1, updated_at = $2, completed_at = $3 WHERE id = $4 AND status = $5`
		args = []any{new, now, now, id, old}
	} else {
		q = `UPDATE events SET status = $1, updated_at = $2 WHERE id = $3 AND status = $4`
		args = []any{new, now, id, old}
	}

	res, err := tx.Exec(ctx, q, args...)
	if err != nil {
		return err
	}

	if res.RowsAffected() == 0 {
		return store.ErrInvalidStateTransition
	}

	return tx.Commit(ctx)
}

func (r *EventRepo) List(ctx context.Context, limit, offset int) ([]*event.Event, error) {
	q := `SELECT id, type, payload, priority, status, attempt, worker_id, created_at, updated_at, completed_at
		  FROM events 
		  ORDER BY created_at DESC 
		  LIMIT $1 OFFSET $2`

	rows, err := r.pool.Query(ctx, q, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []*event.Event
	for rows.Next() {
		var e event.Event
		if err := rows.Scan(&e.ID, &e.Type, &e.Payload, &e.Priority, &e.Status, &e.Attempt, &e.WorkerID, &e.CreatedAt, &e.UpdatedAt, &e.CompletedAt); err != nil {
			return nil, err
		}
		events = append(events, &e)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return events, nil
}

func (r *EventRepo) RecordFailure(ctx context.Context, id uuid.UUID, maxAttempts int) (event.EventStatus, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return "", err
	}
	defer tx.Rollback(ctx)

	var currentStatus event.EventStatus
	var currentAttempt int
	err = tx.QueryRow(ctx, "SELECT status, attempt FROM events WHERE id = $1 FOR UPDATE", id).Scan(&currentStatus, &currentAttempt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return "", store.ErrEventNotFound
		}
		return "", err
	}

	if currentStatus != event.StatusProcessing {
		return "", store.ErrInvalidStateTransition
	}

	newAttempt := currentAttempt + 1
	var nextStatus event.EventStatus

	if newAttempt < maxAttempts {
		nextStatus = event.StatusRetrying
	} else {
		nextStatus = event.StatusFailed
	}

	if !event.IsValidTransition(currentStatus, nextStatus) {
		return "", store.ErrInvalidStateTransition
	}

	now := time.Now()
	var q string
	var args []any

	if nextStatus == event.StatusFailed {
		q = `UPDATE events SET status = $1, attempt = $2, updated_at = $3, completed_at = $4 WHERE id = $5`
		args = []any{nextStatus, newAttempt, now, now, id}
	} else {
		q = `UPDATE events SET status = $1, attempt = $2, updated_at = $3 WHERE id = $4`
		args = []any{nextStatus, newAttempt, now, id}
	}

	_, err = tx.Exec(ctx, q, args...)
	if err != nil {
		return "", err
	}

	return nextStatus, tx.Commit(ctx)
}
