package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"

	"streamforge/internal/event"
	"streamforge/internal/queue"
	"streamforge/internal/store"
)

// EventHandler handles HTTP requests for events.
type EventHandler struct {
	repo      store.EventRepository
	publisher queue.Publisher
}

// NewEventHandler creates a new stateless EventHandler.
func NewEventHandler(repo store.EventRepository, pub queue.Publisher) *EventHandler {
	return &EventHandler{
		repo:      repo,
		publisher: pub,
	}
}

// HandleCreateEvent processes POST /api/v1/events.
// Flow: Decode -> Validate -> Create -> Save -> Publish -> Return 202
func (h *EventHandler) HandleCreateEvent(w http.ResponseWriter, r *http.Request) {
	var req CreateEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"malformed JSON request"}`, http.StatusBadRequest)
		return
	}

	e := &event.Event{
		ID:        uuid.New(),
		Type:      event.EventType(req.Type),
		Payload:   req.Payload,
		Priority:  req.Priority,
		Status:    event.StatusReceived,
		Attempt:   0,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := e.Validate(); err != nil {
		slog.Warn("Event validation failed", "err", err, "event_id", e.ID)
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	// Persist the event BEFORE publishing it.
	if err := h.repo.Create(r.Context(), e); err != nil {
		slog.Error("Failed to persist event", "err", err, "event_id", e.ID)
		http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
		return
	}

	// Publish to the queue
	if err := h.publisher.Publish(r.Context(), e); err != nil {
		slog.Error("Failed to publish event to queue", "err", err, "event_id", e.ID)
		http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
		return
	}

	resp := CreateEventResponse{
		ID:     e.ID,
		Status: string(e.Status),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		slog.Error("Failed to encode response", "err", err, "event_id", e.ID)
	}
}
