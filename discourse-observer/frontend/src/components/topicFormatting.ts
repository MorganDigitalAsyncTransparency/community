import type { Topic } from "../mock/data";

const MILLISECONDS_PER_HOUR = 3_600_000;
const HOURS_PER_DAY = 24;

export function formatAge(isoDate: string): string {
  const elapsedMs = Date.now() - new Date(isoDate).getTime();
  const elapsedHours = Math.floor(elapsedMs / MILLISECONDS_PER_HOUR);

  if (elapsedHours >= HOURS_PER_DAY) {
    return `${Math.floor(elapsedHours / HOURS_PER_DAY)}d`;
  }

  return `${Math.max(1, elapsedHours)}h`;
}

export function sortedByOldest(topics: Topic[]): Topic[] {
  return [...topics].sort(
    (a, b) => new Date(a.createdAt).getTime() - new Date(b.createdAt).getTime()
  );
}
