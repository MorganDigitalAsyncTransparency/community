// Spec: specs/dashboard/url-state.md
// Source: frontend/src/components/urlState.ts

import { describe, expect, it } from "vitest";
import { parseUrlState, buildSearch } from "../../frontend/src/components/urlState";

// ---------------------------------------------------------------------------
// parseUrlState — page (US-5, US-7)
// ---------------------------------------------------------------------------

describe("parseUrlState — page", () => {
  it("parses a valid page parameter", () => {
    expect(parseUrlState("?page=slo").page).toBe("slo");
  });

  it("parses all valid page values", () => {
    const pages = ["queue", "response-metrics", "distribution", "slo", "activity", "tag-flows", "sync-log"] as const;
    for (const page of pages) {
      expect(parseUrlState(`?page=${page}`).page).toBe(page);
    }
  });

  it("falls back to queue for an invalid page", () => {
    expect(parseUrlState("?page=unknown").page).toBe("queue");
  });

  it("falls back to queue when page is missing", () => {
    expect(parseUrlState("").page).toBe("queue");
  });
});

// ---------------------------------------------------------------------------
// parseUrlState — period preset (US-5, US-7)
// ---------------------------------------------------------------------------

describe("parseUrlState — period preset", () => {
  it("parses a valid preset", () => {
    const result = parseUrlState("?period=last30");
    expect(result.period).toEqual({ kind: "preset", preset: "last30" });
  });

  it("parses all valid preset values", () => {
    const presets = ["last7", "last30", "lastYear", "allTime"] as const;
    for (const preset of presets) {
      const result = parseUrlState(`?period=${preset}`);
      expect(result.period).toEqual({ kind: "preset", preset });
    }
  });

  it("falls back to allTime for an invalid preset", () => {
    const result = parseUrlState("?period=bogus");
    expect(result.period).toEqual({ kind: "preset", preset: "allTime" });
  });

  it("falls back to allTime when period is missing", () => {
    const result = parseUrlState("");
    expect(result.period).toEqual({ kind: "preset", preset: "allTime" });
  });
});

// ---------------------------------------------------------------------------
// parseUrlState — custom range (US-5, US-3, US-8)
// ---------------------------------------------------------------------------

describe("parseUrlState — custom range", () => {
  it("parses from and to as custom period", () => {
    const result = parseUrlState("?from=2026-01-01&to=2026-03-01");
    expect(result.period).toEqual({
      kind: "custom",
      range: { from: "2026-01-01", to: "2026-03-01" },
    });
  });

  it("falls back to allTime when only from is present", () => {
    const result = parseUrlState("?from=2026-01-01");
    expect(result.period).toEqual({ kind: "preset", preset: "allTime" });
  });

  it("falls back to allTime when only to is present", () => {
    const result = parseUrlState("?to=2026-03-01");
    expect(result.period).toEqual({ kind: "preset", preset: "allTime" });
  });

  it("falls back to allTime for invalid date format in from", () => {
    const result = parseUrlState("?from=not-a-date&to=2026-03-01");
    expect(result.period).toEqual({ kind: "preset", preset: "allTime" });
  });

  it("falls back to allTime for invalid date format in to", () => {
    const result = parseUrlState("?from=2026-01-01&to=13-2026");
    expect(result.period).toEqual({ kind: "preset", preset: "allTime" });
  });

  it("custom range takes precedence over period preset", () => {
    const result = parseUrlState("?period=last7&from=2026-01-01&to=2026-03-01");
    expect(result.period).toEqual({
      kind: "custom",
      range: { from: "2026-01-01", to: "2026-03-01" },
    });
  });
});

// ---------------------------------------------------------------------------
// parseUrlState — tag and area (US-5)
// ---------------------------------------------------------------------------

describe("parseUrlState — tag and area", () => {
  it("parses tag parameter", () => {
    expect(parseUrlState("?tag=api").tag).toBe("api");
  });

  it("returns null when tag is missing", () => {
    expect(parseUrlState("").tag).toBeNull();
  });

  it("returns null for empty tag value", () => {
    expect(parseUrlState("?tag=").tag).toBeNull();
  });

  it("parses area parameter", () => {
    expect(parseUrlState("?area=Platform").area).toBe("Platform");
  });

  it("returns null when area is missing", () => {
    expect(parseUrlState("").area).toBeNull();
  });

  it("returns null for empty area value", () => {
    expect(parseUrlState("?area=").area).toBeNull();
  });

  it("decodes URL-encoded area names", () => {
    expect(parseUrlState("?area=My%20Area").area).toBe("My Area");
  });
});

// ---------------------------------------------------------------------------
// parseUrlState — all defaults (US-6)
// ---------------------------------------------------------------------------

describe("parseUrlState — defaults", () => {
  it("returns all defaults for empty string", () => {
    const result = parseUrlState("");
    expect(result).toEqual({
      page: "queue",
      period: { kind: "preset", preset: "allTime" },
      tag: null,
      area: null,
    });
  });

  it("returns all defaults for bare question mark", () => {
    const result = parseUrlState("?");
    expect(result).toEqual({
      page: "queue",
      period: { kind: "preset", preset: "allTime" },
      tag: null,
      area: null,
    });
  });
});

// ---------------------------------------------------------------------------
// buildSearch — default omission (US-4)
// ---------------------------------------------------------------------------

describe("buildSearch — defaults", () => {
  it("returns empty string for all-default state", () => {
    const result = buildSearch({
      page: "queue",
      period: { kind: "preset", preset: "allTime" },
      tag: null,
      area: null,
    });
    expect(result).toBe("");
  });
});

// ---------------------------------------------------------------------------
// buildSearch — page (US-2)
// ---------------------------------------------------------------------------

describe("buildSearch — page", () => {
  it("includes page when not default", () => {
    const result = buildSearch({
      page: "slo",
      period: { kind: "preset", preset: "allTime" },
      tag: null,
      area: null,
    });
    expect(result).toBe("?page=slo");
  });

  it("omits page when queue", () => {
    const result = buildSearch({
      page: "queue",
      period: { kind: "preset", preset: "allTime" },
      tag: null,
      area: null,
    });
    expect(result).toBe("");
  });
});

// ---------------------------------------------------------------------------
// buildSearch — period (US-2, US-3)
// ---------------------------------------------------------------------------

describe("buildSearch — period", () => {
  it("includes period for non-default preset", () => {
    const result = buildSearch({
      page: "queue",
      period: { kind: "preset", preset: "last30" },
      tag: null,
      area: null,
    });
    expect(result).toBe("?period=last30");
  });

  it("includes from and to for custom range, omits period", () => {
    const result = buildSearch({
      page: "queue",
      period: { kind: "custom", range: { from: "2026-01-01", to: "2026-03-01" } },
      tag: null,
      area: null,
    });
    expect(result).toBe("?from=2026-01-01&to=2026-03-01");
  });

  it("omits period for allTime", () => {
    const result = buildSearch({
      page: "queue",
      period: { kind: "preset", preset: "allTime" },
      tag: null,
      area: null,
    });
    expect(result).toBe("");
  });
});

// ---------------------------------------------------------------------------
// buildSearch — tag and area (US-2)
// ---------------------------------------------------------------------------

describe("buildSearch — tag and area", () => {
  it("includes tag when set", () => {
    const result = buildSearch({
      page: "queue",
      period: { kind: "preset", preset: "allTime" },
      tag: "api",
      area: null,
    });
    expect(result).toBe("?tag=api");
  });

  it("includes area when set", () => {
    const result = buildSearch({
      page: "queue",
      period: { kind: "preset", preset: "allTime" },
      tag: null,
      area: "Platform",
    });
    expect(result).toBe("?area=Platform");
  });

  it("encodes area names with spaces", () => {
    const result = buildSearch({
      page: "queue",
      period: { kind: "preset", preset: "allTime" },
      tag: null,
      area: "My Area",
    });
    expect(result).toContain("area=My+Area");
  });
});

// ---------------------------------------------------------------------------
// buildSearch — multiple params
// ---------------------------------------------------------------------------

describe("buildSearch — multiple params", () => {
  it("combines multiple non-default params", () => {
    const result = buildSearch({
      page: "slo",
      period: { kind: "preset", preset: "last7" },
      tag: "api",
      area: "Platform",
    });
    const params = new URLSearchParams(result);
    expect(params.get("page")).toBe("slo");
    expect(params.get("period")).toBe("last7");
    expect(params.get("tag")).toBe("api");
    expect(params.get("area")).toBe("Platform");
  });
});

// ---------------------------------------------------------------------------
// Round-trip (US-5)
// ---------------------------------------------------------------------------

describe("round-trip", () => {
  it("parse(build(state)) equals state for all-default", () => {
    const state = {
      page: "queue" as const,
      period: { kind: "preset" as const, preset: "allTime" as const },
      tag: null,
      area: null,
    };
    expect(parseUrlState(buildSearch(state))).toEqual(state);
  });

  it("parse(build(state)) equals state for non-default values", () => {
    const state = {
      page: "activity" as const,
      period: { kind: "preset" as const, preset: "last30" as const },
      tag: "billing",
      area: "Platform",
    };
    expect(parseUrlState(buildSearch(state))).toEqual(state);
  });

  it("parse(build(state)) equals state for custom period", () => {
    const state = {
      page: "distribution" as const,
      period: { kind: "custom" as const, range: { from: "2026-01-15", to: "2026-02-28" } },
      tag: null,
      area: null,
    };
    expect(parseUrlState(buildSearch(state))).toEqual(state);
  });
});
