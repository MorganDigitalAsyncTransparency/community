// Spec: specs/dashboard/topic-intake.md

import { describe, expect, it } from "vitest";
import {
  computeIntakeBuckets,
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
// computeIntakeBuckets — daily (TI-3)
// ---------------------------------------------------------------------------

describe("computeIntakeBuckets — daily", () => {
  // TI-3 — topics on same day produce one bucket
  it("groups topics on the same day into a single bucket", () => {
    const topics = [
      makeTopic({ createdAt: utcNoon(2025, 3, 5) }),
      makeTopic({ id: 2, createdAt: utcNoon(2025, 3, 5) }),
    ];
    const result = computeIntakeBuckets(topics, "daily");
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
    const result = computeIntakeBuckets(topics, "daily");
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
    const result = computeIntakeBuckets(topics, "daily");
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
    const result = computeIntakeBuckets(topics, "daily");
    expect(result[0].bucketKey).toBe("2025-03-05");
    expect(result[1].bucketKey).toBe("2025-03-06");
    expect(result[2].bucketKey).toBe("2025-03-07");
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
    const result = computeIntakeBuckets(topics, "weekly");
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
    const result = computeIntakeBuckets(topics, "weekly");
    expect(result).toHaveLength(1);
  });

  // TI-4 — Sunday and Monday on week boundary land in different buckets
  it("Sunday and next Monday land in different buckets", () => {
    const topics = [
      makeTopic({ createdAt: utcNoon(2025, 3, 9) }),   // Sunday of W10
      makeTopic({ id: 2, createdAt: utcNoon(2025, 3, 10) }), // Monday of W11
    ];
    const result = computeIntakeBuckets(topics, "weekly");
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
    const result = computeIntakeBuckets(topics, "weekly");
    expect(result[0].bucketKey).toBe("2025-03-03");
    expect(result[1].bucketKey).toBe("2025-03-10");
    expect(result[2].bucketKey).toBe("2025-03-17");
  });
});

// ---------------------------------------------------------------------------
// computeIntakeBuckets — shared (TI-1, TI-14)
// ---------------------------------------------------------------------------

describe("computeIntakeBuckets — shared", () => {
  // TI-14 — empty input returns empty array
  it("returns an empty array when given no topics", () => {
    expect(computeIntakeBuckets([], "daily")).toEqual([]);
    expect(computeIntakeBuckets([], "weekly")).toEqual([]);
  });

  // TI-1 — does not mutate input array
  it("does not mutate the input array", () => {
    const topics = [
      makeTopic({ createdAt: utcNoon(2025, 3, 5) }),
      makeTopic({ id: 2, createdAt: utcNoon(2025, 3, 6) }),
    ];
    const original = [...topics];
    computeIntakeBuckets(topics, "daily");
    expect(topics).toEqual(original);
  });
});
