// Spec: specs/dashboard/tag-area-filter.md
// Tests: tests/dashboard/tag-area-filter.unit.test.ts

import type { AppConfig } from "../api/types";

interface TagSelectorProps {
  config: AppConfig;
  activeTag: string | null;
  activeArea: string | null;
  onTagSelect: (tag: string | null) => void;
  onAreaSelect: (area: string | null) => void;
}

function tagsForArea(config: AppConfig, area: string | null): string[] {
  const allTags = Object.keys(config.tags).sort();
  if (area === null) return allTags;

  const areaEntry = config.areas.find((a) => a.name === area);
  const matching = allTags.filter((tag) => config.tags[tag].area === area);

  if (areaEntry && matching.includes(areaEntry.primaryTag)) {
    const rest = matching.filter((t) => t !== areaEntry.primaryTag);
    return [areaEntry.primaryTag, ...rest];
  }

  return matching;
}

function allAreas(config: AppConfig): string[] {
  const named = config.areas.map((a) => a.name);
  const defaultArea = config.defaults.area;
  const hasDefault = Object.values(config.tags).some((t) => t.area === defaultArea);

  if (hasDefault && !named.includes(defaultArea)) {
    return [...named, defaultArea];
  }

  return named;
}

export function TagSelector({
  config,
  activeTag,
  activeArea,
  onTagSelect,
  onAreaSelect,
}: TagSelectorProps) {
  const visibleTags = tagsForArea(config, activeArea);
  const areas = allAreas(config);
  const primaryTags = new Set(config.areas.map((a) => a.primaryTag));

  return (
    <div className="tag-selector">
      <span className="tag-selector-label">Area:</span>
      <select
        className="tag-area-select"
        value={activeArea ?? ""}
        onChange={(e) => onAreaSelect(e.target.value || null)}
      >
        <option value="">All areas</option>
        {areas.map((area) => (
          <option key={area} value={area}>
            {area}
          </option>
        ))}
      </select>

      <span className="tag-selector-label tag-selector-label-tag">Tag:</span>
      <button
        className={`tag-btn${activeTag === null ? " tag-btn-active" : ""}`}
        onClick={() => onTagSelect(null)}
      >
        All
      </button>
      {visibleTags.map((tag) => (
        <button
          key={tag}
          className={`tag-btn${activeTag === tag ? " tag-btn-active" : ""}`}
          onClick={() => onTagSelect(tag)}
        >
          {tag}{primaryTags.has(tag) ? "*" : ""}
        </button>
      ))}
    </div>
  );
}

// Re-export for use by App.tsx
export { tagsForArea, allAreas };
