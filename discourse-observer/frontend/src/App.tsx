// Spec: specs/dashboard/queue-visibility.md, specs/dashboard/response-metrics.md
// Tests: tests/dashboard/queue-visibility.unit.test.ts, tests/dashboard/response-metrics.unit.test.ts

import { useState } from "react";
import "./App.css";
import { MOCK_DATA } from "./mock/data";
import { SummaryCards } from "./components/SummaryCards";
import { UnrepliedTable } from "./components/UnrepliedTable";
import { UntaggedTable } from "./components/UntaggedTable";
import { ResponseMetricsCards } from "./components/ResponseMetricsCards";

type Page = "queue" | "response-metrics";

function formatSyncTime(isoString: string): string {
  return new Date(isoString).toLocaleString(undefined, {
    dateStyle: "short",
    timeStyle: "short",
  });
}

export function App() {
  const [page, setPage] = useState<Page>("queue");

  return (
    <div className="app">
      <header className="app-header">
        <h1 className="app-title">discourse-observer</h1>
        <nav className="nav">
          <button
            className={`nav-link ${page === "queue" ? "nav-link-active" : ""}`}
            onClick={() => setPage("queue")}
          >
            Queue
          </button>
          <button
            className={`nav-link ${page === "response-metrics" ? "nav-link-active" : ""}`}
            onClick={() => setPage("response-metrics")}
          >
            Response metrics
          </button>
        </nav>
        <span className="app-sync-status">
          Last synced: {formatSyncTime(MOCK_DATA.lastSyncedAt)}
        </span>
      </header>

      <main className="app-content">
        {page === "queue" && (
          <>
            <SummaryCards data={MOCK_DATA} />

            <section className="app-section">
              <h2 className="app-section-title">Awaiting reply</h2>
              <UnrepliedTable topics={MOCK_DATA.unrepliedTopics} />
            </section>

            <section className="app-section">
              <h2 className="app-section-title">Untagged topics</h2>
              <UntaggedTable topics={MOCK_DATA.untaggedTopics} />
            </section>
          </>
        )}

        {page === "response-metrics" && (
          <ResponseMetricsCards topics={MOCK_DATA.resolvedTopics} />
        )}
      </main>
    </div>
  );
}
