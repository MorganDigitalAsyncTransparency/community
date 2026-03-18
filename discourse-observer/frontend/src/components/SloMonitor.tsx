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
import { formatDuration, topicUrl } from "./topicFormatting";
import { useTableSort, type SortDirection } from "./useTableSort";

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

type ViolationSortColumn = "topic" | "tag" | "threshold" | "actual" | "excess";

const VIOLATION_DEFAULT_DIRS: Record<ViolationSortColumn, SortDirection> = {
  topic: "asc",
  tag: "asc",
  threshold: "desc",
  actual: "desc",
  excess: "desc",
};

function sortViolations(
  violations: Violation[],
  column: ViolationSortColumn,
  direction: SortDirection,
): Violation[] {
  const sorted = [...violations];
  const dir = direction === "asc" ? 1 : -1;

  if (column === "topic") {
    sorted.sort((a, b) => dir * a.topicTitle.localeCompare(b.topicTitle));
  } else if (column === "tag") {
    sorted.sort((a, b) => dir * a.tag.localeCompare(b.tag));
  } else if (column === "threshold") {
    sorted.sort((a, b) => dir * (a.thresholdMs - b.thresholdMs));
  } else if (column === "actual") {
    sorted.sort((a, b) => dir * (a.actualMs - b.actualMs));
  } else {
    sorted.sort((a, b) => dir * (a.excessMs - b.excessMs));
  }

  return sorted;
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
  const { sortColumn, sortDirection, handleSort, arrow } =
    useTableSort<ViolationSortColumn>("excess", VIOLATION_DEFAULT_DIRS);

  const sorted = sortViolations(violations, sortColumn, sortDirection);

  return (
    <section className="app-section">
      <h3 className="slo-section-title">{title}</h3>
      {sorted.length === 0 ? (
        <p className="slo-empty">No violations</p>
      ) : (
        <table className="slo-table">
          <thead>
            <tr>
              <th
                className="slo-th sortable-header"
                onClick={() => handleSort("topic")}
              >
                Topic{arrow("topic")}
              </th>
              <th
                className="slo-th slo-th-tag sortable-header"
                onClick={() => handleSort("tag")}
              >
                Tag{arrow("tag")}
              </th>
              <th
                className="slo-th slo-th-metric sortable-header"
                onClick={() => handleSort("threshold")}
              >
                Threshold{arrow("threshold")}
              </th>
              <th
                className="slo-th slo-th-metric sortable-header"
                onClick={() => handleSort("actual")}
              >
                Actual{arrow("actual")}
              </th>
              <th
                className="slo-th slo-th-metric sortable-header"
                onClick={() => handleSort("excess")}
              >
                Excess{arrow("excess")}
              </th>
            </tr>
          </thead>
          <tbody>
            {sorted.map((v) => (
              <tr key={`${v.topicId}-${v.tag}`} className="slo-row">
                <td className="slo-td">
                  <a href={topicUrl(v.topicId)} className="topic-link" target="_blank" rel="noreferrer">
                    {v.topicTitle}
                  </a>
                </td>
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

type ComplianceSortColumn = "tag" | "firstReply" | "resolution" | "inactivity";

const COMPLIANCE_DEFAULT_DIRS: Record<ComplianceSortColumn, SortDirection> = {
  tag: "asc",
  firstReply: "asc",
  resolution: "asc",
  inactivity: "asc",
};

function sortCompliance(
  rows: TagCompliance[],
  column: ComplianceSortColumn,
  direction: SortDirection,
): TagCompliance[] {
  const sorted = [...rows];
  const dir = direction === "asc" ? 1 : -1;

  if (column === "tag") {
    sorted.sort((a, b) => dir * a.tag.localeCompare(b.tag));
  } else {
    const key = column === "firstReply" ? "firstReplyPercent"
      : column === "resolution" ? "resolutionPercent"
      : "inactivityPercent";
    sorted.sort((a, b) => {
      const aVal = a[key] ?? -1;
      const bVal = b[key] ?? -1;
      return dir * (aVal - bVal);
    });
  }

  return sorted;
}

function ComplianceTable({
  rows,
  defaultSloTags,
}: {
  rows: TagCompliance[];
  defaultSloTags: Set<string>;
}) {
  const { sortColumn, sortDirection, handleSort, arrow } =
    useTableSort<ComplianceSortColumn>("tag", COMPLIANCE_DEFAULT_DIRS);

  if (rows.length === 0 || rows.every((r) =>
    r.firstReplyPercent === null && r.resolutionPercent === null && r.inactivityPercent === null
  )) {
    return <p className="slo-empty">No data</p>;
  }

  const sorted = sortCompliance(rows, sortColumn, sortDirection);

  return (
    <table className="slo-table">
      <thead>
        <tr>
          <th
            className="slo-th sortable-header"
            onClick={() => handleSort("tag")}
          >
            Tag{arrow("tag")}
          </th>
          <th
            className="slo-th slo-th-metric sortable-header"
            onClick={() => handleSort("firstReply")}
          >
            First reply{arrow("firstReply")}
          </th>
          <th
            className="slo-th slo-th-metric sortable-header"
            onClick={() => handleSort("resolution")}
          >
            Resolution{arrow("resolution")}
          </th>
          <th
            className="slo-th slo-th-metric sortable-header"
            onClick={() => handleSort("inactivity")}
          >
            Inactivity{arrow("inactivity")}
          </th>
        </tr>
      </thead>
      <tbody>
        {sorted.map((r) => (
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
