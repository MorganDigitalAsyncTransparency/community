// Spec: specs/dashboard/queue-visibility.md
// Tests: tests/dashboard/queue-visibility.unit.test.ts

import type { Topic } from "../mock/data";
import { formatAge, formatTags, topicUrl } from "./topicFormatting";
import { useTableSort, type SortDirection } from "./useTableSort";

type SortColumn = "topic" | "tags" | "age";

const DEFAULT_DIRECTIONS: Record<SortColumn, SortDirection> = {
  topic: "asc",
  tags: "asc",
  age: "desc",
};

function sortTopics(
  topics: Topic[],
  column: SortColumn,
  direction: SortDirection,
): Topic[] {
  const sorted = [...topics];
  const dir = direction === "asc" ? 1 : -1;

  if (column === "topic") {
    sorted.sort((a, b) => dir * a.title.localeCompare(b.title));
  } else if (column === "tags") {
    sorted.sort((a, b) => dir * formatTags(a.tags).localeCompare(formatTags(b.tags)));
  } else {
    sorted.sort((a, b) => {
      const aTime = new Date(a.createdAt).getTime();
      const bTime = new Date(b.createdAt).getTime();
      return dir === -1 ? aTime - bTime : bTime - aTime;
    });
  }

  return sorted;
}

interface UnrepliedTableProps {
  topics: Topic[];
}

export function UnrepliedTable({ topics }: UnrepliedTableProps) {
  const { sortColumn, sortDirection, handleSort, arrow } =
    useTableSort<SortColumn>("age", DEFAULT_DIRECTIONS);

  const sorted = sortTopics(topics, sortColumn, sortDirection);

  return (
    <table className="unreplied-table">
      <thead>
        <tr>
          <th
            className="unreplied-header-topic sortable-header"
            onClick={() => handleSort("topic")}
          >
            Topic{arrow("topic")}
          </th>
          <th
            className="unreplied-header-tag sortable-header"
            onClick={() => handleSort("tags")}
          >
            Tags{arrow("tags")}
          </th>
          <th
            className="unreplied-header-age sortable-header"
            onClick={() => handleSort("age")}
          >
            Age{arrow("age")}
          </th>
        </tr>
      </thead>
      <tbody>
        {sorted.map((topic) => (
          <tr key={topic.id} className="unreplied-row">
            <td className="unreplied-cell-topic">
              <a href={topicUrl(topic.id)} className="topic-link" target="_blank" rel="noreferrer">
                {topic.title}
              </a>
            </td>
            <td className="unreplied-cell-tag">{formatTags(topic.tags)}</td>
            <td className="unreplied-cell-age">{formatAge(topic.createdAt)}</td>
          </tr>
        ))}
      </tbody>
    </table>
  );
}
