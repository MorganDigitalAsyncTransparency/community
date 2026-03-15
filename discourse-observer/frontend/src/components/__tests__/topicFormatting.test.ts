import { describe, expect, it, vi, afterEach } from "vitest";
import {
  formatAge,
  formatTags,
  oldestUnrepliedDays,
  sortedByOldest,
} from "../topicFormatting";
import type { Topic } from "../../mock/data";

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

// ---------------------------------------------------------------------------
// formatAge (QV-3, QV-13)
// ---------------------------------------------------------------------------
describe("formatAge", () => {
  afterEach(() => {
    vi.useRealTimers();
  });

  it("returns days when topic is 24 hours or older", () => {
    const now = new Date("2026-03-15T12:00:00Z").getTime();
    vi.setSystemTime(now);

    const twoDaysAgo = new Date(now - 2 * DAY_MS).toISOString();
    expect(formatAge(twoDaysAgo)).toBe("2d");
  });

  it("returns whole days truncated, not rounded", () => {
    const now = new Date("2026-03-15T12:00:00Z").getTime();
    vi.setSystemTime(now);

    // 2.9 days → should show 2d, not 3d
    const almostThreeDays = new Date(now - 2.9 * DAY_MS).toISOString();
    expect(formatAge(almostThreeDays)).toBe("2d");
  });

  it("returns hours when topic is younger than 24 hours", () => {
    const now = new Date("2026-03-15T12:00:00Z").getTime();
    vi.setSystemTime(now);

    const eightHoursAgo = new Date(now - 8 * HOUR_MS).toISOString();
    expect(formatAge(eightHoursAgo)).toBe("8h");
  });

  it("returns minimum 1h for very recent topics", () => {
    const now = new Date("2026-03-15T12:00:00Z").getTime();
    vi.setSystemTime(now);

    const fiveMinutesAgo = new Date(now - 5 * 60_000).toISOString();
    expect(formatAge(fiveMinutesAgo)).toBe("1h");
  });

  it("returns 1d at exactly 24 hours", () => {
    const now = new Date("2026-03-15T12:00:00Z").getTime();
    vi.setSystemTime(now);

    const exactly24h = new Date(now - 24 * HOUR_MS).toISOString();
    expect(formatAge(exactly24h)).toBe("1d");
  });

  it("returns 23h at just under 24 hours", () => {
    const now = new Date("2026-03-15T12:00:00Z").getTime();
    vi.setSystemTime(now);

    const just23h = new Date(now - 23.5 * HOUR_MS).toISOString();
    expect(formatAge(just23h)).toBe("23h");
  });
});

// ---------------------------------------------------------------------------
// sortedByOldest (QV-1, QV-11)
// ---------------------------------------------------------------------------
describe("sortedByOldest", () => {
  it("returns topics sorted ascending by createdAt", () => {
    const topics: Topic[] = [
      makeTopic({ id: 1, createdAt: "2026-03-10T00:00:00Z" }),
      makeTopic({ id: 2, createdAt: "2026-03-01T00:00:00Z" }),
      makeTopic({ id: 3, createdAt: "2026-03-05T00:00:00Z" }),
    ];

    const sorted = sortedByOldest(topics);
    expect(sorted.map((t) => t.id)).toEqual([2, 3, 1]);
  });

  it("does not mutate the original array", () => {
    const topics: Topic[] = [
      makeTopic({ id: 1, createdAt: "2026-03-10T00:00:00Z" }),
      makeTopic({ id: 2, createdAt: "2026-03-01T00:00:00Z" }),
    ];

    const original = [...topics];
    sortedByOldest(topics);
    expect(topics.map((t) => t.id)).toEqual(original.map((t) => t.id));
  });

  it("returns empty array for empty input", () => {
    expect(sortedByOldest([])).toEqual([]);
  });

  it("handles single-element array", () => {
    const topics = [makeTopic({ id: 1, createdAt: "2026-03-10T00:00:00Z" })];
    const sorted = sortedByOldest(topics);
    expect(sorted).toHaveLength(1);
    expect(sorted[0].id).toBe(1);
  });
});

// ---------------------------------------------------------------------------
// oldestUnrepliedDays (QV-6, QV-7)
// ---------------------------------------------------------------------------
describe("oldestUnrepliedDays", () => {
  afterEach(() => {
    vi.useRealTimers();
  });

  it("returns dash for empty list", () => {
    expect(oldestUnrepliedDays([])).toBe("–");
  });

  it("returns correct days for single topic", () => {
    const now = new Date("2026-03-15T12:00:00Z").getTime();
    vi.setSystemTime(now);

    const topics = [makeTopic({ createdAt: new Date(now - 7 * DAY_MS).toISOString() })];
    expect(oldestUnrepliedDays(topics)).toBe("7d");
  });

  it("returns days for the oldest topic when multiple exist", () => {
    const now = new Date("2026-03-15T12:00:00Z").getTime();
    vi.setSystemTime(now);

    const topics = [
      makeTopic({ id: 1, createdAt: new Date(now - 3 * DAY_MS).toISOString() }),
      makeTopic({ id: 2, createdAt: new Date(now - 14 * DAY_MS).toISOString() }),
      makeTopic({ id: 3, createdAt: new Date(now - 5 * DAY_MS).toISOString() }),
    ];
    expect(oldestUnrepliedDays(topics)).toBe("14d");
  });

  it("truncates partial days", () => {
    const now = new Date("2026-03-15T12:00:00Z").getTime();
    vi.setSystemTime(now);

    // 2.8 days → should show 2d
    const topics = [makeTopic({ createdAt: new Date(now - 2.8 * DAY_MS).toISOString() })];
    expect(oldestUnrepliedDays(topics)).toBe("2d");
  });
});

// ---------------------------------------------------------------------------
// formatTags (QV-4)
// ---------------------------------------------------------------------------
describe("formatTags", () => {
  it("joins multiple tags with comma and space", () => {
    expect(formatTags(["authentication", "sso"])).toBe("authentication, sso");
  });

  it("returns single tag as-is", () => {
    expect(formatTags(["webhooks"])).toBe("webhooks");
  });

  it("returns dash for empty array", () => {
    expect(formatTags([])).toBe("–");
  });
});
