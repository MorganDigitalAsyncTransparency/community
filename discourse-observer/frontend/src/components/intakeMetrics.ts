// Spec: specs/dashboard/topic-intake.md
// Tests: tests/dashboard/topic-intake.unit.test.ts

import type { Topic } from "../mock/data";
import type { ActivePeriod } from "./timePeriod";
import { mondayOf } from "./trendMetrics";
import { formatWeekLabel } from "./topicFormatting";

export type IntakeGranularity = "daily" | "weekly";

export interface IntakeBucket {
  label: string;
  count: number;
  bucketKey: string;
}

const GRANULARITY_THRESHOLD_DAYS = 90;

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

export function computeIntakeBuckets(
  topics: Topic[],
  granularity: IntakeGranularity,
): IntakeBucket[] {
  const byBucket = new Map<string, number>();

  for (const topic of topics) {
    const date = new Date(topic.createdAt);
    const key = granularity === "daily" ? dayOf(date) : mondayOf(date);
    byBucket.set(key, (byBucket.get(key) ?? 0) + 1);
  }

  const formatLabel = granularity === "daily" ? formatDayLabel : formatWeekLabel;

  return [...byBucket.entries()]
    .sort(([a], [b]) => a.localeCompare(b))
    .map(([bucketKey, count]) => ({
      label: formatLabel(bucketKey),
      count,
      bucketKey,
    }));
}
