// Spec: specs/dashboard/response-time-trends.md

import { describe, expect, it } from "vitest";
import {
  computeWeeklyTrends,
  parseDurationToHours,
  weeklyTrendsChartData,
} from "../../frontend/src/components/trendMetrics";
import type { Topic } from "../../frontend/src/mock/data";

const DAY_MS = 86_400_000;
const HOUR_MS = 3_600_000;

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
// computeWeeklyTrends (RT-1 – RT-7, RT-10)
// ---------------------------------------------------------------------------

describe("computeWeeklyTrends", () => {
  // RT-1, RT-5
  it("returns an empty array when given no topics", () => {
    expect(computeWeeklyTrends([])).toEqual([]);
  });

  // RT-2, RT-4 — topics in the same week produce one row
  it("groups topics in the same ISO week into a single row", () => {
    // 2025-W10: Mon 2025-03-03 to Sun 2025-03-09
    const topics = [
      makeTopic({ createdAt: utcNoon(2025, 3, 3) }), // Monday
      makeTopic({ id: 2, createdAt: utcNoon(2025, 3, 7) }), // Friday
      makeTopic({ id: 3, createdAt: utcNoon(2025, 3, 9) }), // Sunday
    ];

    const result = computeWeeklyTrends(topics);

    expect(result).toHaveLength(1);
    expect(result[0].weekStart).toBe("2025-03-03");
    expect(result[0].topicCount).toBe(3);
  });

  // RT-1, RT-4 — topics in different weeks produce one row per week
  it("produces one row per distinct week when topics span multiple weeks", () => {
    const topics = [
      makeTopic({ createdAt: utcNoon(2025, 3, 3) }),  // W10
      makeTopic({ id: 2, createdAt: utcNoon(2025, 3, 10) }), // W11
      makeTopic({ id: 3, createdAt: utcNoon(2025, 3, 17) }), // W12
    ];

    expect(computeWeeklyTrends(topics)).toHaveLength(3);
  });

  // RT-3 — newest first
  it("orders rows newest week first", () => {
    const topics = [
      makeTopic({ createdAt: utcNoon(2025, 1, 6) }),  // W02
      makeTopic({ id: 2, createdAt: utcNoon(2025, 1, 20) }), // W04
      makeTopic({ id: 3, createdAt: utcNoon(2025, 1, 13) }), // W03
    ];

    const weeks = computeWeeklyTrends(topics).map((r) => r.weekStart);
    expect(weeks).toEqual(["2025-01-20", "2025-01-13", "2025-01-06"]);
  });

  // RT-4 — topicCount is correct
  it("sets topicCount to the number of topics in each week", () => {
    const topics = [
      makeTopic({ createdAt: utcNoon(2025, 3, 3) }),
      makeTopic({ id: 2, createdAt: utcNoon(2025, 3, 5) }),
      makeTopic({ id: 3, createdAt: utcNoon(2025, 3, 10) }),
    ];

    const result = computeWeeklyTrends(topics);
    const w10 = result.find((r) => r.weekStart === "2025-03-03");
    const w11 = result.find((r) => r.weekStart === "2025-03-10");

    expect(w10?.topicCount).toBe(2);
    expect(w11?.topicCount).toBe(1);
  });

  // RT-6 — no firstReplyAt → "–" for first reply
  it("shows '–' for median first reply when no topics in the week have firstReplyAt", () => {
    const topics = [makeTopic({ createdAt: utcNoon(2025, 3, 3) })];

    const [row] = computeWeeklyTrends(topics);
    expect(row.medianFirstReply).toBe("–");
  });

  // RT-7 — no resolvedAt → "–" for resolution
  it("shows '–' for median resolution when no topics in the week have resolvedAt", () => {
    const topics = [makeTopic({ createdAt: utcNoon(2025, 3, 3) })];

    const [row] = computeWeeklyTrends(topics);
    expect(row.medianResolution).toBe("–");
  });

  // RT-6, RT-10 — firstReplyAt present → formatted duration
  it("computes median first reply from firstReplyAt when present", () => {
    const created = utcNoon(2025, 3, 3);
    const firstReply = new Date(new Date(created).getTime() + 2 * DAY_MS).toISOString();
    const topics = [makeTopic({ createdAt: created, firstReplyAt: firstReply })];

    const [row] = computeWeeklyTrends(topics);
    expect(row.medianFirstReply).toBe("2d");
  });

  // RT-7, RT-10 — resolvedAt present → formatted duration
  it("computes median resolution from resolvedAt when present", () => {
    const created = utcNoon(2025, 3, 3);
    const resolvedAt = new Date(new Date(created).getTime() + 5 * DAY_MS).toISOString();
    const topics = [
      makeTopic({ createdAt: created, resolvedAt, outcome: "solved" }),
    ];

    const [row] = computeWeeklyTrends(topics);
    expect(row.medianResolution).toBe("5d");
  });

  // RT-2 — Monday and Sunday of the same week land in the same bucket
  it("places Monday and Sunday of the same week in the same bucket", () => {
    // 2025-W10: Mon 2025-03-03 and Sun 2025-03-09
    const topics = [
      makeTopic({ createdAt: utcNoon(2025, 3, 3) }), // Monday
      makeTopic({ id: 2, createdAt: utcNoon(2025, 3, 9) }), // Sunday
    ];

    const result = computeWeeklyTrends(topics);
    expect(result).toHaveLength(1);
    expect(result[0].weekStart).toBe("2025-03-03");
  });

  // RT-2 — Sunday and the following Monday land in different buckets
  it("places Sunday and the following Monday in different buckets", () => {
    // 2025-03-09 is Sun (W10), 2025-03-10 is Mon (W11)
    const topics = [
      makeTopic({ createdAt: utcNoon(2025, 3, 9) }),  // Sunday — W10
      makeTopic({ id: 2, createdAt: utcNoon(2025, 3, 10) }), // Monday — W11
    ];

    const result = computeWeeklyTrends(topics);
    expect(result).toHaveLength(2);
    const weeks = result.map((r) => r.weekStart).sort();
    expect(weeks).toEqual(["2025-03-03", "2025-03-10"]);
  });

  // RT-1 — does not mutate the input array
  it("does not mutate the input array", () => {
    const created = utcNoon(2025, 3, 3);
    const topics = [
      makeTopic({ createdAt: created }),
      makeTopic({ id: 2, createdAt: created }),
    ];
    const original = [...topics];

    computeWeeklyTrends(topics);

    expect(topics).toEqual(original);
  });

  // RT-6 — mixed topics in one week: only those with firstReplyAt contribute
  it("excludes topics without firstReplyAt from first reply median in a mixed week", () => {
    const created = utcNoon(2025, 3, 3);
    const topics = [
      makeTopic({ createdAt: created, firstReplyAt: new Date(new Date(created).getTime() + 4 * HOUR_MS).toISOString() }),
      makeTopic({ id: 2, createdAt: created }), // no firstReplyAt
    ];

    const [row] = computeWeeklyTrends(topics);
    expect(row.medianFirstReply).toBe("4h");
    expect(row.topicCount).toBe(2);
  });
});

// ---------------------------------------------------------------------------
// parseDurationToHours (RT-14)
// ---------------------------------------------------------------------------

describe("parseDurationToHours", () => {
  it("converts day durations to hours (e.g. '3d' → 72)", () => {
    expect(parseDurationToHours("3d")).toBe(72);
  });

  it("converts hour durations directly (e.g. '12h' → 12)", () => {
    expect(parseDurationToHours("12h")).toBe(12);
  });

  it("converts '1h' to 1", () => {
    expect(parseDurationToHours("1h")).toBe(1);
  });

  it("converts '1d' to 24", () => {
    expect(parseDurationToHours("1d")).toBe(24);
  });

  // RT-17
  it("returns undefined for '–' (no data)", () => {
    expect(parseDurationToHours("–")).toBeUndefined();
  });

  it("returns undefined for unrecognized format", () => {
    expect(parseDurationToHours("unknown")).toBeUndefined();
  });
});

// ---------------------------------------------------------------------------
// weeklyTrendsChartData (RT-12, RT-13, RT-14, RT-17)
// ---------------------------------------------------------------------------

describe("weeklyTrendsChartData", () => {
  // RT-12
  it("returns an empty array for empty input", () => {
    expect(weeklyTrendsChartData([])).toEqual([]);
  });

  // RT-13 — chronological order (oldest first)
  it("reverses newest-first trends to chronological (oldest-first) order", () => {
    const trends = computeWeeklyTrends([
      makeTopic({ createdAt: utcNoon(2025, 1, 6) }),  // W02
      makeTopic({ id: 2, createdAt: utcNoon(2025, 1, 20) }), // W04
      makeTopic({ id: 3, createdAt: utcNoon(2025, 1, 13) }), // W03
    ]);

    const chartData = weeklyTrendsChartData(trends);
    const labels = chartData.map((p) => p.weekLabel);

    // Oldest week first — locale-formatted labels, so just verify order
    expect(labels.length).toBe(3);
    // First label should correspond to Jan 6, last to Jan 20
    expect(labels[0]).toContain("6");
    expect(labels[2]).toContain("20");
  });

  // RT-14 — numeric conversion
  it("converts formatted durations to numeric hours", () => {
    const created = utcNoon(2025, 3, 3);
    const trends = computeWeeklyTrends([
      makeTopic({
        createdAt: created,
        firstReplyAt: new Date(new Date(created).getTime() + 2 * DAY_MS).toISOString(),
        resolvedAt: new Date(new Date(created).getTime() + 5 * DAY_MS).toISOString(),
        outcome: "solved",
      }),
    ]);

    const [point] = weeklyTrendsChartData(trends);
    expect(point.medianFirstReplyHours).toBe(48); // 2d = 48h
    expect(point.medianResolutionHours).toBe(120); // 5d = 120h
  });

  // RT-17 — "–" becomes undefined
  it("maps '–' values to undefined for chart gaps", () => {
    const trends = computeWeeklyTrends([
      makeTopic({ createdAt: utcNoon(2025, 3, 3) }), // no firstReplyAt, no resolvedAt
    ]);

    const [point] = weeklyTrendsChartData(trends);
    expect(point.medianFirstReplyHours).toBeUndefined();
    expect(point.medianResolutionHours).toBeUndefined();
  });
});
