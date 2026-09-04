# Testing

## Unit Tests

Test:

- state machine
- scoring/processing
- priority ordering
- retry logic
- idempotency
- worker lifecycle

## Integration Tests

Test:

- API + PostgreSQL
- API + Redis
- Worker + Redis
- Worker + PostgreSQL

## WebSocket Tests

Test:

- connection
- disconnect
- broadcast
- slow client
- heartbeat

## Concurrency Tests

Test:

- multiple workers processing simultaneously
- two workers receiving same event
- simultaneous state transitions
- duplicate finalization
- worker failure
- reconnect

## Race Detector

Mandatory:

```bash
go test -race ./...
```

## Failure Tests

Test:

- Redis unavailable
- PostgreSQL unavailable
- worker crash
- network timeout
- malformed event
- duplicate event
- queue unavailable
