# StreamForge

StreamForge is a distributed real-time event processing platform built primarily in Go.

It accepts high-volume events, places them into a distributed queue, schedules them across workers, processes them concurrently, persists their results, and streams processing state to a real-time dashboard.

The project is designed to explore practical backend engineering problems:

- Go concurrency
- Goroutines and channels
- Worker pools
- Event-driven architecture
- WebSockets
- Redis
- PostgreSQL
- State machines
- Retry mechanisms
- Idempotent processing
- Worker health monitoring
- Failure recovery
- Backpressure
- Horizontal scaling
- Kubernetes
- Load testing
- Observability

## Example Events

StreamForge can process events such as:

```
ORDER_CREATED
PAYMENT_SUCCESS
PAYMENT_FAILED
INVENTORY_UPDATED
ORDER_CANCELLED
USER_REGISTERED
```

The system is generic enough to support additional event types.

## Architecture

```
Client
    |
    v
Go API
    |
    v
Event Router
    |
    v
Redis Queue
    |
    v
Go Worker Pool
    |
    +---- Worker 1
    +---- Worker 2
    +---- Worker 3
    |
    v
PostgreSQL
    |
    v
WebSocket Hub
    |
    v
React Dashboard
```

## Main Engineering Goal

The system should process events concurrently while maintaining correctness.

The system must not sacrifice correctness simply to increase throughput.

## Important Principle

The server is authoritative.

Clients cannot decide:

- event status
- processing result
- worker status
- retry state
- execution state

## Local Development

Requirements:

- Go
- Docker
- Docker Compose
- Node.js

Start dependencies:

```bash
docker compose up -d
```

Run API:

```bash
cd backend
go run ./cmd/api
```

Run worker:

```bash
cd backend
go run ./cmd/worker
```

Run dashboard:

```bash
cd dashboard
npm install
npm run dev
```

## Testing

Run:

```bash
go test ./...
```

Race detector:

```bash
go test -race ./...
```

## Project Objective

This is not intended to be a CRUD application.

The primary objective is to demonstrate reasoning about concurrency, distributed systems, failure handling and scalable backend architecture.
