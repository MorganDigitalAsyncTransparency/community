// Spec: specs/dashboard/queue-visibility.md, specs/dashboard/response-metrics.md,
//       specs/dashboard/tag-distribution.md
// Tests: tests/dashboard/queue-visibility.unit.test.ts

const MILLISECONDS_PER_HOUR = 3_600_000;
const HOURS_PER_DAY = 24;

export function formatDuration(ms: number): string {
  const hours = Math.floor(ms / MILLISECONDS_PER_HOUR);

  if (hours >= HOURS_PER_DAY) {
    return `${Math.floor(hours / HOURS_PER_DAY)}d`;
  }

  return `${Math.max(1, hours)}h`;
}

export function formatAge(isoDate: string): string {
  return formatDuration(Date.now() - new Date(isoDate).getTime());
}

export function formatTags(tags: string[]): string {
  return tags.length > 0 ? tags.join(", ") : "–";
}

// Formats a YYYY-MM-DD week-start date (UTC Monday) as a locale-aware short date.
export function formatWeekLabel(isoDate: string): string {
  return new Date(isoDate + "T00:00:00Z").toLocaleDateString(undefined, {
    year: "numeric",
    month: "short",
    day: "numeric",
    timeZone: "UTC",
  });
}
