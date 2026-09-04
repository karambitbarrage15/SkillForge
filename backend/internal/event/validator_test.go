package event

import (
	"encoding/json"
	"errors"
	"testing"
)

func TestEventValidation(t *testing.T) {
	tests := []struct {
		name    string
		event   Event
		wantErr error
	}{
		{
			name: "Valid event",
			event: Event{
				Type:     TypeOrderCreated,
				Priority: PriorityNormal,
				Payload:  json.RawMessage(`{"id": 1}`),
			},
			wantErr: nil,
		},
		{
			name:    "Zero-value event",
			event:   Event{},
			wantErr: ErrInvalidType,
		},
		{
			name: "Unsupported EventType",
			event: Event{
				Type:     EventType("RANDOM_TYPE"),
				Priority: PriorityNormal,
				Payload:  json.RawMessage(`{}`),
			},
			wantErr: ErrInvalidType,
		},
		{
			name: "Invalid priority (negative)",
			event: Event{
				Type:     TypePaymentFailed,
				Priority: -1,
				Payload:  json.RawMessage(`{}`),
			},
			wantErr: ErrInvalidPriority,
		},
		{
			name: "Invalid priority (unknown)",
			event: Event{
				Type:     TypeInventoryUpdated,
				Priority: 7,
				Payload:  json.RawMessage(`{}`),
			},
			wantErr: ErrInvalidPriority,
		},
		{
			name: "Valid priorities - Low",
			event: Event{
				Type:     TypeOrderCancelled,
				Priority: PriorityLow,
				Payload:  json.RawMessage(`{}`),
			},
			wantErr: nil,
		},
		{
			name: "Valid priorities - High",
			event: Event{
				Type:     TypeUserRegistered,
				Priority: PriorityHigh,
				Payload:  json.RawMessage(`{}`),
			},
			wantErr: nil,
		},
		{
			name: "Nil payload",
			event: Event{
				Type:     TypeOrderCreated,
				Priority: PriorityNormal,
				Payload:  nil,
			},
			wantErr: ErrNilPayload,
		},
		{
			name: "Empty payload",
			event: Event{
				Type:     TypeOrderCreated,
				Priority: PriorityNormal,
				Payload:  json.RawMessage(""),
			},
			wantErr: ErrNilPayload,
		},
		{
			name: "Malformed JSON payload",
			event: Event{
				Type:     TypeOrderCreated,
				Priority: PriorityNormal,
				Payload:  json.RawMessage(`{"id": 1`), // Missing closing brace
			},
			wantErr: ErrInvalidPayload,
		},
		{
			name: "Array payload (valid JSON)",
			event: Event{
				Type:     TypeOrderCreated,
				Priority: PriorityNormal,
				Payload:  json.RawMessage(`[1, 2, 3]`),
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.event.Validate()
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
