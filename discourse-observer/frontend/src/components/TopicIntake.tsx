// Spec: specs/dashboard/topic-intake.md
// Tests: tests/dashboard/topic-intake.unit.test.ts

import type { Topic } from "../mock/data";
import { IntakeChart } from "./IntakeChart";
import { type IntakeGranularity, type TimeRange, computeIntakeBuckets } from "./intakeMetrics";

interface TopicIntakeProps {
  topics: Topic[];
  granularity: IntakeGranularity;
  timeRange: TimeRange | null;
}

export function TopicIntake({ topics, granularity, timeRange }: TopicIntakeProps) {
  const buckets = computeIntakeBuckets(topics, granularity, timeRange);

  if (buckets.length === 0) {
    return (
      <section className="intake">
        <h2 className="intake-title">Topic intake</h2>
        <p className="intake-empty">No data</p>
      </section>
    );
  }

  return (
    <section className="intake">
      <h2 className="intake-title">Topic intake</h2>
      <IntakeChart data={buckets} />
    </section>
  );
}
