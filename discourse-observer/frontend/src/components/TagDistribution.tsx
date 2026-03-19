// Spec: specs/dashboard/tag-distribution.md
// Tests: tests/dashboard/tag-distribution.unit.test.ts

import type { TagVolume, TagResolution, TagBacklog, WeeklyBacklog } from "../api/types";
import { formatDuration, formatWeekLabel } from "./topicFormatting";

interface TagDistributionProps {
  volumeRanking: TagVolume[];
  resolutionRanking: TagResolution[];
  backlogRanking: TagBacklog[];
  weeklyBacklog: WeeklyBacklog[];
}

export function TagDistribution({
  volumeRanking,
  resolutionRanking,
  backlogRanking,
  weeklyBacklog,
}: TagDistributionProps) {
  return (
    <>
      <section className="app-section">
        <h2 className="app-section-title">Topics by tag</h2>
        {volumeRanking.length === 0 ? (
          <p className="dist-empty">No data</p>
        ) : (
          <table className="dist-table">
            <thead>
              <tr>
                <th className="dist-th dist-th-tag">Tag</th>
                <th className="dist-th dist-th-count">Topics</th>
              </tr>
            </thead>
            <tbody>
              {volumeRanking.map((row) => (
                <tr key={row.tag} className="dist-row">
                  <td className="dist-td dist-td-tag">{row.tag}</td>
                  <td className="dist-td dist-td-count">{row.topicCount}</td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </section>

      <section className="app-section">
        <h2 className="app-section-title">Resolution time by tag</h2>
        {resolutionRanking.length === 0 ? (
          <p className="dist-empty">No data</p>
        ) : (
          <table className="dist-table">
            <thead>
              <tr>
                <th className="dist-th dist-th-tag">Tag</th>
                <th className="dist-th dist-th-count">Resolved</th>
                <th className="dist-th dist-th-metric">Median resolution</th>
              </tr>
            </thead>
            <tbody>
              {resolutionRanking.map((row) => (
                <tr key={row.tag} className="dist-row">
                  <td className="dist-td dist-td-tag">{row.tag}</td>
                  <td className="dist-td dist-td-count">{row.resolvedCount}</td>
                  <td className="dist-td dist-td-metric">
                    {row.medianResolutionMs === null ? "–" : formatDuration(row.medianResolutionMs)}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </section>

      <section className="app-section">
        <h2 className="app-section-title">Open backlogs by tag</h2>
        {backlogRanking.length === 0 ? (
          <p className="dist-empty">No data</p>
        ) : (
          <table className="dist-table">
            <thead>
              <tr>
                <th className="dist-th dist-th-tag">Tag</th>
                <th className="dist-th dist-th-count">Open topics</th>
              </tr>
            </thead>
            <tbody>
              {backlogRanking.map((row) => (
                <tr key={row.tag} className="dist-row">
                  <td className="dist-td dist-td-tag">{row.tag}</td>
                  <td className="dist-td dist-td-count">{row.openCount}</td>
                </tr>
              ))}
            </tbody>
          </table>
        )}

        <div className="dist-backlog-trend">
          <h3 className="dist-backlog-trend-title">Weekly backlog</h3>
          {weeklyBacklog.length === 0 ? (
            <p className="dist-empty">No data</p>
          ) : (
            <table className="dist-table">
              <thead>
                <tr>
                  <th className="dist-th dist-th-week">Week</th>
                  <th className="dist-th dist-th-count">Created</th>
                  <th className="dist-th dist-th-count">Resolved</th>
                  <th className="dist-th dist-th-count">Still open</th>
                </tr>
              </thead>
              <tbody>
                {weeklyBacklog.map((row) => (
                  <tr key={row.weekStart} className="dist-row">
                    <td className="dist-td dist-td-week">{formatWeekLabel(row.weekStart)}</td>
                    <td className="dist-td dist-td-count">{row.created}</td>
                    <td className="dist-td dist-td-count">{row.resolved}</td>
                    <td className="dist-td dist-td-count">{row.stillOpen}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </div>
      </section>
    </>
  );
}
