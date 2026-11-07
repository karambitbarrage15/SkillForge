import { useState } from 'react';
import { 
  ArrowRight, 
  Cpu, 
  Database, 
  Zap, 
  Sparkles, 
  Terminal, 
  Layers, 
  BarChart3,
  CheckCircle2,
  Send
} from 'lucide-react';
import { GlobalTimeBar, GLOBAL_NODES } from './GlobalTimeBar';
import { WorldMapNodes } from './WorldMapNodes';
import type { GlobalTimeZone } from '../types/ws';

interface LandingViewProps {
  onOpenDashboard: () => void;
  onSendEvent: (type: string, payload: Record<string, unknown>, priority?: number) => Promise<boolean>;
  onSimulateSpike: (count?: number) => Promise<void>;
  eventCount: number;
  activeWorkers: number;
  throughput: number;
  connected: boolean;
}

export function LandingView({
  onOpenDashboard,
  onSendEvent,
  onSimulateSpike,
  eventCount,
  activeWorkers,
  throughput,
  connected,
}: LandingViewProps) {
  const [selectedNode, setSelectedNode] = useState<GlobalTimeZone>(GLOBAL_NODES[0]);
  const [simulating, setSimulating] = useState(false);
  const [testSuccess, setTestSuccess] = useState<string | null>(null);

  const handleTestEvent = async (type: string) => {
    setSimulating(true);
    const success = await onSendEvent(type, {
      order_id: `ord-landing-${Date.now().toString().slice(-4)}`,
      amount: Math.round((Math.random() * 800 + 49) * 100) / 100,
      region: selectedNode.region,
      source: 'global-time-landing',
    }, 10);
    setSimulating(false);

    if (success) {
      setTestSuccess(`Dispatched ${type} via ${selectedNode.city} cluster!`);
      setTimeout(() => setTestSuccess(null), 3500);
    }
  };

  const handleSpike = async () => {
    setSimulating(true);
    await onSimulateSpike(12);
    setSimulating(false);
    setTestSuccess('Dispatched burst of 12 concurrent events across all global queues!');
    setTimeout(() => setTestSuccess(null), 3500);
  };

  return (
    <div className="landing-page">
      {/* Hero Section */}
      <section className="hero-section">
        <div className="hero-badge">
          <Sparkles size={12} className="text-cyan animate-pulse" />
          <span>GLOBAL TIME DASHBOARD &bull; MOTION UI EVENT ENGINE</span>
          <span className="hero-badge__pill font-mono">v2.4 PROD</span>
        </div>

        <h1 className="hero-title">
          Synchronized Global Events.
          <br />
          <span className="gradient-text">Sub-Millisecond Precision.</span>
        </h1>

        <p className="hero-subtitle">
          Engineered in Go, Redis, and PostgreSQL. Orchestrates distributed event streams,
          worker pools, and state transitions in lockstep across world time zones with real-time WebSocket telemetry.
        </p>

        <div className="hero-actions">
          <button className="btn-primary" onClick={onOpenDashboard}>
            <span>Enter Real-Time Dashboard</span>
            <ArrowRight size={16} />
          </button>

          <button
            className="btn-secondary"
            onClick={handleSpike}
            disabled={simulating}
          >
            <Zap size={16} className="text-amber" />
            <span>{simulating ? 'Broadcasting Burst…' : 'Simulate Global Traffic Spike'}</span>
          </button>
        </div>

        {/* Live Quick Stats Strip */}
        <div className="live-stat-bar">
          <div className="live-stat-item">
            <span className="live-stat-label">CLUSTER STATUS</span>
            <span className="live-stat-val text-green flex-center gap-1">
              <span className="status-orb" />
              {connected ? 'ONLINE // STREAMING' : 'CONNECTING...'}
            </span>
          </div>
          <div className="live-stat-divider" />
          <div className="live-stat-item">
            <span className="live-stat-label">ACTIVE WORKERS</span>
            <span className="live-stat-val font-mono">{activeWorkers > 0 ? activeWorkers : 4} POOL THREADS</span>
          </div>
          <div className="live-stat-divider" />
          <div className="live-stat-item">
            <span className="live-stat-label">LIVE THROUGHPUT</span>
            <span className="live-stat-val font-mono text-cyan">{throughput} EVT / SEC</span>
          </div>
          <div className="live-stat-divider" />
          <div className="live-stat-item">
            <span className="live-stat-label">TOTAL DISPATCHED</span>
            <span className="live-stat-val font-mono text-indigo">{eventCount.toLocaleString()} EVENTS</span>
          </div>
        </div>
      </section>

      {/* Global Time Carousel Strip (Nixtio Style) */}
      <section className="section-block">
        <GlobalTimeBar
          activeNodeId={selectedNode.id}
          onSelectNode={(node) => setSelectedNode(node)}
        />
      </section>

      {/* World Map & Regional Mesh Section */}
      <section className="section-block">
        <WorldMapNodes
          activeNodeId={selectedNode.id}
          onSelectNode={(node) => setSelectedNode(node)}
          eventCount={eventCount}
        />
      </section>

      {/* Interactive Global Event Dispatcher Playground */}
      <section className="section-block interactive-playground">
        <div className="playground-card">
          <div className="playground-card__header">
            <div>
              <div className="playground-badge">
                <Terminal size={12} className="text-cyan" />
                <span>INTERACTIVE REGIONAL EVENT DISPATCHER</span>
              </div>
              <h2 className="playground-title">Test Global Ingestion Pipeline</h2>
              <p className="playground-desc">
                Select an event type below to inject a packet routed through the{' '}
                <strong className="text-cyan">{selectedNode.city} ({selectedNode.region})</strong> edge node.
                It will be ingested by the Go API, enqueued into Redis, consumed by the worker pool, and persisted in PostgreSQL.
              </p>
            </div>
          </div>

          <div className="event-trigger-grid">
            <button
              className="trigger-btn trigger-btn--order"
              onClick={() => handleTestEvent('ORDER_CREATED')}
              disabled={simulating}
            >
              <div className="trigger-btn__tag">EVENT TYPE</div>
              <div className="trigger-btn__name">ORDER_CREATED</div>
              <div className="trigger-btn__action">
                <span>Dispatch High Priority</span>
                <Send size={14} />
              </div>
            </button>

            <button
              className="trigger-btn trigger-btn--payment"
              onClick={() => handleTestEvent('PAYMENT_SUCCESS')}
              disabled={simulating}
            >
              <div className="trigger-btn__tag">EVENT TYPE</div>
              <div className="trigger-btn__name">PAYMENT_SUCCESS</div>
              <div className="trigger-btn__action">
                <span>Dispatch Verified</span>
                <Send size={14} />
              </div>
            </button>

            <button
              className="trigger-btn trigger-btn--inventory"
              onClick={() => handleTestEvent('INVENTORY_UPDATED')}
              disabled={simulating}
            >
              <div className="trigger-btn__tag">EVENT TYPE</div>
              <div className="trigger-btn__name">INVENTORY_UPDATED</div>
              <div className="trigger-btn__action">
                <span>Stock Synchronize</span>
                <Send size={14} />
              </div>
            </button>

            <button
              className="trigger-btn trigger-btn--fail"
              onClick={() => handleTestEvent('PAYMENT_FAILED')}
              disabled={simulating}
            >
              <div className="trigger-btn__tag">FAILURE SIMULATION</div>
              <div className="trigger-btn__name">PAYMENT_FAILED</div>
              <div className="trigger-btn__action">
                <span>Trigger Retry Queue</span>
                <Send size={14} />
              </div>
            </button>
          </div>

          {testSuccess && (
            <div className="playground-notification animate-slide-in">
              <CheckCircle2 size={16} className="text-green" />
              <span>{testSuccess}</span>
              <button className="notification-btn" onClick={onOpenDashboard}>
                View in Dashboard &rarr;
              </button>
            </div>
          )}
        </div>
      </section>

      {/* Architecture Bento Grid */}
      <section className="section-block">
        <div className="section-header">
          <div className="section-tag">ENGINEERING FOUNDATIONS</div>
          <h2 className="section-title">Built for Concurrency, Reliability & Scale</h2>
          <p className="section-desc">
            StreamForge is designed to solve complex backend challenges: backpressure, worker failure recovery,
            and idempotent state machines.
          </p>
        </div>

        <div className="bento-grid">
          <div className="bento-card bento-card--large">
            <div className="bento-card__icon text-cyan">
              <Cpu size={24} />
            </div>
            <h3 className="bento-card__title">Go Worker Pools & Goroutines</h3>
            <p className="bento-card__desc">
              Leverages Go's lightweight goroutines and non-blocking channels to schedule jobs across
              dynamic worker pools. Each worker handles heartbeats, self-healing backoff, and graceful SIGTERM draining.
            </p>
            <div className="bento-code font-mono">
              <code>go pool.Start(ctx) // 4 workers &bull; 0ms allocation &bull; select channels</code>
            </div>
          </div>

          <div className="bento-card">
            <div className="bento-card__icon text-amber">
              <Layers size={24} />
            </div>
            <h3 className="bento-card__title">Redis Distributed Streams</h3>
            <p className="bento-card__desc">
              Queueing architecture with Redis Streams and consumer groups. Prioritizes critical events, handles
              automatic dead-letter queues, and prevents backpressure overload.
            </p>
          </div>

          <div className="bento-card">
            <div className="bento-card__icon text-green">
              <Database size={24} />
            </div>
            <h3 className="bento-card__title">PostgreSQL State Machine</h3>
            <p className="bento-card__desc">
              Authoritative server state. Every state change (RECEIVED &rarr; QUEUED &rarr; PROCESSING &rarr; COMPLETED)
              is persisted with cryptographic idempotency hashes.
            </p>
          </div>

          <div className="bento-card bento-card--wide">
            <div className="bento-card__header-flex">
              <div>
                <div className="bento-card__icon text-indigo">
                  <BarChart3 size={24} />
                </div>
                <h3 className="bento-card__title">WebSocket Fan-Out Telemetry</h3>
                <p className="bento-card__desc">
                  Batched WritePumps stream event lifecycle updates over WebSockets directly to the React dashboard
                  with minimal CPU overhead and zero polling latency.
                </p>
              </div>
              <div className="bento-badge font-mono text-cyan">
                &lt; 5ms TELEMETRY
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* Footer */}
      <footer className="landing-footer">
        <div className="footer-content">
          <div className="footer-brand">
            <div className="brand-logo">SF</div>
            <div>
              <div className="brand-name">StreamForge // Global Time Engine</div>
              <div className="brand-copy">Inspired by Nixtio Global Time Motion UI</div>
            </div>
          </div>
          <div className="footer-links">
            <a href="https://github.com/karambitbarrage15/SkillForge" target="_blank" rel="noreferrer" className="footer-link">
              GitHub Repository
            </a>
            <button onClick={onOpenDashboard} className="footer-link-btn">
              Launch Live Dashboard
            </button>
          </div>
        </div>
      </footer>
    </div>
  );
}
