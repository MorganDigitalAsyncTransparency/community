// Spec: specs/dashboard/tag-distribution.md
// Tests: tests/dashboard/tag-distribution.unit.test.ts

import type { Topic } from "../mock/data";
import { formatWeekLabel } from "./topicFormatting";
import {
  tagVolumeRanking,
  tagResolutionRanking,
  tagBacklogRanking,
  computeWeeklyBacklog,
} from "./tagMetrics";

interface TagDistributionProps {
  // Filtered unreplied + resolved combined — UC-9 volume ranking
  allTopics: Topic[];
  // Filtered resolved topics — UC-10 resolution time ranking
  resolvedTopics: Topic[];
  // Filtered unreplied topics — UC-11 per-tag open count
  openTopics: Topic[];
  // TD-23: unfiltered unreplied + resolved — UC-11 weekly trend spans all history
  allTopicsHistory: Topic[];
  // TD-23: unfiltered unreplied — UC-11 weekly trend spans all history
  openTopicsHistory: Topic[];
}

export function TagDistribution({
  allTopics,
  resolvedTopics,
  openTopics,
  allTopicsHistory,
  openTopicsHistory,
}: TagDistributionProps) {
  const volumeRanking = tagVolumeRanking(allTopics);
  const resolutionRanking = tagResolutionRanking(resolvedTopics);
  const backlogRanking = tagBacklogRanking(openTopics);
  const weeklyBacklog = computeWeeklyBacklog(allTopicsHistory, openTopicsHistory);

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
                  <td className="dist-td dist-td-metric">{row.medianResolution}</td>
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
