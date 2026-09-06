import type { RecentEvent } from '../types/ws';

interface RecentEventsProps {
  events: RecentEvent[];
}

function statusClass(status: RecentEvent['status']): string {
  switch (status) {
    case 'RECEIVED':
      return 'event-row__status--received';
    case 'PROCESSING':
      return 'event-row__status--processing';
    case 'COMPLETED':
      return 'event-row__status--completed';
    case 'FAILED':
      return 'event-row__status--failed';
  }
}

function formatTime(ts: number): string {
  const d = new Date(ts);
  return d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit' });
}

function formatExtra(event: RecentEvent): string {
  if (event.processingTimeMs !== undefined) {
    return `${event.processingTimeMs}ms`;
  }
  if (event.workerId) {
    return event.workerId;
  }
  if (event.attempt !== undefined) {
    return `attempt ${event.attempt}`;
  }
  return '';
}

export function RecentEvents({ events }: RecentEventsProps) {
  return (
    <div className="panel">
      <div className="panel__header">
        <span className="panel__title">Recent Events</span>
        <span className="panel__badge">{events.length}</span>
      </div>
      <div className="panel__body">
        {events.length === 0 ? (
          <div className="empty-state">
            <div className="empty-state__icon">📡</div>
            <div>Waiting for events…</div>
          </div>
        ) : (
          events.map(ev => (
            <div key={ev.id} className="event-row">
              <span className="event-row__id" title={ev.eventId}>
                {ev.eventId.slice(0, 8)}…
              </span>
              <span className={`event-row__status ${statusClass(ev.status)}`}>
                {ev.status}
              </span>
              <span className="event-row__time">
                {formatExtra(ev) ? `${formatExtra(ev)} · ` : ''}{formatTime(ev.timestamp)}
              </span>
            </div>
          ))
        )}
      </div>
    </div>
  );
}
