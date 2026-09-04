package event

// EventStatus represents the lifecycle state of an Event.
type EventStatus string

const (
	StatusReceived   EventStatus = "RECEIVED"
	StatusQueued     EventStatus = "QUEUED"
	StatusProcessing EventStatus = "PROCESSING"
	StatusCompleted  EventStatus = "COMPLETED"
	StatusRetrying   EventStatus = "RETRYING"
	StatusFailed     EventStatus = "FAILED"
)

// EventType represents the category of the business event.
type EventType string

const (
	TypeOrderCreated     EventType = "ORDER_CREATED"
	TypePaymentSuccess   EventType = "PAYMENT_SUCCESS"
	TypePaymentFailed    EventType = "PAYMENT_FAILED"
	TypeInventoryUpdated EventType = "INVENTORY_UPDATED"
	TypeOrderCancelled   EventType = "ORDER_CANCELLED"
	TypeUserRegistered   EventType = "USER_REGISTERED" // Mentioned in README
)

// Priorities denote the processing urgency.
const (
	PriorityLow    int = 1
	PriorityNormal int = 5
	PriorityHigh   int = 10
)
