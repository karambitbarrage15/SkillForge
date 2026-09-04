# AGENTS.md

> **IMPORTANT**: Read `ANTIGRAVITY_PROMPT.md` first and follow it exactly. Start with Phase 1 only. Do not implement later phases until Phase 1 is tested and working.

---

## Project Overview

This is a production-oriented Go backend project designed to demonstrate
strong backend engineering, concurrency, system design, and scalability
fundamentals.

The goal is not to build a simple CRUD application. The implementation
should demonstrate careful reasoning about concurrency, state management,
failure handling, performance, and clean architecture.

## Engineering Principles

- Write idiomatic Go.
- Prefer simple designs with explicit ownership.
- Avoid unnecessary abstractions.
- Keep business logic independent from HTTP handlers.
- Use goroutines only when concurrency provides a clear benefit.
- Protect shared mutable state properly.
- Never introduce data races.
- Handle errors explicitly.
- Avoid global mutable state.
- Prefer context.Context for request-scoped cancellation and deadlines.
- Make components testable.
- Write benchmarks for performance-sensitive components.

## Concurrency

Concurrency is a core part of this project.

Before introducing a goroutine, determine:

1. Who owns the data?
2. Who can mutate it?
3. What synchronization is required?
4. What happens if the goroutine fails?
5. How is cancellation handled?
6. How does the goroutine terminate?

Use channels, mutexes, atomic operations, or other primitives only when
appropriate.

Run:

    go test -race ./...

before considering concurrency-related work complete.

## Architecture

Follow the repository's architecture documented in:

- ARCHITECTURE.md
- DESIGN.md
- IMPLEMENTATION.md

Do not move business logic into transport handlers.

Keep interfaces small and define them near the code that consumes them.

## Testing

Every meaningful feature should have tests.

Minimum expectations:

- Unit tests
- Integration tests where applicable
- Concurrency tests where applicable
- Race detector
- Benchmarks for critical paths

Commands:

    go test ./...

    go test -race ./...

    go test -cover ./...

    go vet ./...

## Performance

Do not optimize based on assumptions.

For performance-sensitive code:

1. Establish a baseline.
2. Add a benchmark.
3. Implement the optimization.
4. Benchmark again.
5. Document the trade-off.

Use pprof when investigating CPU or memory behaviour.

## Code Quality

Before completing a task:

    gofmt -w .

    go vet ./...

    go test ./...

    go test -race ./...

Do not leave TODOs for core functionality.

Do not silently swallow errors.

## Documentation

Any architectural decision that materially affects the system should be
documented.

Update the relevant documentation when changing:

- concurrency model
- data flow
- APIs
- storage
- failure handling
- performance characteristics
- system architecture

## Git

Keep commits focused and descriptive.

Do not commit:

- secrets
- `.env` files
- generated binaries
- temporary files
- local IDE configuration

## Definition of Done

A feature is complete only when:

- It works correctly.
- It has appropriate tests.
- It handles failure cases.
- It is safe under concurrency.
- It is documented where necessary.
- `go test ./...` passes.
- `go test -race ./...` passes.
- `go vet ./...` passes.
- The implementation remains understandable to another engineer.
