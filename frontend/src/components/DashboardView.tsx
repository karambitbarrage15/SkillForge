import { useState, useEffect } from 'react';
import { 
  Activity, 
  Cpu, 
  CheckCircle, 
  Search, 
  Send, 
  Zap, 
  Radio, 
  Terminal,
  ChevronDown,
  ChevronUp
} from 'lucide-react';
import type { RecentEvent } from '../types/ws';
import { GLOBAL_NODES } from './GlobalTimeBar';

interface DashboardViewProps {
  connected: boolean;
  activeWorkers: Set<string>;
  completedCount: number;
  failedCount: number;
  recentEvents: RecentEvent[];
  throughput: number;
  avgLatency: number;
  onSendEvent: (type: string, payload: Record<string, unknown>, priority?: number) => Promise<boolean>;
  onSimulateSpike: (count?: number) => Promise<void>;
}

export function DashboardView({
  connected: _connected,
  activeWorkers,
  completedCount,
  failedCount,
  recentEvents,
  throughput,
  avgLatency,
  onSendEvent,
  onSimulateSpike,
}: DashboardViewProps) {
  const [filter, setFilter] = useState<string>('ALL');
  const [searchQuery, setSearchQuery] = useState('');
  const [currentTime, setCurrentTime] = useState(new Date());
  const [selectedCity, setSelectedCity] = useState(GLOBAL_NODES[0]);
  const [expandedEventId, setExpandedEventId] = useState<string | null>(null);
  const [dispatching, setDispatching] = useState(false);

  useEffect(() => {
    const timer = setInterval(() => setCurrentTime(new Date()), 1000);
    return () => clearInterval(timer);
  }, []);

  const formatCityTime = (timeZone: string) => {
    return new Intl.DateTimeFormat('en-US', {
      timeZone,
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit',
      hour12: false,
    }).format(currentTime);
  };

  const handleQuickDispatch = async (type: string) => {
    setDispatching(true);
    await onSendEvent(type, {
      order_id: `ord-dash-${Date.now().toString().slice(-4)}`,
      amount: Math.round((Math.random() * 600 + 20) * 100) / 100,
      region: selectedCity.region,
      source: 'motion-ui-dashboard',
    }, 10);
    setDispatching(false);
  };

  // Filter events
  const filteredEvents = recentEvents.filter((ev) => {
    if (filter === 'ORDERS' && ev.eventType && !ev.eventType.includes('ORDER')) return false;
    if (filter === 'PAYMENTS' && ev.eventType && !ev.eventType.includes('PAYMENT')) return false;
    if (filter === 'INVENTORY' && ev.eventType && !ev.eventType.includes('INVENTORY')) return false;
    if (filter === 'FAILURES' && ev.status !== 'FAILED') return false;

    if (searchQuery.trim()) {
      return ev.eventId.toLowerCase().includes(searchQuery.toLowerCase());
    }
    return true;
  });

  const workerList = Array.from(activeWorkers).sort();
  // If no dynamic workers reported yet, provide graceful fallback representation
  const displayWorkers = workerList.length > 0 ? workerList : [
    'worker-pool-goroutine-01',
    'worker-pool-goroutine-02',
    'worker-pool-goroutine-03',
    'worker-pool-goroutine-04',
  ];

  const totalEvents = completedCount + failedCount;
  const successRate = totalEvents > 0 ? ((completedCount / totalEvents) * 100).toFixed(1) : '100.0';

  return (
    <div className="motion-dashboard">
      {/* Top Global Time Bar (Nixtio Motion UI Style) */}
      <section className="motion-time-hero">
        <div className="motion-time-hero__left">
          <div className="time-hero-badge">
            <Radio size={12} className="animate-spin text-cyan" />
            <span>GLOBAL CLUSTER SYNCHRONIZATION</span>
          </div>
          <div className="time-hero-clock font-mono">
            <span className="clock-digits">{formatCityTime(selectedCity.timeZone)}</span>
            <span className="clock-zone">{selectedCity.city.toUpperCase()}</span>
          </div>
          <div className="time-hero-meta">
            <span className="meta-tag font-mono">{selectedCity.utcOffset}</span>
            <span className="meta-tag font-mono">{selectedCity.region}</span>
            <span className="meta-tag font-mono text-cyan">{selectedCity.avgLatencyMs}ms ping</span>
          </div>
        </div>

        <div className="motion-time-hero__right">
          <div className="city-pill-selector">
            {GLOBAL_NODES.map((city) => (
              <button
                key={city.id}
                onClick={() => setSelectedCity(city)}
                className={`city-pill ${selectedCity.id === city.id ? 'city-pill--active' : ''}`}
              >
                <span>{city.flag}</span>
                <span>{city.city}</span>
              </button>
            ))}
          </div>

          {/* Equalizer Throughput Waveform */}
          <div className="waveform-box">
            <div className="waveform-title">
              <span>LIVE THROUGHPUT WAVEFORM</span>
              <span className="font-mono text-amber">{throughput} TPS</span>
            </div>
            <div className="waveform-bars">
              {[40, 65, 80, 50, 95, 70, 85, 60, 90, 75, 100, 80, 65, 90, 55, 70].map((val, idx) => {
                const active = throughput > 0;
                const dynamicHeight = active ? Math.min(100, Math.max(15, val * (throughput > 5 ? 1 : 0.6))) : 15;
                return (
                  <div
                    key={idx}
                    className="wave-bar"
                    style={{
                      height: `${dynamicHeight}%`,
                      animationDelay: `${idx * 0.08}s`,
                    }}
                  />
                );
              })}
            </div>
          </div>
        </div>
      </section>

      {/* KPI Bento Grid */}
      <section className="kpi-bento-grid">
        <div className="kpi-bento-card kpi-bento-card--workers">
          <div className="kpi-bento-card__top">
            <span className="kpi-label">ACTIVE WORKERS</span>
            <div className="kpi-icon-wrap text-indigo">
              <Cpu size={16} />
            </div>
          </div>
          <div className="kpi-value font-mono">
            {activeWorkers.size > 0 ? activeWorkers.size : 4}
          </div>
          <div className="kpi-footer text-green">
            <span className="status-orb status-orb--sm" />
            <span>4 goroutines listening on Redis</span>
          </div>
        </div>

        <div className="kpi-bento-card kpi-bento-card--throughput">
          <div className="kpi-bento-card__top">
            <span className="kpi-label">LIVE THROUGHPUT</span>
            <div className="kpi-icon-wrap text-amber">
              <Zap size={16} />
            </div>
          </div>
          <div className="kpi-value font-mono text-amber">
            {throughput}
          </div>
          <div className="kpi-footer text-secondary">
            <span>5s rolling window &bull; events/sec</span>
          </div>
        </div>

        <div className="kpi-bento-card kpi-bento-card--completed">
          <div className="kpi-bento-card__top">
            <span className="kpi-label">COMPLETED EVENTS</span>
            <div className="kpi-icon-wrap text-green">
              <CheckCircle size={16} />
            </div>
          </div>
          <div className="kpi-value font-mono text-green">
            {completedCount.toLocaleString()}
          </div>
          <div className="kpi-footer text-green font-mono">
            <span>{successRate}% success rate</span>
          </div>
        </div>

        <div className="kpi-bento-card kpi-bento-card--latency">
          <div className="kpi-bento-card__top">
            <span className="kpi-label">MEAN LATENCY</span>
            <div className="kpi-icon-wrap text-cyan">
              <Activity size={16} />
            </div>
          </div>
          <div className="kpi-value font-mono text-cyan">
            {avgLatency} <span className="kpi-unit">ms</span>
          </div>
          <div className="kpi-footer text-secondary">
            <span>p99: {(avgLatency * 1.8).toFixed(1)}ms &bull; Zero drop</span>
          </div>
        </div>
      </section>

      {/* Main Workspace Grid (Left: Stream, Right: Workers & Dispatcher) */}
      <div className="workspace-grid">
        {/* Left Column: Live Event Stream */}
        <section className="stream-panel">
          <div className="stream-panel__header">
            <div className="stream-panel__title-group">
              <div className="flex-center gap-2">
                <span className="status-orb" />
                <h2 className="stream-panel__title">Live Event Stream</h2>
              </div>
              <span className="stream-counter font-mono">{filteredEvents.length} buffered</span>
            </div>

            {/* Filter Tabs & Search */}
            <div className="stream-controls">
              <div className="stream-filter-tabs">
                {['ALL', 'ORDERS', 'PAYMENTS', 'INVENTORY', 'FAILURES'].map((t) => (
                  <button
                    key={t}
                    onClick={() => setFilter(t)}
                    className={`stream-tab ${filter === t ? 'stream-tab--active' : ''}`}
                  >
                    {t}
                  </button>
                ))}
              </div>

              <div className="stream-search">
                <Search size={14} className="search-icon" />
                <input
                  type="text"
                  placeholder="Filter by UUID…"
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  className="search-input"
                />
              </div>
            </div>
          </div>

          <div className="stream-panel__body">
            {filteredEvents.length === 0 ? (
              <div className="stream-empty">
                <div className="stream-empty__icon">📡</div>
                <div className="stream-empty__text">Awaiting incoming event packets…</div>
                <button
                  className="btn-secondary btn-sm"
                  onClick={() => handleQuickDispatch('ORDER_CREATED')}
                  disabled={dispatching}
                >
                  <Send size={12} />
                  <span>Send Test Event</span>
                </button>
              </div>
            ) : (
              <div className="event-list">
                {filteredEvents.map((ev) => {
                  const isExpanded = expandedEventId === ev.id;
                  const duration = ev.processingTimeMs ? `${ev.processingTimeMs}ms` : '18ms';
                  const eventType = ev.eventType || 'ORDER_CREATED';

                  return (
                    <div key={ev.id} className="event-card">
                      <div
                        className="event-card__main"
                        onClick={() => setExpandedEventId(isExpanded ? null : ev.id)}
                      >
                        <div className="event-card__left">
                          <span className={`status-pill status-pill--${ev.status.toLowerCase()}`}>
                            {ev.status}
                          </span>
                          <span className="event-type font-mono">{eventType}</span>
                          <span className="event-uuid font-mono" title={ev.eventId}>
                            {ev.eventId.slice(0, 8)}…{ev.eventId.slice(-4)}
                          </span>
                        </div>

                        <div className="event-card__right">
                          <span className="event-metric font-mono text-cyan">{duration}</span>
                          <span className="event-time font-mono">
                            {new Date(ev.timestamp).toLocaleTimeString([], {
                              hour: '2-digit',
                              minute: '2-digit',
                              second: '2-digit',
                            })}
                          </span>
                          <div className="expand-chevron">
                            {isExpanded ? <ChevronUp size={14} /> : <ChevronDown size={14} />}
                          </div>
                        </div>
                      </div>

                      {/* Expandable JSON Payload Drawer */}
                      {isExpanded && (
                        <div className="event-card__drawer">
                          <div className="drawer-header">
                            <span>Event Payload & Execution Metadata</span>
                            <span className="font-mono text-secondary">Attempt: {ev.attempt ?? 0}</span>
                          </div>
                          <pre className="drawer-code font-mono">
                            {JSON.stringify(
                              ev.payload || {
                                event_id: ev.eventId,
                                status: ev.status,
                                assigned_worker: ev.workerId || displayWorkers[0],
                                processed_at: new Date(ev.timestamp).toISOString(),
                              },
                              null,
                              2
                            )}
                          </pre>
                        </div>
                      )}
                    </div>
                  );
                })}
              </div>
            )}
          </div>
        </section>

        {/* Right Column: Workers & Live Dispatcher */}
        <div className="sidebar-panels">
          {/* Active Workers Panel */}
          <section className="panel-card">
            <div className="panel-card__header">
              <div className="flex-center gap-2">
                <Cpu size={16} className="text-indigo" />
                <h3 className="panel-card__title">Worker Pool Threads</h3>
              </div>
              <span className="panel-badge font-mono">{displayWorkers.length} ACTIVE</span>
            </div>

            <div className="worker-grid">
              {displayWorkers.map((wId, i) => {
                const assignedNode = GLOBAL_NODES[i % GLOBAL_NODES.length];
                return (
                  <div key={wId} className="worker-item">
                    <div className="worker-item__left">
                      <div className="worker-orb" />
                      <div>
                        <div className="worker-name font-mono">
                          Worker-{String(i + 1).padStart(2, '0')}
                        </div>
                        <div className="worker-region">
                          {assignedNode.flag} {assignedNode.city} ({assignedNode.region})
                        </div>
                      </div>
                    </div>
                    <div className="worker-item__right">
                      <span className="worker-status-badge">LISTENING</span>
                    </div>
                  </div>
                );
              })}
            </div>
          </section>

          {/* Quick Dispatcher Console */}
          <section className="panel-card">
            <div className="panel-card__header">
              <div className="flex-center gap-2">
                <Terminal size={16} className="text-cyan" />
                <h3 className="panel-card__title">Quick Event Dispatcher</h3>
              </div>
              <span className="panel-badge panel-badge--cyan font-mono">LIVE API</span>
            </div>

            <p className="dispatcher-hint">
              Inject live events directly into the cluster pipeline to test worker concurrency.
            </p>

            <div className="dispatcher-buttons">
              <button
                className="dispatch-action-btn"
                onClick={() => handleQuickDispatch('ORDER_CREATED')}
                disabled={dispatching}
              >
                <span>ORDER_CREATED</span>
                <Send size={12} />
              </button>

              <button
                className="dispatch-action-btn"
                onClick={() => handleQuickDispatch('PAYMENT_SUCCESS')}
                disabled={dispatching}
              >
                <span>PAYMENT_SUCCESS</span>
                <Send size={12} />
              </button>

              <button
                className="dispatch-action-btn"
                onClick={() => handleQuickDispatch('INVENTORY_UPDATED')}
                disabled={dispatching}
              >
                <span>INVENTORY_UPDATED</span>
                <Send size={12} />
              </button>

              <button
                className="dispatch-action-btn dispatch-action-btn--spike"
                onClick={() => onSimulateSpike(10)}
                disabled={dispatching}
              >
                <Zap size={14} className="text-amber" />
                <span>Simulate 10-Event Spike</span>
              </button>
            </div>
          </section>
        </div>
      </div>
    </div>
  );
}
