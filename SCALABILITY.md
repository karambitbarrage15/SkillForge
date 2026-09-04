# Scalability

## Target

The project should be capable of being extended toward:

- 1,000 events/sec
- 10,000 events/sec
- 100,000 events/sec

## Important Metrics

**Throughput:**

- events/sec

**Latency:**

- p50
- p95
- p99

**Queue:**

- queue depth
- oldest event age

**Workers:**

- active
- busy
- offline

## Horizontal Scaling

- API servers are stateless.
- Workers are horizontally scalable.
- Redis coordinates shared state.
- PostgreSQL stores durable state.

## Bottleneck Analysis

Do not assume adding workers always increases throughput.

At some point, Redis, PostgreSQL, or CPU may become the bottleneck.

Load tests should identify the bottleneck.

## Engineering Goal

Optimize based on measurements rather than assumptions.
