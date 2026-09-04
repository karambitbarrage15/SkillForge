package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"

	"streamforge/internal/event"
)

// MockRepo tracks calls to assert ordering and passes/fails based on config
type MockRepo struct {
	FailCreate bool
	Created    *event.Event
	CallOrder  *[]string
}

func (m *MockRepo) Create(ctx context.Context, e *event.Event) error {
	if m.CallOrder != nil {
		*m.CallOrder = append(*m.CallOrder, "repo")
	}
	if m.FailCreate {
		return errors.New("db error")
	}
	m.Created = e
	return nil
}

func (m *MockRepo) GetByID(ctx context.Context, id uuid.UUID) (*event.Event, error) { return nil, nil }
func (m *MockRepo) UpdateStatus(ctx context.Context, id uuid.UUID, old, new event.EventStatus) error {
	return nil
}

// MockPublisher tracks calls and can simulate failures
type MockPublisher struct {
	FailPublish bool
	Published   *event.Event
	CallOrder   *[]string
}

func (m *MockPublisher) Publish(ctx context.Context, e *event.Event) error {
	if m.CallOrder != nil {
		*m.CallOrder = append(*m.CallOrder, "publisher")
	}
	if m.FailPublish {
		return errors.New("queue error")
	}
	m.Published = e
	return nil
}

func TestHandleCreateEvent(t *testing.T) {
	tests := []struct {
		name           string
		body           string
		repoFail       bool
		pubFail        bool
		expectedStatus int
		verifyFn       func(*testing.T, *MockRepo, *MockPublisher, *httptest.ResponseRecorder)
	}{
		{
			name:           "Malformed JSON",
			body:           `{"type": "ORDER_CREATED"`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid event type",
			body:           `{"type": "UNKNOWN", "priority": 5, "payload": {}}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid priority",
			body:           `{"type": "ORDER_CREATED", "priority": 2, "payload": {}}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Missing payload",
			body:           `{"type": "ORDER_CREATED", "priority": 5}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Repository failure",
			body:           `{"type": "ORDER_CREATED", "priority": 5, "payload": {}}`,
			repoFail:       true,
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "Publisher failure",
			body:           `{"type": "ORDER_CREATED", "priority": 5, "payload": {}}`,
			pubFail:        true,
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "Success",
			body:           `{"type": "ORDER_CREATED", "priority": 5, "payload": {"foo":"bar"}}`,
			expectedStatus: http.StatusAccepted,
			verifyFn: func(t *testing.T, repo *MockRepo, pub *MockPublisher, rr *httptest.ResponseRecorder) {
				// Verify 1: Returned ID is valid and status is RECEIVED
				var resp CreateEventResponse
				if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}
				if resp.ID == uuid.Nil {
					t.Error("Returned ID is nil UUID")
				}
				if resp.Status != "RECEIVED" {
					t.Errorf("Expected status RECEIVED, got %s", resp.Status)
				}

				// Verify 2: Repository received the correct Event
				if repo.Created == nil {
					t.Fatal("Repository did not receive an event")
				}
				if repo.Created.ID != resp.ID {
					t.Errorf("Repo event ID %s != response ID %s", repo.Created.ID, resp.ID)
				}
				if repo.Created.Status != event.StatusReceived {
					t.Errorf("Repo event status != RECEIVED")
				}

				// Verify 3: Publisher received the same Event
				if pub.Published == nil {
					t.Fatal("Publisher did not receive an event")
				}
				if pub.Published != repo.Created {
					t.Errorf("Publisher received a different event instance than the repository")
				}

				// Verify 4: Ordering (Persistence -> Publish)
				if len(*repo.CallOrder) != 2 || (*repo.CallOrder)[0] != "repo" || (*repo.CallOrder)[1] != "publisher" {
					t.Errorf("Expected order [repo, publisher], got %v", *repo.CallOrder)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var callOrder []string
			repo := &MockRepo{FailCreate: tt.repoFail, CallOrder: &callOrder}
			pub := &MockPublisher{FailPublish: tt.pubFail, CallOrder: &callOrder}

			handler := NewEventHandler(repo, pub)

			req := httptest.NewRequest(http.MethodPost, "/api/v1/events", bytes.NewBufferString(tt.body))
			rr := httptest.NewRecorder()

			handler.HandleCreateEvent(rr, req)

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("Handler returned wrong status code: got %v want %v", status, tt.expectedStatus)
			}

			if tt.verifyFn != nil {
				tt.verifyFn(t, repo, pub, rr)
			}
		})
	}
}
