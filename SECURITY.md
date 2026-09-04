# Security

## Input Validation

Validate:

- event type
- payload size
- priority
- request structure

## Message Size

WebSocket messages must have a maximum size.

## Rate Limiting

Rate limit event ingestion.

Prevent clients from overwhelming the API.

## Authentication

Do not trust user identity supplied in request body.

## Secrets

Never commit:

- database passwords
- Redis passwords
- API keys

Use environment variables.

## SQL

Use parameterized queries.

Never concatenate user input into SQL.

## WebSocket

Validate every incoming message.

Disconnect malformed or abusive clients.
