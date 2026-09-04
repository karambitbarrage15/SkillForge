# Database Design

## events

| Column       | Type      | Constraints             |
|--------------|-----------|-------------------------|
| id           | UUID      | PRIMARY KEY             |
| type         | VARCHAR   | NOT NULL                |
| payload      | JSONB     | NOT NULL                |
| priority     | INTEGER   | NOT NULL                |
| status       | VARCHAR   | NOT NULL                |
| attempt      | INTEGER   | NOT NULL DEFAULT 0      |
| worker_id    | UUID      | NULL                    |
| created_at   | TIMESTAMP | NOT NULL                |
| updated_at   | TIMESTAMP | NOT NULL                |
| completed_at | TIMESTAMP | NULL                    |

## workers

| Column           | Type      | Constraints |
|------------------|-----------|-------------|
| id               | UUID      | PRIMARY KEY |
| status           | VARCHAR   | NOT NULL    |
| last_heartbeat   | TIMESTAMP |             |
| current_event_id | UUID      | NULL        |
| created_at       | TIMESTAMP |             |
| updated_at       | TIMESTAMP |             |

## executions

| Column     | Type      | Constraints |
|------------|-----------|-------------|
| id         | UUID      | PRIMARY KEY |
| event_id   | UUID      | NOT NULL    |
| worker_id  | UUID      | NOT NULL    |
| attempt    | INTEGER   | NOT NULL    |
| started_at | TIMESTAMP |             |
| finished_at| TIMESTAMP |             |
| status     | VARCHAR   |             |
| error      | TEXT      |             |

## idempotency

| Column      | Type      | Constraints |
|-------------|-----------|-------------|
| event_id    | UUID      | PRIMARY KEY |
| result_hash | TEXT      |             |
| created_at  | TIMESTAMP |             |

## Indexes

```sql
CREATE INDEX ON events(status);
CREATE INDEX ON events(priority);
CREATE INDEX ON events(created_at);
CREATE INDEX ON events(worker_id);
CREATE INDEX ON executions(event_id);
CREATE INDEX ON workers(status);
CREATE INDEX ON workers(last_heartbeat);
```

## Requirements

- Use foreign keys where appropriate.
- Use unique constraints where required.
- Use transactions for operations that must remain atomic.
