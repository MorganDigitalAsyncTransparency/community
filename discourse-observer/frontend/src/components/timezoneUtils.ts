// Spec: specs/dashboard/peak-activity.md
// Tests: tests/dashboard/timezone-utils.unit.test.ts

/**
 * Returns the UTC offset in minutes for an IANA timezone at the current moment.
 * Positive values mean ahead of UTC (e.g. CET = +60).
 */
export function utcOffsetMinutes(timeZone: string): number {
  const formatter = new Intl.DateTimeFormat("en-US", {
    timeZone,
    timeZoneName: "shortOffset",
  });
  const parts = formatter.formatToParts(new Date());
  const tzPart = parts.find((p) => p.type === "timeZoneName");
  if (!tzPart) return 0;

  // tzPart.value is like "GMT", "GMT+1", "GMT-5", "GMT+5:30"
  const raw = tzPart.value.replace("GMT", "");
  if (raw === "") return 0;

  const sign = raw.startsWith("-") ? -1 : 1;
  const abs = raw.replace(/^[+-]/, "");
  const [hoursStr, minutesStr] = abs.split(":");
  const hours = parseInt(hoursStr, 10) || 0;
  const minutes = parseInt(minutesStr, 10) || 0;
  return sign * (hours * 60 + minutes);
}

/**
 * Returns the display label for a UTC hour adjusted by an offset in minutes.
 * Whole-hour offsets produce integer strings ("14").
 * Fractional offsets produce "H:MM" strings ("14:30").
 */
export function formatOffsetHour(utcHour: number, offsetMinutes: number): string {
  const totalMinutes = utcHour * 60 + offsetMinutes;
  const wrapped = ((totalMinutes % 1440) + 1440) % 1440;
  const h = Math.floor(wrapped / 60);
  const m = wrapped % 60;
  return m === 0 ? String(h) : `${h}:${String(m).padStart(2, "0")}`;
}

/**
 * Returns the short display name for a timezone (e.g. "CET", "IST").
 */
export function timezoneShortName(timeZone: string): string {
  const formatter = new Intl.DateTimeFormat("en-US", {
    timeZone,
    timeZoneName: "short",
  });
  const parts = formatter.formatToParts(new Date());
  const tzPart = parts.find((p) => p.type === "timeZoneName");
  return tzPart?.value ?? timeZone;
}

/**
 * Returns a human-readable UTC offset string like "+1", "-5", "+5:30".
 */
export function formatUtcOffset(offsetMinutes: number): string {
  if (offsetMinutes === 0) return "+0";
  const sign = offsetMinutes >= 0 ? "+" : "\u2212";
  const abs = Math.abs(offsetMinutes);
  const h = Math.floor(abs / 60);
  const m = abs % 60;
  return m === 0 ? `${sign}${h}` : `${sign}${h}:${String(m).padStart(2, "0")}`;
}

/**
 * Curated list of commonly used IANA timezones, each with a representative
 * IANA id, its short code, and a list of major cities sharing that offset.
 * Sorted by UTC offset (ascending) for display in the picker.
 */
export interface TimezoneEntry {
  id: string;
  code: string;
  cities: string[];
  offsetMinutes: number;
}

/**
 * Builds the timezone list with live offsets. Called once and cached.
 * Each entry's offsetMinutes reflects the current offset (DST-aware).
 */
function buildTimezoneList(): TimezoneEntry[] {
  const raw: { id: string; code: string; cities: string[] }[] = [
    { id: "Pacific/Honolulu", code: "HST", cities: ["Honolulu"] },
    { id: "America/Anchorage", code: "AKST", cities: ["Anchorage"] },
    { id: "America/Los_Angeles", code: "PST", cities: ["Los Angeles", "Vancouver", "Seattle"] },
    { id: "America/Denver", code: "MST", cities: ["Denver", "Phoenix", "Salt Lake City"] },
    { id: "America/Chicago", code: "CST", cities: ["Chicago", "Mexico City", "Houston"] },
    { id: "America/New_York", code: "EST", cities: ["New York", "Toronto", "Montreal"] },
    { id: "America/Halifax", code: "AST", cities: ["Halifax", "Santiago"] },
    { id: "America/Sao_Paulo", code: "BRT", cities: ["São Paulo", "Buenos Aires"] },
    { id: "Atlantic/Reykjavik", code: "GMT", cities: ["Reykjavik"] },
    { id: "Europe/London", code: "GMT/BST", cities: ["London", "Dublin", "Lisbon"] },
    { id: "Europe/Berlin", code: "CET", cities: ["Berlin", "Paris", "Stockholm", "Oslo", "Rome", "Madrid", "Amsterdam", "Brussels", "Prague", "Vienna", "Warsaw", "Zurich"] },
    { id: "Europe/Athens", code: "EET", cities: ["Athens", "Helsinki", "Bucharest", "Kyiv"] },
    { id: "Europe/Moscow", code: "MSK", cities: ["Moscow", "Istanbul"] },
    { id: "Asia/Dubai", code: "GST", cities: ["Dubai"] },
    { id: "Asia/Karachi", code: "PKT", cities: ["Karachi"] },
    { id: "Asia/Kolkata", code: "IST", cities: ["Mumbai", "Delhi", "Kolkata", "Colombo"] },
    { id: "Asia/Kathmandu", code: "NPT", cities: ["Kathmandu"] },
    { id: "Asia/Dhaka", code: "BST", cities: ["Dhaka"] },
    { id: "Asia/Bangkok", code: "ICT", cities: ["Bangkok", "Ho Chi Minh City", "Jakarta"] },
    { id: "Asia/Shanghai", code: "CST", cities: ["Shanghai", "Hong Kong", "Taipei", "Singapore", "Kuala Lumpur", "Manila", "Perth"] },
    { id: "Asia/Tokyo", code: "JST", cities: ["Tokyo", "Seoul"] },
    { id: "Australia/Adelaide", code: "ACST", cities: ["Adelaide", "Darwin"] },
    { id: "Australia/Sydney", code: "AEST", cities: ["Sydney", "Melbourne", "Brisbane", "Guam"] },
    { id: "Pacific/Auckland", code: "NZST", cities: ["Auckland", "Fiji"] },
  ];

  return raw
    .map((entry) => ({
      ...entry,
      offsetMinutes: utcOffsetMinutes(entry.id),
    }))
    .sort((a, b) => a.offsetMinutes - b.offsetMinutes);
}

let _cachedList: TimezoneEntry[] | null = null;

export function getTimezoneList(): TimezoneEntry[] {
  if (!_cachedList) _cachedList = buildTimezoneList();
  return _cachedList;
}
