# Scheduler

## Objective

Assign queued events to available workers.

## Priority

Events are ordered by:

1. `priority DESC`
2. `created_at ASC`

Example:

A HIGH priority event created later should generally be processed before a LOW priority event created earlier.

## Worker Selection

Initial implementation: select available worker.

Future implementation may consider:

- worker capabilities
- current load
- locality
- historical processing time

## Scheduling Loop

1. inspect queue
2. find available worker
3. assign event
4. mark event `PROCESSING`
5. record worker ownership
6. continue

## Important

Do not lose an event between:

```
QUEUE → PROCESSING
```

Use atomic state updates or transactional coordination where required.
