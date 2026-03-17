import { describe, expect, it } from "vitest";
import {
  filterByTag,
  filterByMonitoredTags,
  monitoredTags,
  tagsForArea,
  allAreas,
  resolveTag,
  resolveAllTags,
  extractSloConfig,
  scopeSloConfig,
  sloDefaultTags,
  type TagConfig,
  type ResolvedTag,
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
  defaults: {
    stalledDays: 14,
    area: "Other",
    slo: { firstReplyHours: 24, resolutionHours: 336, inactivityHours: 48 },
  },
  areas: [
    { name: "Integration", primaryTag: "api" },
    { name: "Access", primaryTag: "authentication" },
    { name: "Content", primaryTag: "editor" },
  ],
  tags: {
    api: { area: "Integration", closedTag: "closed", stalledDays: 7, slo: { firstReplyHours: 4, resolutionHours: 48, inactivityHours: 24 } },
    webhooks: { area: "Integration" },
    sso: { area: "Integration" },
    authentication: { area: "Access", stalledDays: 7 },
    ssl: { area: "Access" },
    editor: { area: "Content" },
    search: { area: "Content" },
    plugin: { slo: { firstReplyHours: 8, resolutionHours: 72, inactivityHours: 48 } },
    migration: {},
  },
};

// ---------------------------------------------------------------------------
// monitoredTags (TA-17)
// ---------------------------------------------------------------------------

describe("monitoredTags", () => {
  it("returns all tag keys from config", () => {
    const result = monitoredTags(CONFIG);
    expect(result).toHaveLength(9);
    expect(result).toContain("api");
    expect(result).toContain("webhooks");
    expect(result).toContain("sso");
    expect(result).toContain("authentication");
    expect(result).toContain("ssl");
    expect(result).toContain("editor");
    expect(result).toContain("search");
    expect(result).toContain("plugin");
    expect(result).toContain("migration");
  });

  it("returns empty array for config with no tags", () => {
    const config: TagConfig = { ...CONFIG, tags: {} };
    expect(monitoredTags(config)).toEqual([]);
  });
});

// ---------------------------------------------------------------------------
// resolveTag — default resolution and provenance
// ---------------------------------------------------------------------------

describe("resolveTag", () => {
  it("uses explicit values when present", () => {
    const resolved = resolveTag(
      { area: "Integration", closedTag: "closed", stalledDays: 7, slo: { firstReplyHours: 4, resolutionHours: 48, inactivityHours: 24 } },
      CONFIG.defaults,
    );
    expect(resolved.area).toBe("Integration");
    expect(resolved.areaIsDefault).toBe(false);
    expect(resolved.closedTag).toBe("closed");
    expect(resolved.stalledDays).toBe(7);
    expect(resolved.stalledDaysIsDefault).toBe(false);
    expect(resolved.slo.firstReplyHours).toBe(4);
    expect(resolved.sloIsDefault).toBe(false);
  });

  it("falls back to defaults for absent fields", () => {
    const resolved = resolveTag({}, CONFIG.defaults);
    expect(resolved.area).toBe("Other");
    expect(resolved.areaIsDefault).toBe(true);
    expect(resolved.closedTag).toBeNull();
    expect(resolved.stalledDays).toBe(14);
    expect(resolved.stalledDaysIsDefault).toBe(true);
    expect(resolved.slo).toEqual(CONFIG.defaults.slo);
    expect(resolved.sloIsDefault).toBe(true);
  });

  it("allows partial overrides", () => {
    const resolved = resolveTag({ stalledDays: 3 }, CONFIG.defaults);
    expect(resolved.stalledDays).toBe(3);
    expect(resolved.stalledDaysIsDefault).toBe(false);
    expect(resolved.area).toBe("Other");
    expect(resolved.areaIsDefault).toBe(true);
  });
});

// ---------------------------------------------------------------------------
// resolveAllTags
// ---------------------------------------------------------------------------

describe("resolveAllTags", () => {
  it("resolves every tag in config", () => {
    const resolved = resolveAllTags(CONFIG);
    expect(Object.keys(resolved)).toHaveLength(9);
    expect(resolved.api.area).toBe("Integration");
    expect(resolved.migration.area).toBe("Other");
    expect(resolved.migration.sloIsDefault).toBe(true);
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
  it("returns primary tag first, rest alphabetical for a named area (TA-12)", () => {
    const result = tagsForArea(CONFIG, "Integration");
    expect(result).toEqual(["api", "sso", "webhooks"]);
  });

  it("returns all tags sorted alphabetically when area is null (TA-13)", () => {
    const result = tagsForArea(CONFIG, null);
    expect(result).toEqual([
      "api", "authentication", "editor", "migration", "plugin", "search", "ssl", "sso", "webhooks",
    ]);
  });

  it("returns empty array for unknown area", () => {
    expect(tagsForArea(CONFIG, "Unknown")).toEqual([]);
  });

  it("returns tags in default area sorted alphabetically (no primaryTag)", () => {
    const result = tagsForArea(CONFIG, "Other");
    expect(result).toEqual(["migration", "plugin"]);
  });

  it("handles area with only one tag", () => {
    const config: TagConfig = {
      defaults: CONFIG.defaults,
      areas: [{ name: "Solo", primaryTag: "only" }],
      tags: { only: { area: "Solo" } },
    };
    expect(tagsForArea(config, "Solo")).toEqual(["only"]);
  });
});

// ---------------------------------------------------------------------------
// allAreas
// ---------------------------------------------------------------------------

describe("allAreas", () => {
  it("returns named areas plus default area when tags use it", () => {
    const result = allAreas(CONFIG);
    expect(result).toEqual(["Integration", "Access", "Content", "Other"]);
  });

  it("omits default area when no tags use it", () => {
    const config: TagConfig = {
      defaults: CONFIG.defaults,
      areas: [{ name: "Integration", primaryTag: "api" }],
      tags: { api: { area: "Integration" } },
    };
    const result = allAreas(config);
    expect(result).toEqual(["Integration"]);
  });
});

// ---------------------------------------------------------------------------
// extractSloConfig and sloDefaultTags
// ---------------------------------------------------------------------------

describe("extractSloConfig", () => {
  it("returns SLO thresholds for all tags (explicit and default)", () => {
    const slo = extractSloConfig(CONFIG);
    expect(slo.api.firstReplyHours).toBe(4);
    expect(slo.migration.firstReplyHours).toBe(24); // from defaults
    expect(Object.keys(slo)).toHaveLength(9);
  });
});

describe("scopeSloConfig", () => {
  it("returns only tags in the visible set", () => {
    const full = extractSloConfig(CONFIG);
    const scoped = scopeSloConfig(full, ["api", "webhooks"]);
    expect(Object.keys(scoped).sort()).toEqual(["api", "webhooks"]);
  });

  it("returns empty config when visible tags is empty", () => {
    const full = extractSloConfig(CONFIG);
    expect(scopeSloConfig(full, [])).toEqual({});
  });

  it("ignores tags not in full config", () => {
    const full = extractSloConfig(CONFIG);
    const scoped = scopeSloConfig(full, ["api", "nonexistent"]);
    expect(Object.keys(scoped)).toEqual(["api"]);
  });
});

describe("sloDefaultTags", () => {
  it("returns tags without explicit SLO", () => {
    const defaults = sloDefaultTags(CONFIG);
    expect(defaults.has("migration")).toBe(true);
    expect(defaults.has("webhooks")).toBe(true);
    expect(defaults.has("api")).toBe(false);
    expect(defaults.has("plugin")).toBe(false);
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
