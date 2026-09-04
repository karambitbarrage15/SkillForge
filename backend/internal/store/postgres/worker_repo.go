package postgres

import (
	"context"
	"time"

	"streamforge/internal/store"
)

type WorkerRepo struct {
	pool *DB
}

func NewWorkerRepo(db *DB) *WorkerRepo {
	return &WorkerRepo{pool: db}
}

func (r *WorkerRepo) Upsert(ctx context.Context, w *store.Worker) error {
	q := `INSERT INTO workers (id, status, last_heartbeat, current_event_id, created_at, updated_at)
		  VALUES ($1, $2, $3, $4, $5, $6)
		  ON CONFLICT (id) DO UPDATE SET
		  status = EXCLUDED.status,
		  last_heartbeat = EXCLUDED.last_heartbeat,
		  current_event_id = EXCLUDED.current_event_id,
		  updated_at = EXCLUDED.updated_at`

	_, err := r.pool.pool.Exec(ctx, q, w.ID, w.Status, w.LastHeartbeat, w.CurrentEventID, w.CreatedAt, w.UpdatedAt)
	return err
}

func (r *WorkerRepo) MarkOffline(ctx context.Context, threshold time.Time) (int, error) {
	q := `UPDATE workers SET status = 'OFFLINE', updated_at = NOW()
		  WHERE status != 'OFFLINE' AND last_heartbeat < $1`

	res, err := r.pool.pool.Exec(ctx, q, threshold)
	if err != nil {
		return 0, err
	}
	return int(res.RowsAffected()), nil
}
