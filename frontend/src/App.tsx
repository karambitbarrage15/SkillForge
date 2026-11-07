import { useState } from 'react';
import './index.css';
import { useWebSocket } from './hooks/useWebSocket';
import { Header } from './components/Header';
import { LandingView } from './components/LandingView';
import { DashboardView } from './components/DashboardView';

const WS_URL =
  import.meta.env.VITE_WS_URL ||
  (typeof window !== 'undefined' && window.location.host
    ? (window.location.protocol === 'https:' ? 'wss:' : 'ws:') + '//' + window.location.host + '/ws'
    : 'ws://localhost:8080/ws');

function App() {
  const [activeView, setActiveView] = useState<'landing' | 'dashboard'>('landing');

  const {
    connected,
    activeWorkers,
    completedCount,
    failedCount,
    recentEvents,
    throughput,
    avgLatency,
    sendEvent,
    simulateSpike,
  } = useWebSocket(WS_URL);

  return (
    <div className="streamforge-app">
      <Header
        connected={connected}
        activeView={activeView}
        onChangeView={setActiveView}
        onSimulateSpike={() => simulateSpike(8)}
      />

      <main className="main-content">
        {activeView === 'landing' ? (
          <LandingView
            onOpenDashboard={() => setActiveView('dashboard')}
            onSendEvent={sendEvent}
            onSimulateSpike={simulateSpike}
            eventCount={recentEvents.length > 0 ? recentEvents.length : completedCount}
            activeWorkers={activeWorkers.size}
            throughput={throughput}
            connected={connected}
          />
        ) : (
          <DashboardView
            connected={connected}
            activeWorkers={activeWorkers}
            completedCount={completedCount}
            failedCount={failedCount}
            recentEvents={recentEvents}
            throughput={throughput}
            avgLatency={avgLatency}
            onSendEvent={sendEvent}
            onSimulateSpike={simulateSpike}
          />
        )}
      </main>
    </div>
  );
}

export default App;
