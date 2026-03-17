// Spec: specs/dashboard/response-time-distribution.md
// Tests: tests/dashboard/response-time-distribution.unit.test.ts

import type { Topic } from "../mock/data";

const MILLISECONDS_PER_HOUR = 3_600_000;
const HOURS_PER_DAY = 24;

export interface DistributionBucket {
  label: string;
  count: number;
}

export function formatBucketCeiling(hours: number): string {
  if (hours % HOURS_PER_DAY === 0) {
    return `${hours / HOURS_PER_DAY}d`;
  }
  return `${hours}h`;
}

function buildBucketLabels(ceilings: number[]): string[] {
  const labels: string[] = [];

  labels.push(`< ${formatBucketCeiling(ceilings[0])}`);

  for (let i = 1; i < ceilings.length; i++) {
    labels.push(
      `${formatBucketCeiling(ceilings[i - 1])}–${formatBucketCeiling(ceilings[i])}`,
    );
  }

  labels.push(`> ${formatBucketCeiling(ceilings[ceilings.length - 1])}`);

  return labels;
}

export function bucketDurations(
  durationsMs: number[],
  ceilingsHours: number[],
): DistributionBucket[] {
  const ceilingsMs = ceilingsHours.map((h) => h * MILLISECONDS_PER_HOUR);
  const labels = buildBucketLabels(ceilingsHours);
  const counts = new Array<number>(labels.length).fill(0);

  for (const duration of durationsMs) {
    let placed = false;
    for (let i = 0; i < ceilingsMs.length; i++) {
      if (duration < ceilingsMs[i]) {
        counts[i]++;
        placed = true;
        break;
      }
    }
    if (!placed) {
      counts[counts.length - 1]++;
    }
  }

  return labels.map((label, i) => ({ label, count: counts[i] }));
}

export function firstReplyDurations(topics: Topic[]): number[] {
  return topics
    .filter((t): t is Topic & { firstReplyAt: string } => t.firstReplyAt != null)
    .map((t) => new Date(t.firstReplyAt).getTime() - new Date(t.createdAt).getTime());
}

export function resolutionDurations(topics: Topic[]): number[] {
  return topics
    .filter((t): t is Topic & { resolvedAt: string } => t.resolvedAt != null)
    .map((t) => new Date(t.resolvedAt).getTime() - new Date(t.createdAt).getTime());
}
