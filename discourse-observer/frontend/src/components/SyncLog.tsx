// Spec: specs/api/api-contract.md (AC-33)
// Tests: manual (sync log rendering)

import type { SyncLogResponse, SyncProgress, SyncLogEntry } from "../api/types";

interface SyncLogProps {
  data: SyncLogResponse;
}

function formatTimestamp(iso: string): string {
  return new Date(iso).toLocaleString(undefined, {
    dateStyle: "short",
    timeStyle: "medium",
  });
}

function formatDuration(seconds: number): string {
  if (seconds < 1) return `${Math.round(seconds * 1000)}ms`;
  if (seconds < 60) return `${seconds.toFixed(1)}s`;
  const m = Math.floor(seconds / 60);
  const s = Math.round(seconds % 60);
  return `${m}m${s}s`;
}

function modeClass(mode: string): string {
  switch (mode) {
    case "initial": return "sync-entry sync-entry-initial";
    case "delta":   return "sync-entry sync-entry-delta";
    case "detail":  return "sync-entry sync-entry-detail";
    default:        return "sync-entry";
  }
}

function topicSummary(e: SyncLogEntry): string {
  if (!e.hasChanges) return "up to date";
  return `${e.topics} topics`;
}

function EntryRow({ e }: { e: SyncLogEntry }) {
  return (
    <div className={modeClass(e.mode)}>
      <span className="sync-entry-time">{formatTimestamp(e.timestamp)}</span>
      <span className="sync-entry-mode">{e.mode}</span>
      <span className="sync-entry-stat">{topicSummary(e)}</span>
      <span className="sync-entry-stat">{formatDuration(e.durationSeconds)}</span>
    </div>
  );
}

function formatEta(etaSeconds: number): string {
  const now = new Date();
  const done = new Date(now.getTime() + etaSeconds * 1000);
  return done.toLocaleTimeString(undefined, { timeStyle: "medium" });
}

function ProgressRow({ p }: { p: SyncProgress }) {
  const elapsed = formatDuration(p.elapsedSeconds);
  const mode = p.mode || "sync";
  const hasTotal = p.totalTopics > 0;
  const topicLabel = hasTotal
    ? `${p.topics}/${p.totalTopics} topics`
    : `${p.topics} topics`;

  return (
    <div className="sync-progress">
      <span className="sync-entry-mode">{mode}</span>
      {p.topics === 0
        ? <span className="sync-entry-stat">starting...</span>
        : <span className="sync-entry-stat">{topicLabel}</span>
      }
      <span className="sync-entry-stat">{elapsed}</span>
      {p.etaSeconds > 0 && (
        <span className="sync-entry-stat">done ~{formatEta(p.etaSeconds)}</span>
      )}
    </div>
  );
}

export function SyncLog({ data }: SyncLogProps) {
  return (
    <section>
      <h2 className="app-section-title">Sync log</h2>

      <div className="sync-log-description">
        <p>The sync pipeline keeps local data up to date with the Discourse forum. Three sync types run at different times:</p>
        <dl className="sync-type-list">
          <dt className="sync-type-label sync-type-initial">initial</dt>
          <dd>Full crawl of all topics. Runs once when the database is empty.</dd>
          <dt className="sync-type-label sync-type-delta">delta</dt>
          <dd>Incremental sync of recent changes. Runs every 15 minutes.</dd>
          <dt className="sync-type-label sync-type-detail">detail</dt>
          <dd>Fetches revision history per topic. Runs during low-activity windows.</dd>
        </dl>
        <p className="sync-log-retention">The log keeps the 20 most recent entries per type and persists across restarts.</p>
      </div>

      {data.progress && <ProgressRow p={data.progress} />}

      <div className="sync-log">
        {data.entries.length === 0
          ? <div className="sync-entry">No sync events yet.</div>
          : data.entries.map((e, i) => <EntryRow key={i} e={e} />)
        }
      </div>
    </section>
  );
}
