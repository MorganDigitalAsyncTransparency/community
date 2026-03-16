// Spec: specs/dashboard/time-period-filter.md

import { useState } from "react";
import {
  type ActivePeriod,
  type PeriodPreset,
  PRESET_LABELS,
} from "./timePeriod";

const PRESETS: PeriodPreset[] = ["last7", "last30", "lastYear", "allTime"];

interface PeriodSelectorProps {
  period: ActivePeriod;
  onPeriodChange: (period: ActivePeriod) => void;
}

export function PeriodSelector({ period, onPeriodChange }: PeriodSelectorProps) {
  const [showCustom, setShowCustom] = useState(period.kind === "custom");
  const [customFrom, setCustomFrom] = useState(
    period.kind === "custom" ? period.range.from : ""
  );
  const [customTo, setCustomTo] = useState(
    period.kind === "custom" ? period.range.to : ""
  );

  function handlePresetClick(preset: PeriodPreset) {
    setShowCustom(false);
    onPeriodChange({ kind: "preset", preset });
  }

  function handleCustomClick() {
    setShowCustom(true);
  }

  function handleFromChange(value: string) {
    setCustomFrom(value);
    if (value && customTo) {
      onPeriodChange({ kind: "custom", range: { from: value, to: customTo } });
    }
  }

  function handleToChange(value: string) {
    setCustomTo(value);
    if (customFrom && value) {
      onPeriodChange({ kind: "custom", range: { from: customFrom, to: value } });
    }
  }

  const activePreset = period.kind === "preset" ? period.preset : null;

  return (
    <div className="period-selector">
      <span className="period-selector-label">Period:</span>

      {PRESETS.map((preset) => (
        <button
          key={preset}
          className={`period-btn${activePreset === preset ? " period-btn-active" : ""}`}
          onClick={() => handlePresetClick(preset)}
        >
          {PRESET_LABELS[preset]}
        </button>
      ))}

      <button
        className={`period-btn${showCustom ? " period-btn-active" : ""}`}
        onClick={handleCustomClick}
      >
        Custom
      </button>

      {showCustom && (
        <span className="period-custom-inputs">
          <input
            type="date"
            className="period-date-input"
            value={customFrom}
            onChange={(e) => handleFromChange(e.target.value)}
          />
          <span className="period-custom-separator">–</span>
          <input
            type="date"
            className="period-date-input"
            value={customTo}
            onChange={(e) => handleToChange(e.target.value)}
          />
        </span>
      )}
    </div>
  );
}
