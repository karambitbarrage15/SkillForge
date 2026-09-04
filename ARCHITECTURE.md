# Architecture

```
                    ┌───────────────┐
                    │ React         │
                    │ Dashboard     │
                    └───────┬───────┘
                            │
                         WebSocket
                            │
                    ┌───────▼───────┐
                    │ Go API        │
                    └───────┬───────┘
                            │
                    ┌───────▼───────┐
                    │ Event Router  │
                    └───────┬───────┘
                            │
                    ┌───────▼───────┐
                    │ Redis Queue   │
                    └───────┬───────┘
                            │
              ┌─────────────┼─────────────┐
              │             │             │
        ┌─────▼─────┐ ┌─────▼─────┐ ┌─────▼─────┐
        │ Go Worker │ │ Go Worker │ │ Go Worker │
        └─────┬─────┘ └─────┬─────┘ └─────┬─────┘
              │             │             │
              └─────────────┼─────────────┘
                            │
                    ┌───────▼───────┐
                    │ PostgreSQL    │
                    └───────────────┘
```

## Services

### API

Responsibilities:

- HTTP
- authentication
- event ingestion
- WebSocket connections

### Worker

Responsibilities:

- queue consumption
- event processing
- heartbeat
- retry

### Scheduler

Responsibilities:

- assignment
- recovery
- worker monitoring

### Dashboard

Responsibilities:

- visualization
- real-time monitoring

## Principle

Separate ingestion from processing.
