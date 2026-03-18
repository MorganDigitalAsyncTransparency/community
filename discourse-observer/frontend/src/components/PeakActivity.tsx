// Spec: specs/dashboard/peak-activity.md
// Tests: tests/dashboard/peak-activity.unit.test.ts

import { useState } from "react";
import type { Topic } from "../mock/data";
import { computeHeatmapData, DAY_LABELS } from "./peakActivityMetrics";
import {
  utcOffsetMinutes,
  formatOffsetHour,
  timezoneShortName,
  formatUtcOffset,
} from "./timezoneUtils";
import {
  readTimezoneCookie,
  writeTimezoneCookie,
  readConsentCookie,
  writeConsentCookie,
} from "./timezoneCookies";
import { TimezonePicker } from "./TimezonePicker";
import { CookieConsentModal } from "./CookieConsentModal";
import { HEATMAP_BASE, TEXT_ON_ACCENT } from "./themeColors";

const HOUR_LABELS = Array.from({ length: 24 }, (_, i) => String(i));
const MAX_TZ_ROWS = 3;

interface PeakActivityProps {
  topics: Topic[];
}

type ConsentState = "pending" | "accepted" | "denied";

function initTimezones(): string[] {
  if (readConsentCookie() !== "accepted") return [];
  return readTimezoneCookie().filter((tz) => tz.length > 0).slice(0, MAX_TZ_ROWS);
}

function initConsent(): ConsentState {
  return readConsentCookie() === "accepted" ? "accepted" : "pending";
}

export function PeakActivity({ topics }: PeakActivityProps) {
  const [timezones, setTimezones] = useState<string[]>(initTimezones);
  const [pickerOpen, setPickerOpen] = useState(false);
  const [consentState, setConsentState] = useState<ConsentState>(initConsent);
  const [pendingTz, setPendingTz] = useState<string | null>(null);

  const { cells, maxCount } = computeHeatmapData(topics);
  const heatmapBase = HEATMAP_BASE;
  const heatmapText = TEXT_ON_ACCENT;

  function addTimezone(tz: string) {
    if (timezones.includes(tz) || timezones.length >= MAX_TZ_ROWS) return;
    const next = [...timezones, tz];
    setTimezones(next);
    if (consentState === "accepted") {
      writeTimezoneCookie(next);
    }
  }

  function removeTimezone(tz: string) {
    const next = timezones.filter((t) => t !== tz);
    setTimezones(next);
    if (consentState === "accepted") {
      writeTimezoneCookie(next);
    }
  }

  function handlePickerSelect(tz: string) {
    setPickerOpen(false);
    if (consentState === "pending") {
      setPendingTz(tz);
    } else {
      addTimezone(tz);
    }
  }

  function handleConsentAccept() {
    setConsentState("accepted");
    writeConsentCookie();
    if (pendingTz) {
      const next = timezones.includes(pendingTz) ? timezones : [...timezones, pendingTz];
      setTimezones(next);
      writeTimezoneCookie(next);
      setPendingTz(null);
    }
  }

  function handleConsentDeny() {
    setConsentState("denied");
    if (pendingTz) {
      addTimezone(pendingTz);
      setPendingTz(null);
    }
  }

  if (topics.length === 0) {
    return (
      <section className="peak-section">
        <h2 className="peak-heading">Peak activity</h2>
        <p className="peak-empty">No data</p>
      </section>
    );
  }

  return (
    <section className="peak-section">
      <h2 className="peak-heading">Peak activity</h2>
      <button
        className="peak-add-tz-btn"
        disabled={timezones.length >= MAX_TZ_ROWS}
        title={
          timezones.length >= MAX_TZ_ROWS
            ? "Maximum of 3 timezone rows reached"
            : "Add a timezone header row"
        }
        onClick={() => setPickerOpen(true)}
      >
        + Add timezone
      </button>
      {pickerOpen && (
        <TimezonePicker
          onSelect={handlePickerSelect}
          onClose={() => setPickerOpen(false)}
          excludeTimezones={timezones}
        />
      )}
      <div className="peak-table-wrapper">
        <table className="peak-table">
          <thead>
            {[...timezones].reverse().map((tz) => {
              const offset = utcOffsetMinutes(tz);
              const name = timezoneShortName(tz);
              const offsetLabel = formatUtcOffset(offset);
              return (
                <tr key={tz} className="peak-tz-row">
                  <th className="peak-tz-label">
                    {name} ({offsetLabel})
                  </th>
                  {HOUR_LABELS.map((h) => (
                    <th key={h} className="peak-tz-hour">
                      {formatOffsetHour(Number(h), offset)}
                    </th>
                  ))}
                  <th className="peak-tz-remove-cell">
                    <button
                      className="peak-tz-remove"
                      onClick={() => removeTimezone(tz)}
                      title={`Remove ${name}`}
                    >
                      ×
                    </button>
                  </th>
                </tr>
              );
            })}
            <tr className="peak-header-utc">
              <th className="peak-header-day">UTC</th>
              {HOUR_LABELS.map((h) => (
                <th key={h} className="peak-header-hour">{h}</th>
              ))}
              {timezones.length > 0 && <th className="peak-tz-remove-cell" />}
            </tr>
          </thead>
          <tbody>
            {cells.map((row, dayIndex) => (
              <tr key={dayIndex}>
                <td className="peak-cell-day">{DAY_LABELS[dayIndex]}</td>
                {row.map((cell) => {
                  const alpha = maxCount > 0 ? cell.count / maxCount : 0;
                  const style: React.CSSProperties = {
                    backgroundColor: alpha > 0 ? `rgb(${heatmapBase} / ${alpha})` : undefined,
                    color: alpha > 0.5 ? heatmapText : undefined,
                  };
                  return (
                    <td key={cell.hour} className="peak-cell" style={style}>
                      {cell.count}
                    </td>
                  );
                })}
                {timezones.length > 0 && <td className="peak-tz-remove-cell" />}
              </tr>
            ))}
          </tbody>
        </table>
      </div>
      <div className="peak-legend">
        <span className="peak-legend-label">0</span>
        <div className="peak-legend-bar" />
        <span className="peak-legend-label">{maxCount}</span>
      </div>
      {pendingTz !== null && (
        <CookieConsentModal
          onAccept={handleConsentAccept}
          onDeny={handleConsentDeny}
        />
      )}
    </section>
  );
}
