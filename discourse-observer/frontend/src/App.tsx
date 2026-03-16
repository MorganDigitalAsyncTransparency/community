// Spec: specs/dashboard/queue-visibility.md, specs/dashboard/response-metrics.md,
//       specs/dashboard/time-period-filter.md, specs/dashboard/response-time-trends.md,
//       specs/dashboard/tag-distribution.md, specs/dashboard/slo-monitoring.md
// Tests: tests/dashboard/queue-visibility.unit.test.ts, tests/dashboard/response-metrics.unit.test.ts,
//        tests/dashboard/time-period-filter.unit.test.ts, tests/dashboard/response-time-trends.unit.test.ts,
//        tests/dashboard/tag-distribution.unit.test.ts, tests/dashboard/slo-monitoring.unit.test.ts

import { useState } from "react";
import "./App.css";
import { MOCK_DATA } from "./mock/data";
import { SummaryCards } from "./components/SummaryCards";
import { UnrepliedTable } from "./components/UnrepliedTable";
import { UntaggedTable } from "./components/UntaggedTable";
import { ResponseMetricsCards } from "./components/ResponseMetricsCards";
import { ResponseTimeTrends } from "./components/ResponseTimeTrends";
import { TagDistribution } from "./components/TagDistribution";
import { SloMonitor } from "./components/SloMonitor";
import { PeriodSelector } from "./components/PeriodSelector";
import sloConfig from "../../config/sloThresholds.json";
import {
  type ActivePeriod,
  type CustomRange,
  type PeriodPreset,
  filterByPeriod,
} from "./components/timePeriod";

type Page = "queue" | "response-metrics" | "distribution" | "slo";

function formatSyncTime(isoString: string): string {
  return new Date(isoString).toLocaleString(undefined, {
    dateStyle: "short",
    timeStyle: "short",
  });
}

export function App() {
  const [page, setPage] = useState<Page>("queue");
  const [activePeriod, setActivePeriod] = useState<ActivePeriod>({
    kind: "preset",
    preset: "allTime",
  });
  // customDraft holds the in-progress custom range inputs.
  // null means the custom tab is not visible. An object (possibly with empty strings)
  // means the custom tab is open. The filter is applied only when both dates are set.
  const [customDraft, setCustomDraft] = useState<CustomRange | null>(null);

  function handlePresetSelect(preset: PeriodPreset) {
    setActivePeriod({ kind: "preset", preset });
    setCustomDraft(null);
  }

  function handleCustomOpen() {
    // Restore the current range if already in custom mode, otherwise start empty.
    setCustomDraft(
      activePeriod.kind === "custom" ? activePeriod.range : { from: "", to: "" }
    );
  }

  function handleCustomDraftChange(from: string, to: string) {
    setCustomDraft({ from, to });
    if (from && to) {
      setActivePeriod({ kind: "custom", range: { from, to } });
    }
  }

  const filteredData = {
    ...MOCK_DATA,
    unrepliedTopics: filterByPeriod(MOCK_DATA.unrepliedTopics, activePeriod),
    untaggedTopics: filterByPeriod(MOCK_DATA.untaggedTopics, activePeriod),
    resolvedTopics: filterByPeriod(MOCK_DATA.resolvedTopics, activePeriod),
  };

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
          <button
            className={`nav-link ${page === "distribution" ? "nav-link-active" : ""}`}
            onClick={() => setPage("distribution")}
          >
            Distribution
          </button>
          <button
            className={`nav-link ${page === "slo" ? "nav-link-active" : ""}`}
            onClick={() => setPage("slo")}
          >
            SLO
          </button>
        </nav>
        <span className="app-sync-status">
          Last synced: {formatSyncTime(MOCK_DATA.lastSyncedAt)}
        </span>
      </header>

      <PeriodSelector
        period={activePeriod}
        customDraft={customDraft}
        onPresetSelect={handlePresetSelect}
        onCustomOpen={handleCustomOpen}
        onCustomDraftChange={handleCustomDraftChange}
      />

      <main className="app-content">
        {page === "queue" && (
          <>
            <SummaryCards data={filteredData} />

            <section className="app-section">
              <h2 className="app-section-title">Awaiting reply</h2>
              <UnrepliedTable topics={filteredData.unrepliedTopics} />
            </section>

            <section className="app-section">
              <h2 className="app-section-title">Untagged topics</h2>
              <UntaggedTable topics={filteredData.untaggedTopics} />
            </section>
          </>
        )}

        {page === "response-metrics" && (
          <>
            <ResponseMetricsCards topics={filteredData.resolvedTopics} />
            {/* RT-8: trends always span full history, independent of the active period filter */}
            <ResponseTimeTrends topics={MOCK_DATA.resolvedTopics} />
          </>
        )}

        {page === "distribution" && (
          // TD-23: allTopicsHistory and openTopicsHistory are unfiltered so the
          // weekly backlog trend spans all history regardless of the active period.
          <TagDistribution
            allTopics={[...filteredData.unrepliedTopics, ...filteredData.resolvedTopics]}
            resolvedTopics={filteredData.resolvedTopics}
            openTopics={filteredData.unrepliedTopics}
            allTopicsHistory={[...MOCK_DATA.unrepliedTopics, ...MOCK_DATA.resolvedTopics]}
            openTopicsHistory={MOCK_DATA.unrepliedTopics}
          />
        )}

        {page === "slo" && (
          // SL-9, SL-18: violations and compliance use the filtered topic sets
          <SloMonitor
            resolvedTopics={filteredData.resolvedTopics}
            unrepliedTopics={filteredData.unrepliedTopics}
            sloConfig={sloConfig}
          />
        )}
      </main>
    </div>
  );
}
