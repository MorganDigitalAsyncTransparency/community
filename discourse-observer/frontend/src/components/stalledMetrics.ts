// Spec: specs/dashboard/stalled-topics.md
// Tests: tests/dashboard/stalled-topics.unit.test.ts

import type { Topic } from "../mock/data";
import type { ResolvedTag } from "./tagFilter";

const DAY_MS = 86_400_000;

export function daysSinceLastActivity(topic: Topic, now: Date = new Date()): number {
  const lastActivity = topic.lastActivityAt ?? topic.createdAt;
  return Math.floor((now.getTime() - new Date(lastActivity).getTime()) / DAY_MS);
}

// Finds the strictest (lowest) stalledDays across a topic's configured tags.
// Returns null when the topic has no configured tags.
function strictestStalledDays(
  topicTags: string[],
  resolved: Record<string, ResolvedTag>,
): number | null {
  let best: number | null = null;
  for (const tag of topicTags) {
    const r = resolved[tag];
    if (!r) continue;
    if (best === null || r.stalledDays < best) {
      best = r.stalledDays;
    }
  }
  return best;
}

// Collects all closedTag values from a topic's configured tags.
function closedTagsForTopic(
  topicTags: string[],
  resolved: Record<string, ResolvedTag>,
): Set<string> {
  const result = new Set<string>();
  for (const tag of topicTags) {
    const r = resolved[tag];
    if (r?.closedTag) {
      result.add(r.closedTag);
    }
  }
  return result;
}

export function filterStalledTopics(
  topics: Topic[],
  resolved: Record<string, ResolvedTag>,
  now: Date = new Date(),
): Topic[] {
  return topics
    .filter((t) => {
      const closedTags = closedTagsForTopic(t.tags, resolved);
      if (t.tags.some((tag) => closedTags.has(tag))) return false;

      const threshold = strictestStalledDays(t.tags, resolved);
      if (threshold === null) return false;

      const lastActivity = t.lastActivityAt ?? t.createdAt;
      const elapsed = now.getTime() - new Date(lastActivity).getTime();
      return elapsed > threshold * DAY_MS;
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

// Returns the strictest (lowest) stalledDays for a topic and whether it's a default.
// Returns null when the topic has no configured tags.
export function stalledThresholdForTopic(
  topic: Topic,
  resolved: Record<string, ResolvedTag>,
): { days: number; isDefault: boolean } | null {
  let best: { days: number; isDefault: boolean } | null = null;
  for (const tag of topic.tags) {
    const r = resolved[tag];
    if (!r) continue;
    if (best === null || r.stalledDays < best.days) {
      best = { days: r.stalledDays, isDefault: r.stalledDaysIsDefault };
    }
  }
  return best;
}
