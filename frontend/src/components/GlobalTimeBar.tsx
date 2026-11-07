import { useState, useEffect } from 'react';
import { Sun, Moon, Clock, Globe2, Radio } from 'lucide-react';
import type { GlobalTimeZone } from '../types/ws';

export const GLOBAL_NODES: GlobalTimeZone[] = [
  {
    id: 'nyc',
    city: 'New York',
    country: 'USA',
    region: 'us-east-1',
    timeZone: 'America/New_York',
    utcOffset: 'UTC-5',
    coordinates: { x: 28, y: 35 },
    flag: '🇺🇸',
    activeWorkersCount: 4,
    avgLatencyMs: 14,
  },
  {
    id: 'sfo',
    city: 'San Francisco',
    country: 'USA',
    region: 'us-west-1',
    timeZone: 'America/Los_Angeles',
    utcOffset: 'UTC-8',
    coordinates: { x: 18, y: 38 },
    flag: '🇺🇸',
    activeWorkersCount: 2,
    avgLatencyMs: 22,
  },
  {
    id: 'lon',
    city: 'London',
    country: 'UK',
    region: 'eu-west-1',
    timeZone: 'Europe/London',
    utcOffset: 'UTC+0',
    coordinates: { x: 48, y: 28 },
    flag: '🇬🇧',
    activeWorkersCount: 3,
    avgLatencyMs: 9,
  },
  {
    id: 'fra',
    city: 'Frankfurt',
    country: 'Germany',
    region: 'eu-central-1',
    timeZone: 'Europe/Berlin',
    utcOffset: 'UTC+1',
    coordinates: { x: 52, y: 30 },
    flag: '🇩🇪',
    activeWorkersCount: 4,
    avgLatencyMs: 11,
  },
  {
    id: 'tyo',
    city: 'Tokyo',
    country: 'Japan',
    region: 'ap-northeast-1',
    timeZone: 'Asia/Tokyo',
    utcOffset: 'UTC+9',
    coordinates: { x: 84, y: 38 },
    flag: '🇯🇵',
    activeWorkersCount: 3,
    avgLatencyMs: 29,
  },
  {
    id: 'sin',
    city: 'Singapore',
    country: 'Singapore',
    region: 'ap-southeast-1',
    timeZone: 'Asia/Singapore',
    utcOffset: 'UTC+8',
    coordinates: { x: 76, y: 55 },
    flag: '🇸🇬',
    activeWorkersCount: 2,
    avgLatencyMs: 31,
  },
];

interface GlobalTimeBarProps {
  activeNodeId?: string;
  onSelectNode?: (node: GlobalTimeZone) => void;
}

export function GlobalTimeBar({ activeNodeId = 'nyc', onSelectNode }: GlobalTimeBarProps) {
  const [now, setNow] = useState(new Date());

  useEffect(() => {
    const timer = setInterval(() => setNow(new Date()), 1000);
    return () => clearInterval(timer);
  }, []);

  const formatCityTime = (timeZone: string) => {
    const formatter = new Intl.DateTimeFormat('en-US', {
      timeZone,
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit',
      hour12: false,
    });
    return formatter.format(now);
  };

  const getDayStatus = (timeZone: string) => {
    const hour = parseInt(
      new Intl.DateTimeFormat('en-US', {
        timeZone,
        hour: 'numeric',
        hour12: false,
      }).format(now),
      10
    );
    const isDay = hour >= 6 && hour < 18;
    return { isDay, hour };
  };

  const utcTime = now.toISOString().substring(11, 19);

  return (
    <div className="global-time-strip">
      <div className="global-time-strip__header">
        <div className="global-time-strip__title">
          <Globe2 className="icon-pulse text-cyan" size={16} />
          <span>Synchronized Global Nodes</span>
          <span className="live-pill">
            <Radio size={10} className="animate-spin text-cyan" />
            LIVE ATOMIC SYNC
          </span>
        </div>
        <div className="global-time-strip__utc">
          <Clock size={13} className="text-secondary" />
          <span className="utc-label">UTC REFERENCE:</span>
          <span className="utc-value font-mono">{utcTime}Z</span>
        </div>
      </div>

      <div className="global-time-cards">
        {GLOBAL_NODES.map((node) => {
          const timeStr = formatCityTime(node.timeZone);
          const { isDay } = getDayStatus(node.timeZone);
          const isSelected = node.id === activeNodeId;

          return (
            <div
              key={node.id}
              onClick={() => onSelectNode?.(node)}
              className={`timezone-card ${isSelected ? 'timezone-card--active' : ''}`}
            >
              <div className="timezone-card__top">
                <div className="timezone-card__location">
                  <span className="timezone-card__flag">{node.flag}</span>
                  <div>
                    <div className="timezone-card__city">{node.city}</div>
                    <div className="timezone-card__region">{node.region}</div>
                  </div>
                </div>
                <div className={`circadian-badge ${isDay ? 'circadian-badge--day' : 'circadian-badge--night'}`}>
                  {isDay ? <Sun size={12} /> : <Moon size={12} />}
                  <span>{isDay ? 'DAY' : 'NIGHT'}</span>
                </div>
              </div>

              <div className="timezone-card__time font-mono">
                {timeStr}
              </div>

              <div className="timezone-card__footer">
                <span className="timezone-card__offset">{node.utcOffset}</span>
                <span className="timezone-card__latency font-mono">
                  {node.avgLatencyMs}ms ping
                </span>
                <span className="timezone-card__workers">
                  {node.activeWorkersCount} workers
                </span>
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
}

