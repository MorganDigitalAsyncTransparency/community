import type { DashboardData } from "../mock/data";

const MILLISECONDS_PER_DAY = 86_400_000;

function oldestUnrepliedDays(data: DashboardData): string {
  if (data.unrepliedTopics.length === 0) {
    return "–";
  }

  const oldestMs = data.unrepliedTopics.reduce(
    (oldest, topic) => Math.min(oldest, new Date(topic.createdAt).getTime()),
    Infinity
  );

  const days = Math.floor((Date.now() - oldestMs) / MILLISECONDS_PER_DAY);
  return `${days}d`;
}

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
        label="Väntar på svar"
        value={String(data.unrepliedTopics.length)}
      />
      <SummaryCard
        label="Otaggade"
        value={String(data.untaggedTopics.length)}
      />
      <SummaryCard
        label="Äldsta utan svar"
        value={oldestUnrepliedDays(data)}
      />
    </div>
  );
}
