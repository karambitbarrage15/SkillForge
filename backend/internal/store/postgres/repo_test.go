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

	// 3. Invalid transition (stale old status)
	err = repo.UpdateStatus(ctx, e.ID, event.StatusReceived, event.StatusProcessing)
	if err != store.ErrInvalidStateTransition {
		t.Errorf("Expected ErrInvalidStateTransition, got: %v", err)
	}

	// 4. Event not found
	err = repo.UpdateStatus(ctx, uuid.New(), event.StatusQueued, event.StatusProcessing)
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
