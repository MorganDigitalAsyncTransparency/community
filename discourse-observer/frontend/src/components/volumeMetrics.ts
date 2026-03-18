// Spec: specs/dashboard/response-metrics.md
// Tests: tests/dashboard/volume-metrics.unit.test.ts

import type { Topic } from "../mock/data";
import type { IntakeGranularity, TimeRange } from "./intakeMetrics";
import { mondayOf } from "./trendMetrics";
import { formatWeekLabel } from "./topicFormatting";

export interface VolumeBucket {
  label: string;
  bucketKey: string;
  created: number;
  accepted: number;
  closed: number;
  open: number;
}

function dayOf(date: Date): string {
  return date.toISOString().slice(0, 10);
}

function formatDayLabel(isoDate: string): string {
  return new Date(isoDate + "T00:00:00Z").toLocaleDateString(undefined, {
    year: "numeric",
    month: "short",
    day: "numeric",
    timeZone: "UTC",
  });
}

function bucketKeyFor(date: Date, granularity: IntakeGranularity): string {
  return granularity === "daily" ? dayOf(date) : mondayOf(date);
}

function nextDay(isoDate: string): string {
  const d = new Date(isoDate + "T00:00:00Z");
  d.setUTCDate(d.getUTCDate() + 1);
  return dayOf(d);
}

function nextMonday(isoDate: string): string {
  const d = new Date(isoDate + "T00:00:00Z");
  d.setUTCDate(d.getUTCDate() + 7);
  return dayOf(d);
}

interface TopicSets {
  allTopics: Topic[];
  solvedTopics: Topic[];
  selfClosedTopics: Topic[];
  openTopics: Topic[];
}

function countByBucket(
  topics: Topic[],
  granularity: IntakeGranularity,
): Map<string, number> {
  const counts = new Map<string, number>();
  for (const topic of topics) {
    const key = bucketKeyFor(new Date(topic.createdAt), granularity);
    counts.set(key, (counts.get(key) ?? 0) + 1);
  }
  return counts;
}

export function computeVolumeBuckets(
  sets: TopicSets,
  granularity: IntakeGranularity,
  range: TimeRange | null,
): VolumeBucket[] {
  if (!range) return [];

  const createdCounts = countByBucket(sets.allTopics, granularity);
  const acceptedCounts = countByBucket(sets.solvedTopics, granularity);
  const closedCounts = countByBucket(sets.selfClosedTopics, granularity);
  const openCounts = countByBucket(sets.openTopics, granularity);

  const advance = granularity === "daily" ? nextDay : nextMonday;
  const formatLabel = granularity === "daily" ? formatDayLabel : formatWeekLabel;

  const buckets: VolumeBucket[] = [];
  let current = range.first;
  while (current <= range.last) {
    buckets.push({
      label: formatLabel(current),
      bucketKey: current,
      created: createdCounts.get(current) ?? 0,
      accepted: acceptedCounts.get(current) ?? 0,
      closed: closedCounts.get(current) ?? 0,
      open: openCounts.get(current) ?? 0,
    });
    current = advance(current);
  }

  return buckets;
}
