import { describe, expect, it, vi, afterEach } from "vitest";
import { filterByPeriod } from "../../frontend/src/components/timePeriod";
import type { Topic } from "../../frontend/src/mock/data";

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
// Shared test clock
// ---------------------------------------------------------------------------

const NOW = new Date("2026-03-15T12:00:00Z").getTime();

// ---------------------------------------------------------------------------
// filterByPeriod — preset: allTime (TF-7)
// ---------------------------------------------------------------------------

describe("filterByPeriod — allTime", () => {
  it("returns all topics regardless of age", () => {
    const topics = [
      makeTopic({ id: 1, createdAt: new Date(NOW - 400 * DAY_MS).toISOString() }),
      makeTopic({ id: 2, createdAt: new Date(NOW - 5 * DAY_MS).toISOString() }),
    ];

    const result = filterByPeriod(topics, { kind: "preset", preset: "allTime" });
    expect(result.map((t) => t.id)).toEqual([1, 2]);
  });

  it("does not mutate the input array", () => {
    const topics = [makeTopic({ id: 1, createdAt: new Date(NOW).toISOString() })];
    const original = [...topics];

    filterByPeriod(topics, { kind: "preset", preset: "allTime" });
    expect(topics).toEqual(original);
  });

  it("returns empty array for empty input", () => {
    expect(filterByPeriod([], { kind: "preset", preset: "allTime" })).toEqual([]);
  });
});

// ---------------------------------------------------------------------------
// filterByPeriod — preset: last7 (TF-4)
// ---------------------------------------------------------------------------

describe("filterByPeriod — last7", () => {
  afterEach(() => {
    vi.useRealTimers();
  });

  it("includes topics created within the 7-day window", () => {
    vi.setSystemTime(NOW);

    const topics = [
      makeTopic({ id: 1, createdAt: new Date(NOW - 3 * DAY_MS).toISOString() }),
      makeTopic({ id: 2, createdAt: new Date(NOW - 6 * DAY_MS).toISOString() }),
    ];

    const result = filterByPeriod(topics, { kind: "preset", preset: "last7" });
    expect(result.map((t) => t.id)).toEqual([1, 2]);
  });

  it("excludes topics older than 7 days", () => {
    vi.setSystemTime(NOW);

    const topics = [
      makeTopic({ id: 1, createdAt: new Date(NOW - 8 * DAY_MS).toISOString() }),
      makeTopic({ id: 2, createdAt: new Date(NOW - 30 * DAY_MS).toISOString() }),
    ];

    const result = filterByPeriod(topics, { kind: "preset", preset: "last7" });
    expect(result).toHaveLength(0);
  });

  it("includes a topic created exactly at the 7-day boundary", () => {
    vi.setSystemTime(NOW);

    const topics = [
      makeTopic({ id: 1, createdAt: new Date(NOW - 7 * DAY_MS).toISOString() }),
    ];

    const result = filterByPeriod(topics, { kind: "preset", preset: "last7" });
    expect(result).toHaveLength(1);
  });

  it("returns empty array for empty input", () => {
    vi.setSystemTime(NOW);
    expect(filterByPeriod([], { kind: "preset", preset: "last7" })).toEqual([]);
  });
});

// ---------------------------------------------------------------------------
// filterByPeriod — preset: last30 (TF-5)
// ---------------------------------------------------------------------------

describe("filterByPeriod — last30", () => {
  afterEach(() => {
    vi.useRealTimers();
  });

  it("includes topics created within the 30-day window", () => {
    vi.setSystemTime(NOW);

    const topics = [
      makeTopic({ id: 1, createdAt: new Date(NOW - 15 * DAY_MS).toISOString() }),
      makeTopic({ id: 2, createdAt: new Date(NOW - 29 * DAY_MS).toISOString() }),
    ];

    const result = filterByPeriod(topics, { kind: "preset", preset: "last30" });
    expect(result.map((t) => t.id)).toEqual([1, 2]);
  });

  it("excludes topics older than 30 days", () => {
    vi.setSystemTime(NOW);

    const topics = [
      makeTopic({ id: 1, createdAt: new Date(NOW - 31 * DAY_MS).toISOString() }),
    ];

    const result = filterByPeriod(topics, { kind: "preset", preset: "last30" });
    expect(result).toHaveLength(0);
  });

  it("includes a topic created exactly at the 30-day boundary", () => {
    vi.setSystemTime(NOW);

    const topics = [
      makeTopic({ id: 1, createdAt: new Date(NOW - 30 * DAY_MS).toISOString() }),
    ];

    const result = filterByPeriod(topics, { kind: "preset", preset: "last30" });
    expect(result).toHaveLength(1);
  });
});

// ---------------------------------------------------------------------------
// filterByPeriod — preset: lastYear (TF-6)
// ---------------------------------------------------------------------------

describe("filterByPeriod — lastYear", () => {
  afterEach(() => {
    vi.useRealTimers();
  });

  it("includes topics created within the last 365 days", () => {
    vi.setSystemTime(NOW);

    const topics = [
      makeTopic({ id: 1, createdAt: new Date(NOW - 100 * DAY_MS).toISOString() }),
      makeTopic({ id: 2, createdAt: new Date(NOW - 364 * DAY_MS).toISOString() }),
    ];

    const result = filterByPeriod(topics, { kind: "preset", preset: "lastYear" });
    expect(result.map((t) => t.id)).toEqual([1, 2]);
  });

  it("excludes topics older than 365 days", () => {
    vi.setSystemTime(NOW);

    const topics = [
      makeTopic({ id: 1, createdAt: new Date(NOW - 366 * DAY_MS).toISOString() }),
    ];

    const result = filterByPeriod(topics, { kind: "preset", preset: "lastYear" });
    expect(result).toHaveLength(0);
  });

  it("includes a topic created exactly at the 365-day boundary", () => {
    vi.setSystemTime(NOW);

    const topics = [
      makeTopic({ id: 1, createdAt: new Date(NOW - 365 * DAY_MS).toISOString() }),
    ];

    const result = filterByPeriod(topics, { kind: "preset", preset: "lastYear" });
    expect(result).toHaveLength(1);
  });
});

// ---------------------------------------------------------------------------
// filterByPeriod — custom range (TF-8)
// ---------------------------------------------------------------------------

describe("filterByPeriod — custom range", () => {
  it("includes topics on or after the from date", () => {
    const topics = [
      // 2026-03-10 — on the from date
      makeTopic({ id: 1, createdAt: "2026-03-10T08:00:00Z" }),
      // 2026-03-12 — after the from date
      makeTopic({ id: 2, createdAt: "2026-03-12T08:00:00Z" }),
    ];

    const result = filterByPeriod(topics, {
      kind: "custom",
      range: { from: "2026-03-10", to: "2026-03-14" },
    });
    expect(result.map((t) => t.id)).toEqual([1, 2]);
  });

  it("excludes topics before the from date", () => {
    const topics = [
      makeTopic({ id: 1, createdAt: "2026-03-09T23:59:00Z" }),
    ];

    const result = filterByPeriod(topics, {
      kind: "custom",
      range: { from: "2026-03-10", to: "2026-03-14" },
    });
    expect(result).toHaveLength(0);
  });

  it("includes topics on or before the to date (through end of day)", () => {
    const topics = [
      // 2026-03-14 at 21:00 UTC — within the to-date boundary (23:59:59.999 UTC)
      makeTopic({ id: 1, createdAt: "2026-03-14T21:00:00Z" }),
    ];

    const result = filterByPeriod(topics, {
      kind: "custom",
      range: { from: "2026-03-10", to: "2026-03-14" },
    });
    expect(result).toHaveLength(1);
  });

  it("excludes topics after the to date", () => {
    const topics = [
      makeTopic({ id: 1, createdAt: "2026-03-16T00:00:00Z" }),
    ];

    const result = filterByPeriod(topics, {
      kind: "custom",
      range: { from: "2026-03-10", to: "2026-03-14" },
    });
    expect(result).toHaveLength(0);
  });

  it("returns empty array for empty input", () => {
    const result = filterByPeriod([], {
      kind: "custom",
      range: { from: "2026-03-01", to: "2026-03-14" },
    });
    expect(result).toEqual([]);
  });

  it("does not mutate the input array", () => {
    const topics = [makeTopic({ id: 1, createdAt: "2026-03-12T08:00:00Z" })];
    const original = [...topics];

    filterByPeriod(topics, {
      kind: "custom",
      range: { from: "2026-03-10", to: "2026-03-14" },
    });
    expect(topics).toEqual(original);
  });
});
