# WebSocket Protocol

Endpoint:

```
GET /ws
```

## Server → Client

### EVENT_RECEIVED

```json
{
  "type": "EVENT_RECEIVED",
  "event_id": "uuid"
}
```

### EVENT_PROCESSING

```json
{
  "type": "EVENT_PROCESSING",
  "event_id": "uuid",
  "worker_id": "worker-1"
}
```

### EVENT_COMPLETED

```json
{
  "type": "EVENT_COMPLETED",
  "event_id": "uuid",
  "processing_time_ms": 42
}
```

### EVENT_FAILED

```json
{
  "type": "EVENT_FAILED",
  "event_id": "uuid",
  "attempt": 3
}
```

### WORKER_ONLINE

```json
{
  "type": "WORKER_ONLINE",
  "worker_id": "worker-1"
}
```

### WORKER_OFFLINE

```json
{
  "type": "WORKER_OFFLINE",
  "worker_id": "worker-1"
}
```

## Client → Server

### PING

```json
{
  "type": "PING"
}
```

## Rules

- Messages must be validated.
- Maximum message size must be configured.
- Outgoing buffers must be bounded.
- Slow clients must eventually be disconnected.
