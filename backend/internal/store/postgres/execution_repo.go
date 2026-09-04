package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"

	"streamforge/internal/store"
)

type ExecutionRepo struct {
	pool *DB
}

func NewExecutionRepo(db *DB) *ExecutionRepo {
	return &ExecutionRepo{pool: db}
}

func (r *ExecutionRepo) Create(ctx context.Context, ex *store.Execution) error {
	q := `INSERT INTO executions (id, event_id, worker_id, attempt, status, error, started_at, finished_at)
		  VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err := r.pool.pool.Exec(ctx, q,
		ex.ID, ex.EventID, ex.WorkerID, ex.Attempt, ex.Status, ex.Error, ex.StartedAt, ex.FinishedAt)
	return err
}

func (r *ExecutionRepo) Complete(ctx context.Context, id uuid.UUID, status string, errStr *string) error {
	q := `UPDATE executions SET status = $1, error = $2, finished_at = $3 WHERE id = $4`
	_, err := r.pool.pool.Exec(ctx, q, status, errStr, time.Now(), id)
	return err
}
