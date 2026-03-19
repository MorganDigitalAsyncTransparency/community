// Spec: specs/dashboard/slo-monitoring.md
// Tests: tests/dashboard/slo-monitoring.unit.test.ts

import type { ViolationGroups, Violation, TagCompliance } from "../api/types";
import { formatDuration } from "./topicFormatting";
import { useTableSort, type SortDirection } from "./useTableSort";

interface SloMonitorProps {
  violations: ViolationGroups;
  compliance: TagCompliance[];
}

function DefaultIndicator({ isDefault }: { isDefault: boolean }) {
  if (!isDefault) return null;
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
}: {
  violations: Violation[];
  title: string;
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
                  <a href={v.topicUrl} className="topic-link" target="_blank" rel="noreferrer">
                    {v.topicTitle}
                  </a>
                </td>
                <td className="slo-td slo-td-tag">
                  {v.tag}
                  <DefaultIndicator isDefault={v.thresholdIsDefault} />
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

function ComplianceTable({ rows }: { rows: TagCompliance[] }) {
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
              <DefaultIndicator isDefault={r.thresholdIsDefault} />
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

export function SloMonitor({ violations, compliance }: SloMonitorProps) {
  const hasAnyViolation =
    violations.firstReply.length > 0 ||
    violations.resolution.length > 0 ||
    violations.inactivity.length > 0;
  const hasAnyCompliance = compliance.length > 0;

  if (!hasAnyViolation && !hasAnyCompliance) {
    return <p className="slo-empty">No SLO data</p>;
  }

  return (
    <>
      <section className="app-section">
        <h2 className="app-section-title">Threshold violations</h2>
        <ViolationTable violations={violations.firstReply} title="First reply violations" />
        <ViolationTable violations={violations.resolution} title="Resolution violations" />
        <ViolationTable violations={violations.inactivity} title="Inactivity violations" />
      </section>

      <section className="app-section">
        <h2 className="app-section-title">SLO compliance</h2>
        <ComplianceTable rows={compliance} />
      </section>
    </>
  );
}
