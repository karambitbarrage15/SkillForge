package event

import (
	"errors"
	"time"
)

var (
	ErrInvalidStateTransition = errors.New("invalid state transition")
)

var validTransitions = map[EventStatus]map[EventStatus]bool{
	StatusReceived: {
		StatusQueued: true,
	},
	StatusQueued: {
		StatusProcessing: true,
	},
	StatusProcessing: {
		StatusCompleted: true,
		StatusRetrying:  true,
		StatusFailed:    true,
		StatusQueued:    true, // Worker recovery
	},
	StatusRetrying: {
		StatusQueued: true,
	},
	StatusCompleted: {}, // Terminal
	StatusFailed:    {}, // Terminal
}

// IsValidTransition returns true if the transition from old to new is allowed.
func IsValidTransition(from, to EventStatus) bool {
	allowed, ok := validTransitions[from]
	if !ok {
		return false
	}
	return allowed[to]
}

// TransitionTo applies the state transition to the event if it's valid.
// It also manages timestamps where appropriate (e.g., CompletedAt).
func (e *Event) TransitionTo(to EventStatus) error {
	if !IsValidTransition(e.Status, to) {
		return ErrInvalidStateTransition
	}

	e.Status = to
	e.UpdatedAt = time.Now()

	if to == StatusCompleted {
		now := time.Now()
		e.CompletedAt = &now
	}

	return nil
}
