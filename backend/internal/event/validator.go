package event

import (
	"encoding/json"
	"errors"
	"fmt"
)

var (
	ErrInvalidType     = errors.New("invalid event type")
	ErrInvalidPriority = errors.New("invalid event priority")
	ErrInvalidPayload  = errors.New("invalid event payload")
	ErrNilPayload      = errors.New("event payload cannot be nil or empty")
)

// Validate checks whether an Event has acceptable field values for ingestion.
// This is typically called on newly created events.
func (e *Event) Validate() error {
	switch e.Type {
	case TypeOrderCreated, TypePaymentSuccess, TypePaymentFailed,
		TypeInventoryUpdated, TypeOrderCancelled, TypeUserRegistered:
		// valid
	default:
		return fmt.Errorf("%w: %q", ErrInvalidType, e.Type)
	}

	switch e.Priority {
	case PriorityLow, PriorityNormal, PriorityHigh:
		// valid
	default:
		return fmt.Errorf("%w: %d", ErrInvalidPriority, e.Priority)
	}

	if len(e.Payload) == 0 {
		return ErrNilPayload
	}

	if !json.Valid(e.Payload) {
		return ErrInvalidPayload
	}

	return nil
}
