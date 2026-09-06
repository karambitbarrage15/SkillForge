package worker

import (
	"context"
	"errors"
	"log/slog"
	"strconv"
	"sync"
	"time"

	"github.com/google/uuid"

	"streamforge/internal/event"
	"streamforge/internal/metrics"
	"streamforge/internal/queue"
	"streamforge/internal/store"
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
	AssignEvent(ctx context.Context, id uuid.UUID, workerID string, old, new event.EventStatus) error
	FinalizeEvent(ctx context.Context, id uuid.UUID, resultHash string) error
}

type Broadcaster interface {
	Broadcast(ctx context.Context, payload []byte)
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
	broadcaster Broadcaster

	interval    time.Duration
	ttl         time.Duration
	maxAttempts int
	baseDelay   time.Duration
}

// NewWorker creates a new worker instance.
func NewWorker(id, group string, consumer queue.Consumer, acker queue.Acker, retrier queue.Retrier, hb Heartbeater, proc Processor, recorder EventFailureRecorder, updater EventStatusUpdater, broadcaster Broadcaster, interval, ttl time.Duration, maxAttempts int, baseDelay time.Duration) *Worker {
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
		broadcaster: broadcaster,
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

	if w.broadcaster != nil {
		w.broadcaster.Broadcast(context.Background(), []byte(`{"type":"WORKER_ONLINE","worker_id":"`+w.id+`"}`))
	}

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

	if w.broadcaster != nil {
		w.broadcaster.Broadcast(context.Background(), []byte(`{"type":"WORKER_OFFLINE","worker_id":"`+w.id+`"}`))
	}
}

func (w *Worker) runHeartbeat(ctx context.Context) {
	defer w.wg.Done()

	// Initial ping
	if err := w.heartbeater.Ping(ctx, w.id, w.ttl); err != nil {
		slog.Error("Initial heartbeat failed", "worker_id", w.id, "err", err)
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
				slog.Error("Heartbeat failed", "worker_id", w.id, "err", err)
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
			slog.Error("Consume error", "worker_id", w.id, "err", err)
			select {
			case <-time.After(1 * time.Second):
				continue
			case <-ctx.Done():
				return
			}
		}

		// Process the job
		w.setStatus(StatusBusy)

		// Phase 8 & 14: Database state machine transition to PROCESSING with worker assignment
		if err := w.updater.AssignEvent(ctx, job.Event.ID, w.id, event.StatusQueued, event.StatusProcessing); err != nil {
			if errors.Is(err, store.ErrInvalidStateTransition) {
				slog.Info("Job is no longer QUEUED (state conflict). Acking to discard duplicate.", "worker_id", w.id, "event_id", job.Event.ID)
				w.acker.Ack(ctx, job)
			} else {
				slog.Error("Failed to assign job to PROCESSING due to infra error", "worker_id", w.id, "event_id", job.Event.ID, "err", err)
			}
			w.setStatus(StatusReady)
			continue
		}
		// Wait, user requirements say:
		// "8. If a job is already being processed when shutdown begins, allow the current processor call to finish before the worker reaches STOPPED, as specified by WORKER_POOL.md."
		// Thus, we pass a background-derived context to Processor if we want to allow it to finish, OR we just pass a context that is not cancelled when `ctx` is cancelled.
		// Actually, standard context propagation means if `ctx` is cancelled, `Process(ctx)` might abort early.
		// If we want it to finish cleanly without being interrupted by the shutdown, we should pass context.Background() with an independent timeout (e.g., job timeout).
		// Let's just use context.Background() for the job, to shield it from shutdown cancellation, as we wait for wg.Wait() anyway.
		jobCtx := context.Background()
		startT := time.Now()

		if w.broadcaster != nil {
			w.broadcaster.Broadcast(context.Background(), []byte(`{"type":"EVENT_PROCESSING","event_id":"`+job.Event.ID.String()+`","worker_id":"`+w.id+`"}`))
		}

		err = w.processor.Process(jobCtx, job.Event)
		if err != nil {
			slog.Error("Process error for job", "worker_id", w.id, "event_id", job.Event.ID, "err", err)
			w.handleFailure(ctx, jobCtx, job)
		} else {
			// Phase 11: Idempotent atomic finalization
			// Use a SHA-256 hash of "completed" as the deterministic success marker
			resultHash := "a38c4b12759e669bc01a742880c9261a8f9024f0c4369e8b7f8df1cb52fc4cf2" // SHA-256("completed")
			err := w.updater.FinalizeEvent(jobCtx, job.Event.ID, resultHash)
			if err != nil {
				if errors.Is(err, store.ErrAlreadyFinalized) {
					slog.Info("Job was already finalized. Acknowledging duplicate.", "worker_id", w.id, "event_id", job.Event.ID)
					if err := w.acker.Ack(jobCtx, job); err != nil {
						slog.Error("Ack error for duplicate job", "worker_id", w.id, "event_id", job.Event.ID, "err", err)
					}
				} else {
					slog.Error("Failed to finalize event", "worker_id", w.id, "event_id", job.Event.ID, "err", err)
				}
			} else {
				// Acknowledge the job on success
				if err := w.acker.Ack(jobCtx, job); err != nil {
					slog.Error("Ack error for job", "worker_id", w.id, "event_id", job.Event.ID, "err", err)
				}
				metrics.EventsProcessed.WithLabelValues(string(job.Event.Type)).Inc()
				durSeconds := time.Since(startT).Seconds()
				metrics.EventProcessingDuration.WithLabelValues(string(job.Event.Type)).Observe(durSeconds)

				if w.broadcaster != nil {
					// Broadcast completed
					dur := time.Since(startT).Milliseconds()
					msg := `{"type":"EVENT_COMPLETED","event_id":"` + job.Event.ID.String() + `","processing_time_ms":` + func() string {
						importStr := strconv.FormatInt(dur, 10)
						return importStr
					}() + `}`
					w.broadcaster.Broadcast(context.Background(), []byte(msg))
				}
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
		slog.Error("Failed to record failure for job", "worker_id", w.id, "event_id", job.Event.ID, "err", err)
		return
	}

	job.Event.Attempt++ // keep local object in sync

	if nextStatus == event.StatusFailed {
		// Permanently failed. Ack from Redis to drop it from PEL.
		metrics.EventsFailed.WithLabelValues(string(job.Event.Type)).Inc()
		if err := w.acker.Ack(jobCtx, job); err != nil {
			slog.Error("Ack error for permanently failed job", "worker_id", w.id, "event_id", job.Event.ID, "err", err)
		}
		if w.broadcaster != nil {
			msg := `{"type":"EVENT_FAILED","event_id":"` + job.Event.ID.String() + `","attempt":` + strconv.Itoa(job.Event.Attempt) + `}`
			w.broadcaster.Broadcast(context.Background(), []byte(msg))
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
				slog.Error("Failed to requeue job in Redis", "worker_id", w.id, "event_id", job.Event.ID, "err", err)
				// Do not update Postgres. The event stays in RETRYING and will be recovered by Phase 14.
				return
			}

			// 2. Update PostgreSQL from RETRYING -> QUEUED
			if err := w.updater.UpdateStatus(context.Background(), job.Event.ID, event.StatusRetrying, event.StatusQueued); err != nil {
				slog.Error("Failed to update job to QUEUED after Redis requeue", "worker_id", w.id, "event_id", job.Event.ID, "err", err)
			}
		case <-ctx.Done():
			// Worker is shutting down. Do not block shutdown, just exit.
			// Message remains un-acked in Redis PEL and status remains RETRYING in PostgreSQL.
			slog.Info("Aborting scheduled retry for job due to shutdown", "worker_id", w.id, "event_id", job.Event.ID)
			return
		}
	}()
}
