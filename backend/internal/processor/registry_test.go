package processor

import (
	"context"
	"errors"
	"sync"
	"testing"

	"streamforge/internal/event"
)

type mockProcessor struct {
	called bool
	mu     sync.Mutex
}

func (m *mockProcessor) Process(ctx context.Context, e *event.Event) error {
	m.mu.Lock()
	m.called = true
	m.mu.Unlock()
	return nil
}

func (m *mockProcessor) wasCalled() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.called
}

func TestRegistry_RegisterAndProcess(t *testing.T) {
	r := NewRegistry()
	p1 := &mockProcessor{}
	p2 := &mockProcessor{}

	err := r.Register(event.TypeOrderCreated, p1)
	if err != nil {
		t.Fatalf("Expected nil error, got %v", err)
	}

	err = r.Register(event.TypePaymentSuccess, p2)
	if err != nil {
		t.Fatalf("Expected nil error, got %v", err)
	}

	// Test successful routing
	e1 := &event.Event{Type: event.TypeOrderCreated}
	err = r.Process(context.Background(), e1)
	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}
	if !p1.wasCalled() {
		t.Error("Expected p1 to be called")
	}
	if p2.wasCalled() {
		t.Error("Expected p2 to not be called")
	}

	// Test unknown event type
	eUnknown := &event.Event{Type: event.TypeOrderCancelled}
	err = r.Process(context.Background(), eUnknown)
	if !errors.Is(err, ErrUnknownEventType) {
		t.Errorf("Expected ErrUnknownEventType, got %v", err)
	}

	// Test duplicate registration
	err = r.Register(event.TypeOrderCreated, p2)
	if err == nil {
		t.Error("Expected error on duplicate registration")
	}

	// Test nil processor
	err = r.Register(event.TypeInventoryUpdated, nil)
	if !errors.Is(err, ErrNilProcessor) {
		t.Errorf("Expected ErrNilProcessor, got %v", err)
	}
}

func TestRegistry_ConcurrentProcess(t *testing.T) {
	r := NewRegistry()
	p1 := &mockProcessor{}
	_ = r.Register(event.TypeOrderCreated, p1)

	var wg sync.WaitGroup
	// 100 concurrent workers
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			e := &event.Event{Type: event.TypeOrderCreated}
			_ = r.Process(context.Background(), e)
		}()
	}

	wg.Wait()
	if !p1.wasCalled() {
		t.Error("Expected processor to be called")
	}
}
