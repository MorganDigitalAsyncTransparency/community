// Spec: specs/dashboard/queue-visibility.md
// Tests: tests/dashboard/queue-visibility.unit.test.ts

import type { Topic } from "../mock/data";
import { formatAge, formatTags, sortedByOldest } from "./topicFormatting";

interface UnrepliedTableProps {
  topics: Topic[];
}

export function UnrepliedTable({ topics }: UnrepliedTableProps) {
  const sorted = sortedByOldest(topics);

  return (
    <table className="unreplied-table">
      <thead>
        <tr>
          <th className="unreplied-header-age">Ålder</th>
          <th className="unreplied-header-title">Titel</th>
          <th className="unreplied-header-tag">Tagg</th>
        </tr>
      </thead>
      <tbody>
        {sorted.map((topic) => (
          <tr key={topic.id} className="unreplied-row">
            <td className="unreplied-cell-age">{formatAge(topic.createdAt)}</td>
            <td className="unreplied-cell-title">{topic.title}</td>
            <td className="unreplied-cell-tag">{formatTags(topic.tags)}</td>
          </tr>
        ))}
      </tbody>
    </table>
  );
}
