import { useEffect, useRef, useCallback, useState } from 'react';
import type { WSMessage, RecentEvent } from '../types/ws';

const MAX_RECENT_EVENTS = 100;
const RECONNECT_DELAY_MS = 3000;
const THROUGHPUT_WINDOW_MS = 5000;

export interface DashboardState {
  connected: boolean;
  activeWorkers: Set<string>;
  completedCount: number;
  failedCount: number;
  recentEvents: RecentEvent[];
  throughput: number;
}

/**
 * useWebSocket connects to the Phase 12 WebSocket endpoint and maintains
 * all dashboard state derived from incoming messages.
 *
 * IMPORTANT: The backend's WritePump batches multiple JSON messages into
 * a single WebSocket frame separated by '\n'. We must split on '\n' and
 * parse each segment independently.
 */
export function useWebSocket(url: string): DashboardState {
  const [connected, setConnected] = useState(false);
  const [activeWorkers, setActiveWorkers] = useState<Set<string>>(new Set());
  const [completedCount, setCompletedCount] = useState(0);
  const [failedCount, setFailedCount] = useState(0);
  const [recentEvents, setRecentEvents] = useState<RecentEvent[]>([]);
  const [throughput, setThroughput] = useState(0);

  // Track event timestamps for throughput calculation
  const eventTimestamps = useRef<number[]>([]);
  const wsRef = useRef<WebSocket | null>(null);
  const reconnectTimer = useRef<ReturnType<typeof setTimeout> | null>(null);
  const throughputTimer = useRef<ReturnType<typeof setInterval> | null>(null);
  const mountedRef = useRef(true);
  // Use a monotonic counter for unique RecentEvent IDs
  const idCounter = useRef(0);

  const computeThroughput = useCallback(() => {
    const now = Date.now();
    const cutoff = now - THROUGHPUT_WINDOW_MS;
    eventTimestamps.current = eventTimestamps.current.filter(t => t > cutoff);
    const count = eventTimestamps.current.length;
    // events per second over the window
    setThroughput(Math.round((count / (THROUGHPUT_WINDOW_MS / 1000)) * 10) / 10);
  }, []);

  const handleMessage = useCallback((raw: string) => {
    // Backend may batch multiple JSON messages in one frame separated by '\n'
    const lines = raw.split('\n');
    for (const line of lines) {
      const trimmed = line.trim();
      if (!trimmed) continue;

      let msg: WSMessage;
      try {
        msg = JSON.parse(trimmed);
      } catch {
        continue; // skip malformed
      }

      const now = Date.now();

      switch (msg.type) {
        case 'EVENT_RECEIVED':
          eventTimestamps.current.push(now);
          setRecentEvents(prev => {
            const entry: RecentEvent = {
              id: String(++idCounter.current),
              eventId: msg.event_id,
              status: 'RECEIVED',
              timestamp: now,
            };
            return [entry, ...prev].slice(0, MAX_RECENT_EVENTS);
          });
          break;

        case 'EVENT_PROCESSING':
          setRecentEvents(prev => {
            const entry: RecentEvent = {
              id: String(++idCounter.current),
              eventId: msg.event_id,
              status: 'PROCESSING',
              timestamp: now,
              workerId: msg.worker_id,
            };
            return [entry, ...prev].slice(0, MAX_RECENT_EVENTS);
          });
          break;

        case 'EVENT_COMPLETED':
          setCompletedCount(c => c + 1);
          setRecentEvents(prev => {
            const entry: RecentEvent = {
              id: String(++idCounter.current),
              eventId: msg.event_id,
              status: 'COMPLETED',
              timestamp: now,
              processingTimeMs: msg.processing_time_ms,
            };
            return [entry, ...prev].slice(0, MAX_RECENT_EVENTS);
          });
          break;

        case 'EVENT_FAILED':
          setFailedCount(c => c + 1);
          setRecentEvents(prev => {
            const entry: RecentEvent = {
              id: String(++idCounter.current),
              eventId: msg.event_id,
              status: 'FAILED',
              timestamp: now,
              attempt: msg.attempt,
            };
            return [entry, ...prev].slice(0, MAX_RECENT_EVENTS);
          });
          break;

        case 'WORKER_ONLINE':
          setActiveWorkers(prev => {
            const next = new Set(prev);
            next.add(msg.worker_id);
            return next;
          });
          break;

        case 'WORKER_OFFLINE':
          setActiveWorkers(prev => {
            const next = new Set(prev);
            next.delete(msg.worker_id);
            return next;
          });
          break;
      }
    }
  }, []);

  const connect = useCallback(() => {
    if (!mountedRef.current) return;
    if (wsRef.current && wsRef.current.readyState === WebSocket.OPEN) return;

    const ws = new WebSocket(url);
    wsRef.current = ws;

    ws.onopen = () => {
      if (!mountedRef.current) return;
      setConnected(true);
    };

    ws.onmessage = (ev) => {
      if (!mountedRef.current) return;
      if (typeof ev.data === 'string') {
        handleMessage(ev.data);
      }
    };

    ws.onclose = () => {
      if (!mountedRef.current) return;
      setConnected(false);
      wsRef.current = null;
      // Attempt reconnection
      reconnectTimer.current = setTimeout(connect, RECONNECT_DELAY_MS);
    };

    ws.onerror = () => {
      // onclose will fire after onerror, which handles reconnection
      ws.close();
    };
  }, [url, handleMessage]);

  useEffect(() => {
    mountedRef.current = true;
    connect();

    // Start throughput calculation interval
    throughputTimer.current = setInterval(computeThroughput, 1000);

    return () => {
      mountedRef.current = false;
      if (reconnectTimer.current) {
        clearTimeout(reconnectTimer.current);
        reconnectTimer.current = null;
      }
      if (throughputTimer.current) {
        clearInterval(throughputTimer.current);
        throughputTimer.current = null;
      }
      if (wsRef.current) {
        wsRef.current.onclose = null; // prevent reconnection on unmount
        wsRef.current.close();
        wsRef.current = null;
      }
    };
  }, [connect, computeThroughput]);

  return {
    connected,
    activeWorkers,
    completedCount,
    failedCount,
    recentEvents,
    throughput,
  };
}
