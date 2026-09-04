# Observability

## Logs

Use structured JSON logs.

Every event-related log should include:

- event_id
- worker_id
- event_type
- attempt

## Metrics

Expose Prometheus metrics.

Required:

```
events_received_total
events_completed_total
events_failed_total
events_retried_total
event_processing_duration_seconds
queue_depth
active_workers
worker_failures_total
websocket_connections
```

## Health

```
GET /health
```

Indicates process is alive.

```
GET /ready
```

Indicates dependencies are available.

## Dashboard

Grafana can visualize:

- throughput
- latency
- queue depth
- worker utilization
- failures
