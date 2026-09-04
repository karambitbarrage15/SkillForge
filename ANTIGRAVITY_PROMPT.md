# StreamForge — Antigravity Implementation Instructions

## Step 0 — Read Before You Code

Read the entire repository and all specification files before writing any code.

Pay particular attention to:

- AGENTS.md
- IMPLEMENTATION.md
- ARCHITECTURE.md
- ANTIGRAVITY_PROMPT.md
- API.md
- TESTING.md

Do not implement anything yet.

First:

1. Understand the complete system.
2. Explain the architecture you intend to build.
3. Identify the major components and their responsibilities.
4. Identify the concurrency model.
5. Identify the main data structures and state transitions.
6. Identify the hardest engineering problems.
7. Give the implementation order you recommend.
8. Point out any ambiguity or contradiction in the specification.

Wait for approval before making any code changes.

---

You are implementing StreamForge.

This is a serious backend engineering project designed to demonstrate Go concurrency, distributed systems, WebSockets, queue processing, failure recovery and scalability.

Read ALL documentation in the root before implementing.

Required reading order:

1. README.md
2. PRD.md
3. IMPLEMENTATION.md
4. ARCHITECTURE.md
5. SYSTEM_DESIGN.md
6. GO_CONCURRENCY.md
7. EVENT_LIFECYCLE.md
8. STATE_MACHINE.md
9. WORKER_POOL.md
10. SCHEDULER.md
11. WEBSOCKET_PROTOCOL.md
12. REDIS.md
13. DATABASE.md
14. FAILURE_HANDLING.md
15. IDEMPOTENCY.md
16. SCALABILITY.md
17. LOAD_TESTING.md
18. OBSERVABILITY.md
19. SECURITY.md
20. TESTING.md
21. DEPLOYMENT.md
22. API.md
23. ADR.md
24. TODO.md

---

# CRITICAL RULE

Do not build everything at once.

Implement the project phase-by-phase according to IMPLEMENTATION.md.

After every phase:

1. compile
2. run tests
3. run go vet
4. run race detector
5. fix errors
6. update TODO.md

Do not move to the next phase while the current phase is broken.

---

# Go Requirements

The backend must be written in idiomatic Go.

Use:

- goroutines
- channels
- context.Context
- interfaces
- mutexes only where appropriate
- sync.WaitGroup
- atomic operations where appropriate

Avoid unnecessary abstractions.

Do not create interfaces for every struct just to appear "enterprise."

Prefer simple, understandable Go.

---

# Concurrency

Concurrency correctness is more important than raw performance.

Never introduce:

- data races
- unbounded goroutines
- unbounded channels
- unsafe shared state

The project must pass:

```bash
go test -race ./...
```

---

# Worker Pool

The worker pool must be bounded.

Do not create one goroutine for every event.

Workers should consume jobs through channels or the selected queue abstraction.

Worker count should be configurable.

---

# State Machine

Implement event states explicitly.

Do not replace the state machine with random boolean fields.

Invalid transitions must return errors.

State transition logic must be tested.

---

# Redis

Use Redis for queueing and coordination.

Do not use Redis as the permanent source of truth for business data.

PostgreSQL is the durable database.

---

# WebSockets

Implement:

- read loop
- write loop
- heartbeat
- bounded outgoing buffer
- graceful shutdown

Never allow a slow WebSocket client to block the entire system.

---

# Failure Handling

Assume workers can crash.

Implement heartbeat detection.

If a worker dies while processing an event:

1. detect worker failure
2. determine unfinished event
3. safely return event to queue
4. allow another worker to process it
5. prevent duplicate finalization

---

# Idempotency

Assume at-least-once delivery.

Do not claim exactly-once processing unless it can actually be demonstrated.

Use unique event IDs and database constraints.

Finalization must be safe if repeated.

---

# API

HTTP handlers should remain thin.

Preferred structure:

```
Handler
  ↓
Service
  ↓
Repository
```

Do not place business logic directly in handlers.

---

# Database

Use migrations.

Use transactions for atomic state changes.

Use indexes based on actual query patterns.

Never construct SQL using string concatenation.

---

# Error Handling

Errors must be explicit.

Do not swallow errors.

Do not use panic for ordinary runtime failures.

Return meaningful errors to callers.

Log internal failures with useful context.

---

# Logging

Use structured logging.

Include useful identifiers:

- event_id
- worker_id
- event_type
- attempt

Do not log secrets.

---

# Configuration

Use environment variables.

Provide `.env.example`.

Never commit real credentials.

---

# Testing

Every important component must have tests.

Minimum:

- state machine tests
- worker pool tests
- scheduler tests
- retry tests
- idempotency tests
- WebSocket tests
- repository tests
- integration tests

Also run:

```bash
go vet ./...
go test ./...
go test -race ./...
```

---

# Load Testing

Do not fabricate performance numbers.

Only document measurements actually obtained from the system.

The simulator should allow different event rates.

Measure:

- throughput
- p50
- p95
- p99
- CPU
- memory
- queue depth

---

# Frontend

The dashboard is an observability interface.

It should not contain authoritative backend logic.

The frontend receives server state through WebSockets.

---

# Docker

Use multi-stage builds.

Keep runtime images small.

Do not put secrets in Dockerfiles.

---

# Kubernetes

Create production-oriented manifests.

Include:

- Deployment
- Service
- ConfigMap
- Secret example
- HPA
- health checks

---

# Code Quality

Do not leave:

- TODO hacks
- fake implementations
- placeholder functions
- unused imports
- dead code
- hardcoded secrets
- fake benchmark results

Do not mark TODO.md items complete unless the implementation actually exists.

---

# Documentation

If an implementation differs from the design: update the relevant documentation.

Architecture decisions should be recorded in ADR.md.

---

# Final Requirement

At completion, the repository must be explainable by a developer who understands the code.

Do not optimize for number of files.

Optimize for:

- correctness
- clarity
- concurrency safety
- failure handling
- testability
- measurable performance
