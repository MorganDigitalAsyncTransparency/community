// Spec: specs/dashboard/stalled-topics.md
// Tests: tests/dashboard/stalled-topics.unit.test.ts

import type { Topic } from "../mock/data";
import {
  filterStalledTopics,
  daysSinceLastActivity,
  formatStalledTag,
} from "./stalledMetrics";

interface StalledTopicsProps {
  topics: Topic[];
  stalledDays: number;
  closedTag: string;
  monitoredTags: string[];
}

export function StalledTopics({
  topics,
  stalledDays,
  closedTag,
  monitoredTags,
}: StalledTopicsProps) {
  const stalled = filterStalledTopics(topics, stalledDays, closedTag);
  const monitored = new Set(monitoredTags);

  return (
    <section className="stalled-section">
      <h2 className="stalled-heading">
        Stalled topics (inactive &gt; {stalledDays} days)
      </h2>
      {stalled.length === 0 ? (
        <p className="stalled-empty">No stalled topics</p>
      ) : (
        <table className="stalled-table">
          <thead>
            <tr>
              <th className="stalled-header-title">Title</th>
              <th className="stalled-header-tag">Tag</th>
              <th className="stalled-header-days">Days inactive</th>
            </tr>
          </thead>
          <tbody>
            {stalled.map((topic) => (
              <tr key={topic.id} className="stalled-row">
                <td className="stalled-cell-title">{topic.title}</td>
                <td className="stalled-cell-tag">
                  {formatStalledTag(topic, monitored)}
                </td>
                <td className="stalled-cell-days">
                  {daysSinceLastActivity(topic)}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </section>
  );
}
