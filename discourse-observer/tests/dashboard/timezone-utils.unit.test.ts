// Spec: specs/dashboard/peak-activity.md (PA-17 – PA-19, PA-21, PA-22, PA-25)

import { describe, expect, it, beforeEach, afterEach, vi } from "vitest";
import {
  utcOffsetMinutes,
  formatOffsetHour,
  formatUtcOffset,
  getTimezoneList,
} from "../../frontend/src/components/timezoneUtils";
import {
  readTimezoneCookie,
  writeTimezoneCookie,
  readConsentCookie,
  writeConsentCookie,
} from "../../frontend/src/components/timezoneCookies";

// ---------------------------------------------------------------------------
// utcOffsetMinutes (PA-18)
// ---------------------------------------------------------------------------

describe("utcOffsetMinutes", () => {
  it("returns a whole-hour offset for a standard timezone", () => {
    // Europe/London is either +0 or +1 depending on DST — both are whole-hour
    const offset = utcOffsetMinutes("Europe/London");
    expect(offset % 60).toBe(0);
  });

  it("returns a half-hour offset for Asia/Kolkata", () => {
    // IST is always UTC+5:30
    const offset = utcOffsetMinutes("Asia/Kolkata");
    expect(offset).toBe(330);
  });

  it("returns a quarter-hour offset for Asia/Kathmandu", () => {
    // Nepal is always UTC+5:45
    const offset = utcOffsetMinutes("Asia/Kathmandu");
    expect(offset).toBe(345);
  });

  it("returns 0 for UTC", () => {
    expect(utcOffsetMinutes("UTC")).toBe(0);
  });
});

// ---------------------------------------------------------------------------
// formatOffsetHour (PA-17, PA-19)
// ---------------------------------------------------------------------------

describe("formatOffsetHour", () => {
  it("produces integer label for whole-hour offset", () => {
    // UTC hour 8 with offset +60min (CET) → 9
    expect(formatOffsetHour(8, 60)).toBe("9");
  });

  it("produces H:MM label for fractional offset", () => {
    // UTC hour 8 with offset +330min (IST +5:30) → 13:30
    expect(formatOffsetHour(8, 330)).toBe("13:30");
  });

  it("wraps around 24 correctly", () => {
    // UTC hour 23 with offset +180min (+3) → 2
    expect(formatOffsetHour(23, 180)).toBe("2");
  });

  it("wraps around negative correctly", () => {
    // UTC hour 1 with offset -300min (-5) → 20
    expect(formatOffsetHour(1, -300)).toBe("20");
  });

  it("handles zero offset", () => {
    expect(formatOffsetHour(14, 0)).toBe("14");
  });

  it("handles midnight boundary", () => {
    // UTC hour 0 with offset +0 → 0
    expect(formatOffsetHour(0, 0)).toBe("0");
  });

  it("handles fractional negative offset", () => {
    // UTC hour 0 with offset -330min (-5:30) → 18:30
    expect(formatOffsetHour(0, -330)).toBe("18:30");
  });
});

// ---------------------------------------------------------------------------
// formatUtcOffset (PA-17)
// ---------------------------------------------------------------------------

describe("formatUtcOffset", () => {
  it("formats positive whole-hour offset", () => {
    expect(formatUtcOffset(60)).toBe("+1");
  });

  it("formats negative whole-hour offset", () => {
    expect(formatUtcOffset(-300)).toBe("\u22125");
  });

  it("formats zero offset", () => {
    expect(formatUtcOffset(0)).toBe("+0");
  });

  it("formats positive fractional offset", () => {
    expect(formatUtcOffset(330)).toBe("+5:30");
  });

  it("formats negative fractional offset", () => {
    expect(formatUtcOffset(-210)).toBe("\u22123:30");
  });
});

// ---------------------------------------------------------------------------
// Cookie functions (PA-25)
// ---------------------------------------------------------------------------

describe("cookie functions", () => {
  // Simulate document.cookie with a simple in-memory store.
  // Real document.cookie is a getter/setter that accumulates key=value pairs.
  let cookieStore: string;

  beforeEach(() => {
    cookieStore = "";
    vi.stubGlobal("document", {
      get cookie() {
        return cookieStore;
      },
      set cookie(value: string) {
        // Emulate browser behavior: setting document.cookie adds/replaces one key.
        const name = value.split("=")[0];
        const isDelete = value.includes("max-age=0");
        const pairs = cookieStore
          .split("; ")
          .filter((p) => p && !p.startsWith(`${name}=`));
        if (!isDelete) {
          // Strip attributes (path, max-age, SameSite) for storage
          const kvPart = value.split(";")[0];
          pairs.push(kvPart);
        }
        cookieStore = pairs.join("; ");
      },
    });
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  describe("readTimezoneCookie", () => {
    it("returns empty array when no cookie exists", () => {
      expect(readTimezoneCookie()).toEqual([]);
    });

    it("returns parsed array from valid cookie", () => {
      const value = encodeURIComponent(JSON.stringify(["Europe/Berlin", "Asia/Kolkata"]));
      document.cookie = `peak_tz=${value}; path=/`;
      expect(readTimezoneCookie()).toEqual(["Europe/Berlin", "Asia/Kolkata"]);
    });

    it("returns empty array for malformed cookie", () => {
      document.cookie = "peak_tz=not-valid-json; path=/";
      expect(readTimezoneCookie()).toEqual([]);
    });

    it("filters out non-string values", () => {
      const value = encodeURIComponent(JSON.stringify(["Europe/Berlin", 42, null]));
      document.cookie = `peak_tz=${value}; path=/`;
      expect(readTimezoneCookie()).toEqual(["Europe/Berlin"]);
    });
  });

  describe("writeTimezoneCookie", () => {
    it("writes JSON-encoded array", () => {
      writeTimezoneCookie(["Europe/Berlin"]);
      expect(readTimezoneCookie()).toEqual(["Europe/Berlin"]);
    });

    it("overwrites previous value", () => {
      writeTimezoneCookie(["Europe/Berlin"]);
      writeTimezoneCookie(["Asia/Tokyo"]);
      expect(readTimezoneCookie()).toEqual(["Asia/Tokyo"]);
    });
  });

  describe("consent cookie", () => {
    it("returns null when no consent cookie exists", () => {
      expect(readConsentCookie()).toBeNull();
    });

    it("round-trips accepted consent", () => {
      writeConsentCookie();
      expect(readConsentCookie()).toBe("accepted");
    });
  });
});

// ---------------------------------------------------------------------------
// getTimezoneList (PA-21, PA-22)
// ---------------------------------------------------------------------------

describe("getTimezoneList", () => {
  it("has no duplicate entries", () => {
    const list = getTimezoneList();
    const ids = list.map((tz) => tz.id);
    expect(new Set(ids).size).toBe(ids.length);
  });

  it("every entry is a valid IANA identifier", () => {
    for (const tz of getTimezoneList()) {
      expect(() => {
        new Intl.DateTimeFormat("en-US", { timeZone: tz.id });
      }).not.toThrow();
    }
  });

  it("every entry has a code and at least one city", () => {
    for (const tz of getTimezoneList()) {
      expect(tz.code.length).toBeGreaterThan(0);
      expect(tz.cities.length).toBeGreaterThan(0);
    }
  });

  it("is sorted by offset ascending", () => {
    const list = getTimezoneList();
    for (let i = 1; i < list.length; i++) {
      expect(list[i].offsetMinutes).toBeGreaterThanOrEqual(list[i - 1].offsetMinutes);
    }
  });
});
