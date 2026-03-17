// Spec: specs/dashboard/peak-activity.md
// Tests: tests/dashboard/peak-activity.unit.test.ts

import type { Topic } from "../mock/data";

export interface HeatmapCell {
  day: number;
  hour: number;
  count: number;
}

export interface HeatmapData {
  cells: HeatmapCell[][];
  maxCount: number;
}

export const DAY_LABELS = ["Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"];

function utcDayMondayBased(date: Date): number {
  const jsDay = date.getUTCDay();
  return jsDay === 0 ? 6 : jsDay - 1;
}

export function computeHeatmapData(topics: Topic[]): HeatmapData {
  const cells: HeatmapCell[][] = Array.from({ length: 7 }, (_, day) =>
    Array.from({ length: 24 }, (_, hour) => ({ day, hour, count: 0 })),
  );

  for (const topic of topics) {
    const date = new Date(topic.createdAt);
    const day = utcDayMondayBased(date);
    const hour = date.getUTCHours();
    cells[day][hour].count += 1;
  }

  let maxCount = 0;
  for (const row of cells) {
    for (const cell of row) {
      if (cell.count > maxCount) maxCount = cell.count;
    }
  }

  return { cells, maxCount };
}
