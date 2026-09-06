import './index.css';
import { useWebSocket } from './hooks/useWebSocket';
import { Header } from './components/Header';
import { KpiCards } from './components/KpiCards';
import { RecentEvents } from './components/RecentEvents';
import { WorkerStatus } from './components/WorkerStatus';

const WS_URL = 'ws://localhost:8080/ws';

function App() {
  const {
    connected,
    activeWorkers,
    completedCount,
    failedCount,
    recentEvents,
    throughput,
  } = useWebSocket(WS_URL);

  return (
    <div className="dashboard">
      <Header connected={connected} />

      <KpiCards
        activeWorkers={activeWorkers.size}
        throughput={throughput}
        completed={completedCount}
        failed={failedCount}
      />

      <div className="content-grid">
        <RecentEvents events={recentEvents} />
        <WorkerStatus workers={activeWorkers} />
      </div>
    </div>
  );
}

export default App;
