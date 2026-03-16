// Spec: specs/dashboard/response-time-trends.md
// Tests: tests/dashboard/response-time-trends.unit.test.ts

import type { Topic } from "../mock/data";
import { medianFirstReplyTime, medianResolutionTime } from "./responseMetrics";

export interface WeeklyTrend {
  weekStart: string; // Monday of the week, YYYY-MM-DD (UTC)
  topicCount: number;
  medianFirstReply: string; // formatted duration or "–"
  medianResolution: string; // formatted duration or "–"
}

// Returns the YYYY-MM-DD of the Monday that begins the ISO week containing `date`.
// Week boundaries are computed in UTC so they match how topic timestamps are stored.
function mondayOf(date: Date): string {
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
