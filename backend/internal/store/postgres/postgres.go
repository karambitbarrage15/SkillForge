package postgres

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed schema.sql
var schemaSQL string

// DB wraps a pgxpool.Pool to provide our specific database initialization logic.
type DB struct {
	pool *pgxpool.Pool
}

// Connect initializes the connection pool and applies the embedded schema.
func Connect(ctx context.Context, databaseURL string) (*DB, error) {
	poolConfig, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database url: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Apply schema
	if _, err := pool.Exec(ctx, schemaSQL); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to execute schema migrations: %w", err)
	}

	return &DB{pool: pool}, nil
}

// Close closes the underlying connection pool.
func (db *DB) Close() {
	db.pool.Close()
}
