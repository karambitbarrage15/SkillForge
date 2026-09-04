# Product Requirements

## Product

StreamForge

## Problem

Modern applications generate large numbers of events.

Processing every event synchronously inside the API creates several problems:

- slow API responses
- poor scalability
- blocked request handlers
- difficult retry handling
- difficult failure recovery

StreamForge separates event ingestion from event processing.

## Core Flow

1. Client submits event
2. API validates event
3. Event receives unique ID
4. Event is persisted
5. Event enters queue
6. Scheduler assigns event
7. Worker processes event
8. Result is persisted
9. Event state changes
10. WebSocket broadcasts update

## MVP Event Types

```
ORDER_CREATED
PAYMENT_SUCCESS
PAYMENT_FAILED
INVENTORY_UPDATED
ORDER_CANCELLED
```

## Event Properties

Every event contains:

- id
- type
- payload
- priority
- created_at
- status

## Priority

```
1  = LOW
5  = NORMAL
10 = HIGH
```

Higher priority events should be processed first.

For equal priority, older events should be preferred.

## Requirements

### Event Ingestion

```
POST /api/v1/events
```

The endpoint must:

- validate request
- generate event ID
- persist event
- enqueue event
- return immediately

The API must not synchronously execute the event.

### Processing

Workers consume events from the queue.

Workers process events concurrently.

### Retry

Failed events should be retried.

Maximum attempts: `3`

After maximum attempts: `FAILED`

### Worker Health

Workers send heartbeats.

The system detects workers that stop sending heartbeats.

### Dashboard

Dashboard displays:

- events received
- queue depth
- events processing
- completed events
- failed events
- worker status
- processing latency

### Real-Time Updates

Dashboard receives updates through WebSockets.

## Non-Functional Requirements

The system should:

- avoid data races
- support graceful shutdown
- use bounded queues
- prevent duplicate finalization
- handle worker failures
- provide structured logs
- expose metrics
