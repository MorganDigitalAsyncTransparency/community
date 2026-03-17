// Spec: specs/dashboard/peak-activity.md
// Tests: tests/dashboard/peak-activity.unit.test.ts

import type { Topic } from "../mock/data";
import { computeHeatmapData, DAY_LABELS } from "./peakActivityMetrics";

const HOUR_LABELS = Array.from({ length: 24 }, (_, i) => String(i));

interface PeakActivityProps {
  topics: Topic[];
}

export function PeakActivity({ topics }: PeakActivityProps) {
  const { cells, maxCount } = computeHeatmapData(topics);

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
      <div className="peak-table-wrapper">
        <table className="peak-table">
          <thead>
            <tr>
              <th className="peak-header-day" />
              {HOUR_LABELS.map((h) => (
                <th key={h} className="peak-header-hour">{h}</th>
              ))}
            </tr>
          </thead>
          <tbody>
            {cells.map((row, dayIndex) => (
              <tr key={dayIndex} className="peak-row">
                <td className="peak-cell-day">{DAY_LABELS[dayIndex]}</td>
                {row.map((cell) => {
                  const alpha = maxCount > 0 ? cell.count / maxCount : 0;
                  const style: React.CSSProperties = {
                    backgroundColor: alpha > 0 ? `rgba(59, 130, 246, ${alpha})` : undefined,
                    color: alpha > 0.5 ? "#fff" : undefined,
                  };
                  return (
                    <td key={cell.hour} className="peak-cell" style={style}>
                      {cell.count}
                    </td>
                  );
                })}
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
    </section>
  );
}
