// Spec: specs/api/api-contract.md (AC-12 through AC-28)
// Tests: tests/dashboard/api-integration.unit.test.ts

import { apiFetch, type FilterParams } from "./client";
import type {
  QueueSummary,
  UnrepliedTopic,
  UntaggedTopic,
  StalledTopic,
  MetricsSummary,
  VolumeBucket,
  MedianTrends,
  MetricsDistribution,
  TagVolume,
  TagResolution,
  TagBacklog,
  WeeklyBacklog,
  ViolationGroups,
  TagCompliance,
  Heatmap,
  AppConfig,
  AppStatus,
} from "./types";

// Queue
export const fetchQueueSummary = (f: FilterParams) =>
  apiFetch<QueueSummary>("/queue/summary", f);

export const fetchUnrepliedTopics = (f: FilterParams) =>
  apiFetch<UnrepliedTopic[]>("/queue/unreplied", f);

export const fetchUntaggedTopics = (f: FilterParams) =>
  apiFetch<UntaggedTopic[]>("/queue/untagged", f);

export const fetchStalledTopics = (f: FilterParams) =>
  apiFetch<StalledTopic[]>("/queue/stalled", f);

// Metrics
export const fetchMetricsSummary = (f: FilterParams) =>
  apiFetch<MetricsSummary>("/metrics/summary", f);

export const fetchVolume = (f: FilterParams) =>
  apiFetch<VolumeBucket[]>("/metrics/volume", f);

export const fetchMedianTrends = (f: FilterParams) =>
  apiFetch<MedianTrends>("/metrics/median-trends", f);

export const fetchDistribution = (f: FilterParams) =>
  apiFetch<MetricsDistribution>("/metrics/distribution", f);

// Distribution
export const fetchTagVolume = (f: FilterParams) =>
  apiFetch<TagVolume[]>("/distribution/volume", f);

export const fetchTagResolution = (f: FilterParams) =>
  apiFetch<TagResolution[]>("/distribution/resolution", f);

export const fetchTagBacklog = (f: FilterParams) =>
  apiFetch<TagBacklog[]>("/distribution/backlog", f);

export const fetchBacklogTrend = (f: FilterParams) =>
  apiFetch<WeeklyBacklog[]>("/distribution/backlog-trend", f);

// SLO
export const fetchViolations = (f: FilterParams) =>
  apiFetch<ViolationGroups>("/slo/violations", f);

export const fetchCompliance = (f: FilterParams) =>
  apiFetch<TagCompliance[]>("/slo/compliance", f);

// Activity
export const fetchHeatmap = (f: FilterParams) =>
  apiFetch<Heatmap>("/activity/heatmap", f);

// Config & Status (no filters)
export const fetchConfig = () =>
  apiFetch<AppConfig>("/config");

export const fetchStatus = () =>
  apiFetch<AppStatus>("/status");
