// Spec: specs/dashboard/response-time-distribution.md
// Tests: tests/dashboard/response-time-distribution.unit.test.ts

import type { Topic } from "../mock/data";
import { DistributionChart } from "./DistributionChart";
import {
  bucketDurations,
  firstReplyDurations,
  resolutionDurations,
} from "./distributionMetrics";
import { getThemeColor } from "./themeColors";

interface ResponseTimeDistributionProps {
  topics: Topic[];
  ceilingsHours: number[];
}

export function ResponseTimeDistribution({
  topics,
  ceilingsHours,
}: ResponseTimeDistributionProps) {
  const firstReplyColor = getThemeColor("--color-chart-1");
  const resolutionColor = getThemeColor("--color-chart-2");

  const replyDurations = firstReplyDurations(topics);
  const resolDurations = resolutionDurations(topics);

  const replyBuckets = bucketDurations(replyDurations, ceilingsHours);
  const resolBuckets = bucketDurations(resolDurations, ceilingsHours);

  return (
    <div className="rd-section">
      <section className="rd-histogram">
        <h2 className="rd-heading">First reply distribution</h2>
        {replyDurations.length === 0 ? (
          <p className="rd-empty">No data</p>
        ) : (
          <DistributionChart
            data={replyBuckets}
            color={firstReplyColor}
            name="Topics"
          />
        )}
      </section>

      <section className="rd-histogram">
        <h2 className="rd-heading">Resolution time distribution</h2>
        {resolDurations.length === 0 ? (
          <p className="rd-empty">No data</p>
        ) : (
          <DistributionChart
            data={resolBuckets}
            color={resolutionColor}
            name="Topics"
          />
        )}
      </section>
    </div>
  );
}
