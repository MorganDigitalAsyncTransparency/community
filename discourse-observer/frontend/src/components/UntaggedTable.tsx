// Spec: specs/dashboard/queue-visibility.md
// Tests: tests/dashboard/queue-visibility.unit.test.ts

import type { Topic } from "../mock/data";
import { formatAge, sortedByOldest } from "./topicFormatting";

interface UntaggedTableProps {
  topics: Topic[];
}

export function UntaggedTable({ topics }: UntaggedTableProps) {
  const sorted = sortedByOldest(topics);

  return (
    <table className="untagged-table">
      <thead>
        <tr>
          <th className="untagged-header-age">Ålder</th>
          <th className="untagged-header-title">Titel</th>
          <th className="untagged-header-category">Kategori</th>
        </tr>
      </thead>
      <tbody>
        {sorted.map((topic) => (
          <tr key={topic.id} className="untagged-row">
            <td className="untagged-cell-age">{formatAge(topic.createdAt)}</td>
            <td className="untagged-cell-title">{topic.title}</td>
            <td className="untagged-cell-category">{topic.category}</td>
          </tr>
        ))}
      </tbody>
    </table>
  );
}
