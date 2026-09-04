# Load Testing

The simulator generates events.

## Scenarios

### Scenario 1

100 events/sec

### Scenario 2

1,000 events/sec

### Scenario 3

5,000 events/sec

### Scenario 4

10,000 events/sec

## Metrics

Measure:

- throughput
- p50
- p95
- p99
- queue depth
- CPU
- memory
- failed events
- retry count

## Worker Scaling Experiment

Run with:

- 4 workers
- 8 workers
- 16 workers
- 32 workers

Compare throughput.

## Deliverable

Document actual results.

Never fabricate benchmark numbers.

Example:

| Workers | Throughput | P95      |
|---------|------------|----------|
| 4       | measured   | measured |
| 8       | measured   | measured |
| 16      | measured   | measured |
