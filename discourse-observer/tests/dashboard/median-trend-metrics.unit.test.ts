// Spec: specs/dashboard/response-metrics.md

import { describe, expect, it } from "vitest";
import {
  computeMedianFirstReplyBuckets,
  computeMedianResolutionBuckets,
} from "../../frontend/src/components/medianTrendMetrics";
import type { Topic } from "../../frontend/src/mock/data";
import type { TimeRange } from "../../frontend/src/components/intakeMetrics";

function makeTopic(overrides: Partial<Topic> & { createdAt: string }): Topic {
  return {
    id: 1,
    title: "Test topic",
    tags: [],
    categoryName: "Support",
    replyCount: 0,
    ...overrides,
  };
}

function utcNoon(year: number, month: number, day: number): string {
  return new Date(Date.UTC(year, month - 1, day, 12, 0, 0)).toISOString();
}

function hoursAfter(base: string, hours: number): string {
  return new Date(new Date(base).getTime() + hours * 3_600_000).toISOString();
}

function daysAfter(base: string, days: number): string {
  return new Date(new Date(base).getTime() + days * 86_400_000).toISOString();
}

describe("computeMedianFirstReplyBuckets", () => {
  it("returns empty when range is null", () => {
    expect(computeMedianFirstReplyBuckets([], "daily", null)).toEqual([]);
  });

  it("returns undefined medianHours for buckets with no first reply data", () => {
    const topic = makeTopic({ createdAt: utcNoon(2025, 3, 10) });
    const range: TimeRange = { first: "2025-03-10", last: "2025-03-10" };

    const result = computeMedianFirstReplyBuckets([topic], "daily", range);

    expect(result).toHaveLength(1);
    expect(result[0].medianHours).toBeUndefined();
  });

  it("computes median first reply hours for a bucket with data", () => {
    const created = utcNoon(2025, 3, 10);
    const topics = [
      makeTopic({ id: 1, createdAt: created, firstReplyAt: hoursAfter(created, 6) }),
      makeTopic({ id: 2, createdAt: created, firstReplyAt: hoursAfter(created, 10) }),
    ];
    const range: TimeRange = { first: "2025-03-10", last: "2025-03-10" };

    const result = computeMedianFirstReplyBuckets(topics, "daily", range);

    // median of 6h and 10h = 8h
    expect(result[0].medianHours).toBe(8);
  });

  it("fills gaps between buckets with undefined medianHours", () => {
    const topic = makeTopic({
      createdAt: utcNoon(2025, 3, 10),
      firstReplyAt: hoursAfter(utcNoon(2025, 3, 10), 4),
    });
    const range: TimeRange = { first: "2025-03-10", last: "2025-03-12" };

    const result = computeMedianFirstReplyBuckets([topic], "daily", range);

    expect(result).toHaveLength(3);
    expect(result[0].medianHours).toBe(4);
    expect(result[1].medianHours).toBeUndefined();
    expect(result[2].medianHours).toBeUndefined();
  });
});

describe("computeMedianResolutionBuckets", () => {
  it("returns empty when range is null", () => {
    expect(computeMedianResolutionBuckets([], "daily", null)).toEqual([]);
  });

  it("computes median resolution hours for a bucket with data", () => {
    const created = utcNoon(2025, 3, 10);
    const topics = [
      makeTopic({
        id: 1,
        createdAt: created,
        resolvedAt: daysAfter(created, 2),
        outcome: "solved",
      }),
      makeTopic({
        id: 2,
        createdAt: created,
        resolvedAt: daysAfter(created, 4),
        outcome: "solved",
      }),
    ];
    const range: TimeRange = { first: "2025-03-10", last: "2025-03-10" };

    const result = computeMedianResolutionBuckets(topics, "daily", range);

    // median of 48h and 96h = 72h
    expect(result[0].medianHours).toBe(72);
  });

  it("excludes topics without resolvedAt", () => {
    const created = utcNoon(2025, 3, 10);
    const topics = [
      makeTopic({ id: 1, createdAt: created }), // no resolvedAt
      makeTopic({
        id: 2,
        createdAt: created,
        resolvedAt: daysAfter(created, 3),
        outcome: "solved",
      }),
    ];
    const range: TimeRange = { first: "2025-03-10", last: "2025-03-10" };

    const result = computeMedianResolutionBuckets(topics, "daily", range);

    // Only one topic with resolution: 72h
    expect(result[0].medianHours).toBe(72);
  });

  it("returns undefined for buckets where no topics have resolvedAt", () => {
    const topic = makeTopic({ createdAt: utcNoon(2025, 3, 10) });
    const range: TimeRange = { first: "2025-03-10", last: "2025-03-10" };

    const result = computeMedianResolutionBuckets([topic], "daily", range);

    expect(result[0].medianHours).toBeUndefined();
  });
});
