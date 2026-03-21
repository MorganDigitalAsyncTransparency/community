// Spec: specs/dashboard/dashboard-components.md
// ADR: docs/decisions/0011-dashboard-layout-and-theme.md
// Tests: manual (footer rendering)

interface FooterProps {
  version: string;
  lastSyncedAt: string | null;
  onSyncLogClick?: () => void;
}

function formatSyncTime(isoString: string): string {
  return new Date(isoString).toLocaleString(undefined, {
    dateStyle: "short",
    timeStyle: "short",
  });
}

export function Footer({ version, lastSyncedAt, onSyncLogClick }: FooterProps) {
  return (
    <footer className="footer">
      <div className="footer-inner">
        <span>{version}</span>
        <span>&middot;</span>
        <span>
          {lastSyncedAt
            ? `Last synced ${formatSyncTime(lastSyncedAt)}`
            : "Not yet synced"}
        </span>
        <span>&middot;</span>
        <button className="footer-link" onClick={onSyncLogClick}>
          Sync log
        </button>
        <span>&middot;</span>
        <a
          className="footer-link"
          href="https://github.com/MorganDigitalAsyncTransparency/community"
          target="_blank"
          rel="noopener noreferrer"
        >
          GitHub &#8599;
        </a>
      </div>
    </footer>
  );
}
