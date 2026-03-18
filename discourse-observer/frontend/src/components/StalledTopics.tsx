// Spec: specs/dashboard/stalled-topics.md
// Tests: tests/dashboard/stalled-topics.unit.test.ts

import type { Topic } from "../mock/data";
import type { ResolvedTag } from "./tagFilter";
import { topicUrl } from "./topicFormatting";
import { useTableSort, type SortDirection } from "./useTableSort";
import {
  filterStalledTopics,
  daysSinceLastActivity,
  formatStalledTag,
  stalledThresholdForTopic,
} from "./stalledMetrics";

type SortColumn = "topic" | "tag" | "threshold" | "days";

interface StalledTopicsProps {
  topics: Topic[];
  resolvedTags: Record<string, ResolvedTag>;
  monitoredTags: string[];
}

const DEFAULT_DIRECTIONS: Record<SortColumn, SortDirection> = {
  topic: "asc",
  tag: "asc",
  threshold: "asc",
  days: "desc",
};

function sortTopics(
  topics: Topic[],
  column: SortColumn,
  direction: SortDirection,
  monitored: Set<string>,
  resolvedTags: Record<string, ResolvedTag>,
): Topic[] {
  const sorted = [...topics];
  const dir = direction === "asc" ? 1 : -1;

  if (column === "topic") {
    sorted.sort((a, b) => dir * a.title.localeCompare(b.title));
  } else if (column === "tag") {
    sorted.sort((a, b) => {
      const aTag = formatStalledTag(a, monitored);
      const bTag = formatStalledTag(b, monitored);
      return dir * aTag.localeCompare(bTag);
    });
  } else if (column === "threshold") {
    sorted.sort((a, b) => {
      const aThreshold = stalledThresholdForTopic(a, resolvedTags);
      const bThreshold = stalledThresholdForTopic(b, resolvedTags);
      const aDays = aThreshold ? aThreshold.days : 0;
      const bDays = bThreshold ? bThreshold.days : 0;
      return dir * (aDays - bDays);
    });
  } else {
    sorted.sort((a, b) => {
      const aDate = new Date(a.lastActivityAt ?? a.createdAt).getTime();
      const bDate = new Date(b.lastActivityAt ?? b.createdAt).getTime();
      // "asc" = oldest first (lowest date), "desc" = highest days first (also lowest date)
      // days = now - date, so higher days = lower date
      // For days desc (highest first): lower date first → aDate - bDate
      // For days asc (lowest first): higher date first → bDate - aDate
      return dir === -1 ? aDate - bDate : bDate - aDate;
    });
  }

  return sorted;
}

export function StalledTopics({
  topics,
  resolvedTags,
  monitoredTags,
}: StalledTopicsProps) {
  const { sortColumn, sortDirection, handleSort, arrow } =
    useTableSort<SortColumn>("days", DEFAULT_DIRECTIONS);

  const stalled = filterStalledTopics(topics, resolvedTags);
  const monitored = new Set(monitoredTags);
  const sorted = sortTopics(stalled, sortColumn, sortDirection, monitored, resolvedTags);

  return (
    <section className="stalled-section">
      <h2 className="stalled-heading">Stalled topics</h2>
      {sorted.length === 0 ? (
        <p className="stalled-empty">No stalled topics</p>
      ) : (
        <table className="stalled-table">
          <thead>
            <tr>
              <th
                className="stalled-header-topic sortable-header"
                onClick={() => handleSort("topic")}
              >
                Topic{arrow("topic")}
              </th>
              <th
                className="stalled-header-tag sortable-header"
                onClick={() => handleSort("tag")}
              >
                Tag{arrow("tag")}
              </th>
              <th
                className="stalled-header-threshold sortable-header"
                onClick={() => handleSort("threshold")}
              >
                Threshold{arrow("threshold")}
              </th>
              <th
                className="stalled-header-days sortable-header"
                onClick={() => handleSort("days")}
              >
                Days inactive{arrow("days")}
              </th>
            </tr>
          </thead>
          <tbody>
            {sorted.map((topic) => {
              const tagName = formatStalledTag(topic, monitored);
              const threshold = stalledThresholdForTopic(topic, resolvedTags);
              return (
                <tr key={topic.id} className="stalled-row">
                  <td className="stalled-cell-topic">
                    <a href={topicUrl(topic.id)} className="topic-link" target="_blank" rel="noreferrer">
                      {topic.title}
                    </a>
                  </td>
                  <td className="stalled-cell-tag">{tagName}</td>
                  <td className="stalled-cell-threshold">
                    {threshold ? threshold.days : "–"}
                    {threshold?.isDefault && (
                      <span className="stalled-default-indicator"> (default)</span>
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
