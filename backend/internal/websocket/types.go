package websocket

import "github.com/google/uuid"

type MessageType string

const (
	TypeEventReceived   MessageType = "EVENT_RECEIVED"
	TypeEventProcessing MessageType = "EVENT_PROCESSING"
	TypeEventCompleted  MessageType = "EVENT_COMPLETED"
	TypeEventFailed     MessageType = "EVENT_FAILED"
	TypeWorkerOnline    MessageType = "WORKER_ONLINE"
	TypeWorkerOffline   MessageType = "WORKER_OFFLINE"
	TypePing            MessageType = "PING"
)

type BaseMessage struct {
	Type MessageType `json:"type"`
}

type EventReceivedMsg struct {
	Type    MessageType `json:"type"`
	EventID uuid.UUID   `json:"event_id"`
}

type EventProcessingMsg struct {
	Type     MessageType `json:"type"`
	EventID  uuid.UUID   `json:"event_id"`
	WorkerID string      `json:"worker_id"`
}

type EventCompletedMsg struct {
	Type             MessageType `json:"type"`
	EventID          uuid.UUID   `json:"event_id"`
	ProcessingTimeMs int64       `json:"processing_time_ms"`
}

type EventFailedMsg struct {
	Type    MessageType `json:"type"`
	EventID uuid.UUID   `json:"event_id"`
	Attempt int         `json:"attempt"`
}

type WorkerOnlineMsg struct {
	Type     MessageType `json:"type"`
	WorkerID string      `json:"worker_id"`
}

type WorkerOfflineMsg struct {
	Type     MessageType `json:"type"`
	WorkerID string      `json:"worker_id"`
}

type PingMessage struct {
	Type MessageType `json:"type"`
}
