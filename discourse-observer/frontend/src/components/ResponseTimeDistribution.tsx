// Spec: specs/dashboard/response-time-distribution.md
// Tests: tests/dashboard/response-time-distribution.unit.test.ts

import type { MetricsDistribution } from "../api/types";
import { DistributionChart } from "./DistributionChart";
import { CHART_COLOR_1, CHART_COLOR_2 } from "./themeColors";

interface ResponseTimeDistributionProps {
  data: MetricsDistribution;
}

export function ResponseTimeDistribution({
  data,
}: ResponseTimeDistributionProps) {
  return (
    <div className="rd-section">
      <section>
        <h2 className="rd-heading">First reply distribution</h2>
        {data.firstReply.length === 0 ? (
          <p className="rd-empty">No data</p>
        ) : (
          <DistributionChart
            data={data.firstReply}
            color={CHART_COLOR_1}
            name="Topics"
          />
        )}
      </section>

      <section>
        <h2 className="rd-heading">Resolution time distribution</h2>
        {data.resolution.length === 0 ? (
          <p className="rd-empty">No data</p>
        ) : (
          <DistributionChart
            data={data.resolution}
            color={CHART_COLOR_2}
            name="Topics"
          />
        )}
      </section>
    </div>
  );
}
