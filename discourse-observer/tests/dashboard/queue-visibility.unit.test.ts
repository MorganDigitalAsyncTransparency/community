import { describe, expect, it, vi, afterEach } from "vitest";
import {
  formatAge,
  formatTags,
} from "../../frontend/src/components/topicFormatting";

const HOUR_MS = 3_600_000;
const DAY_MS = 86_400_000;

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
