package event

import (
	"testing"
)

func TestIsValidTransition(t *testing.T) {
	tests := []struct {
		name  string
		from  EventStatus
		to    EventStatus
		valid bool
	}{
		// Valid transitions
		{"Received to Queued", StatusReceived, StatusQueued, true},
		{"Queued to Processing", StatusQueued, StatusProcessing, true},
		{"Processing to Completed", StatusProcessing, StatusCompleted, true},
		{"Processing to Retrying", StatusProcessing, StatusRetrying, true},
		{"Retrying to Queued", StatusRetrying, StatusQueued, true},
		{"Processing to Failed", StatusProcessing, StatusFailed, true},
		{"Processing to Queued", StatusProcessing, StatusQueued, true}, // Recovery

		// Invalid transitions
		{"Received to Completed", StatusReceived, StatusCompleted, false},
		{"Queued to Completed", StatusQueued, StatusCompleted, false},
		{"Completed to Queued", StatusCompleted, StatusQueued, false},
		{"Failed to Queued", StatusFailed, StatusQueued, false},
		{"Failed to Processing", StatusFailed, StatusProcessing, false},
		{"Retrying to Completed", StatusRetrying, StatusCompleted, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := IsValidTransition(tc.from, tc.to); got != tc.valid {
				t.Errorf("IsValidTransition(%q, %q) = %v; want %v", tc.from, tc.to, got, tc.valid)
			}
		})
	}
}

func TestEvent_TransitionTo(t *testing.T) {
	// Valid transition
	e := &Event{Status: StatusQueued}
	err := e.TransitionTo(StatusProcessing)
	if err != nil {
		t.Errorf("Expected nil error for valid transition, got %v", err)
	}
	if e.Status != StatusProcessing {
		t.Errorf("Expected status to be updated to %v, got %v", StatusProcessing, e.Status)
	}
	if e.CompletedAt != nil {
		t.Errorf("Expected CompletedAt to be nil, got %v", e.CompletedAt)
	}

	// Invalid transition
	e2 := &Event{Status: StatusCompleted}
	err = e2.TransitionTo(StatusQueued)
	if err != ErrInvalidStateTransition {
		t.Errorf("Expected ErrInvalidStateTransition, got %v", err)
	}
	if e2.Status != StatusCompleted {
		t.Errorf("Expected status to remain unchanged, got %v", e2.Status)
	}

	// Completion sets CompletedAt
	e3 := &Event{Status: StatusProcessing}
	err = e3.TransitionTo(StatusCompleted)
	if err != nil {
		t.Errorf("Expected nil error for valid transition, got %v", err)
	}
	if e3.Status != StatusCompleted {
		t.Errorf("Expected status to be updated to %v, got %v", StatusCompleted, e3.Status)
	}
	if e3.CompletedAt == nil {
		t.Error("Expected CompletedAt to be set")
	}
}
