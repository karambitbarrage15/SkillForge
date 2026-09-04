# Event State Machine

## States

```
RECEIVED
QUEUED
PROCESSING
COMPLETED
RETRYING
FAILED
```

## Valid Transitions

```
RECEIVED   → QUEUED
QUEUED     → PROCESSING
PROCESSING → COMPLETED
PROCESSING → RETRYING
RETRYING   → QUEUED
PROCESSING → FAILED
```

## Invalid Transitions

```
COMPLETED → PROCESSING
FAILED    → PROCESSING
COMPLETED → QUEUED
```

These must return errors.

## Rules

State transitions must be centralized.

Do not allow arbitrary components to mutate event status.

The state machine must be unit tested.
