// Spec: specs/api/api-contract.md
// Tests: tests/dashboard/api-integration.unit.test.ts

// ---------------------------------------------------------------------------
// Shared
// ---------------------------------------------------------------------------

export interface ApiError {
  error: string;
}

// ---------------------------------------------------------------------------
// AC-12: Queue summary
// ---------------------------------------------------------------------------

export interface QueueSummary {
  unrepliedCount: number;
  untaggedCount: number;
  oldestUnrepliedAgeDays: number | null;
}

// ---------------------------------------------------------------------------
// AC-13: Unreplied topics
// ---------------------------------------------------------------------------

export interface UnrepliedTopic {
  id: number;
  title: string;
  createdAt: string;
  tags: string[];
  topicUrl: string;
}

// ---------------------------------------------------------------------------
// AC-14: Untagged topics
// ---------------------------------------------------------------------------

export interface UntaggedTopic {
  id: number;
  title: string;
  createdAt: string;
  categoryName: string;
  topicUrl: string;
}

// ---------------------------------------------------------------------------
// AC-15: Stalled topics
// ---------------------------------------------------------------------------

export interface StalledTopic {
  id: number;
  title: string;
  createdAt: string;
  tags: string[];
  topicUrl: string;
  strictestTag: string | null;
  thresholdDays: number;
  thresholdIsDefault: boolean;
  daysSinceLastActivity: number;
}

// ---------------------------------------------------------------------------
// AC-16: Metrics summary
// ---------------------------------------------------------------------------

export interface MetricsSummary {
  medianFirstReplyMs: number | null;
  medianResolutionMs: number | null;
  solvedCount: number;
  selfClosedCount: number;
  answerRatePercent: number | null;
}

// ---------------------------------------------------------------------------
// AC-17: Volume trend
// ---------------------------------------------------------------------------

export interface VolumeBucket {
  label: string;
  bucketKey: string;
  created: number;
  accepted: number;
  closed: number;
  open: number;
}

// ---------------------------------------------------------------------------
// AC-18: Median trends
// ---------------------------------------------------------------------------

export interface MedianBucket {
  label: string;
  bucketKey: string;
  medianMs: number | null;
}

export interface MedianTrends {
  firstReply: MedianBucket[];
  resolution: MedianBucket[];
}

// ---------------------------------------------------------------------------
// AC-19: Response time distribution
// ---------------------------------------------------------------------------

export interface DistributionBucket {
  label: string;
  count: number;
}

export interface MetricsDistribution {
  firstReply: DistributionBucket[];
  resolution: DistributionBucket[];
}

// ---------------------------------------------------------------------------
// AC-20: Tag volume
// ---------------------------------------------------------------------------

export interface TagVolume {
  tag: string;
  topicCount: number;
}

// ---------------------------------------------------------------------------
// AC-21: Tag resolution
// ---------------------------------------------------------------------------

export interface TagResolution {
  tag: string;
  resolvedCount: number;
  medianResolutionMs: number | null;
}

// ---------------------------------------------------------------------------
// AC-22: Tag backlog
// ---------------------------------------------------------------------------

export interface TagBacklog {
  tag: string;
  openCount: number;
}

// ---------------------------------------------------------------------------
// AC-23: Weekly backlog trend
// ---------------------------------------------------------------------------

export interface WeeklyBacklog {
  weekStart: string;
  created: number;
  resolved: number;
  stillOpen: number;
}

// ---------------------------------------------------------------------------
// AC-24: SLO violations
// ---------------------------------------------------------------------------

export interface Violation {
  topicId: number;
  topicTitle: string;
  topicUrl: string;
  tag: string;
  thresholdMs: number;
  actualMs: number;
  excessMs: number;
  thresholdIsDefault: boolean;
}

export interface ViolationGroups {
  firstReply: Violation[];
  resolution: Violation[];
  inactivity: Violation[];
}

// ---------------------------------------------------------------------------
// AC-25: SLO compliance
// ---------------------------------------------------------------------------

export interface TagCompliance {
  tag: string;
  firstReplyPercent: number | null;
  resolutionPercent: number | null;
  inactivityPercent: number | null;
  thresholdIsDefault: boolean;
}

// ---------------------------------------------------------------------------
// AC-26: Heatmap
// ---------------------------------------------------------------------------

export interface HeatmapCell {
  day: number;
  hour: number;
  count: number;
}

export interface Heatmap {
  cells: HeatmapCell[][];
  maxCount: number;
}

// ---------------------------------------------------------------------------
// AC-27: Config
// ---------------------------------------------------------------------------

export interface SloThresholds {
  firstReplyHours: number;
  resolutionHours: number;
  inactivityHours: number;
}

export interface ConfigTag {
  area: string;
  areaIsDefault: boolean;
  stalledDays: number;
  stalledDaysIsDefault: boolean;
  slo: SloThresholds;
  sloIsDefault: boolean;
  closedTag: string | null;
}

export interface ConfigArea {
  name: string;
  primaryTag: string;
}

export interface ConfigDefaults {
  stalledDays: number;
  area: string;
  slo: SloThresholds;
}

export interface AppConfig {
  areas: ConfigArea[];
  tags: Record<string, ConfigTag>;
  defaults: ConfigDefaults;
  distributionBucketCeilings: number[];
}

// ---------------------------------------------------------------------------
// AC-28: Status
// ---------------------------------------------------------------------------

export interface AppStatus {
  lastSyncedAt: string | null;
  version: string;
  syncState: string;
  lastSyncDuration: number;
  lastSyncTopics: number;
}

// ---------------------------------------------------------------------------
// Sync log
// ---------------------------------------------------------------------------

export interface SyncLogEntry {
  timestamp: string;
  mode: string;
  topics: number;
  durationSeconds: number;
  hasChanges: boolean;
  error: string;
}

export interface SyncProgress {
  mode: string;
  topics: number;
  totalTopics: number;
  elapsedSeconds: number;
  etaSeconds: number;
}

export interface SyncLogResponse {
  progress: SyncProgress | null;
  entries: SyncLogEntry[];
}
