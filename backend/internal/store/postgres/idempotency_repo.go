package postgres

import (
	"context"

	"streamforge/internal/store"
)

type IdempotencyRepo struct {
	pool *DB
}

func NewIdempotencyRepo(db *DB) *IdempotencyRepo {
	return &IdempotencyRepo{pool: db}
}

func (r *IdempotencyRepo) CheckOrInsert(ctx context.Context, record *store.IdempotencyRecord) (bool, error) {
	q := `INSERT INTO idempotency (event_id, result_hash, created_at)
		  VALUES ($1, $2, $3)
		  ON CONFLICT (event_id) DO NOTHING`

	res, err := r.pool.pool.Exec(ctx, q, record.EventID, record.ResultHash, record.CreatedAt)
	if err != nil {
		return false, err
	}

	return res.RowsAffected() > 0, nil
}
