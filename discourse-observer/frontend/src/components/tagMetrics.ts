// Spec: specs/dashboard/tag-distribution.md
// Tests: tests/dashboard/tag-distribution.unit.test.ts

import type { Topic } from "../mock/data";
import { formatDuration } from "./topicFormatting";
import { mondayOf } from "./trendMetrics";

export interface TagVolume {
  tag: string;
  topicCount: number;
}

export interface TagResolutionTime {
  tag: string;
  resolvedCount: number;
  medianResolution: string; // formatted duration or "–"
}

export interface TagBacklog {
  tag: string;
  openCount: number;
}

export interface WeeklyBacklog {
  weekStart: string; // Monday of the week, YYYY-MM-DD (UTC)
  created: number;
  resolved: number;
  stillOpen: number;
}

// Groups topics by tag. A topic with multiple tags appears in each tag's group.
// Topics with no tags produce no entries.
function topicsByTag(topics: Topic[]): Map<string, Topic[]> {
  const byTag = new Map<string, Topic[]>();
  for (const topic of topics) {
    for (const tag of topic.tags) {
      const bucket = byTag.get(tag) ?? [];
      bucket.push(topic);
      byTag.set(tag, bucket);
    }
  }
  return byTag;
}

// Returns the median of a sorted numeric array, truncated to a whole number.
// Caller is responsible for sorting. Returns -1 for empty arrays (sentinel for "no data").
function medianMs(sorted: number[]): number {
  if (sorted.length === 0) return -1;
  const mid = Math.floor(sorted.length / 2);
  return sorted.length % 2 === 1
    ? sorted[mid]
    : Math.trunc((sorted[mid - 1] + sorted[mid]) / 2);
}

// Returns all tags ranked by total topic count, highest first (TD-1 – TD-3).
export function tagVolumeRanking(topics: Topic[]): TagVolume[] {
  const byTag = topicsByTag(topics);
  return [...byTag.entries()]
    .map(([tag, tagTopics]) => ({ tag, topicCount: tagTopics.length }))
    .sort((a, b) => b.topicCount - a.topicCount || a.tag.localeCompare(b.tag));
}

// Returns all tags from resolvedTopics ranked by median resolution time,
// slowest first. Tags with no resolvedAt values sort to the bottom (TD-6 – TD-10).
export function tagResolutionRanking(resolvedTopics: Topic[]): TagResolutionTime[] {
  const byTag = topicsByTag(resolvedTopics);

  const entries = [...byTag.entries()].map(([tag, tagTopics]) => {
    const durations = tagTopics
      .filter((t) => t.resolvedAt)
      .map((t) => new Date(t.resolvedAt!).getTime() - new Date(t.createdAt).getTime())
      .sort((a, b) => a - b);
    const ms = medianMs(durations);
    return {
      tag,
      resolvedCount: durations.length,
      medianResolution: ms >= 0 ? formatDuration(ms) : "–",
      _sortKey: ms,
    };
  });

  return entries
    .sort((a, b) => {
      if (a._sortKey < 0 && b._sortKey < 0) return a.tag.localeCompare(b.tag);
      if (a._sortKey < 0) return 1;
      if (b._sortKey < 0) return -1;
      return b._sortKey - a._sortKey; // slowest first
    })
    .map(({ tag, resolvedCount, medianResolution }) => ({
      tag,
      resolvedCount,
      medianResolution,
    }));
}

// Returns all tags from openTopics ranked by open topic count, highest first (TD-12 – TD-13).
export function tagBacklogRanking(openTopics: Topic[]): TagBacklog[] {
  const byTag = topicsByTag(openTopics);
  return [...byTag.entries()]
    .map(([tag, tagTopics]) => ({ tag, openCount: tagTopics.length }))
    .sort((a, b) => b.openCount - a.openCount || a.tag.localeCompare(b.tag));
}

// Returns one WeeklyBacklog per calendar week that contains at least one topic,
// ordered newest first. allTopics = unreplied + resolved combined (unfiltered).
// openTopics = unreplied topics (unfiltered). The ID-based lookup correctly
// attributes each topic to "still open" or "resolved" (TD-16 – TD-23).
export function computeWeeklyBacklog(
  allTopics: Topic[],
  openTopics: Topic[]
): WeeklyBacklog[] {
  const openIds = new Set(openTopics.map((t) => t.id));

  const byWeek = new Map<string, { created: number; resolved: number; stillOpen: number }>();

  for (const topic of allTopics) {
    const week = mondayOf(new Date(topic.createdAt));
    const bucket = byWeek.get(week) ?? { created: 0, resolved: 0, stillOpen: 0 };
    bucket.created += 1;
    if (openIds.has(topic.id)) {
      bucket.stillOpen += 1;
    } else {
      bucket.resolved += 1;
    }
    byWeek.set(week, bucket);
  }

  return [...byWeek.entries()]
    .sort(([a], [b]) => b.localeCompare(a)) // newest first
    .map(([weekStart, counts]) => ({ weekStart, ...counts }));
}
