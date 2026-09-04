# System Design

## Goal

Design StreamForge to scale horizontally.

## Initial

```
One API
One Worker
One Redis
One PostgreSQL
```

## Scaled

```
                    Load Balancer
                         │
             ┌───────────┼───────────┐
             ▼           ▼           ▼
          API-1       API-2       API-3
             │           │           │
             └───────────┼───────────┘
                         │
                       Redis
                         │
             ┌───────────┼───────────┐
             ▼           ▼           ▼
         Worker-1    Worker-2    Worker-3
             │           │           │
             └───────────┼───────────┘
                         │
                    PostgreSQL
```

## Bottlenecks

Potential bottlenecks:

- PostgreSQL writes
- Redis throughput
- WebSocket connections
- worker CPU
- slow processors

## Scaling Workers

Workers can be horizontally scaled.

Scale based on:

- queue depth
- CPU
- processing latency

## Backpressure

If incoming event rate exceeds processing capacity, queue depth grows.

The system should expose this through metrics.

## WebSockets

WebSocket connections are long-lived.

Use:

- bounded outbound buffers
- heartbeat
- connection limits
- horizontal scaling

## Database

Use connection pooling.

Avoid querying PostgreSQL on every realtime update.

Use asynchronous persistence where correctness permits.

## Future Improvements

- partition queues
- event sharding
- Kafka
- database partitioning
- regional workers
- autoscaling based on queue depth
