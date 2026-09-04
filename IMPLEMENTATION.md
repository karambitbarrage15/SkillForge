# Implementation Plan

Implement the project incrementally.

DO NOT generate the entire system at once.

After every phase:

1. compile
2. run tests
3. run race detector
4. fix errors
5. update TODO.md
6. update documentation if architecture changed

---

# Phase 1 — Go Foundation

Create:

```
backend/go.mod

cmd/api/main.go
cmd/worker/main.go
cmd/simulator/main.go
```

Implement:

- configuration
- HTTP server
- graceful shutdown
- /health
- /ready

Tests must pass.

---

# Phase 2 — Event Model

Create:

```
event/model.go
event/types.go
event/validator.go
```

Implement Event struct.

Fields:

- ID
- Type
- Payload
- Priority
- Status
- CreatedAt
- UpdatedAt
- Attempt

Implement validation.

---

# Phase 3 — PostgreSQL

Implement:

- postgres connection
- event repository
- migrations

Create tables:

- events
- workers
- executions
- idempotency

Use transactions where required.

---

# Phase 4 — Event API

Implement:

```
POST /api/v1/events
```

Flow:

```
request
→ validation
→ create event
→ database
→ queue
→ response
```

Do not execute the event inside the HTTP handler.

---

# Phase 5 — Queue

Create queue interface.

Implement Redis-backed queue.

Required operations:

- Publish
- Consume
- Ack
- Retry

Queue must support priority.

---

# Phase 6 — Worker

Create Worker.

Worker lifecycle:

```
STARTING
READY
BUSY
STOPPING
STOPPED
```

Worker should:

1. register
2. start heartbeat
3. consume jobs
4. process jobs
5. acknowledge jobs
6. update status
7. gracefully shut down

---

# Phase 7 — Worker Pool

Create bounded worker pool.

Configuration:

- worker_count
- queue_size

Workers must not be created indefinitely.

Use goroutines and channels correctly.

---

# Phase 8 — State Machine

Implement explicit event state machine.

States:

```
RECEIVED
QUEUED
PROCESSING
COMPLETED
RETRYING
FAILED
```

Only valid transitions are allowed.

---

# Phase 9 — Processor

Create processor registry.

Example:

```
ORDER_CREATED     → order processor
PAYMENT_SUCCESS   → payment processor
INVENTORY_UPDATED → inventory processor
```

Each processor should implement:

```go
Process(ctx, event)
```

Do not create one huge switch statement.

---

# Phase 10 — Retry

If processing fails:

```
attempt++
```

If `attempt < maxAttempts`:

```
RETRYING → QUEUED
```

Otherwise:

```
FAILED
```

Implement exponential backoff.

---

# Phase 11 — Idempotency

Every event has a unique ID.

An event must not be finalized twice.

Use PostgreSQL constraints and application-level checks.

The finalization operation must be atomic.

---

# Phase 12 — WebSockets

Create WebSocket hub.

Clients connect to:

```
/ws
```

Broadcast:

```
EVENT_RECEIVED
EVENT_PROCESSING
EVENT_COMPLETED
EVENT_FAILED
WORKER_ONLINE
WORKER_OFFLINE
```

Implement:

- read loop
- write loop
- heartbeat
- bounded outgoing queue
- graceful disconnect

---

# Phase 13 — Dashboard

Build React dashboard.

Display:

- event rate
- queue depth
- worker count
- active workers
- completed events
- failed events
- recent events

Update in real time.

---

# Phase 14 — Failure Recovery

Implement worker heartbeat.

If:

```
current_time - last_heartbeat > timeout
```

mark worker `OFFLINE`.

Find jobs owned by worker.

Return unfinished jobs to queue.

Do not recover jobs already finalized.

---

# Phase 15 — Observability

Add Prometheus metrics.

Metrics:

```
events_received_total
events_processed_total
events_failed_total
event_processing_duration
queue_depth
active_workers
worker_failures_total
websocket_connections
```

Add structured logs.

---

# Phase 16 — Docker

Create production Dockerfiles.

Use multi-stage builds.

Do not include development tooling in runtime images.

---

# Phase 17 — Kubernetes

Create:

- API Deployment
- Worker Deployment
- Services
- ConfigMap
- Secrets
- HPA

Workers should scale horizontally.

---

# Phase 18 — Load Testing

Create simulator.

Generate:

- 1,000 events/sec
- 5,000 events/sec
- 10,000 events/sec

Measure:

- throughput
- p50 latency
- p95 latency
- p99 latency
- CPU
- memory
- queue depth

---

# Phase 19 — Final Hardening

Run:

```bash
go test ./...
go test -race ./...
go vet ./...
```

Run load tests.

Fix race conditions.

Test worker failure.

Test Redis failure.

Test PostgreSQL failure.

Test graceful shutdown.

Update documentation.
