// Spec: specs/dashboard/queue-visibility.md
// Tests: tests/dashboard/queue-visibility.unit.test.ts

import type { DashboardData } from "../mock/data";
import { oldestUnrepliedDays } from "./topicFormatting";

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
  data: DashboardData;
}

export function SummaryCards({ data }: SummaryCardsProps) {
  return (
    <div className="summary-cards">
      <SummaryCard
        label="Awaiting reply"
        value={String(data.unrepliedTopics.length)}
      />
      <SummaryCard
        label="Untagged"
        value={String(data.untaggedTopics.length)}
      />
      <SummaryCard
        label="Oldest unreplied"
        value={oldestUnrepliedDays(data.unrepliedTopics)}
      />
    </div>
  );
}
