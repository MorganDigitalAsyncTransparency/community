// Spec: specs/dashboard/tag-area-filter.md, specs/dashboard/stalled-topics.md
// Tests: tests/dashboard/tag-area-filter.unit.test.ts

import type { Topic } from "../mock/data";

// ---------------------------------------------------------------------------
// Configuration types — mirrors config/tagConfig.json
// ---------------------------------------------------------------------------

export interface SloThresholds {
  firstReplyHours: number;
  resolutionHours: number;
  inactivityHours: number;
}

export interface TagEntry {
  area?: string;
  closedTag?: string;
  stalledDays?: number;
  slo?: SloThresholds;
}

export interface AreaEntry {
  name: string;
  primaryTag: string;
}

export interface DefaultsEntry {
  stalledDays: number;
  area: string;
  slo: SloThresholds;
}

export interface TagConfig {
  defaults: DefaultsEntry;
  areas: AreaEntry[];
  tags: Record<string, TagEntry>;
}

// ---------------------------------------------------------------------------
// Resolved tag — after applying defaults, with provenance tracking
// ---------------------------------------------------------------------------

export interface ResolvedTag {
  area: string;
  areaIsDefault: boolean;
  closedTag: string | null;
  stalledDays: number;
  stalledDaysIsDefault: boolean;
  slo: SloThresholds;
  sloIsDefault: boolean;
}

export function resolveTag(entry: TagEntry, defaults: DefaultsEntry): ResolvedTag {
  return {
    area: entry.area ?? defaults.area,
    areaIsDefault: entry.area === undefined,
    closedTag: entry.closedTag ?? null,
    stalledDays: entry.stalledDays ?? defaults.stalledDays,
    stalledDaysIsDefault: entry.stalledDays === undefined,
    slo: entry.slo ?? defaults.slo,
    sloIsDefault: entry.slo === undefined,
  };
}

export function resolveAllTags(
  config: TagConfig,
): Record<string, ResolvedTag> {
  const result: Record<string, ResolvedTag> = {};
  for (const [tag, entry] of Object.entries(config.tags)) {
    result[tag] = resolveTag(entry, config.defaults);
  }
  return result;
}

// ---------------------------------------------------------------------------
// Tag and area queries
// ---------------------------------------------------------------------------

export function monitoredTags(config: TagConfig): string[] {
  return Object.keys(config.tags);
}

export function tagsForArea(
  config: TagConfig,
  area: string | null,
): string[] {
  const resolved = resolveAllTags(config);

  if (area === null) {
    return Object.keys(resolved).sort();
  }

  const areaEntry = config.areas.find((a) => a.name === area);

  const matching = Object.entries(resolved)
    .filter(([, r]) => r.area === area)
    .map(([tag]) => tag);

  if (areaEntry) {
    const rest = matching
      .filter((t) => t !== areaEntry.primaryTag)
      .sort();
    return [areaEntry.primaryTag, ...rest];
  }

  return matching.sort();
}

export function allAreas(config: TagConfig): string[] {
  const resolved = resolveAllTags(config);
  const defaultArea = config.defaults.area;

  const named = config.areas.map((a) => a.name);
  const hasDefault = Object.values(resolved).some((r) => r.area === defaultArea);

  if (hasDefault && !named.includes(defaultArea)) {
    return [...named, defaultArea];
  }

  return named;
}

// ---------------------------------------------------------------------------
// Topic filtering
// ---------------------------------------------------------------------------

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

// ---------------------------------------------------------------------------
// SLO config extraction
// ---------------------------------------------------------------------------

export type SloConfig = Record<string, SloThresholds>;

export function extractSloConfig(config: TagConfig): SloConfig {
  const resolved = resolveAllTags(config);
  const result: SloConfig = {};
  for (const [tag, r] of Object.entries(resolved)) {
    result[tag] = r.slo;
  }
  return result;
}

export function scopeSloConfig(
  full: SloConfig,
  visibleTags: string[],
): SloConfig {
  const set = new Set(visibleTags);
  const result: SloConfig = {};
  for (const [tag, thresholds] of Object.entries(full)) {
    if (set.has(tag)) {
      result[tag] = thresholds;
    }
  }
  return result;
}

export function sloDefaultTags(config: TagConfig): Set<string> {
  const result = new Set<string>();
  for (const [tag, entry] of Object.entries(config.tags)) {
    if (entry.slo === undefined) {
      result.add(tag);
    }
  }
  return result;
}
