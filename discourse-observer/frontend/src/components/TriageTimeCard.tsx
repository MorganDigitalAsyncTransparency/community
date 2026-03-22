// Spec: specs/api/triage-time.md (TT-13)
// Tests: backend/api/triage-time_contract_test.go

import type { TriageTime } from "../api/types";

interface TriageTimeCardProps {
  data: TriageTime;
}

function formatHours(hours: number | null): string {
  if (hours === null) return "–";
  if (hours < 1) return `${Math.round(hours * 60)}m`;
  if (hours < 24) return `${hours.toFixed(1)}h`;
  const days = hours / 24;
  return `${days.toFixed(1)}d`;
}

export function TriageTimeCard({ data }: TriageTimeCardProps) {
  return (
    <div className="response-card">
      <span className="response-card-value">
        {formatHours(data.overall.medianHours)}
      </span>
      <span className="response-card-label">
        Median triage time ({data.overall.count} topics)
      </span>
    </div>
  );
}
