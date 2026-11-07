import { Clock, LayoutDashboard, Globe2, Zap } from 'lucide-react';
import { useState, useEffect } from 'react';

interface HeaderProps {
  connected: boolean;
  activeView: 'landing' | 'dashboard';
  onChangeView: (view: 'landing' | 'dashboard') => void;
  onSimulateSpike: () => void;
}

export function Header({ connected, activeView, onChangeView, onSimulateSpike }: HeaderProps) {
  const [utcTime, setUtcTime] = useState('');

  useEffect(() => {
    const updateTime = () => {
      const now = new Date();
      setUtcTime(now.toISOString().substring(11, 19));
    };
    updateTime();
    const interval = setInterval(updateTime, 1000);
    return () => clearInterval(interval);
  }, []);

  return (
    <header className="motion-header">
      <div className="motion-header__left">
        <div className="motion-brand" onClick={() => onChangeView('landing')}>
          <div className="motion-brand__logo">SF</div>
          <div>
            <div className="motion-brand__name">StreamForge</div>
            <div className="motion-brand__sub font-mono">GLOBAL TIME &bull; MOTION ENGINE</div>
          </div>
        </div>

        {/* View Switcher Tabs */}
        <nav className="motion-nav">
          <button
            className={`motion-nav__tab ${activeView === 'landing' ? 'motion-nav__tab--active' : ''}`}
            onClick={() => onChangeView('landing')}
          >
            <Globe2 size={14} />
            <span>Overview & Mesh</span>
          </button>

          <button
            className={`motion-nav__tab ${activeView === 'dashboard' ? 'motion-nav__tab--active' : ''}`}
            onClick={() => onChangeView('dashboard')}
          >
            <LayoutDashboard size={14} />
            <span>Live Console</span>
          </button>
        </nav>
      </div>

      <div className="motion-header__right">
        {/* Real-time UTC clock */}
        <div className="header-utc font-mono">
          <Clock size={13} className="text-secondary" />
          <span className="text-muted">UTC</span>
          <span className="text-cyan font-semibold">{utcTime}Z</span>
        </div>

        {/* Quick Simulator Trigger */}
        <button
          className="header-trigger-btn"
          onClick={onSimulateSpike}
          title="Inject quick event burst"
        >
          <Zap size={13} className="text-amber" />
          <span>Quick Burst</span>
        </button>

        {/* Status Pill */}
        <div className="header-status-pill">
          <span
            className={`status-radar ${
              connected ? 'status-radar--online' : 'status-radar--offline'
            }`}
          />
          <span className={`status-text ${connected ? 'text-green' : 'text-red'} font-mono`}>
            {connected ? 'WS LIVE' : 'DISCONNECTED'}
          </span>
        </div>
      </div>
    </header>
  );
}
