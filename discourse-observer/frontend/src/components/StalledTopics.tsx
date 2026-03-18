// Spec: specs/dashboard/stalled-topics.md
// Tests: tests/dashboard/stalled-topics.unit.test.ts

import { useState } from "react";
import type { Topic } from "../mock/data";
import type { ResolvedTag } from "./tagFilter";
import {
  filterStalledTopics,
  daysSinceLastActivity,
  formatStalledTag,
  stalledThresholdForTopic,
} from "./stalledMetrics";

type SortColumn = "tag" | "days";
type SortDirection = "asc" | "desc";

interface StalledTopicsProps {
  topics: Topic[];
  resolvedTags: Record<string, ResolvedTag>;
  monitoredTags: string[];
}

const DEFAULT_DIRECTIONS: Record<SortColumn, SortDirection> = {
  tag: "asc",
  days: "desc",
};

function sortTopics(
  topics: Topic[],
  column: SortColumn,
  direction: SortDirection,
  monitored: Set<string>,
): Topic[] {
  const sorted = [...topics];
  const dir = direction === "asc" ? 1 : -1;

  if (column === "tag") {
    sorted.sort((a, b) => {
      const aTag = formatStalledTag(a, monitored);
      const bTag = formatStalledTag(b, monitored);
      return dir * aTag.localeCompare(bTag);
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
  const [sortColumn, setSortColumn] = useState<SortColumn>("days");
  const [sortDirection, setSortDirection] = useState<SortDirection>("desc");

  const stalled = filterStalledTopics(topics, resolvedTags);
  const monitored = new Set(monitoredTags);
  const sorted = sortTopics(stalled, sortColumn, sortDirection, monitored);

  function handleSort(column: SortColumn) {
    if (column === sortColumn) {
      setSortDirection((d) => (d === "asc" ? "desc" : "asc"));
    } else {
      setSortColumn(column);
      setSortDirection(DEFAULT_DIRECTIONS[column]);
    }
  }

  function arrow(column: SortColumn): string {
    if (column !== sortColumn) return "";
    return sortDirection === "asc" ? " ▲" : " ▼";
  }

  return (
    <section className="stalled-section">
      <h2 className="stalled-heading">Stalled topics</h2>
      {sorted.length === 0 ? (
        <p className="stalled-empty">No stalled topics</p>
      ) : (
        <table className="stalled-table">
          <thead>
            <tr>
              <th className="stalled-header-title">Title</th>
              <th
                className="stalled-header-tag stalled-header-sortable"
                onClick={() => handleSort("tag")}
              >
                Tag{arrow("tag")}
              </th>
              <th className="stalled-header-threshold">Threshold</th>
              <th
                className="stalled-header-days stalled-header-sortable"
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
                  <td className="stalled-cell-title">{topic.title}</td>
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
