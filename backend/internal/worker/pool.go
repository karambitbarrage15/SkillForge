package worker

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"streamforge/internal/queue"
)

// Pool manages a bounded set of workers.
type Pool struct {
	count   int
	workers []*Worker

	startMu sync.Mutex
	started bool
	stopped bool
}

// NewPool creates a new Worker Pool with exactly `count` workers.
// It explicitly prevents unbounded worker creation by statically allocating the workers.
func NewPool(count int, groupName string, consumer queue.Consumer, acker queue.Acker, retrier queue.Retrier, hb Heartbeater, proc Processor, recorder EventFailureRecorder, updater EventStatusUpdater, broadcaster Broadcaster, interval, ttl time.Duration, maxAttempts int, baseDelay time.Duration) *Pool {
	if count <= 0 {
		count = 4 // Enforce a minimum safe limit
	}

	workers := make([]*Worker, count)
	for i := 0; i < count; i++ {
		workerID := fmt.Sprintf("worker-%d", i+1)
		workers[i] = NewWorker(workerID, groupName, consumer, acker, retrier, hb, proc, recorder, updater, broadcaster, interval, ttl, maxAttempts, baseDelay)
	}

	return &Pool{
		count:   count,
		workers: workers,
	}
}

// Start boots all exactly WORKER_COUNT workers. It is safe against concurrent calls.
func (p *Pool) Start(ctx context.Context) error {
	p.startMu.Lock()
	defer p.startMu.Unlock()

	if p.started || p.stopped {
		return errors.New("pool already started or stopped")
	}
	p.started = true

	for _, w := range p.workers {
		if err := w.Start(ctx); err != nil {
			return fmt.Errorf("failed to start worker %s: %w", w.id, err)
		}
	}
	return nil
}

// Stop gracefully stops all workers concurrently, blocking until every child Worker
// has completely reached STOPPED. It is safe against concurrent or repeated calls.
func (p *Pool) Stop() {
	p.startMu.Lock()
	if !p.started || p.stopped {
		p.startMu.Unlock()
		return
	}
	p.stopped = true
	p.startMu.Unlock()

	var wg sync.WaitGroup
	// Stop workers concurrently rather than sequentially
	for _, w := range p.workers {
		wg.Add(1)
		go func(worker *Worker) {
			defer wg.Done()
			worker.Stop() // Worker.Stop() itself blocks until that worker's goroutines finish
		}(w)
	}

	wg.Wait()
}

// GetWorkers returns a copy of the slice of workers, strictly for testing/inspection.
func (p *Pool) GetWorkers() []*Worker {
	return append([]*Worker(nil), p.workers...)
}
