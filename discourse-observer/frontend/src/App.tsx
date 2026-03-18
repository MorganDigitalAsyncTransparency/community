// Spec: specs/dashboard/queue-visibility.md, specs/dashboard/response-metrics.md,
//       specs/dashboard/time-period-filter.md, specs/dashboard/response-time-trends.md,
//       specs/dashboard/tag-distribution.md, specs/dashboard/slo-monitoring.md,
//       specs/dashboard/tag-area-filter.md, specs/dashboard/topic-intake.md,
//       specs/dashboard/stalled-topics.md, specs/dashboard/peak-activity.md,
//       specs/dashboard/response-time-distribution.md
// Tests: tests/dashboard/queue-visibility.unit.test.ts, tests/dashboard/response-metrics.unit.test.ts,
//        tests/dashboard/time-period-filter.unit.test.ts, tests/dashboard/response-time-trends.unit.test.ts,
//        tests/dashboard/tag-distribution.unit.test.ts, tests/dashboard/slo-monitoring.unit.test.ts,
//        tests/dashboard/tag-area-filter.unit.test.ts, tests/dashboard/topic-intake.unit.test.ts,
//        tests/dashboard/stalled-topics.unit.test.ts, tests/dashboard/peak-activity.unit.test.ts,
//        tests/dashboard/response-time-distribution.unit.test.ts

import { useState } from "react";
import "./App.css";
import { MOCK_DATA, type Topic } from "./mock/data";
import { SummaryCards } from "./components/SummaryCards";
import { UnrepliedTable } from "./components/UnrepliedTable";
import { UntaggedTable } from "./components/UntaggedTable";
import { ResponseMetricsCards } from "./components/ResponseMetricsCards";
import { ResponseTimeTrends } from "./components/ResponseTimeTrends";
import { TagDistribution } from "./components/TagDistribution";
import { SloMonitor } from "./components/SloMonitor";
import { TopicIntake } from "./components/TopicIntake";
import { StalledTopics } from "./components/StalledTopics";
import { PeakActivity } from "./components/PeakActivity";
import { ResponseTimeDistribution } from "./components/ResponseTimeDistribution";
import { PeriodSelector } from "./components/PeriodSelector";
import { TagSelector } from "./components/TagSelector";
import { Sidebar } from "./components/Sidebar";
import { Footer } from "./components/Footer";
import tagConfigJson from "../../config/tagConfig.json";
import distributionConfig from "../../config/distributionBuckets.json";
import {
  type ActivePeriod,
  type CustomRange,
  type PeriodPreset,
  filterByPeriod,
} from "./components/timePeriod";
import {
  type TagConfig,
  filterByMonitoredTags,
  monitoredTags,
  tagsForArea,
  extractSloConfig,
  scopeSloConfig,
  sloDefaultTags,
  resolveAllTags,
} from "./components/tagFilter";
import { intakeGranularity, computeTimeRange } from "./components/intakeMetrics";
import type { Page } from "./types";

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
  const [activeTag, setActiveTag] = useState<string | null>(null);
  const [activeArea, setActiveArea] = useState<string | null>(null);
  const [sidebarOpen, setSidebarOpen] = useState(false);

  const typedTagConfig = tagConfigJson as TagConfig;
  const monitored = monitoredTags(typedTagConfig);
  const sloConfig = extractSloConfig(typedTagConfig);
  const defaultSloTags = sloDefaultTags(typedTagConfig);
  const resolvedTags = resolveAllTags(typedTagConfig);

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

  // TA-4, TA-17: visible tags depend on area and tag selection.
  // Specific tag → just that tag. "All" within area → area's tags. No area → all monitored.
  const visibleTags = activeTag !== null
    ? [activeTag]
    : tagsForArea(typedTagConfig, activeArea);

  const applyTagFilter = (topics: Topic[]) =>
    filterByMonitoredTags(topics, visibleTags);

  const allTopicsUnfiltered = [...MOCK_DATA.unrepliedTopics, ...MOCK_DATA.resolvedTopics];

  // Period-filtered topics (before tag filter) — used for global time range.
  const periodFiltered = {
    unrepliedTopics: filterByPeriod(MOCK_DATA.unrepliedTopics, activePeriod),
    resolvedTopics: filterByPeriod(MOCK_DATA.resolvedTopics, activePeriod),
    repliedOpenTopics: filterByPeriod(MOCK_DATA.repliedOpenTopics, activePeriod),
  };

  const filteredData = {
    ...MOCK_DATA,
    unrepliedTopics: applyTagFilter(periodFiltered.unrepliedTopics),
    // TA-6: untagged topics are empty when a tag is selected
    untaggedTopics: activeTag !== null
      ? []
      : filterByPeriod(MOCK_DATA.untaggedTopics, activePeriod),
    resolvedTopics: applyTagFilter(periodFiltered.resolvedTopics),
    // ST-8: period filter applies; ST-9: tag filter applies
    repliedOpenTopics: applyTagFilter(periodFiltered.repliedOpenTopics),
  };

  const allFilteredTopics = [...filteredData.unrepliedTopics, ...filteredData.resolvedTopics];

  const hasActiveFilters =
    activePeriod.kind !== "preset" ||
    activePeriod.preset !== "allTime" ||
    activeTag !== null ||
    activeArea !== null;

  // TI-8a: global time range from period-filtered + monitored-tag topics.
  // Uses all monitored tags (not the active tag) so the x-axis stays
  // consistent when switching between tags within the same period.
  const granularity = intakeGranularity(activePeriod);
  const allPeriodFiltered = [...periodFiltered.unrepliedTopics, ...periodFiltered.resolvedTopics];
  const intakeTimeRange = computeTimeRange(
    filterByMonitoredTags(allPeriodFiltered, monitored),
    granularity,
  );

  return (
    <div className="shell">
      <button
        className="hamburger"
        onClick={() => setSidebarOpen(true)}
        aria-label="Open navigation"
      >
        &#9776;
      </button>
      <Sidebar
        activePage={page}
        onNavigate={setPage}
        mobileOpen={sidebarOpen}
        onMobileClose={() => setSidebarOpen(false)}
      />

      <div className="filter-bar">
        <PeriodSelector
          period={activePeriod}
          customDraft={customDraft}
          onPresetSelect={handlePresetSelect}
          onCustomOpen={handleCustomOpen}
          onCustomDraftChange={handleCustomDraftChange}
        />

        <TagSelector
          config={typedTagConfig}
          activeTag={activeTag}
          activeArea={activeArea}
          onTagSelect={setActiveTag}
          onAreaSelect={setActiveArea}
        />

        {hasActiveFilters && (
          <button
            className="clear-filters-btn"
            onClick={() => {
              setActivePeriod({ kind: "preset", preset: "allTime" });
              setCustomDraft(null);
              setActiveTag(null);
              setActiveArea(null);
            }}
          >
            Clear all filters
          </button>
        )}
      </div>

      <main className="content">
        <div className="app-content">
          {page === "queue" && (
            <>
              <SummaryCards data={filteredData} />

              <section>
                <h2 className="app-section-title">Awaiting reply</h2>
                <UnrepliedTable topics={filteredData.unrepliedTopics} />
              </section>

              <section>
                <h2 className="app-section-title">Untagged topics</h2>
                <UntaggedTable topics={filteredData.untaggedTopics} />
              </section>
            </>
          )}

          {page === "response-metrics" && (
            <>
              <ResponseMetricsCards topics={filteredData.resolvedTopics} />
              {/* RT-8: trends span full history (no period filter).
                  TA-7: tag filter applies to trends — scope is a tag decision. */}
              <ResponseTimeTrends topics={applyTagFilter(MOCK_DATA.resolvedTopics)} />
              {/* RD-12: period filter applies; RD-13: tag filter applies */}
              <ResponseTimeDistribution
                topics={filteredData.resolvedTopics}
                ceilingsHours={distributionConfig.bucketCeilingsHours}
              />
            </>
          )}

          {page === "distribution" && (
            // TD-23: allTopicsHistory and openTopicsHistory skip period filter.
            // TA-7: tag filter applies to history — scope is a tag decision.
            <TagDistribution
              allTopics={allFilteredTopics}
              resolvedTopics={filteredData.resolvedTopics}
              openTopics={filteredData.unrepliedTopics}
              allTopicsHistory={applyTagFilter(allTopicsUnfiltered)}
              openTopicsHistory={applyTagFilter(MOCK_DATA.unrepliedTopics)}
            />
          )}

          {page === "slo" && (
            // SL-9, SL-18: violations and compliance use the filtered topic sets
            <SloMonitor
              resolvedTopics={filteredData.resolvedTopics}
              unrepliedTopics={filteredData.unrepliedTopics}
              sloConfig={scopeSloConfig(sloConfig, visibleTags)}
              defaultSloTags={defaultSloTags}
            />
          )}

          {page === "volume" && (
            // TI-5: period filter applies; TI-6: tag filter applies; TI-7: all topics (unreplied + resolved)
            <TopicIntake
              topics={allFilteredTopics}
              granularity={granularity}
              timeRange={intakeTimeRange}
            />
          )}

          {page === "activity" && (
            <>
              {/* ST-8: period filter applies; ST-9: tag filter applies */}
              <StalledTopics
                topics={filteredData.repliedOpenTopics}
                resolvedTags={resolvedTags}
                monitoredTags={monitored}
              />
              {/* PA-8: all topics (unreplied + resolved); PA-11: period filter; PA-12: tag filter */}
              <PeakActivity
                topics={allFilteredTopics}
              />
            </>
          )}
        </div>
      </main>

      <Footer version="v0.1.0" lastSyncedAt={MOCK_DATA.lastSyncedAt} />
    </div>
  );
}
