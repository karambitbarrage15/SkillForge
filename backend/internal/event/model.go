package event

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Event is the core domain model representing a unit of work.
type Event struct {
	ID          uuid.UUID
	Type        EventType
	Payload     json.RawMessage
	Priority    int
	Status      EventStatus
	Attempt     int
	WorkerID    *uuid.UUID
	CreatedAt   time.Time
	UpdatedAt   time.Time
	CompletedAt *time.Time
}
