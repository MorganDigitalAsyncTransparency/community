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

function formatEntry(e: SyncLogEntry): string {
  return `${formatTimestamp(e.timestamp)}  ${e.mode}  ${e.pages} pages  ${e.topics} topics  ${formatDuration(e.durationSeconds)}`;
}

function formatProgress(p: SyncProgress): string {
  const elapsed = formatDuration(p.elapsedSeconds);
  const mode = p.mode || "sync";
  if (p.pages === 0) {
    return `${mode} running... fetching first page  ${elapsed}`;
  }
  return `${mode} running... ${p.pages} pages  ${p.topics} topics  ${elapsed}`;
}

export function SyncLog({ data }: SyncLogProps) {
  return (
    <section>
      <h2 className="app-section-title">Sync log</h2>

      {data.progress && (
        <div className="sync-progress">{formatProgress(data.progress)}</div>
      )}

      <pre className="sync-log">{
        data.entries.length === 0
          ? "No sync events yet."
          : data.entries.map(formatEntry).join("\n")
      }</pre>
    </section>
  );
}
