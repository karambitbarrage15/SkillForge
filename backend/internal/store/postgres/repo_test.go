package postgres

import (
	"context"
	"encoding/json"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"streamforge/internal/event"
	"streamforge/internal/store"
)

func setupTestDB(t *testing.T) *DB {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		t.Skip("DATABASE_URL not set. Skipping integration tests.")
	}

	ctx := context.Background()
	db, err := Connect(ctx, dbURL)
	if err != nil {
		t.Fatalf("Failed to connect to test db: %v", err)
	}

	// Truncate tables for a clean slate
	_, err = db.pool.Exec(ctx, "TRUNCATE TABLE events, workers, executions, idempotency RESTART IDENTITY CASCADE")
	if err != nil {
		t.Fatalf("Failed to truncate tables: %v", err)
	}

	return db
}

func TestEventRepo_UpdateStatus(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := NewEventRepo(db)
	ctx := context.Background()

	// 1. Create a test event
	e := &event.Event{
		ID:        uuid.New(),
		Type:      event.TypeOrderCreated,
		Payload:   json.RawMessage(`{"val": 1}`),
		Priority:  event.PriorityNormal,
		Status:    event.StatusReceived,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := repo.Create(ctx, e)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// 2. Valid transition
	err = repo.UpdateStatus(ctx, e.ID, event.StatusReceived, event.StatusQueued)
	if err != nil {
		t.Errorf("Expected valid transition to succeed, got: %v", err)
	}

	// 3. Invalid transition (rejected by state machine)
	err = repo.UpdateStatus(ctx, e.ID, event.StatusQueued, event.StatusCompleted)
	if err != store.ErrInvalidStateTransition {
		t.Errorf("Expected ErrInvalidStateTransition from state machine, got: %v", err)
	}

	// 4. Stale old status (rejected by DB optimistic lock)
	err = repo.UpdateStatus(ctx, e.ID, event.StatusReceived, event.StatusQueued)
	if err != store.ErrInvalidStateTransition {
		t.Errorf("Expected ErrInvalidStateTransition from DB check, got: %v", err)
	}

	// 5. Valid transition to PROCESSING
	err = repo.UpdateStatus(ctx, e.ID, event.StatusQueued, event.StatusProcessing)
	if err != nil {
		t.Errorf("Expected transition to PROCESSING to succeed, got: %v", err)
	}

	// 6. Valid transition to COMPLETED (should set completed_at)
	err = repo.UpdateStatus(ctx, e.ID, event.StatusProcessing, event.StatusCompleted)
	if err != nil {
		t.Errorf("Expected transition to COMPLETED to succeed, got: %v", err)
	}

	updatedEvent, _ := repo.GetByID(ctx, e.ID)
	if updatedEvent.CompletedAt == nil {
		t.Error("Expected CompletedAt to be set after transition to COMPLETED")
	}

	// 7. Event not found
	err = repo.UpdateStatus(ctx, uuid.New(), event.StatusReceived, event.StatusQueued)
	if err != store.ErrEventNotFound {
		t.Errorf("Expected ErrEventNotFound, got: %v", err)
	}
}

func TestEventRepo_ConcurrentUpdateStatus(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := NewEventRepo(db)
	ctx := context.Background()

	e := &event.Event{
		ID:        uuid.New(),
		Type:      event.TypeOrderCreated,
		Payload:   json.RawMessage(`{"val": 2}`),
		Priority:  event.PriorityNormal,
		Status:    event.StatusQueued,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	_ = repo.Create(ctx, e)

	// Simulate 10 workers trying to claim the same queued event simultaneously
	numWorkers := 10
	var wg sync.WaitGroup
	successCount := 0
	var mu sync.Mutex

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := repo.UpdateStatus(context.Background(), e.ID, event.StatusQueued, event.StatusProcessing)
			if err == nil {
				mu.Lock()
				successCount++
				mu.Unlock()
			}
		}()
	}

	wg.Wait()

	if successCount != 1 {
		t.Errorf("Expected exactly 1 worker to succeed, but %d succeeded", successCount)
	}
}

func TestIdempotencyRepo_CheckOrInsert(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := NewIdempotencyRepo(db)
	ctx := context.Background()

	record := &store.IdempotencyRecord{
		EventID:    uuid.New(),
		ResultHash: "hash123",
		CreatedAt:  time.Now(),
	}

	// First insert should succeed
	inserted, err := repo.CheckOrInsert(ctx, record)
	if err != nil {
		t.Fatalf("CheckOrInsert failed: %v", err)
	}
	if !inserted {
		t.Error("Expected first insert to return true")
	}

	// Second insert of same ID should fail gracefully and return false
	inserted, err = repo.CheckOrInsert(ctx, record)
	if err != nil {
		t.Fatalf("Second CheckOrInsert returned error: %v", err)
	}
	if inserted {
		t.Error("Expected second insert to return false")
	}
}

func TestContextCancellation(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := NewEventRepo(db)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	err := repo.UpdateStatus(ctx, uuid.New(), event.StatusReceived, event.StatusQueued)
	if err == nil {
		t.Error("Expected error from cancelled context, got nil")
	}
}

func TestEventRepo_List(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := NewEventRepo(db)
	ctx := context.Background()

	// Insert 3 events
	for i := 0; i < 3; i++ {
		e := &event.Event{
			ID:        uuid.New(),
			Type:      event.TypeOrderCreated,
			Payload:   json.RawMessage(`{}`),
			Priority:  event.PriorityNormal,
			Status:    event.StatusReceived,
			CreatedAt: time.Now().Add(time.Duration(i) * time.Second), // i=2 is newest
			UpdatedAt: time.Now(),
		}
		if err := repo.Create(ctx, e); err != nil {
			t.Fatalf("Failed to create event: %v", err)
		}
	}

	// Test List with limit 2, offset 0
	events, err := repo.List(ctx, 2, 0)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(events) != 2 {
		t.Fatalf("Expected 2 events, got %d", len(events))
	}
	// Verify descending order (newest first)
	if events[0].CreatedAt.Before(events[1].CreatedAt) {
		t.Errorf("Expected descending order, but events[0] is older than events[1]")
	}

	// Test offset 2 (should return the remaining 1 event)
	events, err = repo.List(ctx, 2, 2)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(events) != 1 {
		t.Errorf("Expected 1 event, got %d", len(events))
	}
}
