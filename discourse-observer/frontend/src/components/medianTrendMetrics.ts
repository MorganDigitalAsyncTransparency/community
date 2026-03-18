// Spec: specs/dashboard/response-metrics.md
// Tests: tests/dashboard/median-trend-metrics.unit.test.ts

import type { Topic } from "../mock/data";
import type { IntakeGranularity, TimeRange } from "./intakeMetrics";
import { median } from "./responseMetrics";
import {
  advanceBucket,
  bucketKeyFor,
  formatBucketLabel,
} from "./bucketHelpers";

export interface MedianBucket {
  label: string;
  bucketKey: string;
  medianHours: number | undefined;
}

const MILLISECONDS_PER_HOUR = 3_600_000;

type DurationExtractor = (topic: Topic) => number | null;

function firstReplyDuration(topic: Topic): number | null {
  if (!topic.firstReplyAt) return null;
  return new Date(topic.firstReplyAt).getTime() - new Date(topic.createdAt).getTime();
}

function resolutionDuration(topic: Topic): number | null {
  if (!topic.resolvedAt) return null;
  return new Date(topic.resolvedAt).getTime() - new Date(topic.createdAt).getTime();
}

function computeMedianPerBucket(
  topics: Topic[],
  extractor: DurationExtractor,
  granularity: IntakeGranularity,
  range: TimeRange,
): MedianBucket[] {
  const byBucket = new Map<string, number[]>();

  for (const topic of topics) {
    const key = bucketKeyFor(new Date(topic.createdAt), granularity);
    const duration = extractor(topic);
    if (duration !== null) {
      const bucket = byBucket.get(key) ?? [];
      bucket.push(duration);
      byBucket.set(key, bucket);
    }
  }

  const advance = advanceBucket(granularity);
  const formatLabel = formatBucketLabel(granularity);

  const buckets: MedianBucket[] = [];
  let current = range.first;
  while (current <= range.last) {
    const durations = byBucket.get(current);
    const medianHours = durations
      ? median(durations.sort((a, b) => a - b)) / MILLISECONDS_PER_HOUR
      : undefined;
    buckets.push({
      label: formatLabel(current),
      bucketKey: current,
      medianHours,
    });
    current = advance(current);
  }

  return buckets;
}

export function computeMedianFirstReplyBuckets(
  topics: Topic[],
  granularity: IntakeGranularity,
  range: TimeRange | null,
): MedianBucket[] {
  if (!range) return [];
  return computeMedianPerBucket(topics, firstReplyDuration, granularity, range);
}

export function computeMedianResolutionBuckets(
  topics: Topic[],
  granularity: IntakeGranularity,
  range: TimeRange | null,
): MedianBucket[] {
  if (!range) return [];
  return computeMedianPerBucket(topics, resolutionDuration, granularity, range);
}
