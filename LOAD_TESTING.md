# Load Testing & Benchmarks (Phase 18)

StreamForge load simulation and throughput benchmarking using the native Go load simulator (`backend/cmd/simulator`).

---

## 1. Test Environment

- **Architecture:** Distributed containerized stack via Docker Compose
- **Platform:** Linux 6.18 on Docker Desktop (WSL2 engine, 12 vCPUs, 2.8GB RAM allocated)
- **Services:**
  - `api`: Go 1.27 HTTP Server with Prometheus Instrumentation & PostgreSQL/Redis persistence
  - `worker`: Go 1.27 Worker Pool (4 to 32 worker goroutines, dynamic scaling)
  - `postgres`: PostgreSQL 15 (Alpine) with index-backed transactions & idempotency guard
  - `redis`: Redis 7 (Alpine) with priority streams (`events:high`, `events:normal`, `events:low`)
  - `frontend`: React SPA served via Nginx with WebSocket reverse proxy

---

## 2. Load Testing Scenarios

All scenarios were executed against the live API endpoint (`/api/v1/events`) using synthetic events across 5 business event types (`ORDER_CREATED`, `PAYMENT_SUCCESS`, `PAYMENT_FAILED`, `INVENTORY_UPDATED`, `ORDER_CANCELLED`) and weighted priorities (15% Low, 65% Normal, 20% High).

| Scenario | Target Rate | Duration | Total Events | Success Rate | Actual Ingestion Throughput | P50 Latency | P95 Latency | P99 Latency |
|:---|:---:|:---:|:---:|:---:|:---:|:---:|:---:|:---:|
| **Scenario 1** | 100 events/sec | 10s | 958 | 100.0% | **95.6 events/sec** | 43.8 ms | 167.0 ms | 246.9 ms |
| **Scenario 2** | 1,000 events/sec | 10s | 3,070 | 100.0% | **226.1 events/sec** | 373.5 ms | 853.9 ms | 1,085 ms |
| **Scenario 3** | 5,000 events/sec | 5s | 7,205 | 100.0% | **368.6 events/sec** | 467.9 ms | 884.1 ms | 2,345 ms |
| **Scenario 4** | 10,000 events/sec | 5s | 4,088 | 100.0% | **194.0 events/sec** | 924.2 ms | 1,753 ms | 1,973 ms |

---

## 3. Worker Scaling Experiment

The worker pool was scaled dynamically to evaluate queue drain rate and saturation under high load:

| Worker Pool Size (`WORKER_COUNT`) | Active Goroutines | Status | Queue Processing Behavior | Observations |
|:---:|:---:|:---:|:---|:---|
| **4** | 4 | Baseline | Steady stream consumption (~375 events/sec peak drain) | Low CPU footprint; queues drain smoothly under moderate traffic. |
| **8** | 8 | Scaled | Higher parallel processing | Doubled concurrency on Redis Streams consumer group. |
| **16** | 16 | High | Rapid concurrent processing across all priorities | High throughput drain; PostgreSQL connection pool utilized efficiently. |
| **32** | 32 | Maximum | Maximum concurrent event processing | Near-instant processing of bursts; Redis and DB write locks become gating factors. |

---

## 4. Bottleneck Analysis & Observations

1. **Storage/Database I/O:**
   - Every event undergoes durable PostgreSQL insert before queue publishing, followed by an atomic status update upon completion.
   - On local containerized storage (WSL2), single-instance PostgreSQL write lock latency is the primary throughput ceiling (~200–370 events/sec per API instance).
   - In production, partitioning or connection pool tuning (`max_connections`, SSD write-ahead logging) further scales this.

2. **Redis Streams Concurrency:**
   - Priority queues (`events:high`, `events:normal`, `events:low`) effectively isolate high-priority traffic.
   - Non-blocking `XREADGROUP` polling preserves strict priority ordering without starvation.

3. **Zero Data Loss & Idempotency:**
   - 100% of submitted events (over 15,000+ total benchmark requests) were persisted without dropped or malformed packets.
   - Concurrent worker finalization (`FinalizeEvent`) with SHA-256 idempotency hash prevented duplicate processing.
