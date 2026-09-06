package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// EventsReceived tracks the number of successfully accepted/validated event submissions.
	EventsReceived = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "events_received_total",
		Help: "Total number of successfully received events",
	}, []string{"type"})

	// EventsProcessed tracks the number of successfully processed events.
	EventsProcessed = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "events_processed_total",
		Help: "Total number of successfully processed events",
	}, []string{"type"})

	// EventsFailed tracks the number of failed event processing attempts.
	EventsFailed = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "events_failed_total",
		Help: "Total number of failed events",
	}, []string{"type"})

	// EventProcessingDuration tracks the duration of successful event processing.
	EventProcessingDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "event_processing_duration_seconds",
		Help:    "Duration of event processing in seconds",
		Buckets: prometheus.DefBuckets,
	}, []string{"type"})

	// QueueDepth tracks the number of pending events in the queue (if identifiable).
	QueueDepth = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "queue_depth",
		Help: "Current depth of the event queue",
	}, []string{"queue"})

	// ActiveWorkers tracks the number of active workers based on heartbeats.
	ActiveWorkers = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "active_workers",
		Help: "Current number of active workers",
	})

	// WorkerFailures tracks the number of unexpected worker failures.
	WorkerFailures = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "worker_failures_total",
		Help: "Total number of unexpected worker failures",
	}, []string{"reason"})

	// WebsocketConnections tracks the number of active WebSocket connections.
	WebsocketConnections = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "websocket_connections",
		Help: "Current number of active WebSocket connections",
	})
)
