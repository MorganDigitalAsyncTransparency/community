// Spec: specs/api/tag-flows.md (TF-18, TF-19, TF-20, TF-21)
// Tests: backend/api/tag-flows_contract_test.go

import type { TagFlows } from "../api/types";

interface TagFlowsPageProps {
  data: TagFlows;
}

function formatHours(hours: number | null): string {
  if (hours === null) return "–";
  if (hours < 1) return `${Math.round(hours * 60)}m`;
  if (hours < 24) return `${hours.toFixed(1)}h`;
  return `${(hours / 24).toFixed(1)}d`;
}

function formatTagSet(tags: string[]): string {
  if (tags.length === 0) return "(none)";
  return tags.join(", ");
}

function formatRatio(part: number, total: number): string {
  if (total === 0) return "0 / 0";
  const pct = Math.round((part / total) * 100);
  return `${part} / ${total} (${pct}%)`;
}

export function TagFlowsPage({ data }: TagFlowsPageProps) {
  const { summary, transitions, tagPairs } = data;

  return (
    <>
      <div className="response-cards">
        <div className="response-card">
          <span className="response-card-value">
            {formatRatio(summary.topicsWithTagChanges, summary.totalTopics)}
          </span>
          <span className="response-card-label">Topics with tag changes</span>
        </div>
        <div className="response-card">
          <span className="response-card-value">
            {summary.medianChangesPerTopic !== null
              ? summary.medianChangesPerTopic.toFixed(1)
              : "–"}
          </span>
          <span className="response-card-label">Median changes per topic</span>
        </div>
        <div className="response-card">
          <span className="response-card-value">
            {summary.mostCommonFirstTag ?? "–"}
          </span>
          <span className="response-card-label">Most common first tag</span>
        </div>
        <div className="response-card">
          <span className="response-card-value">
            {summary.mostUnstableTag ?? "–"}
          </span>
          <span className="response-card-label">Most unstable tag</span>
        </div>
      </div>

      <section>
        <h2 className="app-section-title">Tag transitions</h2>
        {transitions.length === 0 ? (
          <p className="chart-empty">No tag transitions recorded</p>
        ) : (
          <div className="table-scroll">
            <table className="data-table">
              <thead>
                <tr>
                  <th>From</th>
                  <th>To</th>
                  <th>Count</th>
                  <th>Median duration</th>
                </tr>
              </thead>
              <tbody>
                {transitions.map((tr, i) => (
                  <tr key={i}>
                    <td>{formatTagSet(tr.from)}</td>
                    <td>{formatTagSet(tr.to)}</td>
                    <td>{tr.count}</td>
                    <td>{formatHours(tr.medianDurationHours)}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </section>

      <section>
        <h2 className="app-section-title">Tag pairs</h2>
        {tagPairs.length === 0 ? (
          <p className="chart-empty">No tag pairs recorded</p>
        ) : (
          <div className="table-scroll">
            <table className="data-table">
              <thead>
                <tr>
                  <th>Tags</th>
                  <th>Count</th>
                </tr>
              </thead>
              <tbody>
                {tagPairs.map((pair, i) => (
                  <tr key={i}>
                    <td>{pair.tags.join(" + ")}</td>
                    <td>{pair.count}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </section>
    </>
  );
}
