package worker

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"streamforge/internal/event"
	"streamforge/internal/queue"
)

// --- Mocks ---

type MockConsumer struct {
	consumeFunc func(ctx context.Context, group, consumerName string) (*queue.Job, error)
}

func (m *MockConsumer) Consume(ctx context.Context, group, consumerName string) (*queue.Job, error) {
	if m.consumeFunc != nil {
		return m.consumeFunc(ctx, group, consumerName)
	}
	<-ctx.Done()
	return nil, ctx.Err()
}

type MockAcker struct {
	acked int32
}

func (m *MockAcker) Ack(ctx context.Context, j *queue.Job) error {
	atomic.AddInt32(&m.acked, 1)
	return nil
}

type MockHeartbeater struct {
	pings int32
	err   error
}

func (m *MockHeartbeater) Ping(ctx context.Context, workerID string, ttl time.Duration) error {
	atomic.AddInt32(&m.pings, 1)
	return m.err
}

type MockProcessor struct {
	processed int32
	delay     time.Duration
}

func (m *MockProcessor) Process(ctx context.Context, e *event.Event) error {
	if m.delay > 0 {
		time.Sleep(m.delay)
	}
	atomic.AddInt32(&m.processed, 1)
	return nil
}

type MockRetrier struct {
	retries int32
}

func (m *MockRetrier) Retry(ctx context.Context, j *queue.Job) error {
	atomic.AddInt32(&m.retries, 1)
	return nil
}

type MockEventFailureRecorder struct {
	status event.EventStatus
	err    error
}

func (m *MockEventFailureRecorder) RecordFailure(ctx context.Context, id uuid.UUID, maxAttempts int) (event.EventStatus, error) {
	return m.status, m.err
}

type MockEventStatusUpdater struct {
	updated int32
}

func (m *MockEventStatusUpdater) UpdateStatus(ctx context.Context, id uuid.UUID, old, new event.EventStatus) error {
	atomic.AddInt32(&m.updated, 1)
	return nil
}

func (m *MockEventStatusUpdater) FinalizeEvent(ctx context.Context, id uuid.UUID, resultHash string) error {
	atomic.AddInt32(&m.updated, 1)
	return nil
}

// --- Tests ---

func TestWorker_Lifecycle_And_Repeated_Start_Stop(t *testing.T) {
	w := NewWorker("w1", "g1", &MockConsumer{}, &MockAcker{}, &MockRetrier{}, &MockHeartbeater{}, &MockProcessor{}, &MockEventFailureRecorder{status: event.StatusFailed}, &MockEventStatusUpdater{}, nil, 50*time.Millisecond, time.Second, 3, time.Millisecond)

	if w.Status() != StatusStopped {
		t.Fatalf("expected STOPPED, got %s", w.Status())
	}

	err := w.Start(context.Background())
	if err != nil {
		t.Fatalf("unexpected error on start: %v", err)
	}

	// Repeated start should fail
	err = w.Start(context.Background())
	if err == nil {
		t.Fatal("expected error on repeated start")
	}

	// Wait briefly to allow transition to READY
	time.Sleep(10 * time.Millisecond)
	if w.Status() != StatusReady {
		t.Errorf("expected READY, got %s", w.Status())
	}

	// Stop the worker
	w.Stop()
	if w.Status() != StatusStopped {
		t.Fatalf("expected STOPPED, got %s", w.Status())
	}

	// Repeated stop should be a no-op and not panic
	w.Stop()
}

func TestWorker_SuccessfulProcessing_And_Ack(t *testing.T) {
	ev := &event.Event{ID: uuid.New()}
	job := &queue.Job{Event: ev, Receipt: "rcpt-1", Queue: "q1"}

	consumedCount := int32(0)
	consumer := &MockConsumer{
		consumeFunc: func(ctx context.Context, group, consumerName string) (*queue.Job, error) {
			if atomic.AddInt32(&consumedCount, 1) == 1 {
				return job, nil
			}
			<-ctx.Done()
			return nil, ctx.Err()
		},
	}
	acker := &MockAcker{}
	processor := &MockProcessor{}

	w := NewWorker("w1", "g1", consumer, acker, &MockRetrier{}, &MockHeartbeater{}, processor, &MockEventFailureRecorder{status: event.StatusFailed}, &MockEventStatusUpdater{}, nil, time.Second, time.Second, 3, time.Millisecond)
	w.Start(context.Background())

	// Wait for processing
	time.Sleep(50 * time.Millisecond)

	w.Stop()

	if atomic.LoadInt32(&processor.processed) != 1 {
		t.Errorf("expected 1 job processed, got %d", processor.processed)
	}
	if atomic.LoadInt32(&acker.acked) != 1 {
		t.Errorf("expected 1 job acked, got %d", acker.acked)
	}
}

func TestWorker_Heartbeat_SuccessAndFailure(t *testing.T) {
	hb := &MockHeartbeater{}

	w := NewWorker("w1", "g1", &MockConsumer{}, &MockAcker{}, &MockRetrier{}, hb, &MockProcessor{}, &MockEventFailureRecorder{status: event.StatusFailed}, &MockEventStatusUpdater{}, nil, 20*time.Millisecond, time.Second, 3, time.Millisecond)
	w.Start(context.Background())

	time.Sleep(70 * time.Millisecond)
	w.Stop()

	pings := atomic.LoadInt32(&hb.pings)
	if pings < 2 {
		t.Errorf("expected multiple pings, got %d", pings)
	}

	// Test heartbeat failure doesn't crash worker
	hbFailing := &MockHeartbeater{err: errors.New("redis dead")}
	w2 := NewWorker("w2", "g1", &MockConsumer{}, &MockAcker{}, &MockRetrier{}, hbFailing, &MockProcessor{}, &MockEventFailureRecorder{status: event.StatusFailed}, &MockEventStatusUpdater{}, nil, 20*time.Millisecond, time.Second, 3, time.Millisecond)
	w2.Start(context.Background())

	time.Sleep(50 * time.Millisecond)
	// If it crashed, Stop() might hang or we'd see a panic. It should just gracefully stop.
	w2.Stop()

	if atomic.LoadInt32(&hbFailing.pings) < 1 {
		t.Errorf("expected pings despite errors, got %d", hbFailing.pings)
	}
}

func TestWorker_ConsumeFailure_Backoff(t *testing.T) {
	consumerCalls := int32(0)
	consumer := &MockConsumer{
		consumeFunc: func(ctx context.Context, group, consumerName string) (*queue.Job, error) {
			atomic.AddInt32(&consumerCalls, 1)
			return nil, errors.New("network error")
		},
	}

	w := NewWorker("w1", "g1", consumer, &MockAcker{}, &MockRetrier{}, &MockHeartbeater{}, &MockProcessor{}, &MockEventFailureRecorder{status: event.StatusFailed}, &MockEventStatusUpdater{}, nil, time.Second, time.Second, 3, time.Millisecond)
	w.Start(context.Background())

	// Wait 1.5 seconds. With a 1-second backoff, we should see 2 calls max.
	time.Sleep(1500 * time.Millisecond)
	w.Stop()

	calls := atomic.LoadInt32(&consumerCalls)
	if calls > 3 {
		t.Errorf("expected backoff to limit calls, got %d calls", calls)
	}
}

func TestWorker_ConcurrentStatusAccess(t *testing.T) {
	w := NewWorker("w1", "g1", &MockConsumer{}, &MockAcker{}, &MockRetrier{}, &MockHeartbeater{}, &MockProcessor{}, &MockEventFailureRecorder{status: event.StatusFailed}, &MockEventStatusUpdater{}, nil, time.Second, time.Second, 3, time.Millisecond)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			w.Status()
			time.Sleep(1 * time.Millisecond)
		}
	}()

	w.Start(context.Background())
	w.Stop()
	wg.Wait()
	// If no data race panic occurs, test passes.
}

func TestWorker_Shutdown_WhileProcessing(t *testing.T) {
	jobStarted := make(chan struct{})

	consumer := &MockConsumer{
		consumeFunc: func(ctx context.Context, group, consumerName string) (*queue.Job, error) {
			select {
			case <-jobStarted:
				<-ctx.Done()
				return nil, ctx.Err()
			default:
				close(jobStarted)
				return &queue.Job{Event: &event.Event{}}, nil
			}
		},
	}

	processor := &MockProcessor{delay: 200 * time.Millisecond}
	w := NewWorker("w1", "g1", consumer, &MockAcker{}, &MockRetrier{}, &MockHeartbeater{}, processor, &MockEventFailureRecorder{status: event.StatusFailed}, &MockEventStatusUpdater{}, nil, time.Second, time.Second, 3, time.Millisecond)

	w.Start(context.Background())

	// Wait until job has actually started processing
	<-jobStarted

	// Ensure status is BUSY
	// give it a tiny moment to transition state after consume returns
	time.Sleep(10 * time.Millisecond)
	if w.Status() != StatusBusy {
		t.Errorf("expected BUSY, got %s", w.Status())
	}

	// Trigger shutdown
	stopStart := time.Now()
	w.Stop()
	stopDuration := time.Since(stopStart)

	// Stop() should have blocked until the 200ms processing finished
	if stopDuration < 150*time.Millisecond {
		t.Errorf("expected Stop() to wait for processing to finish, took %v", stopDuration)
	}

	if atomic.LoadInt32(&processor.processed) != 1 {
		t.Errorf("expected job to finish processing, got %d", processor.processed)
	}

	if w.Status() != StatusStopped {
		t.Errorf("expected STOPPED, got %s", w.Status())
	}
}
