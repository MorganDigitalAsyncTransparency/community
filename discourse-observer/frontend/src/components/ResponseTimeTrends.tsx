// Spec: specs/dashboard/response-time-trends.md
// Tests: tests/dashboard/response-time-trends.unit.test.ts

import type { Topic } from "../mock/data";
import { formatWeekLabel } from "./topicFormatting";
import { computeWeeklyTrends } from "./trendMetrics";

interface ResponseTimeTrendsProps {
  topics: Topic[]; // all resolved topics, unfiltered — trend always spans full history
}

export function ResponseTimeTrends({ topics }: ResponseTimeTrendsProps) {
  const trends = computeWeeklyTrends(topics);

  if (trends.length === 0) {
    return (
      <section className="trends">
        <h2 className="trends-title">Weekly trends</h2>
        <p className="trends-empty">No data</p>
      </section>
    );
  }

  return (
    <section className="trends">
      <h2 className="trends-title">Weekly trends</h2>
      <table className="trends-table">
        <thead>
          <tr>
            <th className="trends-th trends-th-week">Week</th>
            <th className="trends-th trends-th-count">Topics</th>
            <th className="trends-th trends-th-metric">Median first reply</th>
            <th className="trends-th trends-th-metric">Median resolution</th>
          </tr>
        </thead>
        <tbody>
          {trends.map((row) => (
            <tr key={row.weekStart} className="trends-row">
              <td className="trends-td trends-td-week">{formatWeekLabel(row.weekStart)}</td>
              <td className="trends-td trends-td-count">{row.topicCount}</td>
              <td className="trends-td trends-td-metric">{row.medianFirstReply}</td>
              <td className="trends-td trends-td-metric">{row.medianResolution}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </section>
  );
}
