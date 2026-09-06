package api

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	gorillaws "github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"

	"streamforge/internal/event"
	"streamforge/internal/metrics"
	"streamforge/internal/queue"
	"streamforge/internal/store"
	"streamforge/internal/websocket"
)

var upgrader = gorillaws.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for now
	},
}

// EventHandler handles HTTP requests for events.
type EventHandler struct {
	repo      store.EventRepository
	publisher queue.Publisher
	hub       *websocket.Hub
	rdb       *redis.Client
}

// NewEventHandler creates a new stateless EventHandler.
func NewEventHandler(repo store.EventRepository, pub queue.Publisher, hub *websocket.Hub, rdb *redis.Client) *EventHandler {
	return &EventHandler{
		repo:      repo,
		publisher: pub,
		hub:       hub,
		rdb:       rdb,
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

	// Phase 8: Transition from RECEIVED to QUEUED
	if err := h.repo.UpdateStatus(r.Context(), e.ID, event.StatusReceived, event.StatusQueued); err != nil {
		slog.Error("Failed to transition event to QUEUED", "err", err, "event_id", e.ID)
		// Non-fatal for the client, the worker will eventually pick it up
	}

	// Non-critical: Broadcast EVENT_RECEIVED
	go func(id uuid.UUID) {
		if h.rdb == nil {
			return
		}
		msg := websocket.EventReceivedMsg{
			Type:    websocket.TypeEventReceived,
			EventID: id,
		}
		b, _ := json.Marshal(msg)
		// Publish to Redis PubSub for all hubs
		h.rdb.Publish(context.Background(), websocket.RedisPubSubChannel, string(b))
	}(e.ID)

	// Phase 15: Observability
	// Increment successfully accepted/validated event submissions.
	metrics.EventsReceived.WithLabelValues(string(e.Type)).Inc()

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

// HandleGetEventByID processes GET /api/v1/events/{id}
func (h *EventHandler) HandleGetEventByID(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, `{"error":"invalid event ID format"}`, http.StatusBadRequest)
		return
	}

	e, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		if err == store.ErrEventNotFound {
			http.Error(w, `{"error":"event not found"}`, http.StatusNotFound)
			return
		}
		slog.Error("Failed to get event", "err", err, "event_id", id)
		http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(e); err != nil {
		slog.Error("Failed to encode response", "err", err, "event_id", id)
	}
}

// HandleListEvents processes GET /api/v1/events
func (h *EventHandler) HandleListEvents(w http.ResponseWriter, r *http.Request) {
	limit := 50
	offset := 0

	if l := r.URL.Query().Get("limit"); l != "" {
		parsed, err := strconv.Atoi(l)
		if err != nil || parsed < 1 || parsed > 100 {
			http.Error(w, `{"error":"invalid limit parameter"}`, http.StatusBadRequest)
			return
		}
		limit = parsed
	}

	if o := r.URL.Query().Get("offset"); o != "" {
		parsed, err := strconv.Atoi(o)
		if err != nil || parsed < 0 {
			http.Error(w, `{"error":"invalid offset parameter"}`, http.StatusBadRequest)
			return
		}
		offset = parsed
	}

	events, err := h.repo.List(r.Context(), limit, offset)
	if err != nil {
		slog.Error("Failed to list events", "err", err)
		http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
		return
	}

	// Ensure we return an empty array instead of null if no events
	if events == nil {
		events = make([]*event.Event, 0)
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(events); err != nil {
		slog.Error("Failed to encode response", "err", err)
	}
}

// ServeWs handles websocket requests from the peer.
func (h *EventHandler) ServeWs(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("Failed to upgrade websocket", "err", err)
		return
	}
	client := websocket.NewClient(h.hub, conn)
	client.HubRegister() // Wrapper or just access channel directly

	go client.WritePump()
	go client.ReadPump()
}
