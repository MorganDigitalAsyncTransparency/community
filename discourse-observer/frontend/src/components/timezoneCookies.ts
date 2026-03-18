// Spec: specs/dashboard/peak-activity.md
// Tests: tests/dashboard/timezone-utils.unit.test.ts

const TZ_COOKIE = "peak_tz";
const CONSENT_COOKIE = "peak_tz_consent";
const MAX_AGE_DAYS = 365;

export function readTimezoneCookie(): string[] {
  const match = document.cookie
    .split("; ")
    .find((row) => row.startsWith(`${TZ_COOKIE}=`));
  if (!match) return [];
  try {
    const value = decodeURIComponent(match.split("=")[1]);
    const parsed: unknown = JSON.parse(value);
    if (!Array.isArray(parsed)) return [];
    return parsed.filter((v): v is string => typeof v === "string");
  } catch {
    return [];
  }
}

export function writeTimezoneCookie(timezones: string[]): void {
  const value = encodeURIComponent(JSON.stringify(timezones));
  const maxAge = MAX_AGE_DAYS * 24 * 60 * 60;
  document.cookie = `${TZ_COOKIE}=${value}; max-age=${maxAge}; SameSite=Lax; path=/`;
}

export function readConsentCookie(): "accepted" | null {
  const match = document.cookie
    .split("; ")
    .find((row) => row.startsWith(`${CONSENT_COOKIE}=`));
  if (!match) return null;
  return match.split("=")[1] === "accepted" ? "accepted" : null;
}

export function writeConsentCookie(): void {
  const maxAge = MAX_AGE_DAYS * 24 * 60 * 60;
  document.cookie = `${CONSENT_COOKIE}=accepted; max-age=${maxAge}; SameSite=Lax; path=/`;
}
