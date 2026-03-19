// Spec: specs/dashboard/stalled-topics.md
// Tests: backend/api/contract_test.go

import type { StalledTopic } from "../api/types";
import { useTableSort, type SortDirection } from "./useTableSort";

type SortColumn = "topic" | "tag" | "threshold" | "days";

interface StalledTopicsProps {
  topics: StalledTopic[];
}

const DEFAULT_DIRECTIONS: Record<SortColumn, SortDirection> = {
  topic: "asc",
  tag: "asc",
  threshold: "asc",
  days: "desc",
};

function sortTopics(
  topics: StalledTopic[],
  column: SortColumn,
  direction: SortDirection,
): StalledTopic[] {
  const sorted = [...topics];
  const dir = direction === "asc" ? 1 : -1;

  if (column === "topic") {
    sorted.sort((a, b) => dir * a.title.localeCompare(b.title));
  } else if (column === "tag") {
    sorted.sort((a, b) => dir * (a.strictestTag ?? "").localeCompare(b.strictestTag ?? ""));
  } else if (column === "threshold") {
    sorted.sort((a, b) => dir * (a.thresholdDays - b.thresholdDays));
  } else {
    sorted.sort((a, b) => dir * (a.daysSinceLastActivity - b.daysSinceLastActivity));
  }

  return sorted;
}

export function StalledTopics({ topics }: StalledTopicsProps) {
  const { sortColumn, sortDirection, handleSort, arrow } =
    useTableSort<SortColumn>("days", DEFAULT_DIRECTIONS);

  const sorted = sortTopics(topics, sortColumn, sortDirection);

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
            {sorted.map((topic) => (
              <tr key={topic.id} className="stalled-row">
                <td className="stalled-cell-topic">
                  <a href={topic.topicUrl} className="topic-link" target="_blank" rel="noreferrer">
                    {topic.title}
                  </a>
                </td>
                <td className="stalled-cell-tag">{topic.strictestTag ?? "–"}</td>
                <td className="stalled-cell-threshold">
                  {topic.thresholdDays}
                  {topic.thresholdIsDefault && (
                    <span className="stalled-default-indicator"> (default)</span>
                  )}
                </td>
                <td className="stalled-cell-days">
                  {topic.daysSinceLastActivity}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </section>
  );
}
