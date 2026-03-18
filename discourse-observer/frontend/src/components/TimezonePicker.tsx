// Spec: specs/dashboard/peak-activity.md (PA-21, PA-22)
// Tests: tests/dashboard/timezone-utils.unit.test.ts

import { useState } from "react";
import {
  TIMEZONE_LIST,
  utcOffsetMinutes,
  formatUtcOffset,
} from "./timezoneUtils";

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

  const filtered = TIMEZONE_LIST.filter(
    (tz) =>
      !excluded.has(tz.id) &&
      (tz.id.toLowerCase().includes(query) ||
        tz.region.toLowerCase().includes(query)),
  );

  const grouped = new Map<string, typeof filtered>();
  for (const tz of filtered) {
    const list = grouped.get(tz.region) ?? [];
    list.push(tz);
    grouped.set(tz.region, list);
  }

  function cityName(id: string): string {
    const parts = id.split("/");
    return (parts[parts.length - 1] ?? id).replace(/_/g, " ");
  }

  return (
    <div className="peak-tz-picker">
      <input
        className="peak-tz-picker-search"
        type="text"
        placeholder="Search timezones..."
        value={search}
        onChange={(e) => setSearch(e.target.value)}
        autoFocus
      />
      <div className="peak-tz-picker-list">
        {[...grouped.entries()].map(([region, entries]) => (
          <div key={region} className="peak-tz-picker-group">
            <div className="peak-tz-picker-region">{region}</div>
            {entries.map((tz) => {
              const offset = utcOffsetMinutes(tz.id);
              return (
                <button
                  key={tz.id}
                  className="peak-tz-picker-item"
                  onClick={() => {
                    onSelect(tz.id);
                    onClose();
                  }}
                >
                  <span className="peak-tz-picker-city">{cityName(tz.id)}</span>
                  <span className="peak-tz-picker-offset">
                    {formatUtcOffset(offset)}
                  </span>
                </button>
              );
            })}
          </div>
        ))}
        {filtered.length === 0 && (
          <p className="peak-tz-picker-empty">No timezones found</p>
        )}
      </div>
      <button className="peak-tz-picker-close" onClick={onClose}>
        Cancel
      </button>
    </div>
  );
}
