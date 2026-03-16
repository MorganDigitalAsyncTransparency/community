// Spec: specs/dashboard/time-period-filter.md
// Tests: tests/dashboard/time-period-filter.unit.test.ts

import {
  type ActivePeriod,
  type CustomRange,
  type PeriodPreset,
  PRESET_LABELS,
} from "./timePeriod";

const PRESETS: PeriodPreset[] = ["last7", "last30", "lastYear", "allTime"];

interface PeriodSelectorProps {
  period: ActivePeriod;
  customDraft: CustomRange | null; // null = custom inputs not visible
  onPresetSelect: (preset: PeriodPreset) => void;
  onCustomOpen: () => void;
  onCustomDraftChange: (from: string, to: string) => void;
}

export function PeriodSelector({
  period,
  customDraft,
  onPresetSelect,
  onCustomOpen,
  onCustomDraftChange,
}: PeriodSelectorProps) {
  const activePreset = period.kind === "preset" ? period.preset : null;
  const customActive = customDraft !== null;

  return (
    <div className="period-selector">
      <span className="period-selector-label">Period:</span>

      {PRESETS.map((preset) => (
        <button
          key={preset}
          className={`period-btn${activePreset === preset ? " period-btn-active" : ""}`}
          onClick={() => onPresetSelect(preset)}
        >
          {PRESET_LABELS[preset]}
        </button>
      ))}

      <button
        className={`period-btn${customActive ? " period-btn-active" : ""}`}
        onClick={onCustomOpen}
      >
        Custom
      </button>

      {customDraft !== null && (
        <span className="period-custom-inputs">
          <input
            type="date"
            className="period-date-input"
            value={customDraft.from}
            onChange={(e) => onCustomDraftChange(e.target.value, customDraft.to)}
          />
          <span className="period-custom-separator">–</span>
          <input
            type="date"
            className="period-date-input"
            value={customDraft.to}
            onChange={(e) => onCustomDraftChange(customDraft.from, e.target.value)}
          />
        </span>
      )}
    </div>
  );
}
