# Redis Design

Redis is used for high-speed coordination.

## Uses

- event queue
- temporary worker state
- heartbeats
- distributed coordination
- Pub/Sub where required

## Durable Data

PostgreSQL remains the durable source of truth.

Redis must not be the only location containing important business data.

## Queue

Use Redis Streams or another appropriate Redis primitive.

The implementation must document why the selected primitive was chosen.

## Worker Heartbeats

Key:

```
worker:{worker_id}:heartbeat
```

TTL should be used.

If key expires: worker is considered unhealthy.

## Distributed State

Do not assume Redis operations are automatically atomic across multiple operations.

Use appropriate Redis transactions/scripts where required.
