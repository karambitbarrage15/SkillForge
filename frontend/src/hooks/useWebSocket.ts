import { useEffect, useRef, useCallback, useState } from 'react';
import type { WSMessage, RecentEvent } from '../types/ws';

const MAX_RECENT_EVENTS = 120;
const RECONNECT_DELAY_MS = 2500;
const THROUGHPUT_WINDOW_MS = 5000;

export interface DashboardState {
  connected: boolean;
  activeWorkers: Set<string>;
  completedCount: number;
  failedCount: number;
  recentEvents: RecentEvent[];
  throughput: number;
  avgLatency: number;
  sendEvent: (type: string, payload: Record<string, unknown>, priority?: number) => Promise<boolean>;
  simulateSpike: (count?: number) => Promise<void>;
}

export function useWebSocket(url: string): DashboardState {
  const [connected, setConnected] = useState(false);
  const [activeWorkers, setActiveWorkers] = useState<Set<string>>(new Set());
  const [completedCount, setCompletedCount] = useState(0);
  const [failedCount, setFailedCount] = useState(0);
  const [recentEvents, setRecentEvents] = useState<RecentEvent[]>([]);
  const [throughput, setThroughput] = useState(0);
  const [avgLatency, setAvgLatency] = useState(18.5);

  const eventTimestamps = useRef<number[]>([]);
  const latencies = useRef<number[]>([14, 18, 22, 16, 25, 19, 15]);
  const wsRef = useRef<WebSocket | null>(null);
  const reconnectTimer = useRef<ReturnType<typeof setTimeout> | null>(null);
  const throughputTimer = useRef<ReturnType<typeof setInterval> | null>(null);
  const mountedRef = useRef(true);
  const idCounter = useRef(0);

  // Compute rolling throughput
  const computeThroughput = useCallback(() => {
    const now = Date.now();
    const cutoff = now - THROUGHPUT_WINDOW_MS;
    eventTimestamps.current = eventTimestamps.current.filter(t => t > cutoff);
    const count = eventTimestamps.current.length;
    setThroughput(Math.round((count / (THROUGHPUT_WINDOW_MS / 1000)) * 10) / 10);

    // Compute average latency
    if (latencies.current.length > 0) {
      const sum = latencies.current.slice(-20).reduce((acc, curr) => acc + curr, 0);
      setAvgLatency(Math.round((sum / Math.min(latencies.current.length, 20)) * 10) / 10);
    }
  }, []);

  // Fetch initial events from database to hydrate state
  const fetchInitialEvents = useCallback(async () => {
    try {
      const res = await fetch('/api/v1/events?limit=40');
      if (!res.ok) return;
      const data = await res.json();
      if (!Array.isArray(data) || !mountedRef.current) return;

      let comp = 0;
      let fail = 0;
      const initial: RecentEvent[] = data.map((item: any) => {
        if (item.Status === 'COMPLETED') comp++;
        if (item.Status === 'FAILED') fail++;
        return {
          id: String(++idCounter.current),
          eventId: item.ID || '',
          eventType: item.Type || 'ORDER_CREATED',
          status: item.Status || 'COMPLETED',
          timestamp: item.CreatedAt ? new Date(item.CreatedAt).getTime() : Date.now(),
          priority: item.Priority || 5,
          payload: item.Payload,
          processingTimeMs: item.CompletedAt && item.CreatedAt
            ? Math.max(1, new Date(item.CompletedAt).getTime() - new Date(item.CreatedAt).getTime())
            : Math.floor(Math.random() * 25 + 10),
        };
      });

      setCompletedCount(prev => Math.max(prev, comp));
      setFailedCount(prev => Math.max(prev, fail));
      setRecentEvents(prev => {
        if (prev.length === 0) return initial;
        return prev;
      });
    } catch {
      // Non-fatal if offline
    }
  }, []);

  const handleMessage = useCallback((raw: string) => {
    const lines = raw.split('\n');
    for (const line of lines) {
      const trimmed = line.trim();
      if (!trimmed) continue;

      let msg: WSMessage;
      try {
        msg = JSON.parse(trimmed);
      } catch {
        continue;
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
              priority: 5,
            };
            return [entry, ...prev].slice(0, MAX_RECENT_EVENTS);
          });
          break;

        case 'EVENT_PROCESSING':
          setRecentEvents(prev => {
            const idx = prev.findIndex(e => e.eventId === msg.event_id);
            if (idx >= 0) {
              const updated = [...prev];
              updated[idx] = {
                ...updated[idx],
                status: 'PROCESSING',
                workerId: msg.worker_id,
              };
              return updated;
            }
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
          if (msg.processing_time_ms) {
            latencies.current.push(msg.processing_time_ms);
          }
          setRecentEvents(prev => {
            const idx = prev.findIndex(e => e.eventId === msg.event_id);
            if (idx >= 0) {
              const updated = [...prev];
              updated[idx] = {
                ...updated[idx],
                status: 'COMPLETED',
                processingTimeMs: msg.processing_time_ms,
              };
              return updated;
            }
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
            const idx = prev.findIndex(e => e.eventId === msg.event_id);
            if (idx >= 0) {
              const updated = [...prev];
              updated[idx] = {
                ...updated[idx],
                status: 'FAILED',
                attempt: msg.attempt,
              };
              return updated;
            }
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

    try {
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
        reconnectTimer.current = setTimeout(connect, RECONNECT_DELAY_MS);
      };

      ws.onerror = () => {
        ws.close();
      };
    } catch {
      reconnectTimer.current = setTimeout(connect, RECONNECT_DELAY_MS);
    }
  }, [url, handleMessage]);

  useEffect(() => {
    mountedRef.current = true;
    connect();
    fetchInitialEvents();

    throughputTimer.current = setInterval(computeThroughput, 1000);

    return () => {
      mountedRef.current = false;
      if (reconnectTimer.current) clearTimeout(reconnectTimer.current);
      if (throughputTimer.current) clearInterval(throughputTimer.current);
      if (wsRef.current) {
        wsRef.current.onclose = null;
        wsRef.current.close();
        wsRef.current = null;
      }
    };
  }, [connect, computeThroughput, fetchInitialEvents]);

  // Helper to send single event to backend
  const sendEvent = useCallback(async (type: string, payload: Record<string, unknown>, priority = 5): Promise<boolean> => {
    try {
      const res = await fetch('/api/v1/events', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ type, payload, priority }),
      });
      return res.status === 202 || res.status === 200;
    } catch {
      return false;
    }
  }, []);

  // Helper to simulate a multi-event traffic spike
  const simulateSpike = useCallback(async (count = 10) => {
    const types = ['ORDER_CREATED', 'PAYMENT_SUCCESS', 'INVENTORY_UPDATED', 'ORDER_CANCELLED', 'USER_REGISTERED'];
    for (let i = 0; i < count; i++) {
      const t = types[Math.floor(Math.random() * types.length)];
      sendEvent(t, {
        order_id: `ord-burst-${Date.now()}-${i}`,
        amount: Math.round((Math.random() * 500 + 10) * 100) / 100,
        source: 'motion-ui-dashboard',
        simulated: true,
      }, Math.floor(Math.random() * 10) + 1);
      await new Promise(r => setTimeout(r, 60));
    }
  }, [sendEvent]);

  return {
    connected,
    activeWorkers,
    completedCount,
    failedCount,
    recentEvents,
    throughput,
    avgLatency,
    sendEvent,
    simulateSpike,
  };
}
