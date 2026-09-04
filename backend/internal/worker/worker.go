package worker

import (
	"context"
	"errors"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"

	"streamforge/internal/event"
	"streamforge/internal/queue"
)

// Status represents the lifecycle state of a Worker.
type Status string

const (
	StatusStarting Status = "STARTING"
	StatusReady    Status = "READY"
	StatusBusy     Status = "BUSY"
	StatusStopping Status = "STOPPING"
	StatusStopped  Status = "STOPPED"
)

// Processor is the minimal interface required for Phase 6.
// In Phase 9, this will be expanded to the full processor registry.
type Processor interface {
	Process(ctx context.Context, e *event.Event) error
}

type EventFailureRecorder interface {
	RecordFailure(ctx context.Context, id uuid.UUID, maxAttempts int) (event.EventStatus, error)
}

type EventStatusUpdater interface {
	UpdateStatus(ctx context.Context, id uuid.UUID, old, new event.EventStatus) error
}

// Worker represents a single unit of execution.
type Worker struct {
	id    string
	group string

	status   Status
	statusMu sync.RWMutex

	startMu sync.Mutex // Protects against concurrent Start/Stop calls
	cancel  context.CancelFunc
	wg      sync.WaitGroup

	consumer    queue.Consumer
	acker       queue.Acker
	retrier     queue.Retrier
	heartbeater Heartbeater
	processor   Processor
	recorder    EventFailureRecorder
	updater     EventStatusUpdater

	interval    time.Duration
	ttl         time.Duration
	maxAttempts int
	baseDelay   time.Duration
}

// NewWorker creates a new worker instance.
func NewWorker(id, group string, consumer queue.Consumer, acker queue.Acker, retrier queue.Retrier, hb Heartbeater, proc Processor, recorder EventFailureRecorder, updater EventStatusUpdater, interval, ttl time.Duration, maxAttempts int, baseDelay time.Duration) *Worker {
	return &Worker{
		id:          id,
		group:       group,
		status:      StatusStopped,
		consumer:    consumer,
		acker:       acker,
		retrier:     retrier,
		heartbeater: hb,
		processor:   proc,
		recorder:    recorder,
		updater:     updater,
		interval:    interval,
		ttl:         ttl,
		maxAttempts: maxAttempts,
		baseDelay:   baseDelay,
	}
}

// Status returns the current status of the worker safely.
func (w *Worker) Status() Status {
	w.statusMu.RLock()
	defer w.statusMu.RUnlock()
	return w.status
}

func (w *Worker) setStatus(s Status) {
	w.statusMu.Lock()
	defer w.statusMu.Unlock()
	w.status = s
}

// Start boots the worker's internal goroutines. Safe for concurrent calls (ignores if already started).
func (w *Worker) Start(ctx context.Context) error {
	w.startMu.Lock()
	defer w.startMu.Unlock()

	if w.Status() != StatusStopped {
		return errors.New("worker is already running or stopping")
	}

	w.setStatus(StatusStarting)

	// Create a cancellable context for the worker's internal operations
	ctx, cancel := context.WithCancel(ctx)
	w.cancel = cancel

	w.setStatus(StatusReady)

	w.wg.Add(2)
	go w.runHeartbeat(ctx)
	go w.runJobLoop(ctx)

	return nil
}

// Stop gracefully signals the worker to stop accepting new jobs,
// wait for the current job to finish, and transition to STOPPED.
func (w *Worker) Stop() {
	w.startMu.Lock()

	if w.Status() == StatusStopped || w.Status() == StatusStopping {
		w.startMu.Unlock()
		return
	}

	w.setStatus(StatusStopping)
	if w.cancel != nil {
		w.cancel()
	}
	w.startMu.Unlock()

	// Wait for goroutines to finish
	w.wg.Wait()
	w.setStatus(StatusStopped)
}

func (w *Worker) runHeartbeat(ctx context.Context) {
	defer w.wg.Done()

	// Initial ping
	if err := w.heartbeater.Ping(ctx, w.id, w.ttl); err != nil {
		log.Printf("[Worker %s] Initial heartbeat failed: %v", w.id, err)
	}

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := w.heartbeater.Ping(ctx, w.id, w.ttl); err != nil {
				// We log but do NOT crash the worker.
				log.Printf("[Worker %s] Heartbeat failed: %v", w.id, err)
			}
		}
	}
}

func (w *Worker) runJobLoop(ctx context.Context) {
	defer w.wg.Done()

	for {
		// Stop accepting new jobs if context is cancelled
		if ctx.Err() != nil {
			return
		}

		job, err := w.consumer.Consume(ctx, w.group, w.id)
		if err != nil {
			// If error is context cancellation, exit cleanly
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return
			}

			// Otherwise log and apply a short backoff to prevent CPU spinning on persistent errors
			log.Printf("[Worker %s] Consume error: %v", w.id, err)
			select {
			case <-time.After(1 * time.Second):
				continue
			case <-ctx.Done():
				return
			}
		}

		// Process the job
		w.setStatus(StatusBusy)

		// Create a timeout context for the processor, bound by the worker's context
		// Wait, user requirements say:
		// "8. If a job is already being processed when shutdown begins, allow the current processor call to finish before the worker reaches STOPPED, as specified by WORKER_POOL.md."
		// Thus, we pass a background-derived context to Processor if we want to allow it to finish, OR we just pass a context that is not cancelled when `ctx` is cancelled.
		// Actually, standard context propagation means if `ctx` is cancelled, `Process(ctx)` might abort early.
		// If we want it to finish cleanly without being interrupted by the shutdown, we should pass context.Background() with an independent timeout (e.g., job timeout).
		// Let's just use context.Background() for the job, to shield it from shutdown cancellation, as we wait for wg.Wait() anyway.
		jobCtx := context.Background()

		err = w.processor.Process(jobCtx, job.Event)
		if err != nil {
			log.Printf("[Worker %s] Process error for job %s: %v", w.id, job.Event.ID, err)
			w.handleFailure(ctx, jobCtx, job)
		} else {
			// Acknowledge the job on success
			if err := w.acker.Ack(jobCtx, job); err != nil {
				log.Printf("[Worker %s] Ack error for job %s: %v", w.id, job.Event.ID, err)
			}
		}

		// Only return to READY if we aren't stopping
		if w.Status() != StatusStopping {
			w.setStatus(StatusReady)
		}
	}
}

func (w *Worker) handleFailure(ctx, jobCtx context.Context, job *queue.Job) {
	// 1. Record the failure in PostgreSQL. This is atomic.
	nextStatus, err := w.recorder.RecordFailure(jobCtx, job.Event.ID, w.maxAttempts)
	if err != nil {
		log.Printf("[Worker %s] Failed to record failure for job %s: %v", w.id, job.Event.ID, err)
		return
	}

	job.Event.Attempt++ // keep local object in sync

	if nextStatus == event.StatusFailed {
		// Permanently failed. Ack from Redis to drop it from PEL.
		if err := w.acker.Ack(jobCtx, job); err != nil {
			log.Printf("[Worker %s] Ack error for permanently failed job %s: %v", w.id, job.Event.ID, err)
		}
		return
	}

	// It's RETRYING. We schedule the requeue.
	delay := CalculateBackoff(job.Event.Attempt, w.baseDelay)

	w.wg.Add(1)
	go func() {
		defer w.wg.Done()

		select {
		case <-time.After(delay):
			// Proceed with requeue.
			// IMPORTANT ORDERING as requested:
			// 1. Publish to Redis & Ack old (using a background context so it's not cancelled by Worker shutdown if we got this far)
			if err := w.retrier.Retry(context.Background(), job); err != nil {
				log.Printf("[Worker %s] Failed to requeue job %s in Redis: %v", w.id, job.Event.ID, err)
				// Do not update Postgres. The event stays in RETRYING and will be recovered by Phase 14.
				return
			}

			// 2. Update PostgreSQL from RETRYING -> QUEUED
			if err := w.updater.UpdateStatus(context.Background(), job.Event.ID, event.StatusRetrying, event.StatusQueued); err != nil {
				log.Printf("[Worker %s] Failed to update job %s to QUEUED after Redis requeue: %v", w.id, job.Event.ID, err)
			}
		case <-ctx.Done():
			// Worker is shutting down. Do not block shutdown, just exit.
			// Message remains un-acked in Redis PEL and status remains RETRYING in PostgreSQL.
			log.Printf("[Worker %s] Aborting scheduled retry for job %s due to shutdown", w.id, job.Event.ID)
			return
		}
	}()
}
