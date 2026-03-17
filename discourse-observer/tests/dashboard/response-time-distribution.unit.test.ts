// Spec: specs/dashboard/response-time-distribution.md

import { describe, expect, it } from "vitest";
import {
  bucketDurations,
  firstReplyDurations,
  formatBucketCeiling,
  resolutionDurations,
} from "../../frontend/src/components/distributionMetrics";
import type { Topic } from "../../frontend/src/mock/data";

const HOUR_MS = 3_600_000;

function makeTopic(overrides: Partial<Topic> & { id: number }): Topic {
  return {
    title: "Test topic",
    createdAt: "2025-01-06T10:00:00Z",
    tags: ["api"],
    category: "Support",
    replyCount: 0,
    ...overrides,
  };
}

// ---------------------------------------------------------------------------
// formatBucketCeiling (RD-4)
// ---------------------------------------------------------------------------

describe("formatBucketCeiling", () => {
  it("returns 'Xh' for hours less than 24", () => {
    expect(formatBucketCeiling(1)).toBe("1h");
    expect(formatBucketCeiling(4)).toBe("4h");
    expect(formatBucketCeiling(12)).toBe("12h");
  });

  it("returns 'Xd' for multiples of 24", () => {
    expect(formatBucketCeiling(24)).toBe("1d");
    expect(formatBucketCeiling(48)).toBe("2d");
    expect(formatBucketCeiling(168)).toBe("7d");
  });

  it("returns 'Xh' for non-multiples of 24 above 24", () => {
    expect(formatBucketCeiling(36)).toBe("36h");
  });
});

// ---------------------------------------------------------------------------
// bucketDurations (RD-2, RD-3, RD-4)
// ---------------------------------------------------------------------------

describe("bucketDurations", () => {
  const ceilings = [1, 4, 12, 24, 48, 96, 168];

  it("returns all-zero buckets for empty input", () => {
    const result = bucketDurations([], ceilings);
    expect(result.length).toBe(8);
    for (const bucket of result) {
      expect(bucket.count).toBe(0);
    }
  });

  it("places durations in correct buckets", () => {
    const durations = [
      0.5 * HOUR_MS,  // < 1h
      2 * HOUR_MS,    // 1–4h
      6 * HOUR_MS,    // 4–12h
      20 * HOUR_MS,   // 12h–1d
      30 * HOUR_MS,   // 1–2d
      72 * HOUR_MS,   // 2–4d
      120 * HOUR_MS,  // 4–7d
      200 * HOUR_MS,  // > 7d
    ];
    const result = bucketDurations(durations, ceilings);

    expect(result[0]).toEqual({ label: "< 1h", count: 1 });
    expect(result[1]).toEqual({ label: "1h–4h", count: 1 });
    expect(result[2]).toEqual({ label: "4h–12h", count: 1 });
    expect(result[3]).toEqual({ label: "12h–1d", count: 1 });
    expect(result[4]).toEqual({ label: "1d–2d", count: 1 });
    expect(result[5]).toEqual({ label: "2d–4d", count: 1 });
    expect(result[6]).toEqual({ label: "4d–7d", count: 1 });
    expect(result[7]).toEqual({ label: "> 7d", count: 1 });
  });

  it("places duration exceeding all ceilings in last bucket", () => {
    const result = bucketDurations([500 * HOUR_MS], ceilings);
    expect(result[7].count).toBe(1);
    expect(result[7].label).toBe("> 7d");
  });

  it("places duration exactly on a ceiling boundary in the lower bucket", () => {
    // Exactly 1 hour = 1 * HOUR_MS — this is NOT less than ceiling[0] (1h)
    // so it should fall into the next bucket (1–4h)
    const result = bucketDurations([1 * HOUR_MS], ceilings);
    expect(result[0].count).toBe(0); // < 1h
    expect(result[1].count).toBe(1); // 1–4h
  });

  it("generates correct labels", () => {
    const result = bucketDurations([], ceilings);
    const labels = result.map((b) => b.label);
    expect(labels).toEqual([
      "< 1h",
      "1h–4h",
      "4h–12h",
      "12h–1d",
      "1d–2d",
      "2d–4d",
      "4d–7d",
      "> 7d",
    ]);
  });

  it("produces two buckets for a single ceiling", () => {
    const result = bucketDurations([0.5 * HOUR_MS, 5 * HOUR_MS], [4]);
    expect(result.length).toBe(2);
    expect(result[0]).toEqual({ label: "< 4h", count: 1 });
    expect(result[1]).toEqual({ label: "> 4h", count: 1 });
  });

  it("does not mutate input array", () => {
    const durations = [2 * HOUR_MS, 50 * HOUR_MS];
    const copy = [...durations];
    bucketDurations(durations, ceilings);
    expect(durations).toEqual(copy);
  });
});

// ---------------------------------------------------------------------------
// firstReplyDurations (RD-5)
// ---------------------------------------------------------------------------

describe("firstReplyDurations", () => {
  it("excludes topics without firstReplyAt", () => {
    const topics = [
      makeTopic({ id: 1 }),
      makeTopic({ id: 2, firstReplyAt: "2025-01-06T12:00:00Z" }),
    ];
    const result = firstReplyDurations(topics);
    expect(result.length).toBe(1);
  });

  it("computes correct durations in ms", () => {
    const topics = [
      makeTopic({
        id: 1,
        createdAt: "2025-01-06T10:00:00Z",
        firstReplyAt: "2025-01-06T13:00:00Z",
      }),
    ];
    const result = firstReplyDurations(topics);
    expect(result[0]).toBe(3 * HOUR_MS);
  });
});

// ---------------------------------------------------------------------------
// resolutionDurations (RD-6)
// ---------------------------------------------------------------------------

describe("resolutionDurations", () => {
  it("excludes topics without resolvedAt", () => {
    const topics = [
      makeTopic({ id: 1 }),
      makeTopic({ id: 2, resolvedAt: "2025-01-07T10:00:00Z" }),
    ];
    const result = resolutionDurations(topics);
    expect(result.length).toBe(1);
  });

  it("computes correct durations in ms", () => {
    const topics = [
      makeTopic({
        id: 1,
        createdAt: "2025-01-06T10:00:00Z",
        resolvedAt: "2025-01-08T10:00:00Z",
      }),
    ];
    const result = resolutionDurations(topics);
    expect(result[0]).toBe(48 * HOUR_MS);
  });
});
