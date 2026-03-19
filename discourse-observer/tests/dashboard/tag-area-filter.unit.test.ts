import { describe, expect, it } from "vitest";
import {
  tagsForArea,
  allAreas,
} from "../../frontend/src/components/TagSelector";
import type { AppConfig } from "../../frontend/src/api/types";

// Minimal config fixture matching the API /config response shape
const CONFIG: AppConfig = {
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
    api: { area: "Integration", areaIsDefault: false, stalledDays: 7, stalledDaysIsDefault: false, slo: { firstReplyHours: 4, resolutionHours: 48, inactivityHours: 24 }, sloIsDefault: false, closedTag: "closed" },
    webhooks: { area: "Integration", areaIsDefault: false, stalledDays: 14, stalledDaysIsDefault: true, slo: { firstReplyHours: 24, resolutionHours: 336, inactivityHours: 48 }, sloIsDefault: true, closedTag: null },
    sso: { area: "Integration", areaIsDefault: false, stalledDays: 14, stalledDaysIsDefault: true, slo: { firstReplyHours: 24, resolutionHours: 336, inactivityHours: 48 }, sloIsDefault: true, closedTag: null },
    authentication: { area: "Access", areaIsDefault: false, stalledDays: 7, stalledDaysIsDefault: false, slo: { firstReplyHours: 24, resolutionHours: 336, inactivityHours: 48 }, sloIsDefault: true, closedTag: null },
    ssl: { area: "Access", areaIsDefault: false, stalledDays: 14, stalledDaysIsDefault: true, slo: { firstReplyHours: 24, resolutionHours: 336, inactivityHours: 48 }, sloIsDefault: true, closedTag: null },
    editor: { area: "Content", areaIsDefault: false, stalledDays: 14, stalledDaysIsDefault: true, slo: { firstReplyHours: 24, resolutionHours: 336, inactivityHours: 48 }, sloIsDefault: true, closedTag: null },
    search: { area: "Content", areaIsDefault: false, stalledDays: 14, stalledDaysIsDefault: true, slo: { firstReplyHours: 24, resolutionHours: 336, inactivityHours: 48 }, sloIsDefault: true, closedTag: null },
    plugin: { area: "Other", areaIsDefault: true, stalledDays: 14, stalledDaysIsDefault: true, slo: { firstReplyHours: 8, resolutionHours: 72, inactivityHours: 48 }, sloIsDefault: false, closedTag: null },
    migration: { area: "Other", areaIsDefault: true, stalledDays: 14, stalledDaysIsDefault: true, slo: { firstReplyHours: 24, resolutionHours: 336, inactivityHours: 48 }, sloIsDefault: true, closedTag: null },
  },
  distributionBucketCeilings: [1, 4, 12, 24, 48, 96, 168],
};

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
    const config: AppConfig = {
      ...CONFIG,
      areas: [{ name: "Solo", primaryTag: "only" }],
      tags: { only: { area: "Solo", areaIsDefault: false, stalledDays: 14, stalledDaysIsDefault: true, slo: CONFIG.defaults.slo, sloIsDefault: true, closedTag: null } },
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
    const config: AppConfig = {
      ...CONFIG,
      areas: [{ name: "Integration", primaryTag: "api" }],
      tags: { api: { area: "Integration", areaIsDefault: false, stalledDays: 14, stalledDaysIsDefault: true, slo: CONFIG.defaults.slo, sloIsDefault: true, closedTag: null } },
    };
    const result = allAreas(config);
    expect(result).toEqual(["Integration"]);
  });
});
