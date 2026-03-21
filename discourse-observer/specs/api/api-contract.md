# API Contract

This specification defines the HTTP API that the backend exposes and the frontend (and future consumers such as MCP servers) consume. It is the contract between data producers and data consumers.

The API responsibility model is decided in [ADR 0012](../../docs/decisions/0012-api-responsibility-model.md). This spec implements the chosen model.

---

## Scope

The API serves all data needed by the dashboard's five pages (Queue, Response Metrics, Distribution, SLO, Activity) and the global filter bar. It replaces the mock data currently used by the frontend.

Every use case (UC-1 through UC-20, UC-24) must be satisfiable through the endpoints defined here. UC-24 (URL state) is frontend-only and requires no API support beyond stable query parameter semantics.

---

## Requirements

Requirements use the prefix **AC** (API Contract).

### General

**AC-1.** The API is served under the path prefix `/api/v1/`.

**AC-2.** All endpoints accept `GET` requests only. The API is read-only.

**AC-3.** All timestamps in responses are ISO 8601 UTC strings (e.g., `2026-03-18T14:32:00Z`).

**AC-4.** All duration values in responses are integers representing milliseconds.

**AC-5.** When a filter produces an empty dataset, the endpoint returns the normal response structure with zero counts, empty arrays, or null aggregates — not a 404.

**AC-6.** The API returns `application/json` responses with UTF-8 encoding.

**AC-7.** Error responses use a consistent structure: `{ "error": "<message>" }` with an appropriate HTTP status code (400 for invalid parameters, 500 for server errors).

### Filtering parameters

These query parameters are shared across endpoints that support filtering.

**AC-8.** `period` — one of `7d`, `30d`, `1y`, `all`. Scopes data to topics created within the specified rolling window. Default: `all`. Applies to `createdAt`.

**AC-9.** `from` and `to` — ISO 8601 date strings (`YYYY-MM-DD`). Define a custom date range. Both must be provided together; providing only one produces a 400 error. When both are present, `period` is ignored. Both are inclusive (UTC day boundaries: `from` at 00:00:00Z, `to` at 23:59:59.999Z). If `from` is after `to`, the endpoint returns a 400 error.

**AC-10.** `tag` — a single tag name. Scopes data to topics carrying this tag. When absent, all monitored topics are included.

**AC-11.** Invalid filter values (unknown period, malformed dates, unknown tag) produce a 400 response with a descriptive error message.

### Endpoint: Queue summary

**AC-12.** `GET /api/v1/queue/summary` — Returns summary counts for the queue page.

Response fields:

- `unrepliedCount` (integer): number of unreplied monitored topics
- `untaggedCount` (integer): number of topics with no tags at all
- `oldestUnrepliedAgeDays` (integer or null): days since the oldest unreplied topic was created; null if no unreplied topics exist

Supports: `period`, `from`/`to`, `tag`.

Serves: UC-1, UC-2, UC-3.

### Endpoint: Unreplied topics

**AC-13.** `GET /api/v1/queue/unreplied` — Returns the list of unreplied monitored topics.

Response: array of objects, each with:

- `id` (integer)
- `title` (string)
- `createdAt` (timestamp)
- `tags` (string array)
- `topicUrl` (string): full URL to the Discourse topic

Default sort: oldest first (ascending `createdAt`).

Supports: `period`, `from`/`to`, `tag`.

Serves: UC-1, UC-2.

### Endpoint: Untagged topics

**AC-14.** `GET /api/v1/queue/untagged` — Returns topics with no tags.

Response: array of objects, each with:

- `id` (integer)
- `title` (string)
- `createdAt` (timestamp)
- `categoryName` (string)
- `topicUrl` (string)

Default sort: oldest first (ascending `createdAt`).

Supports: `period`, `from`/`to`. Tag filter does not apply (untagged topics have no tags).

Serves: UC-3.

### Endpoint: Stalled topics

**AC-15.** `GET /api/v1/queue/stalled` — Returns topics that have at least one reply but have gone quiet beyond their configured stalled threshold, without being resolved.

Response: array of objects, each with:

- `id` (integer)
- `title` (string)
- `createdAt` (timestamp)
- `tags` (string array)
- `topicUrl` (string)
- `strictestTag` (string or null): the monitored tag with the strictest stalled-days threshold; null if no monitored tags
- `thresholdDays` (integer): the stalled-days threshold applied
- `thresholdIsDefault` (boolean): whether the threshold comes from config defaults
- `daysSinceLastActivity` (integer): days since the topic's last activity

Default sort: most inactive first (descending `daysSinceLastActivity`).

Supports: `period`, `from`/`to`, `tag`.

Serves: UC-18.

### Endpoint: Response metrics summary

**AC-16.** `GET /api/v1/metrics/summary` — Returns aggregate response metrics for the selected period.

Response fields:

- `medianFirstReplyMs` (integer or null): median time to first reply in milliseconds
- `medianResolutionMs` (integer or null): median time to resolution in milliseconds
- `solvedCount` (integer): topics with outcome "solved"
- `selfClosedCount` (integer): topics with outcome "self-closed"
- `answerRatePercent` (integer or null): percentage of resolved topics that were solved; null if no resolved topics

Supports: `period`, `from`/`to`, `tag`.

Serves: UC-4, UC-5, UC-6, UC-7.

### Endpoint: Volume trend

**AC-17.** `GET /api/v1/metrics/volume` — Returns topic volume bucketed over time.

Response: array of objects (one per time bucket, ordered chronologically), each with:

- `label` (string): human-readable bucket label (e.g., "Mar 10" or "Week of Mar 10")
- `bucketKey` (string): machine-sortable key (ISO date of bucket start)
- `created` (integer): topics created in this bucket
- `accepted` (integer): topics with solved outcome created in this bucket
- `closed` (integer): topics with self-closed outcome created in this bucket
- `open` (integer): unreplied + replied-open topics created in this bucket

Granularity is determined by the period: daily for periods under 90 days, weekly for 90 days and above.

Supports: `period`, `from`/`to`, `tag`.

Serves: UC-17.

### Endpoint: Median trends

**AC-18.** `GET /api/v1/metrics/median-trends` — Returns median first-reply and resolution times bucketed over time.

Response fields:

- `firstReply` (array of bucket objects)
- `resolution` (array of bucket objects)

Each bucket object:

- `label` (string)
- `bucketKey` (string)
- `medianMs` (integer or null): median duration in milliseconds; null if no data in bucket

Granularity follows the same rule as AC-17.

Supports: `period`, `from`/`to`, `tag`.

Serves: UC-8.

### Endpoint: Response time distribution

**AC-19.** `GET /api/v1/metrics/distribution` — Returns response time distribution histograms.

Response fields:

- `firstReply` (array of bucket objects)
- `resolution` (array of bucket objects)

Each bucket object:

- `label` (string): human-readable range label (e.g., "< 1h", "1–4h", "> 7d")
- `count` (integer): number of topics in this bucket

Bucket boundaries are defined by `config/distributionBuckets.json`.

Supports: `period`, `from`/`to`, `tag`.

Serves: UC-20.

### Endpoint: Tag distribution — volume

**AC-20.** `GET /api/v1/distribution/volume` — Returns tags ranked by topic count.

Response: array of objects (sorted by count descending, then tag ascending), each with:

- `tag` (string)
- `topicCount` (integer)

Supports: `period`, `from`/`to`, `tag`.

Serves: UC-9.

### Endpoint: Tag distribution — resolution time

**AC-21.** `GET /api/v1/distribution/resolution` — Returns tags ranked by median resolution time.

Response: array of objects (sorted by median descending, tags with no data last), each with:

- `tag` (string)
- `resolvedCount` (integer)
- `medianResolutionMs` (integer or null)

Supports: `period`, `from`/`to`, `tag`.

Serves: UC-10.

### Endpoint: Tag distribution — backlog

**AC-22.** `GET /api/v1/distribution/backlog` — Returns tags ranked by unreplied topic count.

Response: array of objects (sorted by count descending, then tag ascending), each with:

- `tag` (string)
- `openCount` (integer)

Supports: `period`, `from`/`to`, `tag`.

Serves: UC-11.

### Endpoint: Weekly backlog trend

**AC-23.** `GET /api/v1/distribution/backlog-trend` — Returns weekly backlog trend data. This endpoint always uses full history regardless of period filter, but tag filter applies.

Response: array of objects (one per calendar week, sorted newest first), each with:

- `weekStart` (string): Monday of the week, `YYYY-MM-DD`
- `created` (integer): topics created that week
- `resolved` (integer): topics resolved that week
- `stillOpen` (integer): topics from that week still open

Supports: `tag`. Period filter does not apply (full history always shown).

Serves: UC-11.

### Endpoint: SLO violations

**AC-24.** `GET /api/v1/slo/violations` — Returns topics that exceed SLO thresholds, grouped by violation type.

Response fields:

- `firstReply` (array of violation objects)
- `resolution` (array of violation objects)
- `inactivity` (array of violation objects)

Each violation object:

- `topicId` (integer)
- `topicTitle` (string)
- `topicUrl` (string)
- `tag` (string): the tag with the strictest threshold
- `thresholdMs` (integer): configured threshold in milliseconds
- `actualMs` (integer): actual duration in milliseconds
- `excessMs` (integer): amount exceeding threshold in milliseconds
- `thresholdIsDefault` (boolean): whether the threshold comes from config defaults

Each array is sorted by `excessMs` descending.

Supports: `period`, `from`/`to`, `tag`.

Serves: UC-13.

### Endpoint: SLO compliance

**AC-25.** `GET /api/v1/slo/compliance` — Returns per-tag SLO compliance rates.

Response: array of objects (sorted by tag ascending), each with:

- `tag` (string)
- `firstReplyPercent` (integer or null): percentage of topics meeting first-reply SLO; null if no eligible topics
- `resolutionPercent` (integer or null)
- `inactivityPercent` (integer or null)
- `thresholdIsDefault` (boolean)

Supports: `period`, `from`/`to`, `tag`.

Serves: UC-14.

### Endpoint: Peak activity heatmap

**AC-26.** `GET /api/v1/activity/heatmap` — Returns topic creation counts by day of week and hour of day (UTC).

Response fields:

- `cells` (7×24 nested array): `cells[day][hour]` where day 0 = Monday, each cell is `{ "day": N, "hour": N, "count": N }`
- `maxCount` (integer): highest count across all cells

Supports: `period`, `from`/`to`, `tag`.

Serves: UC-19.

### Endpoint: Configuration

**AC-27.** `GET /api/v1/config` — Returns the resolved tag configuration needed for UI rendering (area navigation, tag labels, SLO threshold display).

Response fields:

- `areas` (array): `[{ "name": string, "primaryTag": string }]`
- `tags` (object): keyed by tag name, each value:
  - `area` (string)
  - `areaIsDefault` (boolean)
  - `stalledDays` (integer)
  - `stalledDaysIsDefault` (boolean)
  - `slo` (`{ "firstReplyHours": N, "resolutionHours": N, "inactivityHours": N }`)
  - `sloIsDefault` (boolean)
  - `closedTag` (string or null)
- `defaults` (object): `{ "stalledDays": N, "area": string, "slo": { ... } }`
- `distributionBucketCeilings` (integer array): hour values for histogram bucket boundaries

This endpoint is not filtered. It returns the full configuration.

Serves: UC-15, UC-16 (area navigation, tag display).

### Endpoint: Sync status

**AC-28.** `GET /api/v1/status` — Returns system status information.

Response fields:

- `lastSyncedAt` (timestamp or null): when the last successful sync completed; null if no sync has run
- `version` (string): application version
- `syncState` (string): `"idle"`, `"running"`, or `"disabled"` (when no Discourse credentials are configured)
- `lastSyncDuration` (float): duration of the last completed sync in seconds; 0 if no sync has run
- `lastSyncTopics` (integer): number of topics upserted in the last completed sync; 0 if no sync has run

This endpoint is not filtered.

### Endpoint: Sync log

**AC-33.** `GET /api/v1/sync-log` — Returns sync log and live progress.

Response object:

- `progress` (object or null): present when a sync is running
  - `mode` (string): `"initial"` or `"delta"`
  - `topics` (integer): topics upserted so far
  - `totalTopics` (integer): estimated total from `/about.json` (0 if unknown)
  - `elapsedSeconds` (float): time since sync started
  - `etaSeconds` (float): estimated seconds remaining (0 if unknown)
- `entries` (array): most recent completed syncs, newest first, up to 20 per type
  - `timestamp` (string): ISO 8601 UTC timestamp of sync completion
  - `mode` (string): `"initial"`, `"delta"`, or `"detail"`
  - `topics` (integer): number of topics upserted
  - `durationSeconds` (float): sync duration in seconds
  - `hasChanges` (boolean): true if new or updated data was found

The log is persisted in SQLite and survives restarts. Each sync type retains its own 20 most recent entries, so infrequent events (like initial sync) are never displaced by frequent ones (like delta sync). No-change entries are deduplicated: only the most recent per type is kept. Returns empty entries array and null progress when sync is disabled. This endpoint is not filtered.

### Cross-cutting

**AC-29.** All list endpoints (AC-13, AC-14, AC-15, AC-24) include a `topicUrl` field that is a full, clickable URL to the topic on the Discourse forum.

**AC-30.** The `tag` filter (AC-10) applies to all endpoints except AC-14 (untagged topics have no tags), AC-27 (configuration), and AC-28 (status).

**AC-31.** The `period`/`from`/`to` filters (AC-8, AC-9) apply to all endpoints except AC-23 (full history), AC-27 (configuration), and AC-28 (status).

**AC-32.** "Open" has context-dependent meaning: in the volume endpoint (AC-17), "open" means topics with no outcome (unreplied + replied-open). In the backlog endpoint (AC-22), it means unreplied topics only. Endpoint response field descriptions are authoritative.

---

## Use case traceability

| Use case | Endpoint(s) | Requirements |
|----------|-------------|--------------|
| UC-1 | queue/summary, queue/unreplied | AC-12, AC-13 |
| UC-2 | queue/summary, queue/unreplied | AC-12, AC-13 |
| UC-3 | queue/summary, queue/untagged | AC-12, AC-14 |
| UC-4 | metrics/summary | AC-16 |
| UC-5 | metrics/summary | AC-16 |
| UC-6 | metrics/summary | AC-16 |
| UC-7 | metrics/summary | AC-16 |
| UC-8 | metrics/median-trends | AC-18 |
| UC-9 | distribution/volume | AC-20 |
| UC-10 | distribution/resolution | AC-21 |
| UC-11 | distribution/backlog, distribution/backlog-trend | AC-22, AC-23 |
| UC-12 | (all filtered endpoints) | AC-8, AC-9, AC-11 |
| UC-13 | slo/violations | AC-24 |
| UC-14 | slo/compliance | AC-25 |
| UC-15 | (tag filter on all endpoints) | AC-10, AC-30 |
| UC-16 | config | AC-27 |
| UC-17 | metrics/volume | AC-17 |
| UC-18 | queue/stalled | AC-15 |
| UC-19 | activity/heatmap | AC-26 |
| UC-20 | metrics/distribution | AC-19 |
| UC-24 | (frontend-only) | — |
