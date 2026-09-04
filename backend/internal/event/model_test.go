package event

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestEventSerialization(t *testing.T) {
	workerID := uuid.New()
	completedAt := time.Now().Round(time.Millisecond) // Round for deep equal stability with some json encodings

	e := Event{
		ID:          uuid.New(),
		Type:        TypeOrderCreated,
		Payload:     json.RawMessage(`{"order_id":"abc"}`),
		Priority:    PriorityHigh,
		Status:      StatusCompleted,
		Attempt:     1,
		WorkerID:    &workerID,
		CreatedAt:   time.Now().Round(time.Millisecond),
		UpdatedAt:   time.Now().Round(time.Millisecond),
		CompletedAt: &completedAt,
	}

	b, err := json.Marshal(e)
	if err != nil {
		t.Fatalf("Failed to marshal event: %v", err)
	}

	var decoded Event
	if err := json.Unmarshal(b, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal event: %v", err)
	}

	// UUIDs and times should match exactly
	if !reflect.DeepEqual(e, decoded) {
		t.Errorf("Decoded event does not match original.\nGot: %+v\nWant: %+v", decoded, e)
	}
}

func TestEventOptionalFieldsNil(t *testing.T) {
	e := Event{
		ID:        uuid.New(),
		Type:      TypePaymentSuccess,
		Payload:   json.RawMessage(`{}`),
		Priority:  PriorityNormal,
		Status:    StatusQueued,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	b, err := json.Marshal(e)
	if err != nil {
		t.Fatalf("Failed to marshal event: %v", err)
	}

	var decoded Event
	if err := json.Unmarshal(b, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal event: %v", err)
	}

	if decoded.WorkerID != nil {
		t.Errorf("Expected WorkerID to be nil, got %v", decoded.WorkerID)
	}
	if decoded.CompletedAt != nil {
		t.Errorf("Expected CompletedAt to be nil, got %v", decoded.CompletedAt)
	}
}
