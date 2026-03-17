import { describe, expect, it } from "vitest";
import {
  filterByTag,
  filterByMonitoredTags,
  monitoredTags,
  tagsForArea,
  type TagConfig,
} from "../../frontend/src/components/tagFilter";
import { filterByPeriod } from "../../frontend/src/components/timePeriod";
import type { Topic } from "../../frontend/src/mock/data";

function makeTopic(overrides: Partial<Topic> & { id: number }): Topic {
  return {
    title: "Test topic",
    createdAt: "2026-03-10T12:00:00Z",
    tags: [],
    category: "Support",
    replyCount: 0,
    ...overrides,
  };
}

const CONFIG: TagConfig = {
  closedTag: "closed",
  stalledDays: 14,
  areas: [
    { name: "Integration", primaryTag: "api", tags: ["api", "webhooks", "sso"] },
    { name: "Access", primaryTag: "authentication", tags: ["authentication", "ssl"] },
    { name: "Content", primaryTag: "editor", tags: ["editor", "search"] },
  ],
};

// ---------------------------------------------------------------------------
// monitoredTags (TA-17)
// ---------------------------------------------------------------------------

describe("monitoredTags", () => {
  it("returns deduplicated union of all area tags", () => {
    const result = monitoredTags(CONFIG);
    expect(result).toHaveLength(7);
    expect(result).toContain("api");
    expect(result).toContain("webhooks");
    expect(result).toContain("sso");
    expect(result).toContain("authentication");
    expect(result).toContain("ssl");
    expect(result).toContain("editor");
    expect(result).toContain("search");
  });

  it("deduplicates tags appearing in multiple areas", () => {
    const config: TagConfig = {
      closedTag: "closed",
      stalledDays: 14,
      areas: [
        { name: "A", primaryTag: "shared", tags: ["shared", "x"] },
        { name: "B", primaryTag: "shared", tags: ["shared", "y"] },
      ],
    };
    const result = monitoredTags(config);
    expect(result.filter((t) => t === "shared")).toHaveLength(1);
    expect(result).toHaveLength(3);
  });

  it("returns empty array for empty config", () => {
    expect(monitoredTags({ closedTag: "closed", stalledDays: 14, areas: [] })).toEqual([]);
  });
});

// ---------------------------------------------------------------------------
// filterByTag (TA-2, TA-5, TA-6)
// ---------------------------------------------------------------------------

describe("filterByTag", () => {
  const topics = [
    makeTopic({ id: 1, tags: ["api"] }),
    makeTopic({ id: 2, tags: ["webhooks"] }),
    makeTopic({ id: 3, tags: ["api", "authentication"] }),
    makeTopic({ id: 4, tags: [] }),
  ];

  it("returns topics carrying the selected tag (TA-5)", () => {
    const result = filterByTag(topics, "api");
    expect(result.map((t) => t.id)).toEqual([1, 3]);
  });

  it("excludes topics without the selected tag (TA-5)", () => {
    const result = filterByTag(topics, "api");
    expect(result.find((t) => t.id === 2)).toBeUndefined();
    expect(result.find((t) => t.id === 4)).toBeUndefined();
  });

  it("includes multi-tag topics when any tag matches (TA-5)", () => {
    const result = filterByTag(topics, "authentication");
    expect(result.map((t) => t.id)).toEqual([3]);
  });

  it("returns all topics when tag is null (TA-2)", () => {
    const result = filterByTag(topics, null);
    expect(result).toBe(topics); // same reference — no filtering
  });

  it("returns empty array when no topics match", () => {
    expect(filterByTag(topics, "nonexistent")).toEqual([]);
  });

  it("does not mutate input array", () => {
    const original = [...topics];
    filterByTag(topics, "api");
    expect(topics).toEqual(original);
  });
});

// ---------------------------------------------------------------------------
// filterByMonitoredTags (TA-17)
// ---------------------------------------------------------------------------

describe("filterByMonitoredTags", () => {
  const monitored = ["api", "webhooks", "authentication"];

  it("includes topics with at least one monitored tag", () => {
    const topics = [
      makeTopic({ id: 1, tags: ["api"] }),
      makeTopic({ id: 2, tags: ["unknown"] }),
      makeTopic({ id: 3, tags: ["api", "unknown"] }),
    ];
    const result = filterByMonitoredTags(topics, monitored);
    expect(result.map((t) => t.id)).toEqual([1, 3]);
  });

  it("excludes untagged topics (TA-6, TA-17)", () => {
    const topics = [
      makeTopic({ id: 1, tags: [] }),
      makeTopic({ id: 2, tags: ["api"] }),
    ];
    const result = filterByMonitoredTags(topics, monitored);
    expect(result.map((t) => t.id)).toEqual([2]);
  });

  it("excludes topics with only non-monitored tags", () => {
    const topics = [makeTopic({ id: 1, tags: ["random", "other"] })];
    expect(filterByMonitoredTags(topics, monitored)).toEqual([]);
  });

  it("returns empty array for empty input", () => {
    expect(filterByMonitoredTags([], monitored)).toEqual([]);
  });

  it("does not mutate input array", () => {
    const topics = [makeTopic({ id: 1, tags: ["api"] })];
    const original = [...topics];
    filterByMonitoredTags(topics, monitored);
    expect(topics).toEqual(original);
  });
});

// ---------------------------------------------------------------------------
// tagsForArea (TA-12, TA-13)
// ---------------------------------------------------------------------------

describe("tagsForArea", () => {
  it("returns primary tag first, rest alphabetical for a given area (TA-12)", () => {
    const result = tagsForArea(CONFIG, "Integration");
    expect(result).toEqual(["api", "sso", "webhooks"]);
  });

  it("returns all tags sorted alphabetically when area is null (TA-13)", () => {
    const result = tagsForArea(CONFIG, null);
    expect(result).toEqual([
      "api", "authentication", "editor", "search", "ssl", "sso", "webhooks",
    ]);
  });

  it("returns empty array for unknown area", () => {
    expect(tagsForArea(CONFIG, "Unknown")).toEqual([]);
  });

  it("handles area with only one tag", () => {
    const config: TagConfig = {
      closedTag: "closed",
      stalledDays: 14,
      areas: [
        { name: "Solo", primaryTag: "only", tags: ["only"] },
      ],
    };
    expect(tagsForArea(config, "Solo")).toEqual(["only"]);
  });
});

// ---------------------------------------------------------------------------
// Filter composition (TA-4)
// ---------------------------------------------------------------------------

describe("filterByTag composed with filterByPeriod", () => {
  const NOW = new Date("2026-03-15T12:00:00Z").getTime();
  const DAY_MS = 86_400_000;

  const topics = [
    makeTopic({
      id: 1,
      tags: ["api"],
      createdAt: new Date(NOW - 5 * DAY_MS).toISOString(),
    }),
    makeTopic({
      id: 2,
      tags: ["api"],
      createdAt: new Date(NOW - 40 * DAY_MS).toISOString(),
    }),
    makeTopic({
      id: 3,
      tags: ["webhooks"],
      createdAt: new Date(NOW - 5 * DAY_MS).toISOString(),
    }),
  ];

  it("both filters apply — only topics matching tag AND period remain", () => {
    const afterPeriod = filterByPeriod(topics, { kind: "preset", preset: "last30" });
    const afterTag = filterByTag(afterPeriod, "api");
    expect(afterTag.map((t) => t.id)).toEqual([1]);
  });
});
