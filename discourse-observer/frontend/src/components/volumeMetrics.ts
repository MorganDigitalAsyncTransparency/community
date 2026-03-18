// Spec: specs/dashboard/response-metrics.md
// Tests: tests/dashboard/volume-metrics.unit.test.ts

import type { Topic } from "../mock/data";
import type { IntakeGranularity, TimeRange } from "./intakeMetrics";
import {
  advanceBucket,
  bucketKeyFor,
  formatBucketLabel,
} from "./bucketHelpers";

export interface VolumeBucket {
  label: string;
  bucketKey: string;
  created: number;
  accepted: number;
  closed: number;
  open: number;
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

  const advance = advanceBucket(granularity);
  const formatLabel = formatBucketLabel(granularity);

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
