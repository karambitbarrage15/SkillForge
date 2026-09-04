package worker

import (
	"context"
	"sync"
	"testing"
	"time"

	"streamforge/internal/event"
	"streamforge/internal/queue"
)

func TestPool_CreatesExactlyWorkerCount(t *testing.T) {
	pool := NewPool(5, "g1", &MockConsumer{}, &MockAcker{}, &MockRetrier{}, &MockHeartbeater{}, &MockProcessor{}, &MockEventFailureRecorder{status: event.StatusFailed}, &MockEventStatusUpdater{}, nil, time.Second, time.Second, 3, time.Millisecond)

	workers := pool.GetWorkers()
	if len(workers) != 5 {
		t.Fatalf("expected exactly 5 workers, got %d", len(workers))
	}
}

func TestPool_StartAndStop_AllWorkers(t *testing.T) {
	pool := NewPool(3, "g1", &MockConsumer{}, &MockAcker{}, &MockRetrier{}, &MockHeartbeater{}, &MockProcessor{}, &MockEventFailureRecorder{status: event.StatusFailed}, &MockEventStatusUpdater{}, nil, 50*time.Millisecond, time.Second, 3, time.Millisecond)

	err := pool.Start(context.Background())
	if err != nil {
		t.Fatalf("unexpected error starting pool: %v", err)
	}

	// Brief wait for workers to shift out of STARTING if they race
	time.Sleep(10 * time.Millisecond)

	// Verify all started
	for _, w := range pool.GetWorkers() {
		if w.Status() == StatusStopped {
			t.Errorf("worker %s not started", w.id)
		}
	}

	// Stop pool
	pool.Stop()

	// Verify all stopped
	for _, w := range pool.GetWorkers() {
		if w.Status() != StatusStopped {
			t.Errorf("worker %s not stopped, status is %s", w.id, w.Status())
		}
	}
}

func TestPool_ConcurrentStartStopSafe(t *testing.T) {
	pool := NewPool(4, "g1", &MockConsumer{}, &MockAcker{}, &MockRetrier{}, &MockHeartbeater{}, &MockProcessor{}, &MockEventFailureRecorder{status: event.StatusFailed}, &MockEventStatusUpdater{}, nil, time.Second, time.Second, 3, time.Millisecond)

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		pool.Start(context.Background())
	}()

	go func() {
		defer wg.Done()
		// Start and stop race, but both should complete without panic or deadlock
		pool.Stop()
	}()

	wg.Wait()

	// Regardless of which won the race, second Stop shouldn't hang
	pool.Stop()
}

func TestPool_GracefulShutdownWaitsForAllWorkers(t *testing.T) {
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
	pool := NewPool(2, "g1", consumer, &MockAcker{}, &MockRetrier{}, &MockHeartbeater{}, processor, &MockEventFailureRecorder{status: event.StatusFailed}, &MockEventStatusUpdater{}, nil, time.Second, time.Second, 3, time.Millisecond)

	pool.Start(context.Background())
	<-jobStarted                      // Wait for a job to be pulled
	time.Sleep(10 * time.Millisecond) // Give consumer time to transition to processing

	startStop := time.Now()
	pool.Stop()
	duration := time.Since(startStop)

	if duration < 150*time.Millisecond {
		t.Errorf("expected Stop to block waiting for processor, took %v", duration)
	}
}
