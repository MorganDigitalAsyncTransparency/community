// Spec: specs/dashboard/response-metrics.md
// Tests: tests/dashboard/response-metrics.unit.test.ts

import type { Topic } from "../mock/data";

const MILLISECONDS_PER_HOUR = 3_600_000;
const HOURS_PER_DAY = 24;

export function formatDuration(ms: number): string {
  const hours = Math.floor(ms / MILLISECONDS_PER_HOUR);

  if (hours >= HOURS_PER_DAY) {
    return `${Math.floor(hours / HOURS_PER_DAY)}d`;
  }

  return `${Math.max(1, hours)}h`;
}

function median(sorted: number[]): number {
  const mid = Math.floor(sorted.length / 2);

  if (sorted.length % 2 !== 0) {
    return sorted[mid];
  }

  return Math.trunc((sorted[mid - 1] + sorted[mid]) / 2);
}

export function medianFirstReplyTime(topics: Topic[]): string {
  const durations = topics
    .filter((t): t is Topic & { firstReplyAt: string } => t.firstReplyAt != null)
    .map((t) => new Date(t.firstReplyAt).getTime() - new Date(t.createdAt).getTime())
    .sort((a, b) => a - b);

  if (durations.length === 0) {
    return "–";
  }

  return formatDuration(median(durations));
}

export function medianResolutionTime(topics: Topic[]): string {
  if (topics.length === 0) {
    return "–";
  }

  const durations = topics
    .filter((t): t is Topic & { resolvedAt: string } => t.resolvedAt != null)
    .map((t) => new Date(t.resolvedAt).getTime() - new Date(t.createdAt).getTime())
    .sort((a, b) => a - b);

  if (durations.length === 0) {
    return "–";
  }

  return formatDuration(median(durations));
}

export interface OutcomeCounts {
  solved: number;
  selfClosed: number;
}

export function outcomeCounts(topics: Topic[]): OutcomeCounts {
  let solved = 0;
  let selfClosed = 0;

  for (const topic of topics) {
    if (topic.outcome === "solved") {
      solved++;
    } else if (topic.outcome === "self-closed") {
      selfClosed++;
    }
  }

  return { solved, selfClosed };
}

export function answerRate(topics: Topic[]): string {
  if (topics.length === 0) {
    return "–";
  }

  const { solved } = outcomeCounts(topics);
  return `${Math.round((solved / topics.length) * 100)}%`;
}
