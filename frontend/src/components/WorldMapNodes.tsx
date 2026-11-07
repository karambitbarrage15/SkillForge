import { useState, useEffect } from 'react';
import { GLOBAL_NODES } from './GlobalTimeBar';
import type { GlobalTimeZone } from '../types/ws';
import { Activity, ShieldCheck, Zap } from 'lucide-react';

interface WorldMapNodesProps {
  activeNodeId?: string;
  onSelectNode?: (node: GlobalTimeZone) => void;
  eventCount?: number;
}

export function WorldMapNodes({ activeNodeId = 'nyc', onSelectNode, eventCount = 0 }: WorldMapNodesProps) {
  const [pulseIndex, setPulseIndex] = useState(0);

  useEffect(() => {
    const timer = setInterval(() => {
      setPulseIndex((prev) => (prev + 1) % GLOBAL_NODES.length);
    }, 2400);
    return () => clearInterval(timer);
  }, []);

  return (
    <div className="world-map-container">
      <div className="world-map-header">
        <div className="world-map-header__title">
          <Activity size={16} className="text-cyan animate-pulse" />
          <span>Global Event Distribution Mesh</span>
        </div>
        <div className="world-map-header__meta">
          <span className="mesh-stat">
            <Zap size={12} className="text-amber" />
            <strong className="font-mono">{eventCount}</strong> packets routed
          </span>
          <span className="mesh-stat">
            <ShieldCheck size={12} className="text-green" />
            Zero-drop backpressure
          </span>
        </div>
      </div>

      <div className="world-map-canvas">
        <svg viewBox="0 0 1000 480" className="world-svg" preserveAspectRatio="xMidYMid meet">
          <defs>
            {/* Grid pattern for high-tech matrix */}
            <pattern id="dot-grid" x="0" y="0" width="20" height="20" patternUnits="userSpaceOnUse">
              <circle cx="2" cy="2" r="1.2" fill="rgba(255, 255, 255, 0.08)" />
            </pattern>

            {/* Glowing gradient for connection routes */}
            <linearGradient id="route-gradient-1" x1="0%" y1="0%" x2="100%" y2="100%">
              <stop offset="0%" stopColor="#00f2fe" stopOpacity="0.8" />
              <stop offset="50%" stopColor="#818cf8" stopOpacity="0.6" />
              <stop offset="100%" stopColor="#10b981" stopOpacity="0.8" />
            </linearGradient>

            <filter id="glow-cyan" x="-20%" y="-20%" width="140%" height="140%">
              <feGaussianBlur stdDeviation="4" result="blur" />
              <feComposite in="SourceGraphic" in2="blur" operator="over" />
            </filter>
          </defs>

          {/* Background dot matrix */}
          <rect width="1000" height="480" fill="url(#dot-grid)" />

          {/* Stylized continent landmass silhouettes */}
          <g className="continent-paths" fill="rgba(255, 255, 255, 0.03)" stroke="rgba(255, 255, 255, 0.06)" strokeWidth="1">
            {/* North America */}
            <path d="M 120,90 Q 200,80 290,110 Q 330,170 280,240 Q 220,260 170,220 Q 110,170 120,90 Z" />
            {/* South America */}
            <path d="M 270,250 Q 340,280 320,380 Q 270,440 250,380 Q 230,310 270,250 Z" />
            {/* Europe */}
            <path d="M 460,90 Q 550,80 580,140 Q 540,200 460,190 Q 430,140 460,90 Z" />
            {/* Africa */}
            <path d="M 460,200 Q 560,200 580,310 Q 530,420 480,410 Q 430,320 460,200 Z" />
            {/* Asia */}
            <path d="M 580,80 Q 820,70 880,180 Q 820,300 680,270 Q 590,220 580,80 Z" />
            {/* Australia */}
            <path d="M 780,320 Q 880,310 880,410 Q 800,430 760,380 Z" />
          </g>

          {/* Animated Connecting Arcs between global hubs */}
          <g className="route-arcs">
            {/* SFO to NYC */}
            <path d="M 180,182 Q 230,140 280,168" fill="none" stroke="url(#route-gradient-1)" strokeWidth="1.5" strokeDasharray="4 4" className="flow-dash" />
            {/* NYC to London */}
            <path d="M 280,168 Q 380,100 480,134" fill="none" stroke="url(#route-gradient-1)" strokeWidth="2" strokeDasharray="6 4" className="flow-dash" />
            {/* London to Frankfurt */}
            <path d="M 480,134 Q 500,130 520,144" fill="none" stroke="url(#route-gradient-1)" strokeWidth="1.5" />
            {/* Frankfurt to Singapore */}
            <path d="M 520,144 Q 640,180 760,264" fill="none" stroke="url(#route-gradient-1)" strokeWidth="1.8" strokeDasharray="6 4" className="flow-dash" />
            {/* Singapore to Tokyo */}
            <path d="M 760,264 Q 820,220 840,182" fill="none" stroke="url(#route-gradient-1)" strokeWidth="1.5" strokeDasharray="4 4" className="flow-dash" />
            {/* Tokyo to SFO (Transpacific arc) */}
            <path d="M 840,182 Q 950,110 990,130" fill="none" stroke="url(#route-gradient-1)" strokeWidth="1.2" strokeDasharray="3 3" />
            <path d="M 10,130 Q 80,140 180,182" fill="none" stroke="url(#route-gradient-1)" strokeWidth="1.2" strokeDasharray="3 3" />
          </g>

          {/* Global Nodes */}
          {GLOBAL_NODES.map((node, i) => {
            const posX = node.coordinates.x * 10;
            const posY = node.coordinates.y * 4.8;
            const isSelected = node.id === activeNodeId;
            const isPulsing = pulseIndex === i;

            return (
              <g
                key={node.id}
                className="map-node-group"
                onClick={() => onSelectNode?.(node)}
                style={{ cursor: 'pointer' }}
              >
                {/* Ripple ring */}
                {(isSelected || isPulsing) && (
                  <circle
                    cx={posX}
                    cy={posY}
                    r="18"
                    fill="none"
                    stroke={isSelected ? '#00f2fe' : '#10b981'}
                    strokeWidth="1.5"
                    className="ping-circle"
                  />
                )}

                {/* Outer Glow */}
                <circle
                  cx={posX}
                  cy={posY}
                  r={isSelected ? '9' : '6'}
                  fill={isSelected ? '#00f2fe' : '#818cf8'}
                  filter="url(#glow-cyan)"
                  opacity="0.85"
                />

                {/* Center Core Dot */}
                <circle
                  cx={posX}
                  cy={posY}
                  r="3.5"
                  fill="#ffffff"
                />

                {/* Node Label Tag */}
                <g transform={`translate(${posX + 10}, ${posY - 10})`}>
                  <rect
                    x="0"
                    y="-12"
                    width={node.city.length * 8 + 44}
                    height="20"
                    rx="5"
                    fill="rgba(12, 16, 28, 0.88)"
                    stroke={isSelected ? '#00f2fe' : 'rgba(255, 255, 255, 0.12)'}
                    strokeWidth="1"
                  />
                  <text
                    x="8"
                    y="2"
                    fill={isSelected ? '#00f2fe' : '#f1f5f9'}
                    fontSize="10"
                    fontWeight="600"
                    fontFamily="Inter, sans-serif"
                  >
                    {node.city}
                  </text>
                  <text
                    x={node.city.length * 8 + 14}
                    y="2"
                    fill="#10b981"
                    fontSize="8.5"
                    fontWeight="700"
                    fontFamily="JetBrains Mono, monospace"
                  >
                    {node.avgLatencyMs}ms
                  </text>
                </g>
              </g>
            );
          })}
        </svg>

        {/* Floating Active Node Card */}
        <div className="active-node-hud">
          {(() => {
            const current = GLOBAL_NODES.find((n) => n.id === activeNodeId) || GLOBAL_NODES[0];
            return (
              <div className="hud-card">
                <div className="hud-card__badge">FOCUSED REGION</div>
                <div className="hud-card__main">
                  <span className="hud-card__flag">{current.flag}</span>
                  <div>
                    <span className="hud-card__city">{current.city}</span>
                    <span className="hud-card__region font-mono">{current.region}</span>
                  </div>
                </div>
                <div className="hud-card__grid">
                  <div>
                    <span className="hud-label">TIMEZONE</span>
                    <span className="hud-val font-mono">{current.utcOffset}</span>
                  </div>
                  <div>
                    <span className="hud-label">AVG LATENCY</span>
                    <span className="hud-val font-mono text-cyan">{current.avgLatencyMs} ms</span>
                  </div>
                  <div>
                    <span className="hud-label">WORKERS</span>
                    <span className="hud-val font-mono text-green">{current.activeWorkersCount} ACTIVE</span>
                  </div>
                </div>
              </div>
            );
          })()}
        </div>
      </div>
    </div>
  );
}
