package api

import (
	"encoding/json"

	"github.com/google/uuid"
)

// CreateEventRequest is the incoming payload for POST /api/v1/events
type CreateEventRequest struct {
	Type     string          `json:"type"`
	Priority int             `json:"priority"`
	Payload  json.RawMessage `json:"payload"`
}

// CreateEventResponse is the success response
type CreateEventResponse struct {
	ID     uuid.UUID `json:"id"`
	Status string    `json:"status"`
}
