// Spec: specs/dashboard/time-period-filter.md
// Tests: tests/dashboard/url-state.unit.test.ts

export type PeriodPreset = "last7" | "last30" | "lastYear" | "allTime";

export interface CustomRange {
  from: string; // YYYY-MM-DD
  to: string;   // YYYY-MM-DD
}

export type ActivePeriod =
  | { kind: "preset"; preset: PeriodPreset }
  | { kind: "custom"; range: CustomRange };

export const PRESET_LABELS: Record<PeriodPreset, string> = {
  last7: "Last 7 days",
  last30: "Last 30 days",
  lastYear: "Last year",
  allTime: "All time",
};
