// Spec: specs/dashboard/topic-intake.md (TI-2, TI-3, TI-4)
// Tests: tests/dashboard/topic-intake.unit.test.ts (bucketing via computeIntakeBuckets)

import type { IntakeGranularity } from "./intakeMetrics";
import { mondayOf } from "./trendMetrics";
import { formatWeekLabel } from "./topicFormatting";

export function dayOf(date: Date): string {
  return date.toISOString().slice(0, 10);
}

export function formatDayLabel(isoDate: string): string {
  return new Date(isoDate + "T00:00:00Z").toLocaleDateString(undefined, {
    year: "numeric",
    month: "short",
    day: "numeric",
    timeZone: "UTC",
  });
}

export function bucketKeyFor(date: Date, granularity: IntakeGranularity): string {
  return granularity === "daily" ? dayOf(date) : mondayOf(date);
}

export function nextDay(isoDate: string): string {
  const d = new Date(isoDate + "T00:00:00Z");
  d.setUTCDate(d.getUTCDate() + 1);
  return dayOf(d);
}

export function nextMonday(isoDate: string): string {
  const d = new Date(isoDate + "T00:00:00Z");
  d.setUTCDate(d.getUTCDate() + 7);
  return dayOf(d);
}

export function formatBucketLabel(
  granularity: IntakeGranularity,
): (isoDate: string) => string {
  return granularity === "daily" ? formatDayLabel : formatWeekLabel;
}

export function advanceBucket(
  granularity: IntakeGranularity,
): (isoDate: string) => string {
  return granularity === "daily" ? nextDay : nextMonday;
}
