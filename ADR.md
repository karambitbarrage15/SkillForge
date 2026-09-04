# Architecture Decisions

## ADR-001: Go

Go is the primary backend language.

Reason:

- concurrency
- networking
- low overhead
- simple deployment
- strong standard library

## ADR-002: Redis

Redis is used for high-speed queueing and coordination.

PostgreSQL remains the durable database.

## ADR-003: WebSockets

WebSockets are used for real-time dashboard updates.

Polling would create unnecessary requests and latency.

## ADR-004: At-Least-Once Processing

The system prefers reliability over pretending to provide exactly-once execution.

Idempotent finalization is used to maintain correctness.

## ADR-005: Server Authority

The backend owns event state.

Clients only observe state.

## ADR-006: Bounded Concurrency

The worker pool is bounded.

The system must apply backpressure rather than creating unlimited goroutines.
