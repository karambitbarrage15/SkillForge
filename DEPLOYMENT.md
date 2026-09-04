# Deployment

## Local

Docker Compose:

- API
- Worker
- Redis
- PostgreSQL
- Dashboard

## Production

Kubernetes.

Components:

- API Deployment
- Worker Deployment
- Redis
- PostgreSQL
- Load Balancer

## API

API should be horizontally scalable.

## Worker

Workers should be horizontally scalable.

## Health Checks

Kubernetes readiness:

```
GET /ready
```

Liveness:

```
GET /health
```

## HPA

Potential scaling signals:

- CPU
- memory
- queue depth

## Graceful Shutdown

Kubernetes sends `SIGTERM`.

Application must:

1. stop accepting work
2. finish safe work
3. close connections
4. shutdown cleanly

## CI

GitHub Actions:

- lint
- test
- race test
- build
- Docker build
