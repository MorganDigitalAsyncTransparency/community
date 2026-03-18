// Spec: specs/dashboard/peak-activity.md (PA-21, PA-22)
// Tests: tests/dashboard/timezone-utils.unit.test.ts

import { useState } from "react";
import { getTimezoneList, formatUtcOffset } from "./timezoneUtils";

interface TimezonePickerProps {
  onSelect: (tz: string) => void;
  onClose: () => void;
  excludeTimezones: string[];
}

export function TimezonePicker({
  onSelect,
  onClose,
  excludeTimezones,
}: TimezonePickerProps) {
  const [search, setSearch] = useState("");

  const excluded = new Set(excludeTimezones);
  const query = search.toLowerCase();

  const entries = getTimezoneList().filter((tz) => {
    if (excluded.has(tz.id)) return false;
    if (query === "") return true;
    const offsetStr = formatUtcOffset(tz.offsetMinutes).toLowerCase();
    return (
      tz.code.toLowerCase().includes(query) ||
      tz.cities.some((c) => c.toLowerCase().includes(query)) ||
      offsetStr.includes(query) ||
      tz.id.toLowerCase().includes(query)
    );
  });

  return (
    <div className="peak-tz-picker">
      <input
        className="peak-tz-picker-search"
        type="text"
        placeholder="Search by code, city, or offset..."
        value={search}
        onChange={(e) => setSearch(e.target.value)}
        autoFocus
      />
      <div className="peak-tz-picker-list">
        {entries.map((tz) => (
          <button
            key={tz.id}
            className="peak-tz-picker-item"
            onClick={() => {
              onSelect(tz.id);
              onClose();
            }}
          >
            <span className="peak-tz-picker-offset">
              {formatUtcOffset(tz.offsetMinutes)}
            </span>
            <span className="peak-tz-picker-code">{tz.code}</span>
            <span className="peak-tz-picker-cities">
              {tz.cities.join(", ")}
            </span>
          </button>
        ))}
        {entries.length === 0 && (
          <p className="peak-tz-picker-empty">No timezones found</p>
        )}
      </div>
      <button className="peak-tz-picker-close" onClick={onClose}>
        Cancel
      </button>
    </div>
  );
}
