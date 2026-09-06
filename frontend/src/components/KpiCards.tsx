interface KpiCardsProps {
  activeWorkers: number;
  throughput: number;
  completed: number;
  failed: number;
}

export function KpiCards({ activeWorkers, throughput, completed, failed }: KpiCardsProps) {
  return (
    <section className="kpi-grid">
      <div className="kpi-card kpi-card--workers">
        <div className="kpi-card__label">Active Workers</div>
        <div className="kpi-card__value">{activeWorkers}</div>
        <div className="kpi-card__unit">online</div>
      </div>

      <div className="kpi-card kpi-card--throughput">
        <div className="kpi-card__label">Throughput</div>
        <div className="kpi-card__value">{throughput}</div>
        <div className="kpi-card__unit">events/sec</div>
      </div>

      <div className="kpi-card kpi-card--completed">
        <div className="kpi-card__label">Completed</div>
        <div className="kpi-card__value">{completed.toLocaleString()}</div>
        <div className="kpi-card__unit">events</div>
      </div>

      <div className="kpi-card kpi-card--failed">
        <div className="kpi-card__label">Failed</div>
        <div className="kpi-card__value">{failed.toLocaleString()}</div>
        <div className="kpi-card__unit">events</div>
      </div>
    </section>
  );
}
