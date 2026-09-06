interface WorkerStatusProps {
  workers: Set<string>;
}

export function WorkerStatus({ workers }: WorkerStatusProps) {
  const workerList = Array.from(workers).sort();

  return (
    <div className="panel">
      <div className="panel__header">
        <span className="panel__title">Workers</span>
        <span className="panel__badge">{workerList.length} online</span>
      </div>
      <div className="panel__body">
        {workerList.length === 0 ? (
          <div className="empty-state">
            <div className="empty-state__icon">⚙️</div>
            <div>No workers online</div>
          </div>
        ) : (
          workerList.map(id => (
            <div key={id} className="worker-row">
              <div className="worker-row__dot" />
              <span className="worker-row__id">{id}</span>
              <span className="worker-row__label">Online</span>
            </div>
          ))
        )}
      </div>
    </div>
  );
}
