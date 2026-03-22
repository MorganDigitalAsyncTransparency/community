// Spec: specs/dashboard/queue-visibility.md, specs/dashboard/response-metrics.md,
//       specs/dashboard/time-period-filter.md,
//       specs/dashboard/tag-distribution.md, specs/dashboard/slo-monitoring.md,
//       specs/dashboard/tag-area-filter.md, specs/dashboard/url-state.md,
//       specs/dashboard/stalled-topics.md, specs/dashboard/peak-activity.md,
//       specs/dashboard/response-time-distribution.md
// Tests: tests/dashboard/queue-visibility.unit.test.ts, tests/dashboard/tag-area-filter.unit.test.ts,
//        tests/dashboard/url-state.unit.test.ts, backend/api/contract_test.go

import { useState, useEffect, useCallback } from "react";
import "./App.css";
import type {
  AppConfig,
  AppStatus,
  QueueSummary,
  UnrepliedTopic,
  UntaggedTopic,
  StalledTopic,
  MetricsSummary,
  VolumeBucket,
  MedianTrends,
  MetricsDistribution,
  TriageTime,
  TagFlows,
  TagVolume,
  TagResolution,
  TagBacklog,
  WeeklyBacklog,
  ViolationGroups,
  TagCompliance,
  Heatmap,
  SyncLogResponse,
} from "./api/types";
import type { FilterParams } from "./api/client";
import type { Page } from "./types";
import {
  fetchConfig,
  fetchStatus,
  fetchQueueSummary,
  fetchUnrepliedTopics,
  fetchUntaggedTopics,
  fetchStalledTopics,
  fetchMetricsSummary,
  fetchVolume,
  fetchMedianTrends,
  fetchDistribution,
  fetchTriageTime,
  fetchTagFlows,
  fetchTagVolume,
  fetchTagResolution,
  fetchTagBacklog,
  fetchBacklogTrend,
  fetchViolations,
  fetchCompliance,
  fetchHeatmap,
  fetchSyncLog,
} from "./api/endpoints";
import { SummaryCards } from "./components/SummaryCards";
import { UnrepliedTable } from "./components/UnrepliedTable";
import { UntaggedTable } from "./components/UntaggedTable";
import { ResponseMetricsCards } from "./components/ResponseMetricsCards";
import { VolumeChart } from "./components/VolumeChart";
import { MedianTrendChart } from "./components/MedianTrendChart";
import { TagDistribution } from "./components/TagDistribution";
import { SloMonitor } from "./components/SloMonitor";
import { StalledTopics } from "./components/StalledTopics";
import { PeakActivity } from "./components/PeakActivity";
import { ResponseTimeDistribution } from "./components/ResponseTimeDistribution";
import { PeriodSelector } from "./components/PeriodSelector";
import { TagSelector } from "./components/TagSelector";
import { Sidebar } from "./components/Sidebar";
import { SyncLog } from "./components/SyncLog";
import { TriageTimeCard } from "./components/TriageTimeCard";
import { TagFlowsPage } from "./components/TagFlowsPage";
import {
  type CustomRange,
  type PeriodPreset,
} from "./components/timePeriod";
import { CHART_COLOR_1, CHART_COLOR_2 } from "./components/themeColors";
import { useUrlState } from "./components/useUrlState";

// ---------------------------------------------------------------------------
// Page data types — each page has its own data shape
// ---------------------------------------------------------------------------

interface QueueData {
  summary: QueueSummary;
  unreplied: UnrepliedTopic[];
  untagged: UntaggedTopic[];
  stalled: StalledTopic[];
}

interface ResponseMetricsData {
  summary: MetricsSummary;
  volume: VolumeBucket[];
  medianTrends: MedianTrends;
  distribution: MetricsDistribution;
  triageTime: TriageTime;
}

interface DistributionData {
  volume: TagVolume[];
  resolution: TagResolution[];
  backlog: TagBacklog[];
  backlogTrend: WeeklyBacklog[];
}

interface SloData {
  violations: ViolationGroups;
  compliance: TagCompliance[];
}

// ---------------------------------------------------------------------------
// Data fetchers per page
// ---------------------------------------------------------------------------

async function loadQueueData(f: FilterParams): Promise<QueueData> {
  const [summary, unreplied, untagged, stalled] = await Promise.all([
    fetchQueueSummary(f),
    fetchUnrepliedTopics(f),
    fetchUntaggedTopics(f),
    fetchStalledTopics(f),
  ]);
  return { summary, unreplied, untagged, stalled };
}

async function loadResponseMetricsData(f: FilterParams): Promise<ResponseMetricsData> {
  const [summary, volume, medianTrends, distribution, triageTime] = await Promise.all([
    fetchMetricsSummary(f),
    fetchVolume(f),
    fetchMedianTrends(f),
    fetchDistribution(f),
    fetchTriageTime(f),
  ]);
  return { summary, volume, medianTrends, distribution, triageTime };
}

async function loadDistributionData(f: FilterParams): Promise<DistributionData> {
  const [volume, resolution, backlog, backlogTrend] = await Promise.all([
    fetchTagVolume(f),
    fetchTagResolution(f),
    fetchTagBacklog(f),
    fetchBacklogTrend(f),
  ]);
  return { volume, resolution, backlog, backlogTrend };
}

async function loadSloData(f: FilterParams): Promise<SloData> {
  const [violations, compliance] = await Promise.all([
    fetchViolations(f),
    fetchCompliance(f),
  ]);
  return { violations, compliance };
}

// ---------------------------------------------------------------------------
// App
// ---------------------------------------------------------------------------

export function App() {
  const {
    page, period: activePeriod, tag: activeTag, area: activeArea,
    setPage, setPeriod, setTag: setActiveTag, setArea: setActiveArea, clearAll,
  } = useUrlState();

  const [customDraft, setCustomDraft] = useState<CustomRange | null>(
    activePeriod.kind === "custom" ? activePeriod.range : null,
  );
  const [sidebarOpen, setSidebarOpen] = useState(false);

  // Global config and status — fetched once on mount
  const [config, setConfig] = useState<AppConfig | null>(null);
  const [status, setStatus] = useState<AppStatus | null>(null);

  // Per-page data
  const [queueData, setQueueData] = useState<QueueData | null>(null);
  const [metricsData, setMetricsData] = useState<ResponseMetricsData | null>(null);
  const [distData, setDistData] = useState<DistributionData | null>(null);
  const [sloData, setSloData] = useState<SloData | null>(null);
  const [heatmapData, setHeatmapData] = useState<Heatmap | null>(null);
  const [tagFlowsData, setTagFlowsData] = useState<TagFlows | null>(null);
  const [syncLogData, setSyncLogData] = useState<SyncLogResponse | null>(null);

  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Fetch config and status on mount
  useEffect(() => {
    Promise.all([fetchConfig(), fetchStatus()])
      .then(([cfg, st]) => { setConfig(cfg); setStatus(st); })
      .catch((err) => setError(err.message));
  }, []);

  // Fetch page data when page or filters change
  const loadPageData = useCallback(async (targetPage: Page, f: FilterParams) => {
    setLoading(true);
    setError(null);
    try {
      switch (targetPage) {
        case "queue":
          setQueueData(await loadQueueData(f));
          break;
        case "response-metrics":
          setMetricsData(await loadResponseMetricsData(f));
          break;
        case "distribution":
          setDistData(await loadDistributionData(f));
          break;
        case "slo":
          setSloData(await loadSloData(f));
          break;
        case "activity":
          setHeatmapData(await fetchHeatmap(f));
          break;
        case "tag-flows":
          setTagFlowsData(await fetchTagFlows(f));
          break;
        case "sync-log":
          setSyncLogData(await fetchSyncLog());
          break;
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load data");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    loadPageData(page, { period: activePeriod, tag: activeTag });
  }, [page, activePeriod, activeTag, loadPageData]);

  // Auto-refresh sync log every 10s while on the page
  useEffect(() => {
    if (page !== "sync-log") return;
    const id = setInterval(() => {
      fetchSyncLog().then(setSyncLogData).catch(() => {});
    }, 10_000);
    return () => clearInterval(id);
  }, [page]);

  function handlePresetSelect(preset: PeriodPreset) {
    setPeriod({ kind: "preset", preset });
    setCustomDraft(null);
  }

  function handleCustomOpen() {
    setCustomDraft(
      activePeriod.kind === "custom" ? activePeriod.range : { from: "", to: "" }
    );
  }

  function handleCustomDraftChange(from: string, to: string) {
    setCustomDraft({ from, to });
    if (from && to) {
      setPeriod({ kind: "custom", range: { from, to } });
    }
  }

  const hasActiveFilters =
    activePeriod.kind !== "preset" ||
    activePeriod.preset !== "allTime" ||
    activeTag !== null ||
    activeArea !== null;

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
        version={status?.version ?? ""}
        lastSyncedAt={status?.lastSyncedAt ?? null}
        onSyncLogClick={() => setPage("sync-log")}
      />

      {page !== "sync-log" && <div className="filter-bar">
        <PeriodSelector
          period={activePeriod}
          customDraft={customDraft}
          onPresetSelect={handlePresetSelect}
          onCustomOpen={handleCustomOpen}
          onCustomDraftChange={handleCustomDraftChange}
        />

        {config && (
          <TagSelector
            config={config}
            activeTag={activeTag}
            activeArea={activeArea}
            onTagSelect={setActiveTag}
            onAreaSelect={setActiveArea}
          />
        )}

        {hasActiveFilters && (
          <button
            className="clear-filters-btn"
            onClick={() => {
              clearAll();
              setCustomDraft(null);
            }}
          >
            Clear all filters
          </button>
        )}
      </div>}

      <main className="content">
        <div className="app-content">
          {error && <p className="app-error">{error}</p>}
          {loading && <p className="app-loading">Loading…</p>}

          {page === "queue" && queueData && (
            <>
              <SummaryCards data={queueData.summary} />
              <StalledTopics topics={queueData.stalled} />
              <section>
                <h2 className="app-section-title">Awaiting reply</h2>
                <UnrepliedTable topics={queueData.unreplied} />
              </section>
              <section>
                <h2 className="app-section-title">Untagged topics</h2>
                <UntaggedTable topics={queueData.untagged} />
              </section>
            </>
          )}

          {page === "response-metrics" && metricsData && (
            <>
              <ResponseMetricsCards data={metricsData.summary} />
              <TriageTimeCard data={metricsData.triageTime} />

              <section>
                <h2 className="app-section-title">Topic volume</h2>
                {metricsData.volume.length === 0 ? (
                  <p className="chart-empty">No data</p>
                ) : (
                  <VolumeChart data={metricsData.volume} />
                )}
              </section>

              <section>
                <h2 className="app-section-title">Median first reply</h2>
                {metricsData.medianTrends.firstReply.length === 0 ? (
                  <p className="chart-empty">No data</p>
                ) : (
                  <MedianTrendChart
                    data={metricsData.medianTrends.firstReply}
                    color={CHART_COLOR_1}
                    name="Median first reply"
                  />
                )}
              </section>

              <section>
                <h2 className="app-section-title">Median first resolution</h2>
                {metricsData.medianTrends.resolution.length === 0 ? (
                  <p className="chart-empty">No data</p>
                ) : (
                  <MedianTrendChart
                    data={metricsData.medianTrends.resolution}
                    color={CHART_COLOR_2}
                    name="Median resolution"
                  />
                )}
              </section>

              <ResponseTimeDistribution data={metricsData.distribution} />
            </>
          )}

          {page === "distribution" && distData && (
            <TagDistribution
              volumeRanking={distData.volume}
              resolutionRanking={distData.resolution}
              backlogRanking={distData.backlog}
              weeklyBacklog={distData.backlogTrend}
            />
          )}

          {page === "slo" && sloData && (
            <SloMonitor
              violations={sloData.violations}
              compliance={sloData.compliance}
            />
          )}

          {page === "activity" && heatmapData && (
            <PeakActivity data={heatmapData} />
          )}

          {page === "tag-flows" && tagFlowsData && (
            <TagFlowsPage data={tagFlowsData} />
          )}

          {page === "sync-log" && syncLogData && (
            <SyncLog data={syncLogData} />
          )}
        </div>
      </main>

    </div>
  );
}
