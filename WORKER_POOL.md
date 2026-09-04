# Worker Pool

## Objective

Process events concurrently using a bounded set of workers.

## Bounded Pool

The worker pool is fixed in size.

Worker count is set at startup via configuration:

```
WORKER_COUNT=4
```

Do not spawn unlimited goroutines. Do not create one goroutine per event.

## Architecture

```
Shutdown signal
    │
    ▼
Worker Pool (manages N workers)
    │
    ├── Worker 1
    │   ├── Job goroutine  (blocks on queue.Consume)
    │   └── Heartbeat goroutine
    │
    ├── Worker 2
    │   ├── Job goroutine
    │   └── Heartbeat goroutine
    │
    └── Worker N
        ├── Job goroutine
        └── Heartbeat goroutine
```

## Worker Lifecycle

```
STARTING → READY → BUSY → READY (loop)
READY    → STOPPING → STOPPED   (shutdown)
BUSY     → STOPPING → STOPPED   (finish current job first)
```

## Job Goroutine

Each worker goroutine:

1. Blocks on `queue.Consume(ctx)` — waiting for next event
2. Transitions worker status: `READY → BUSY`
3. Delegates to processor: `Processor.Process(ctx, event)`
4. Handles result (finalize or retry)
5. Transitions worker status: `BUSY → READY`
6. Loops back to step 1

## Heartbeat Goroutine

Each worker has a concurrent heartbeat goroutine that:

1. Writes `worker:{id}:heartbeat` to Redis with a TTL
2. Sleeps for `heartbeat_interval`
3. Repeats until context is cancelled

If the heartbeat goroutine cannot write (Redis failure), it logs the error and continues — the TTL will expire naturally, causing the scheduler to eventually detect the worker as unhealthy.

## Graceful Shutdown

Shutdown is triggered by cancelling the root context (on SIGTERM or SIGINT).

Sequence:

1. Root context cancelled
2. All `queue.Consume()` calls unblock and return
3. Workers in BUSY state finish their current job
4. Heartbeat goroutines stop
5. All workers transition to STOPPED
6. `sync.WaitGroup.Wait()` unblocks
7. Pool reports shutdown complete

Workers must not accept new jobs after the context is cancelled.

## Synchronization

| Resource | Mechanism |
|---|---|
| Worker status field | Mutex per worker |
| Goroutine lifecycle | `context.Context` for cancellation |
| Shutdown ordering | `sync.WaitGroup` |
| Worker registration | Mutex on pool |

## Configuration

| Env Variable | Default | Description |
|---|---|---|
| `WORKER_COUNT` | `4` | Number of concurrent workers |
| `HEARTBEAT_INTERVAL` | `5s` | How often heartbeat is written |
| `HEARTBEAT_TTL` | `15s` | Redis TTL for heartbeat key |
| `JOB_TIMEOUT` | `30s` | Max time allowed per job |

## Constraints

- Worker count must not be 0.
- Worker count must not be unbounded at runtime.
- Every goroutine spawned by the pool must terminate before the process exits.
- The race detector must pass: `go test -race ./...`
