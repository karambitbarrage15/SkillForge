interface HeaderProps {
  connected: boolean;
}

export function Header({ connected }: HeaderProps) {
  return (
    <header className="header">
      <div className="header__brand">
        <div className="header__logo">SF</div>
        <h1 className="header__title">StreamForge</h1>
      </div>
      <div className="header__status">
        <span
          className={`header__status-dot ${
            connected
              ? 'header__status-dot--connected'
              : 'header__status-dot--disconnected'
          }`}
        />
        <span
          className={
            connected
              ? 'header__status-text--connected'
              : 'header__status-text--disconnected'
          }
        >
          {connected ? 'Connected' : 'Disconnected'}
        </span>
      </div>
    </header>
  );
}
