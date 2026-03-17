// Spec: specs/dashboard/stalled-topics.md

import { describe, expect, it } from "vitest";
import {
  filterStalledTopics,
  daysSinceLastActivity,
  formatStalledTag,
} from "../../frontend/src/components/stalledMetrics";
import type { Topic } from "../../frontend/src/mock/data";

const DAY_MS = 86_400_000;

function makeTopic(overrides: Partial<Topic> & { id: number }): Topic {
  return {
    title: "Test topic",
    createdAt: "2025-01-01T12:00:00Z",
    tags: ["api"],
    category: "Support",
    replyCount: 2,
    ...overrides,
  };
}

function daysAgo(days: number, from: Date): string {
  return new Date(from.getTime() - days * DAY_MS).toISOString();
}

const NOW = new Date("2025-03-01T12:00:00Z");

// ---------------------------------------------------------------------------
// filterStalledTopics (ST-2, ST-4, ST-5, ST-6, ST-13)
// ---------------------------------------------------------------------------

describe("filterStalledTopics", () => {
  const STALLED_DAYS = 14;
  const CLOSED_TAG = "closed";

  // ST-5 — excludes topics with closed tag
  it("excludes topics carrying the closed tag", () => {
    const topics = [
      makeTopic({
        id: 1,
        tags: ["api", "closed"],
        lastActivityAt: daysAgo(30, NOW),
      }),
    ];
    const result = filterStalledTopics(topics, STALLED_DAYS, CLOSED_TAG, NOW);
    expect(result).toHaveLength(0);
  });

  // ST-2, ST-4 — excludes topics with recent activity within threshold
  it("excludes topics with activity within threshold", () => {
    const topics = [
      makeTopic({
        id: 1,
        lastActivityAt: daysAgo(10, NOW),
      }),
    ];
    const result = filterStalledTopics(topics, STALLED_DAYS, CLOSED_TAG, NOW);
    expect(result).toHaveLength(0);
  });

  // ST-2, ST-4 — includes topics with activity older than threshold
  it("includes topics with activity older than threshold", () => {
    const topics = [
      makeTopic({
        id: 1,
        lastActivityAt: daysAgo(20, NOW),
      }),
    ];
    const result = filterStalledTopics(topics, STALLED_DAYS, CLOSED_TAG, NOW);
    expect(result).toHaveLength(1);
    expect(result[0].id).toBe(1);
  });

  // ST-4 — boundary: exactly stalledDays ago is not stalled
  it("does not include topic with activity exactly at threshold boundary", () => {
    const topics = [
      makeTopic({
        id: 1,
        lastActivityAt: daysAgo(14, NOW),
      }),
    ];
    const result = filterStalledTopics(topics, STALLED_DAYS, CLOSED_TAG, NOW);
    expect(result).toHaveLength(0);
  });

  // ST-4 — boundary: stalledDays + 1 day is stalled
  it("includes topic with activity one day beyond threshold", () => {
    const topics = [
      makeTopic({
        id: 1,
        lastActivityAt: daysAgo(15, NOW),
      }),
    ];
    const result = filterStalledTopics(topics, STALLED_DAYS, CLOSED_TAG, NOW);
    expect(result).toHaveLength(1);
  });

  // ST-6 — sorts by lastActivityAt ascending (oldest first)
  it("sorts by lastActivityAt ascending", () => {
    const topics = [
      makeTopic({ id: 1, lastActivityAt: daysAgo(20, NOW) }),
      makeTopic({ id: 2, lastActivityAt: daysAgo(30, NOW) }),
      makeTopic({ id: 3, lastActivityAt: daysAgo(25, NOW) }),
    ];
    const result = filterStalledTopics(topics, STALLED_DAYS, CLOSED_TAG, NOW);
    expect(result.map((t) => t.id)).toEqual([2, 3, 1]);
  });

  // ST-13 — empty input returns empty array
  it("returns empty array for empty input", () => {
    expect(filterStalledTopics([], STALLED_DAYS, CLOSED_TAG, NOW)).toEqual([]);
  });

  // ST-2 — does not mutate input array
  it("does not mutate the input array", () => {
    const topics = [
      makeTopic({ id: 1, lastActivityAt: daysAgo(20, NOW) }),
      makeTopic({ id: 2, lastActivityAt: daysAgo(30, NOW) }),
    ];
    const original = [...topics];
    filterStalledTopics(topics, STALLED_DAYS, CLOSED_TAG, NOW);
    expect(topics).toEqual(original);
  });

  // ST-3 — falls back to createdAt when lastActivityAt is absent
  it("falls back to createdAt when lastActivityAt is absent", () => {
    const topics = [
      makeTopic({
        id: 1,
        createdAt: daysAgo(20, NOW),
        lastActivityAt: undefined,
      }),
    ];
    const result = filterStalledTopics(topics, STALLED_DAYS, CLOSED_TAG, NOW);
    expect(result).toHaveLength(1);
  });
});

// ---------------------------------------------------------------------------
// daysSinceLastActivity (ST-3, ST-7)
// ---------------------------------------------------------------------------

describe("daysSinceLastActivity", () => {
  // ST-7 — returns whole days truncated
  it("returns whole days truncated", () => {
    const topic = makeTopic({
      id: 1,
      lastActivityAt: new Date(NOW.getTime() - 15.7 * DAY_MS).toISOString(),
    });
    expect(daysSinceLastActivity(topic, NOW)).toBe(15);
  });

  // ST-3 — uses lastActivityAt when present
  it("uses lastActivityAt when present", () => {
    const topic = makeTopic({
      id: 1,
      createdAt: daysAgo(30, NOW),
      lastActivityAt: daysAgo(10, NOW),
    });
    expect(daysSinceLastActivity(topic, NOW)).toBe(10);
  });

  // ST-3 — falls back to createdAt when lastActivityAt absent
  it("falls back to createdAt when lastActivityAt is absent", () => {
    const topic = makeTopic({
      id: 1,
      createdAt: daysAgo(25, NOW),
      lastActivityAt: undefined,
    });
    expect(daysSinceLastActivity(topic, NOW)).toBe(25);
  });
});

// ---------------------------------------------------------------------------
// formatStalledTag (ST-7)
// ---------------------------------------------------------------------------

describe("formatStalledTag", () => {
  const monitored = new Set(["api", "webhooks", "authentication"]);

  // ST-7 — returns first monitored tag
  it("returns the first monitored tag found", () => {
    const topic = makeTopic({ id: 1, tags: ["api", "webhooks"] });
    expect(formatStalledTag(topic, monitored)).toBe("api");
  });

  // ST-7 — returns dash when no monitored tag
  it("returns dash when topic has no monitored tag", () => {
    const topic = makeTopic({ id: 1, tags: ["unrelated"] });
    expect(formatStalledTag(topic, monitored)).toBe("–");
  });

  it("returns dash for empty tags", () => {
    const topic = makeTopic({ id: 1, tags: [] });
    expect(formatStalledTag(topic, monitored)).toBe("–");
  });
});
