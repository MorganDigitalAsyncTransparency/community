// Spec: specs/dashboard/response-time-trends.md
// Tests: tests/dashboard/response-time-trends.unit.test.ts

import type { Topic } from "../mock/data";
import { medianFirstReplyTime, medianResolutionTime } from "./responseMetrics";
import { formatWeekLabel } from "./topicFormatting";

export interface WeeklyTrend {
  weekStart: string; // Monday of the week, YYYY-MM-DD (UTC)
  topicCount: number;
  medianFirstReply: string; // formatted duration or "–"
  medianResolution: string; // formatted duration or "–"
}

export interface TrendChartPoint {
  weekLabel: string; // locale-formatted Monday date for X-axis
  medianFirstReplyHours: number | undefined; // hours, or undefined for gaps
  medianResolutionHours: number | undefined; // hours, or undefined for gaps
}

// Returns the YYYY-MM-DD of the Monday that begins the ISO week containing `date`.
// Week boundaries are computed in UTC so they match how topic timestamps are stored.
// Exported so tagMetrics.ts can reuse the same bucketing logic without duplication.
export function mondayOf(date: Date): string {
  const dayOfWeek = date.getUTCDay(); // 0 = Sunday, 1 = Monday, …, 6 = Saturday
  const daysToMonday = dayOfWeek === 0 ? 6 : dayOfWeek - 1;
  const monday = new Date(date);
  monday.setUTCDate(date.getUTCDate() - daysToMonday);
  monday.setUTCHours(0, 0, 0, 0);
  return monday.toISOString().slice(0, 10);
}

// Returns one WeeklyTrend per calendar week that contains at least one topic,
// ordered newest first. Accepts a pre-filtered Topic[] so callers can scope to
// a tag, category, or any other subset without this function needing to know why.
export function computeWeeklyTrends(topics: Topic[]): WeeklyTrend[] {
  const byWeek = new Map<string, Topic[]>();

  for (const topic of topics) {
    const week = mondayOf(new Date(topic.createdAt));
    const bucket = byWeek.get(week) ?? [];
    bucket.push(topic);
    byWeek.set(week, bucket);
  }

  return [...byWeek.entries()]
    .sort(([a], [b]) => b.localeCompare(a)) // newest first
    .map(([weekStart, weekTopics]) => ({
      weekStart,
      topicCount: weekTopics.length,
      medianFirstReply: medianFirstReplyTime(weekTopics),
      medianResolution: medianResolutionTime(weekTopics),
    }));
}

// Converts a formatted duration string ("3d", "12h") to numeric hours.
// Returns undefined for "–" (no data), which Recharts renders as a line gap.
export function parseDurationToHours(formatted: string): number | undefined {
  if (formatted === "–") {
    return undefined;
  }

  const match = formatted.match(/^(\d+)(d|h)$/);
  if (!match) {
    return undefined;
  }

  const value = Number(match[1]);
  return match[2] === "d" ? value * 24 : value;
}

// Transforms WeeklyTrend[] (newest-first) into TrendChartPoint[] (oldest-first)
// with numeric hour values suitable for Recharts line chart plotting.
export function weeklyTrendsChartData(trends: WeeklyTrend[]): TrendChartPoint[] {
  return [...trends].reverse().map((trend) => ({
    weekLabel: formatWeekLabel(trend.weekStart),
    medianFirstReplyHours: parseDurationToHours(trend.medianFirstReply),
    medianResolutionHours: parseDurationToHours(trend.medianResolution),
  }));
}
