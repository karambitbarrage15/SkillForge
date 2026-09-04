# Failure Handling

Distributed systems fail.

StreamForge must assume components can fail.

## Worker Failure

Worker sends heartbeat.

If heartbeat expires:

- worker becomes `OFFLINE`
- scheduler finds unfinished event
- event returns to `QUEUED`
- another worker can process it

## Redis Failure

API should fail safely.

Do not acknowledge successful event ingestion if the event cannot be safely queued.

Document recovery behavior.

## PostgreSQL Failure

Do not claim event persistence succeeded.

Return appropriate error.

## Client Disconnect

Client WebSocket disconnect should not affect event processing.

Processing is server-side.

## Worker Crash During Processing

Possible sequence:

```
event = PROCESSING
worker crashes
heartbeat expires
scheduler detects worker failure
event is requeued
another worker processes event
```

This creates potential duplicate execution.

Therefore idempotency is required.
