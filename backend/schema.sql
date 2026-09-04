CREATE TABLE IF NOT EXISTS events (
    id UUID PRIMARY KEY,
    type VARCHAR NOT NULL,
    payload JSONB NOT NULL,
    priority INTEGER NOT NULL,
    status VARCHAR NOT NULL,
    attempt INTEGER NOT NULL DEFAULT 0,
    worker_id UUID NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    completed_at TIMESTAMP NULL
);

CREATE TABLE IF NOT EXISTS workers (
    id UUID PRIMARY KEY,
    status VARCHAR NOT NULL,
    last_heartbeat TIMESTAMP,
    current_event_id UUID NULL,
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);

CREATE TABLE IF NOT EXISTS executions (
    id UUID PRIMARY KEY,
    event_id UUID NOT NULL,
    worker_id UUID NOT NULL,
    attempt INTEGER NOT NULL,
    started_at TIMESTAMP,
    finished_at TIMESTAMP,
    status VARCHAR,
    error TEXT
);

CREATE TABLE IF NOT EXISTS idempotency (
    event_id UUID PRIMARY KEY,
    result_hash TEXT,
    created_at TIMESTAMP
);
