// Spec: specs/observer/mock-server-service.md
// Tests: manual (sync log rendering)

import type { SyncLogEntry } from "../api/types";

interface SyncLogProps {
  entries: SyncLogEntry[];
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

export function SyncLog({ entries }: SyncLogProps) {
  return (
    <section>
      <h2 className="app-section-title">Sync log</h2>
      <pre className="sync-log">{
        entries.length === 0
          ? "No sync events yet."
          : entries.map(formatEntry).join("\n")
      }</pre>
    </section>
  );
}
