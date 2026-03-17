// Spec: specs/dashboard/topic-intake.md

import { describe, expect, it } from "vitest";
import {
  computeIntakeBuckets,
  computeTimeRange,
  intakeGranularity,
} from "../../frontend/src/components/intakeMetrics";
import type { Topic } from "../../frontend/src/mock/data";
import type { ActivePeriod } from "../../frontend/src/components/timePeriod";

function makeTopic(overrides: Partial<Topic> & { createdAt: string }): Topic {
  return {
    id: 1,
    title: "Test topic",
    tags: [],
    category: "Support",
    replyCount: 0,
    ...overrides,
  };
}

// Returns an ISO string for a UTC date with time set to noon to avoid
// date-boundary artefacts in UTC-based bucketing.
function utcNoon(year: number, month: number, day: number): string {
  return new Date(Date.UTC(year, month - 1, day, 12, 0, 0)).toISOString();
}

// Helper: compute buckets using the topics' own time range (no external range).
function bucketsFromTopics(topics: Topic[], granularity: "daily" | "weekly") {
  return computeIntakeBuckets(topics, granularity, computeTimeRange(topics, granularity));
}

// ---------------------------------------------------------------------------
// intakeGranularity (TI-12)
// ---------------------------------------------------------------------------

describe("intakeGranularity", () => {
  // TI-12 — "last7" → daily
  it("returns daily for last7 preset", () => {
    const period: ActivePeriod = { kind: "preset", preset: "last7" };
    expect(intakeGranularity(period)).toBe("daily");
  });

  // TI-12 — "last30" → daily
  it("returns daily for last30 preset", () => {
    const period: ActivePeriod = { kind: "preset", preset: "last30" };
    expect(intakeGranularity(period)).toBe("daily");
  });

  // TI-12 — "lastYear" → weekly
  it("returns weekly for lastYear preset", () => {
    const period: ActivePeriod = { kind: "preset", preset: "lastYear" };
    expect(intakeGranularity(period)).toBe("weekly");
  });

  // TI-12 — "allTime" → weekly
  it("returns weekly for allTime preset", () => {
    const period: ActivePeriod = { kind: "preset", preset: "allTime" };
    expect(intakeGranularity(period)).toBe("weekly");
  });

  // TI-12 — custom range under 90 days → daily
  it("returns daily for custom range under 90 days", () => {
    const period: ActivePeriod = {
      kind: "custom",
      range: { from: "2025-01-01", to: "2025-03-01" },
    };
    expect(intakeGranularity(period)).toBe("daily");
  });

  // TI-12 — custom range of exactly 90 days → weekly
  it("returns weekly for custom range of exactly 90 days", () => {
    // Jan 1 to Apr 1 = 90 full days (from 00:00 Jan 1 to 23:59:59.999 Apr 1)
    const period: ActivePeriod = {
      kind: "custom",
      range: { from: "2025-01-01", to: "2025-04-01" },
    };
    expect(intakeGranularity(period)).toBe("weekly");
  });

  // TI-12 — custom range over 90 days → weekly
  it("returns weekly for custom range over 90 days", () => {
    const period: ActivePeriod = {
      kind: "custom",
      range: { from: "2025-01-01", to: "2025-06-01" },
    };
    expect(intakeGranularity(period)).toBe("weekly");
  });
});

// ---------------------------------------------------------------------------
// computeTimeRange
// ---------------------------------------------------------------------------

describe("computeTimeRange", () => {
  it("returns null for empty topics", () => {
    expect(computeTimeRange([], "daily")).toBeNull();
  });

  it("returns same first/last for single topic (daily)", () => {
    const topics = [makeTopic({ createdAt: utcNoon(2025, 3, 5) })];
    const range = computeTimeRange(topics, "daily");
    expect(range).toEqual({ first: "2025-03-05", last: "2025-03-05" });
  });

  it("returns earliest/latest bucket keys (weekly)", () => {
    const topics = [
      makeTopic({ createdAt: utcNoon(2025, 3, 3) }),  // W10
      makeTopic({ id: 2, createdAt: utcNoon(2025, 3, 17) }), // W12
    ];
    const range = computeTimeRange(topics, "weekly");
    expect(range).toEqual({ first: "2025-03-03", last: "2025-03-17" });
  });
});

// ---------------------------------------------------------------------------
// computeIntakeBuckets — daily (TI-3)
// ---------------------------------------------------------------------------

describe("computeIntakeBuckets — daily", () => {
  // TI-3 — topics on same day produce one bucket
  it("groups topics on the same day into a single bucket", () => {
    const topics = [
      makeTopic({ createdAt: utcNoon(2025, 3, 5) }),
      makeTopic({ id: 2, createdAt: utcNoon(2025, 3, 5) }),
    ];
    const result = bucketsFromTopics(topics, "daily");
    expect(result).toHaveLength(1);
    expect(result[0].count).toBe(2);
    expect(result[0].bucketKey).toBe("2025-03-05");
  });

  // TI-3 — topics on different days produce one bucket per day
  it("produces one bucket per distinct day", () => {
    const topics = [
      makeTopic({ createdAt: utcNoon(2025, 3, 5) }),
      makeTopic({ id: 2, createdAt: utcNoon(2025, 3, 6) }),
      makeTopic({ id: 3, createdAt: utcNoon(2025, 3, 7) }),
    ];
    const result = bucketsFromTopics(topics, "daily");
    expect(result).toHaveLength(3);
  });

  // TI-1 — bucket count matches topic count per day
  it("count matches the number of topics per day", () => {
    const topics = [
      makeTopic({ createdAt: utcNoon(2025, 3, 5) }),
      makeTopic({ id: 2, createdAt: utcNoon(2025, 3, 5) }),
      makeTopic({ id: 3, createdAt: utcNoon(2025, 3, 5) }),
      makeTopic({ id: 4, createdAt: utcNoon(2025, 3, 6) }),
    ];
    const result = bucketsFromTopics(topics, "daily");
    expect(result[0].count).toBe(3);
    expect(result[1].count).toBe(1);
  });

  // TI-9 — buckets are in chronological order
  it("returns buckets in chronological order (oldest first)", () => {
    const topics = [
      makeTopic({ createdAt: utcNoon(2025, 3, 7) }),
      makeTopic({ id: 2, createdAt: utcNoon(2025, 3, 5) }),
      makeTopic({ id: 3, createdAt: utcNoon(2025, 3, 6) }),
    ];
    const result = bucketsFromTopics(topics, "daily");
    expect(result[0].bucketKey).toBe("2025-03-05");
    expect(result[1].bucketKey).toBe("2025-03-06");
    expect(result[2].bucketKey).toBe("2025-03-07");
  });

  // Gap filling — days with no topics get count 0
  it("fills gaps between days with zero-count buckets", () => {
    const topics = [
      makeTopic({ createdAt: utcNoon(2025, 3, 5) }),
      makeTopic({ id: 2, createdAt: utcNoon(2025, 3, 8) }),
    ];
    const result = bucketsFromTopics(topics, "daily");
    expect(result).toHaveLength(4); // Mar 5, 6, 7, 8
    expect(result[0].bucketKey).toBe("2025-03-05");
    expect(result[0].count).toBe(1);
    expect(result[1].bucketKey).toBe("2025-03-06");
    expect(result[1].count).toBe(0);
    expect(result[2].bucketKey).toBe("2025-03-07");
    expect(result[2].count).toBe(0);
    expect(result[3].bucketKey).toBe("2025-03-08");
    expect(result[3].count).toBe(1);
  });
});

// ---------------------------------------------------------------------------
// computeIntakeBuckets — weekly (TI-4)
// ---------------------------------------------------------------------------

describe("computeIntakeBuckets — weekly", () => {
  // TI-4 — topics in same week produce one bucket
  it("groups topics in the same ISO week into a single bucket", () => {
    // 2025-W10: Mon 2025-03-03 to Sun 2025-03-09
    const topics = [
      makeTopic({ createdAt: utcNoon(2025, 3, 3) }),
      makeTopic({ id: 2, createdAt: utcNoon(2025, 3, 7) }),
    ];
    const result = bucketsFromTopics(topics, "weekly");
    expect(result).toHaveLength(1);
    expect(result[0].count).toBe(2);
    expect(result[0].bucketKey).toBe("2025-03-03");
  });

  // TI-4 — Monday and Sunday of same week land in same bucket
  it("Monday and Sunday of the same week land in the same bucket", () => {
    const topics = [
      makeTopic({ createdAt: utcNoon(2025, 3, 3) }),  // Monday
      makeTopic({ id: 2, createdAt: utcNoon(2025, 3, 9) }),  // Sunday
    ];
    const result = bucketsFromTopics(topics, "weekly");
    expect(result).toHaveLength(1);
    expect(result[0].count).toBe(2);
  });

  // TI-4 — Sunday and Monday on week boundary land in different buckets
  it("Sunday and next Monday land in different buckets", () => {
    const topics = [
      makeTopic({ createdAt: utcNoon(2025, 3, 9) }),   // Sunday of W10
      makeTopic({ id: 2, createdAt: utcNoon(2025, 3, 10) }), // Monday of W11
    ];
    const result = bucketsFromTopics(topics, "weekly");
    expect(result).toHaveLength(2);
    expect(result[0].bucketKey).toBe("2025-03-03"); // W10 Monday
    expect(result[1].bucketKey).toBe("2025-03-10"); // W11 Monday
  });

  // TI-9 — buckets are in chronological order
  it("returns buckets in chronological order", () => {
    const topics = [
      makeTopic({ createdAt: utcNoon(2025, 3, 17) }),  // W12
      makeTopic({ id: 2, createdAt: utcNoon(2025, 3, 3) }),  // W10
      makeTopic({ id: 3, createdAt: utcNoon(2025, 3, 10) }), // W11
    ];
    const result = bucketsFromTopics(topics, "weekly");
    expect(result[0].bucketKey).toBe("2025-03-03");
    expect(result[1].bucketKey).toBe("2025-03-10");
    expect(result[2].bucketKey).toBe("2025-03-17");
  });

  // Gap filling — weeks with no topics get count 0
  it("fills gaps between weeks with zero-count buckets", () => {
    // W10 (Mar 3) and W13 (Mar 24) — W11 and W12 should be filled
    const topics = [
      makeTopic({ createdAt: utcNoon(2025, 3, 3) }),   // W10
      makeTopic({ id: 2, createdAt: utcNoon(2025, 3, 24) }), // W13
    ];
    const result = bucketsFromTopics(topics, "weekly");
    expect(result).toHaveLength(4); // W10, W11, W12, W13
    expect(result[0].bucketKey).toBe("2025-03-03");
    expect(result[0].count).toBe(1);
    expect(result[1].bucketKey).toBe("2025-03-10");
    expect(result[1].count).toBe(0);
    expect(result[2].bucketKey).toBe("2025-03-17");
    expect(result[2].count).toBe(0);
    expect(result[3].bucketKey).toBe("2025-03-24");
    expect(result[3].count).toBe(1);
  });
});

// ---------------------------------------------------------------------------
// computeIntakeBuckets — global time range (TI-8a)
// ---------------------------------------------------------------------------

describe("computeIntakeBuckets — global time range", () => {
  // TI-8a — x-axis spans full range even when tag-filtered topics are sparse
  it("uses global range to fill zeros beyond tag-filtered data", () => {
    // Global range spans Mar 3 to Mar 10, but tag-filtered topics only on Mar 5
    const globalRange = { first: "2025-03-03", last: "2025-03-10" };
    const tagTopics = [makeTopic({ createdAt: utcNoon(2025, 3, 5) })];

    const result = computeIntakeBuckets(tagTopics, "daily", globalRange);
    expect(result).toHaveLength(8); // Mar 3–10 inclusive
    expect(result[0].bucketKey).toBe("2025-03-03");
    expect(result[0].count).toBe(0);
    expect(result[2].bucketKey).toBe("2025-03-05");
    expect(result[2].count).toBe(1);
    expect(result[7].bucketKey).toBe("2025-03-10");
    expect(result[7].count).toBe(0);
  });

  // TI-8a — empty tag-filtered topics still produce zero-filled range
  it("returns all-zero buckets when tag has no topics in global range", () => {
    const globalRange = { first: "2025-03-03", last: "2025-03-05" };
    const result = computeIntakeBuckets([], "daily", globalRange);
    expect(result).toHaveLength(3);
    expect(result.every((b) => b.count === 0)).toBe(true);
  });

  // TI-14 — null range returns empty array
  it("returns empty array when range is null", () => {
    expect(computeIntakeBuckets([], "daily", null)).toEqual([]);
    expect(computeIntakeBuckets([], "weekly", null)).toEqual([]);
  });
});

// ---------------------------------------------------------------------------
// computeIntakeBuckets — shared (TI-1)
// ---------------------------------------------------------------------------

describe("computeIntakeBuckets — shared", () => {
  // TI-1 — does not mutate input array
  it("does not mutate the input array", () => {
    const topics = [
      makeTopic({ createdAt: utcNoon(2025, 3, 5) }),
      makeTopic({ id: 2, createdAt: utcNoon(2025, 3, 6) }),
    ];
    const original = [...topics];
    bucketsFromTopics(topics, "daily");
    expect(topics).toEqual(original);
  });
});
