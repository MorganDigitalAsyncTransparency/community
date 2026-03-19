// Spec: specs/dashboard/queue-visibility.md
// Tests: tests/dashboard/queue-visibility.unit.test.ts

import type { QueueSummary } from "../api/types";

interface SummaryCardProps {
  label: string;
  value: string;
}

function SummaryCard({ label, value }: SummaryCardProps) {
  return (
    <div className="summary-card">
      <span className="summary-card-value">{value}</span>
      <span className="summary-card-label">{label}</span>
    </div>
  );
}

interface SummaryCardsProps {
  data: QueueSummary;
}

export function SummaryCards({ data }: SummaryCardsProps) {
  const oldest = data.oldestUnrepliedAgeDays;
  return (
    <div className="summary-cards">
      <SummaryCard
        label="Awaiting reply"
        value={String(data.unrepliedCount)}
      />
      <SummaryCard
        label="Untagged"
        value={String(data.untaggedCount)}
      />
      <SummaryCard
        label="Oldest unreplied"
        value={oldest === null ? "–" : `${oldest}d`}
      />
    </div>
  );
}
