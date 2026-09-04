# Idempotency

Distributed processing generally cannot guarantee that a job will execute exactly once.

StreamForge therefore uses:

```
at-least-once delivery
+
idempotent finalization
```

## Event ID

Every event has a globally unique ID.

## Execution Attempts

Each processing attempt gets:

- event_id
- attempt_number
- worker_id

## Duplicate Finalization

Before marking an event `COMPLETED`:

verify that it has not already been finalized.

Use database constraints and transactions.

## Important Scenario

**Worker A:**

1. processes event
2. crashes before acknowledging queue

**Worker B:**

1. receives same event
2. processes event

Both workers may have executed the operation.

Therefore processors should be designed to be idempotent where possible.

The database finalization must never occur twice.
