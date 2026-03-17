// Spec: specs/dashboard/stalled-topics.md
// Tests: tests/dashboard/stalled-topics.unit.test.ts

import type { Topic } from "../mock/data";
import type { ResolvedTag } from "./tagFilter";
import {
  filterStalledTopics,
  daysSinceLastActivity,
  formatStalledTag,
  minimumStalledDays,
} from "./stalledMetrics";

interface StalledTopicsProps {
  topics: Topic[];
  resolvedTags: Record<string, ResolvedTag>;
  monitoredTags: string[];
}

export function StalledTopics({
  topics,
  resolvedTags,
  monitoredTags,
}: StalledTopicsProps) {
  const stalled = filterStalledTopics(topics, resolvedTags);
  const monitored = new Set(monitoredTags);
  const minDays = minimumStalledDays(resolvedTags);

  return (
    <section className="stalled-section">
      <h2 className="stalled-heading">
        Stalled topics (inactive &gt; {minDays} days)
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
            {stalled.map((topic) => {
              const tagName = formatStalledTag(topic, monitored);
              const isDefault = tagName !== "–" && resolvedTags[tagName]?.stalledDaysIsDefault;
              return (
                <tr key={topic.id} className="stalled-row">
                  <td className="stalled-cell-title">{topic.title}</td>
                  <td className="stalled-cell-tag">
                    {tagName}
                    {isDefault && (
                      <span className="stalled-default-indicator"> (default threshold)</span>
                    )}
                  </td>
                  <td className="stalled-cell-days">
                    {daysSinceLastActivity(topic)}
                  </td>
                </tr>
              );
            })}
          </tbody>
        </table>
      )}
    </section>
  );
}
