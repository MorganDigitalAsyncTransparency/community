// Spec: specs/dashboard/stalled-topics.md
// Tests: tests/dashboard/stalled-topics.unit.test.ts

import type { Topic } from "../mock/data";

const DAY_MS = 86_400_000;

export function daysSinceLastActivity(topic: Topic, now: Date = new Date()): number {
  const lastActivity = topic.lastActivityAt ?? topic.createdAt;
  return Math.floor((now.getTime() - new Date(lastActivity).getTime()) / DAY_MS);
}

export function filterStalledTopics(
  topics: Topic[],
  stalledDays: number,
  closedTag: string,
  now: Date = new Date(),
): Topic[] {
  const thresholdMs = stalledDays * DAY_MS;

  return topics
    .filter((t) => {
      if (t.tags.includes(closedTag)) return false;
      const lastActivity = t.lastActivityAt ?? t.createdAt;
      const elapsed = now.getTime() - new Date(lastActivity).getTime();
      return elapsed > thresholdMs;
    })
    .sort((a, b) => {
      const aTime = new Date(a.lastActivityAt ?? a.createdAt).getTime();
      const bTime = new Date(b.lastActivityAt ?? b.createdAt).getTime();
      return aTime - bTime;
    });
}

export function formatStalledTag(topic: Topic, monitored: Set<string>): string {
  const found = topic.tags.find((t) => monitored.has(t));
  return found ?? "–";
}
