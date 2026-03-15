// Spec: specs/dashboard/response-metrics.md
// Tests: tests/dashboard/response-metrics.unit.test.ts

import type { Topic } from "../mock/data";
import {
  medianFirstReplyTime,
  medianResolutionTime,
  outcomeCounts,
  formatOutcomes,
  answerRate,
} from "./responseMetrics";

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

interface ResponseMetricsCardsProps {
  topics: Topic[];
}

export function ResponseMetricsCards({ topics }: ResponseMetricsCardsProps) {
  const outcomeDisplay = formatOutcomes(outcomeCounts(topics));

  return (
    <div className="response-cards">
      <MetricCard label="Median first reply" value={medianFirstReplyTime(topics)} />
      <MetricCard label="Median resolution" value={medianResolutionTime(topics)} />
      <MetricCard label="Outcomes" value={outcomeDisplay} />
      <MetricCard label="Answer rate" value={answerRate(topics)} />
    </div>
  );
}
