// Spec: specs/dashboard/time-period-filter.md
// Tests: tests/dashboard/time-period-filter.unit.test.ts

import type { Topic } from "../mock/data";

export type PeriodPreset = "last7" | "last30" | "lastYear" | "allTime";

export interface CustomRange {
  from: string; // YYYY-MM-DD
  to: string;   // YYYY-MM-DD
}

export type ActivePeriod =
  | { kind: "preset"; preset: PeriodPreset }
  | { kind: "custom"; range: CustomRange };

const MS_PER_DAY = 86_400_000;

const PRESET_DAYS: Record<Exclude<PeriodPreset, "allTime">, number> = {
  last7: 7,
  last30: 30,
  lastYear: 365,
};

export const PRESET_LABELS: Record<PeriodPreset, string> = {
  last7: "Last 7 days",
  last30: "Last 30 days",
  lastYear: "Last year",
  allTime: "All time",
};

export function filterByPeriod(topics: Topic[], period: ActivePeriod): Topic[] {
  if (period.kind === "preset") {
    if (period.preset === "allTime") {
      return topics;
    }

    const cutoff = Date.now() - PRESET_DAYS[period.preset] * MS_PER_DAY;
    return topics.filter((t) => new Date(t.createdAt).getTime() >= cutoff);
  }

  // Custom range: boundaries are in UTC so they match how topic timestamps are stored
  const from = new Date(period.range.from + "T00:00:00Z").getTime();
  const to = new Date(period.range.to + "T23:59:59.999Z").getTime();
  return topics.filter((t) => {
    const created = new Date(t.createdAt).getTime();
    return created >= from && created <= to;
  });
}
