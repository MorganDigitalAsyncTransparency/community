// Spec: specs/dashboard/slo-monitoring.md
// Tests: tests/dashboard/slo-monitoring.unit.test.ts

import type { Topic } from "../mock/data";
import {
  findViolations,
  computeCompliance,
  type SloConfig,
  type Violation,
  type TagCompliance,
} from "./sloMetrics";
import { formatDuration } from "./topicFormatting";

interface SloMonitorProps {
  resolvedTopics: Topic[];
  unrepliedTopics: Topic[];
  sloConfig: SloConfig;
  defaultSloTags: Set<string>;
}

function DefaultIndicator({ tag, defaultTags }: { tag: string; defaultTags: Set<string> }) {
  if (!defaultTags.has(tag)) return null;
  return (
    <span className="slo-default-indicator"> (default thresholds)</span>
  );
}

function ViolationTable({
  violations,
  title,
  defaultSloTags,
}: {
  violations: Violation[];
  title: string;
  defaultSloTags: Set<string>;
}) {
  return (
    <section className="app-section">
      <h3 className="slo-section-title">{title}</h3>
      {violations.length === 0 ? (
        <p className="slo-empty">No violations</p>
      ) : (
        <table className="slo-table">
          <thead>
            <tr>
              <th className="slo-th">Title</th>
              <th className="slo-th slo-th-tag">Tag</th>
              <th className="slo-th slo-th-metric">Threshold</th>
              <th className="slo-th slo-th-metric">Actual</th>
              <th className="slo-th slo-th-metric">Excess</th>
            </tr>
          </thead>
          <tbody>
            {violations.map((v) => (
              <tr key={`${v.topicId}-${v.tag}`} className="slo-row">
                <td className="slo-td">{v.topicTitle}</td>
                <td className="slo-td slo-td-tag">
                  {v.tag}
                  <DefaultIndicator tag={v.tag} defaultTags={defaultSloTags} />
                </td>
                <td className="slo-td slo-td-metric">{formatDuration(v.thresholdMs)}</td>
                <td className="slo-td slo-td-metric">{formatDuration(v.actualMs)}</td>
                <td className="slo-td slo-td-metric">{formatDuration(v.excessMs)}</td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </section>
  );
}

function ComplianceTable({
  rows,
  defaultSloTags,
}: {
  rows: TagCompliance[];
  defaultSloTags: Set<string>;
}) {
  if (rows.length === 0 || rows.every((r) =>
    r.firstReplyPercent === null && r.resolutionPercent === null && r.inactivityPercent === null
  )) {
    return <p className="slo-empty">No data</p>;
  }

  return (
    <table className="slo-table">
      <thead>
        <tr>
          <th className="slo-th">Tag</th>
          <th className="slo-th slo-th-metric">First reply</th>
          <th className="slo-th slo-th-metric">Resolution</th>
          <th className="slo-th slo-th-metric">Inactivity</th>
        </tr>
      </thead>
      <tbody>
        {rows.map((r) => (
          <tr key={r.tag} className="slo-row">
            <td className="slo-td">
              {r.tag}
              <DefaultIndicator tag={r.tag} defaultTags={defaultSloTags} />
            </td>
            <td className="slo-td slo-td-metric">
              {r.firstReplyPercent === null ? "–" : `${r.firstReplyPercent}%`}
            </td>
            <td className="slo-td slo-td-metric">
              {r.resolutionPercent === null ? "–" : `${r.resolutionPercent}%`}
            </td>
            <td className="slo-td slo-td-metric">
              {r.inactivityPercent === null ? "–" : `${r.inactivityPercent}%`}
            </td>
          </tr>
        ))}
      </tbody>
    </table>
  );
}

export function SloMonitor({
  resolvedTopics,
  unrepliedTopics,
  sloConfig,
  defaultSloTags,
}: SloMonitorProps) {
  const configuredTags = Object.keys(sloConfig);

  // SL-25: empty config
  if (configuredTags.length === 0) {
    return <p className="slo-empty">No SLO thresholds configured</p>;
  }

  const now = Date.now();
  const violations = findViolations(resolvedTopics, unrepliedTopics, sloConfig, now);
  const compliance = computeCompliance(resolvedTopics, unrepliedTopics, sloConfig, now);

  return (
    <>
      <section className="app-section">
        <h2 className="app-section-title">Threshold violations</h2>
        <ViolationTable violations={violations.firstReply} title="First reply violations" defaultSloTags={defaultSloTags} />
        <ViolationTable violations={violations.resolution} title="Resolution violations" defaultSloTags={defaultSloTags} />
        <ViolationTable violations={violations.inactivity} title="Inactivity violations" defaultSloTags={defaultSloTags} />
      </section>

      <section className="app-section">
        <h2 className="app-section-title">SLO compliance</h2>
        <ComplianceTable rows={compliance} defaultSloTags={defaultSloTags} />
      </section>
    </>
  );
}
