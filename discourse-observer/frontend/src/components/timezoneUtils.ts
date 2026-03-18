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
 * Curated list of commonly used IANA timezones grouped by region.
 */
export interface TimezoneEntry {
  id: string;
  region: string;
}

export const TIMEZONE_LIST: TimezoneEntry[] = [
  // Africa
  { id: "Africa/Cairo", region: "Africa" },
  { id: "Africa/Casablanca", region: "Africa" },
  { id: "Africa/Johannesburg", region: "Africa" },
  { id: "Africa/Lagos", region: "Africa" },
  { id: "Africa/Nairobi", region: "Africa" },
  // America
  { id: "America/Anchorage", region: "America" },
  { id: "America/Argentina/Buenos_Aires", region: "America" },
  { id: "America/Bogota", region: "America" },
  { id: "America/Chicago", region: "America" },
  { id: "America/Denver", region: "America" },
  { id: "America/Halifax", region: "America" },
  { id: "America/Lima", region: "America" },
  { id: "America/Los_Angeles", region: "America" },
  { id: "America/Mexico_City", region: "America" },
  { id: "America/New_York", region: "America" },
  { id: "America/Phoenix", region: "America" },
  { id: "America/Santiago", region: "America" },
  { id: "America/Sao_Paulo", region: "America" },
  { id: "America/Toronto", region: "America" },
  { id: "America/Vancouver", region: "America" },
  // Asia
  { id: "Asia/Bangkok", region: "Asia" },
  { id: "Asia/Colombo", region: "Asia" },
  { id: "Asia/Dhaka", region: "Asia" },
  { id: "Asia/Dubai", region: "Asia" },
  { id: "Asia/Ho_Chi_Minh", region: "Asia" },
  { id: "Asia/Hong_Kong", region: "Asia" },
  { id: "Asia/Istanbul", region: "Asia" },
  { id: "Asia/Jakarta", region: "Asia" },
  { id: "Asia/Karachi", region: "Asia" },
  { id: "Asia/Kathmandu", region: "Asia" },
  { id: "Asia/Kolkata", region: "Asia" },
  { id: "Asia/Kuala_Lumpur", region: "Asia" },
  { id: "Asia/Manila", region: "Asia" },
  { id: "Asia/Riyadh", region: "Asia" },
  { id: "Asia/Seoul", region: "Asia" },
  { id: "Asia/Shanghai", region: "Asia" },
  { id: "Asia/Singapore", region: "Asia" },
  { id: "Asia/Taipei", region: "Asia" },
  { id: "Asia/Tehran", region: "Asia" },
  { id: "Asia/Tokyo", region: "Asia" },
  // Atlantic
  { id: "Atlantic/Reykjavik", region: "Atlantic" },
  // Australia
  { id: "Australia/Adelaide", region: "Australia" },
  { id: "Australia/Brisbane", region: "Australia" },
  { id: "Australia/Darwin", region: "Australia" },
  { id: "Australia/Melbourne", region: "Australia" },
  { id: "Australia/Perth", region: "Australia" },
  { id: "Australia/Sydney", region: "Australia" },
  // Europe
  { id: "Europe/Amsterdam", region: "Europe" },
  { id: "Europe/Athens", region: "Europe" },
  { id: "Europe/Berlin", region: "Europe" },
  { id: "Europe/Brussels", region: "Europe" },
  { id: "Europe/Bucharest", region: "Europe" },
  { id: "Europe/Dublin", region: "Europe" },
  { id: "Europe/Helsinki", region: "Europe" },
  { id: "Europe/Kiev", region: "Europe" },
  { id: "Europe/Lisbon", region: "Europe" },
  { id: "Europe/London", region: "Europe" },
  { id: "Europe/Madrid", region: "Europe" },
  { id: "Europe/Moscow", region: "Europe" },
  { id: "Europe/Oslo", region: "Europe" },
  { id: "Europe/Paris", region: "Europe" },
  { id: "Europe/Prague", region: "Europe" },
  { id: "Europe/Rome", region: "Europe" },
  { id: "Europe/Stockholm", region: "Europe" },
  { id: "Europe/Vienna", region: "Europe" },
  { id: "Europe/Warsaw", region: "Europe" },
  { id: "Europe/Zurich", region: "Europe" },
  // Pacific
  { id: "Pacific/Auckland", region: "Pacific" },
  { id: "Pacific/Fiji", region: "Pacific" },
  { id: "Pacific/Guam", region: "Pacific" },
  { id: "Pacific/Honolulu", region: "Pacific" },
];
