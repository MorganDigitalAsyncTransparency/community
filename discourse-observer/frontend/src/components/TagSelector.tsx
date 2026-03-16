// Spec: specs/dashboard/tag-area-filter.md
// Tests: tests/dashboard/tag-area-filter.unit.test.ts

import { type TagConfig, tagsForArea } from "./tagFilter";

interface TagSelectorProps {
  config: TagConfig;
  activeTag: string | null;
  activeArea: string | null;
  onTagSelect: (tag: string | null) => void;
  onAreaSelect: (area: string | null) => void;
}

export function TagSelector({
  config,
  activeTag,
  activeArea,
  onTagSelect,
  onAreaSelect,
}: TagSelectorProps) {
  const visibleTags = tagsForArea(config, activeArea);

  return (
    <div className="tag-selector">
      <span className="tag-selector-label">Area:</span>
      <select
        className="tag-area-select"
        value={activeArea ?? ""}
        onChange={(e) => onAreaSelect(e.target.value || null)}
      >
        <option value="">All areas</option>
        {config.map((area) => (
          <option key={area.name} value={area.name}>
            {area.name}
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
          {tag}
        </button>
      ))}
    </div>
  );
}
