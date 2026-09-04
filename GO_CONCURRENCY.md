# Go Concurrency Design

Concurrency is a central part of StreamForge.

## Goroutines

Use goroutines for:

- worker loops
- heartbeat
- WebSocket read/write
- background recovery
- asynchronous metric collection

## Channels

Use channels for:

- worker job distribution
- worker pool coordination
- WebSocket outbound messages

## Worker Pool

Do not spawn unlimited goroutines.

Use bounded workers.

Conceptual architecture:

```
Producer
    |
    v
Buffered Channel
    |
    +--- Worker
    +--- Worker
    +--- Worker
```

## Mutex

Use mutexes only for shared in-memory state.

Prefer message passing when practical.

## Context

Every long-running operation should accept `context.Context`.

Context is used for:

- cancellation
- shutdown
- request deadlines

## Graceful Shutdown

Shutdown sequence:

1. stop accepting new work
2. stop scheduler
3. stop consuming new queue messages
4. finish current jobs when possible
5. close WebSocket connections
6. close Redis
7. close PostgreSQL
8. exit

## Race Detection

All code must pass:

```bash
go test -race ./...
```

Never ignore race detector warnings.
