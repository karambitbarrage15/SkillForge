# Event Lifecycle

Every event passes through a defined lifecycle from ingestion to finalization.

## Lifecycle Stages

### 1. Ingestion
- Client sends `POST /api/v1/events`
- API validates request (type, payload, priority)
- API generates a globally unique UUID
- Event is persisted to PostgreSQL with status `RECEIVED`
- Event is published to Redis queue
- Event transitions to `QUEUED`
- API returns `{id, status: "QUEUED"}` immediately

The API does not execute the event. It returns as soon as the event is safely persisted and enqueued.

### 2. Queueing
- Event sits in Redis queue ordered by:
  1. `priority DESC` (10=HIGH before 5=NORMAL before 1=LOW)
  2. `created_at ASC` (older events first within same priority)

### 3. Assignment
- Scheduler inspects available workers
- A READY worker is selected
- Event is assigned: `worker_id` is recorded on the event
- Status transitions: `QUEUED → PROCESSING`
- An `executions` row is created for this attempt

### 4. Processing
- Worker executes the appropriate processor for the event type
- Worker holds a Redis Streams pending entry for this event
- Processing is bounded by `context.Context` with a deadline

### 5. Success path
- Processor completes without error
- Within a single PostgreSQL transaction:
  - `INSERT INTO idempotency (event_id) ON CONFLICT DO NOTHING`
  - `UPDATE events SET status=COMPLETED, completed_at=now()`
  - `UPDATE executions SET status=COMPLETED, finished_at=now()`
- Redis Streams `XACK` is sent
- WebSocket hub broadcasts `EVENT_COMPLETED`

### 6. Failure path
- Processor returns an error
- `attempt` is incremented
- If `attempt < maxAttempts (3)`:
  - Status: `PROCESSING → RETRYING`
  - Exponential backoff applied
  - Status: `RETRYING → QUEUED`
  - Event is re-published to queue
- If `attempt >= maxAttempts`:
  - Status: `PROCESSING → FAILED`
  - Redis Streams `XACK` is sent
  - WebSocket hub broadcasts `EVENT_FAILED`

### 7. Recovery (worker failure)
- Worker stops sending heartbeats
- Redis TTL on `worker:{id}:heartbeat` expires
- Scheduler detects expired heartbeat
- Worker is marked `OFFLINE`
- Events in `PROCESSING` owned by this worker are identified
- Events not yet finalized are transitioned back to `QUEUED`
- Another worker picks them up (idempotency guard prevents double-finalization)

## Terminal States

Once an event reaches `COMPLETED` or `FAILED` it cannot be reprocessed.

These are terminal states. Any attempt to transition out of them returns an error.

## Observability

At every stage transition, the WebSocket hub broadcasts an update.

Every transition is logged with `event_id`, `worker_id`, `event_type`, `attempt`.
