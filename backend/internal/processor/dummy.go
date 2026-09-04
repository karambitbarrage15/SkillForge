package processor

import (
	"context"
	"log/slog"

	"streamforge/internal/event"
)

// DummyProcessor is a placeholder for real domain logic.
// It logs its execution and returns nil.
type DummyProcessor struct {
	Name string
}

func (p *DummyProcessor) Process(ctx context.Context, e *event.Event) error {
	slog.Info("DummyProcessor executed", "processor", p.Name, "event_id", e.ID, "event_type", e.Type)
	return nil
}

// NewOrderProcessor creates a dummy processor for ORDER_CREATED events.
func NewOrderProcessor() *DummyProcessor {
	return &DummyProcessor{Name: "OrderProcessor"}
}

// NewPaymentSuccessProcessor creates a dummy processor for PAYMENT_SUCCESS events.
func NewPaymentSuccessProcessor() *DummyProcessor {
	return &DummyProcessor{Name: "PaymentSuccessProcessor"}
}

// NewPaymentFailedProcessor creates a dummy processor for PAYMENT_FAILED events.
func NewPaymentFailedProcessor() *DummyProcessor {
	return &DummyProcessor{Name: "PaymentFailedProcessor"}
}

// NewInventoryProcessor creates a dummy processor for INVENTORY_UPDATED events.
func NewInventoryProcessor() *DummyProcessor {
	return &DummyProcessor{Name: "InventoryProcessor"}
}

// NewOrderCancelledProcessor creates a dummy processor for ORDER_CANCELLED events.
func NewOrderCancelledProcessor() *DummyProcessor {
	return &DummyProcessor{Name: "OrderCancelledProcessor"}
}

// NewUserRegisteredProcessor creates a dummy processor for USER_REGISTERED events.
func NewUserRegisteredProcessor() *DummyProcessor {
	return &DummyProcessor{Name: "UserRegisteredProcessor"}
}
