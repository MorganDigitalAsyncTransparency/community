// Spec: specs/dashboard/tag-area-filter.md, specs/dashboard/stalled-topics.md
// Tests: tests/dashboard/tag-area-filter.unit.test.ts

import type { Topic } from "../mock/data";

export interface AreaConfig {
  name: string;
  primaryTag: string;
  tags: string[];
}

export interface TagConfig {
  closedTag: string;
  stalledDays: number;
  areas: AreaConfig[];
}

export function monitoredTags(config: TagConfig): string[] {
  const seen = new Set<string>();
  for (const area of config.areas) {
    for (const tag of area.tags) {
      seen.add(tag);
    }
  }
  return [...seen];
}

export function filterByTag(topics: Topic[], tag: string | null): Topic[] {
  if (tag === null) return topics;
  return topics.filter((t) => t.tags.includes(tag));
}

export function filterByMonitoredTags(
  topics: Topic[],
  monitored: string[],
): Topic[] {
  const set = new Set(monitored);
  return topics.filter((t) => t.tags.some((tag) => set.has(tag)));
}

export function tagsForArea(
  config: TagConfig,
  area: string | null,
): string[] {
  if (area === null) {
    return monitoredTags(config).sort();
  }

  const found = config.areas.find((a) => a.name === area);
  if (!found) return [];

  const rest = found.tags
    .filter((t) => t !== found.primaryTag)
    .sort();
  return [found.primaryTag, ...rest];
}
