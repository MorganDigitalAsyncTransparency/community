// Spec: specs/dashboard/response-metrics.md
// Tests: tests/dashboard/response-metrics.unit.test.ts

import type { MetricsSummary } from "../api/types";
import { formatDuration } from "./topicFormatting";

interface MetricCardProps {
  label: string;
  value: string;
}

function MetricCard({ label, value }: MetricCardProps) {
  return (
    <div className="response-card">
      <span className="response-card-value">{value}</span>
      <span className="response-card-label">{label}</span>
    </div>
  );
}

function formatOutcomes(solved: number, selfClosed: number): string {
  return `${solved} solved / ${selfClosed} self-closed`;
}

interface ResponseMetricsCardsProps {
  data: MetricsSummary;
}

export function ResponseMetricsCards({ data }: ResponseMetricsCardsProps) {
  const firstReply = data.medianFirstReplyMs === null
    ? "–" : formatDuration(data.medianFirstReplyMs);
  const resolution = data.medianResolutionMs === null
    ? "–" : formatDuration(data.medianResolutionMs);
  const answerRate = data.answerRatePercent === null
    ? "–" : `${data.answerRatePercent}%`;

  return (
    <div className="response-cards">
      <MetricCard label="Median first reply" value={firstReply} />
      <MetricCard label="Median resolution" value={resolution} />
      <MetricCard label="Outcomes" value={formatOutcomes(data.solvedCount, data.selfClosedCount)} />
      <MetricCard label="Answer rate" value={answerRate} />
    </div>
  );
}
