import { describe, expect, it } from "vitest";
import {
  medianFirstReplyTime,
  medianResolutionTime,
  outcomeCounts,
  formatOutcomes,
  answerRate,
  formatDuration,
} from "../../frontend/src/components/responseMetrics";
import type { Topic } from "../../frontend/src/mock/data";

const HOUR_MS = 3_600_000;
const DAY_MS = 86_400_000;

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

function isoPlus(base: string, ms: number): string {
  return new Date(new Date(base).getTime() + ms).toISOString();
}

// ---------------------------------------------------------------------------
// formatDuration (RM-13)
// ---------------------------------------------------------------------------
describe("formatDuration", () => {
  it("returns days when duration is 24 hours or more", () => {
    expect(formatDuration(2 * DAY_MS)).toBe("2d");
  });

  it("truncates partial days", () => {
    expect(formatDuration(2.9 * DAY_MS)).toBe("2d");
  });

  it("returns hours when duration is less than 24 hours", () => {
    expect(formatDuration(8 * HOUR_MS)).toBe("8h");
  });

  it("returns minimum 1h for very short durations", () => {
    expect(formatDuration(5 * 60_000)).toBe("1h");
  });

  it("returns 1d at exactly 24 hours", () => {
    expect(formatDuration(24 * HOUR_MS)).toBe("1d");
  });
});

// ---------------------------------------------------------------------------
// medianFirstReplyTime (RM-1, RM-2, RM-12)
// ---------------------------------------------------------------------------
describe("medianFirstReplyTime", () => {
  it("returns dash for empty list", () => {
    expect(medianFirstReplyTime([])).toBe("–");
  });

  it("excludes topics without firstReplyAt", () => {
    const base = "2026-03-01T00:00:00Z";
    const topics = [
      makeTopic({
        id: 1,
        createdAt: base,
        resolvedAt: isoPlus(base, 3 * DAY_MS),
        outcome: "self-closed",
        // no firstReplyAt
      }),
    ];
    expect(medianFirstReplyTime(topics)).toBe("–");
  });

  it("returns formatted median for single topic with reply", () => {
    const base = "2026-03-01T00:00:00Z";
    const topics = [
      makeTopic({
        id: 1,
        createdAt: base,
        firstReplyAt: isoPlus(base, 3 * HOUR_MS),
        resolvedAt: isoPlus(base, DAY_MS),
        outcome: "solved",
      }),
    ];
    expect(medianFirstReplyTime(topics)).toBe("3h");
  });

  it("returns median for odd number of topics", () => {
    const base = "2026-03-01T00:00:00Z";
    const topics = [
      makeTopic({
        id: 1,
        createdAt: base,
        firstReplyAt: isoPlus(base, 2 * HOUR_MS),
        resolvedAt: isoPlus(base, DAY_MS),
        outcome: "solved",
      }),
      makeTopic({
        id: 2,
        createdAt: base,
        firstReplyAt: isoPlus(base, 6 * HOUR_MS),
        resolvedAt: isoPlus(base, DAY_MS),
        outcome: "solved",
      }),
      makeTopic({
        id: 3,
        createdAt: base,
        firstReplyAt: isoPlus(base, 48 * HOUR_MS),
        resolvedAt: isoPlus(base, 3 * DAY_MS),
        outcome: "solved",
      }),
    ];
    // sorted durations: 2h, 6h, 48h → median = 6h
    expect(medianFirstReplyTime(topics)).toBe("6h");
  });

  it("returns median for even number of topics", () => {
    const base = "2026-03-01T00:00:00Z";
    const topics = [
      makeTopic({
        id: 1,
        createdAt: base,
        firstReplyAt: isoPlus(base, 2 * HOUR_MS),
        resolvedAt: isoPlus(base, DAY_MS),
        outcome: "solved",
      }),
      makeTopic({
        id: 2,
        createdAt: base,
        firstReplyAt: isoPlus(base, 6 * HOUR_MS),
        resolvedAt: isoPlus(base, DAY_MS),
        outcome: "solved",
      }),
    ];
    // sorted durations: 2h, 6h → median = average = 4h
    expect(medianFirstReplyTime(topics)).toBe("4h");
  });
});

// ---------------------------------------------------------------------------
// medianResolutionTime (RM-3, RM-4, RM-12)
// ---------------------------------------------------------------------------
describe("medianResolutionTime", () => {
  it("returns dash for empty list", () => {
    expect(medianResolutionTime([])).toBe("–");
  });

  it("includes both solved and self-closed topics", () => {
    const base = "2026-03-01T00:00:00Z";
    const topics = [
      makeTopic({
        id: 1,
        createdAt: base,
        resolvedAt: isoPlus(base, 2 * DAY_MS),
        outcome: "solved",
        firstReplyAt: isoPlus(base, HOUR_MS),
      }),
      makeTopic({
        id: 2,
        createdAt: base,
        resolvedAt: isoPlus(base, 4 * DAY_MS),
        outcome: "self-closed",
      }),
    ];
    // sorted durations: 2d, 4d → median = 3d
    expect(medianResolutionTime(topics)).toBe("3d");
  });

  it("returns formatted median for single topic", () => {
    const base = "2026-03-01T00:00:00Z";
    const topics = [
      makeTopic({
        id: 1,
        createdAt: base,
        resolvedAt: isoPlus(base, 5 * DAY_MS),
        outcome: "solved",
        firstReplyAt: isoPlus(base, HOUR_MS),
      }),
    ];
    expect(medianResolutionTime(topics)).toBe("5d");
  });
});

// ---------------------------------------------------------------------------
// outcomeCounts (RM-5)
// ---------------------------------------------------------------------------
describe("outcomeCounts", () => {
  it("returns zero counts for empty list", () => {
    expect(outcomeCounts([])).toEqual({ solved: 0, selfClosed: 0 });
  });

  it("counts solved and self-closed correctly", () => {
    const base = "2026-03-01T00:00:00Z";
    const topics = [
      makeTopic({ id: 1, createdAt: base, outcome: "solved", resolvedAt: isoPlus(base, DAY_MS), firstReplyAt: isoPlus(base, HOUR_MS) }),
      makeTopic({ id: 2, createdAt: base, outcome: "solved", resolvedAt: isoPlus(base, DAY_MS), firstReplyAt: isoPlus(base, HOUR_MS) }),
      makeTopic({ id: 3, createdAt: base, outcome: "self-closed", resolvedAt: isoPlus(base, DAY_MS) }),
    ];
    expect(outcomeCounts(topics)).toEqual({ solved: 2, selfClosed: 1 });
  });

  it("handles all solved", () => {
    const base = "2026-03-01T00:00:00Z";
    const topics = [
      makeTopic({ id: 1, createdAt: base, outcome: "solved", resolvedAt: isoPlus(base, DAY_MS), firstReplyAt: isoPlus(base, HOUR_MS) }),
    ];
    expect(outcomeCounts(topics)).toEqual({ solved: 1, selfClosed: 0 });
  });
});

// ---------------------------------------------------------------------------
// formatOutcomes (RM-6, RM-7)
// ---------------------------------------------------------------------------
describe("formatOutcomes", () => {
  it("formats solved and self-closed counts", () => {
    expect(formatOutcomes({ solved: 12, selfClosed: 5 })).toBe("12 solved / 5 self-closed");
  });

  it("formats zero counts for empty input", () => {
    expect(formatOutcomes({ solved: 0, selfClosed: 0 })).toBe("0 solved / 0 self-closed");
  });
});

// ---------------------------------------------------------------------------
// answerRate (RM-8, RM-9)
// ---------------------------------------------------------------------------
describe("answerRate", () => {
  it("returns dash for empty list", () => {
    expect(answerRate([])).toBe("–");
  });

  it("returns percentage of solved topics", () => {
    const base = "2026-03-01T00:00:00Z";
    const topics = [
      makeTopic({ id: 1, createdAt: base, outcome: "solved", resolvedAt: isoPlus(base, DAY_MS), firstReplyAt: isoPlus(base, HOUR_MS) }),
      makeTopic({ id: 2, createdAt: base, outcome: "solved", resolvedAt: isoPlus(base, DAY_MS), firstReplyAt: isoPlus(base, HOUR_MS) }),
      makeTopic({ id: 3, createdAt: base, outcome: "self-closed", resolvedAt: isoPlus(base, DAY_MS) }),
    ];
    // 2/3 = 66.67% → rounded to 67%
    expect(answerRate(topics)).toBe("67%");
  });

  it("returns 100% when all solved", () => {
    const base = "2026-03-01T00:00:00Z";
    const topics = [
      makeTopic({ id: 1, createdAt: base, outcome: "solved", resolvedAt: isoPlus(base, DAY_MS), firstReplyAt: isoPlus(base, HOUR_MS) }),
    ];
    expect(answerRate(topics)).toBe("100%");
  });

  it("returns 0% when all self-closed", () => {
    const base = "2026-03-01T00:00:00Z";
    const topics = [
      makeTopic({ id: 1, createdAt: base, outcome: "self-closed", resolvedAt: isoPlus(base, DAY_MS) }),
    ];
    expect(answerRate(topics)).toBe("0%");
  });
});
