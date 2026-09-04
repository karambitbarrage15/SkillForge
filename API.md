# API

## Health

```
GET /health
```

Response:

```json
{
  "status": "ok"
}
```

## Readiness

```
GET /ready
```

## Create Event

```
POST /api/v1/events
```

Request:

```json
{
  "type": "ORDER_CREATED",
  "priority": 5,
  "payload": {
    "order_id": "123"
  }
}
```

Response:

```json
{
  "id": "uuid",
  "status": "QUEUED"
}
```

## Get Event

```
GET /api/v1/events/:id
```

## List Events

```
GET /api/v1/events
```

Query parameters:

| Parameter | Description          |
|-----------|----------------------|
| status    | Filter by status     |
| type      | Filter by event type |
| limit     | Page size            |
| offset    | Page offset          |

## WebSocket

```
GET /ws
```
