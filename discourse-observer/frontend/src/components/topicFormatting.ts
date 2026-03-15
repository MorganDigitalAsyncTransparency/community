// Spec: specs/dashboard/queue-visibility.md
// Tests: tests/dashboard/queue-visibility.unit.test.ts

import type { Topic } from "../mock/data";
import { formatDuration } from "./responseMetrics";

const MILLISECONDS_PER_DAY = 86_400_000;

export function formatAge(isoDate: string): string {
  return formatDuration(Date.now() - new Date(isoDate).getTime());
}

export function sortedByOldest(topics: Topic[]): Topic[] {
  return [...topics].sort(
    (a, b) => new Date(a.createdAt).getTime() - new Date(b.createdAt).getTime()
  );
}

export function oldestUnrepliedDays(topics: Topic[]): string {
  if (topics.length === 0) {
    return "–";
  }

  const oldestMs = topics.reduce(
    (oldest, topic) => Math.min(oldest, new Date(topic.createdAt).getTime()),
    Infinity
  );

  const days = Math.floor((Date.now() - oldestMs) / MILLISECONDS_PER_DAY);
  return `${days}d`;
}

export function formatTags(tags: string[]): string {
  return tags.length > 0 ? tags.join(", ") : "–";
}
