// Spec: specs/dashboard/queue-visibility.md
// Tests: tests/dashboard/queue-visibility.unit.test.ts

import type { UntaggedTopic } from "../api/types";
import { formatAge } from "./topicFormatting";
import { useTableSort, type SortDirection } from "./useTableSort";

type SortColumn = "topic" | "categoryName" | "age";

const DEFAULT_DIRECTIONS: Record<SortColumn, SortDirection> = {
  topic: "asc",
  categoryName: "asc",
  age: "desc",
};

function sortTopics(
  topics: UntaggedTopic[],
  column: SortColumn,
  direction: SortDirection,
): UntaggedTopic[] {
  const sorted = [...topics];
  const dir = direction === "asc" ? 1 : -1;

  if (column === "topic") {
    sorted.sort((a, b) => dir * a.title.localeCompare(b.title));
  } else if (column === "categoryName") {
    sorted.sort((a, b) => dir * a.categoryName.localeCompare(b.categoryName));
  } else {
    sorted.sort((a, b) => {
      const aTime = new Date(a.createdAt).getTime();
      const bTime = new Date(b.createdAt).getTime();
      return dir === -1 ? aTime - bTime : bTime - aTime;
    });
  }

  return sorted;
}

interface UntaggedTableProps {
  topics: UntaggedTopic[];
}

export function UntaggedTable({ topics }: UntaggedTableProps) {
  const { sortColumn, sortDirection, handleSort, arrow } =
    useTableSort<SortColumn>("age", DEFAULT_DIRECTIONS);

  const sorted = sortTopics(topics, sortColumn, sortDirection);

  return (
    <table className="untagged-table">
      <thead>
        <tr>
          <th
            className="untagged-header-topic sortable-header"
            onClick={() => handleSort("topic")}
          >
            Topic{arrow("topic")}
          </th>
          <th
            className="untagged-header-category sortable-header"
            onClick={() => handleSort("categoryName")}
          >
            Category{arrow("categoryName")}
          </th>
          <th
            className="untagged-header-age sortable-header"
            onClick={() => handleSort("age")}
          >
            Age{arrow("age")}
          </th>
        </tr>
      </thead>
      <tbody>
        {sorted.map((topic) => (
          <tr key={topic.id} className="untagged-row">
            <td className="untagged-cell-topic">
              <a href={topic.topicUrl} className="topic-link" target="_blank" rel="noreferrer">
                {topic.title}
              </a>
            </td>
            <td className="untagged-cell-category">{topic.categoryName}</td>
            <td className="untagged-cell-age">{formatAge(topic.createdAt)}</td>
          </tr>
        ))}
      </tbody>
    </table>
  );
}
