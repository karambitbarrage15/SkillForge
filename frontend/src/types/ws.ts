/**
 * WebSocket message types matching the Phase 12 backend contract.
 * See: backend/internal/websocket/types.go
 */

export type MessageType =
  | 'EVENT_RECEIVED'
  | 'EVENT_PROCESSING'
  | 'EVENT_COMPLETED'
  | 'EVENT_FAILED'
  | 'WORKER_ONLINE'
  | 'WORKER_OFFLINE';

export interface BaseMessage {
  type: MessageType;
}

export interface EventReceivedMsg extends BaseMessage {
  type: 'EVENT_RECEIVED';
  event_id: string;
}

export interface EventProcessingMsg extends BaseMessage {
  type: 'EVENT_PROCESSING';
  event_id: string;
  worker_id: string;
}

export interface EventCompletedMsg extends BaseMessage {
  type: 'EVENT_COMPLETED';
  event_id: string;
  processing_time_ms: number;
}

export interface EventFailedMsg extends BaseMessage {
  type: 'EVENT_FAILED';
  event_id: string;
  attempt: number;
}

export interface WorkerOnlineMsg extends BaseMessage {
  type: 'WORKER_ONLINE';
  worker_id: string;
}

export interface WorkerOfflineMsg extends BaseMessage {
  type: 'WORKER_OFFLINE';
  worker_id: string;
}

export type WSMessage =
  | EventReceivedMsg
  | EventProcessingMsg
  | EventCompletedMsg
  | EventFailedMsg
  | WorkerOnlineMsg
  | WorkerOfflineMsg;

/**
 * Represents a recent event tracked in the dashboard feed.
 */
export interface RecentEvent {
  id: string;
  eventId: string;
  eventType?: string;
  status: 'RECEIVED' | 'PROCESSING' | 'COMPLETED' | 'FAILED';
  timestamp: number;
  processingTimeMs?: number;
  workerId?: string;
  attempt?: number;
  priority?: number;
  payload?: Record<string, unknown>;
}

export interface GlobalTimeZone {
  id: string;
  city: string;
  country: string;
  region: string;
  timeZone: string;
  utcOffset: string;
  coordinates: { x: number; y: number }; // Percentage for world map
  flag: string;
  activeWorkersCount: number;
  avgLatencyMs: number;
}
