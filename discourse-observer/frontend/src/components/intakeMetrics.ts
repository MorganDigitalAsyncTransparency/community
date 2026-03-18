// Spec: specs/dashboard/topic-intake.md
// Tests: tests/dashboard/topic-intake.unit.test.ts

import type { Topic } from "../mock/data";
import type { ActivePeriod } from "./timePeriod";
import {
  advanceBucket,
  bucketKeyFor,
  formatBucketLabel,
} from "./bucketHelpers";

export type IntakeGranularity = "daily" | "weekly";

export interface IntakeBucket {
  label: string;
  count: number;
  bucketKey: string;
}

const GRANULARITY_THRESHOLD_DAYS = 90;

export function intakeGranularity(period: ActivePeriod): IntakeGranularity {
  if (period.kind === "preset") {
    return period.preset === "last7" || period.preset === "last30"
      ? "daily"
      : "weekly";
  }

  const from = new Date(period.range.from + "T00:00:00Z").getTime();
  const to = new Date(period.range.to + "T23:59:59.999Z").getTime();
  const spanDays = (to - from) / 86_400_000;
  return spanDays < GRANULARITY_THRESHOLD_DAYS ? "daily" : "weekly";
}

export interface TimeRange {
  first: string; // YYYY-MM-DD bucket key
  last: string;  // YYYY-MM-DD bucket key
}

export function computeTimeRange(
  topics: Topic[],
  granularity: IntakeGranularity,
): TimeRange | null {
  if (topics.length === 0) return null;

  let earliest = "";
  let latest = "";

  for (const topic of topics) {
    const key = bucketKeyFor(new Date(topic.createdAt), granularity);
    if (earliest === "" || key < earliest) earliest = key;
    if (latest === "" || key > latest) latest = key;
  }

  return { first: earliest, last: latest };
}

function fillRange(
  byBucket: Map<string, number>,
  granularity: IntakeGranularity,
  range: TimeRange,
): Map<string, number> {
  if (range.first > range.last) return new Map();

  const advance = advanceBucket(granularity);

  const filled = new Map<string, number>();
  let current = range.first;
  while (current <= range.last) {
    filled.set(current, byBucket.get(current) ?? 0);
    current = advance(current);
  }
  return filled;
}

export function computeIntakeBuckets(
  topics: Topic[],
  granularity: IntakeGranularity,
  range: TimeRange | null,
): IntakeBucket[] {
  if (!range) return [];

  const byBucket = new Map<string, number>();

  for (const topic of topics) {
    const key = bucketKeyFor(new Date(topic.createdAt), granularity);
    byBucket.set(key, (byBucket.get(key) ?? 0) + 1);
  }

  const filled = fillRange(byBucket, granularity, range);
  const formatLabel = formatBucketLabel(granularity);

  return [...filled.entries()]
    .sort(([a], [b]) => a.localeCompare(b))
    .map(([bucketKey, count]) => ({
      label: formatLabel(bucketKey),
      count,
      bucketKey,
    }));
}
