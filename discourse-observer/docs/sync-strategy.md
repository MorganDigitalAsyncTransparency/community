# Sync Strategy

This document describes how discourse-observer fetches and maintains topic data from a Discourse forum. It covers initial sync, delta sync, scheduling, resume after interruption, and observability.

The strategy is recorded in [ADR 0013](decisions/0013-sync-strategy.md). This document describes the operational flow.

## Overview

```text
                    ┌─────────────┐
                    │  Scheduler  │
                    └──────┬──────┘
                           │ triggers sync cycle
                           ▼
                    ┌─────────────┐
                    │  Observer   │
                    └──────┬──────┘
                           │
              ┌────────────┼────────────┐
              ▼            ▼            ▼
        /categories   /latest.json   Store
         .json         ?page=N       (upsert)
         (1 req)      (paginated)
```

Every sync cycle follows the same sequence:

1. Fetch categories (single request).
2. Paginate `/latest.json` — either to exhaustion (initial) or to the watermark (delta).
3. Normalize and upsert each page of topics into SQLite.
4. Update the high-water mark.

## Initial sync

On first run the database is empty and there is no stored watermark. The observer performs a full crawl.

### Flow

1. Fetch `/categories.json` → store category map.
2. Set `page = 0`.
3. Fetch `/latest.json?page={page}`.
4. Normalize the topics on the page and upsert into SQLite.
5. Record `page` as the last completed page (for resume).
6. If the response contains `more_topics_url`: wait the configured delay, increment `page`, go to step 3.
7. If `more_topics_url` is absent: all topics fetched. Store `max(bumped_at)` as the high-water mark.

### Timing

With ~30 topics per page and a 20-second delay between requests:

| Topics | Pages | Time     |
|--------|-------|----------|
| 1 000  |    34 | ~11 min  |
| 3 000  |   100 | ~33 min  |
| 5 000  |   167 | ~56 min  |

The default delay of 20 seconds yields ~3 requests per minute — well below Discourse's default rate limit of 200 req/min.

## Delta sync

After the initial sync, subsequent runs only need to fetch topics that changed since the last watermark.

### Flow

1. Load the stored high-water mark (`last_bumped_at`).
2. Fetch `/categories.json` → update category map.
3. Set `page = 0`.
4. Fetch `/latest.json?page={page}`.
5. Normalize and upsert topics from the page.
6. Check: are all topics on this page older than or equal to `last_bumped_at`?
   - **Yes** → this page contains no new changes. Stop pagination.
   - **No** → wait the configured delay, increment `page`, go to step 4.
7. Store `max(bumped_at)` from all fetched topics as the new high-water mark.

### Why this works

`/latest.json` is ordered by `bumped_at` descending. Once we reach a page where every topic has `bumped_at ≤ last_bumped_at`, all subsequent pages contain only topics we have already seen in their current state. Re-fetching them would be safe (upsert) but wasteful.

### Typical delta volume

On a moderately active forum (~50 topic changes per working day), delta sync fetches 1–3 pages — completing in under a minute even with throttling.

## Scheduling

### Adaptive frequency

Sync frequency adapts to expected activity:

| Period | Interval | Rationale |
|--------|----------|-----------|
| Working hours (Mon–Fri, 08–18 local) | Every 15 minutes | Most forum activity happens here |
| Evenings (Mon–Fri, 18–22 local) | Every 30 minutes | Reduced activity |
| Nights and weekends | Every 60 minutes | Minimal activity |

These intervals are configuration defaults. The actual values are set in the deployment config.

### Jitter

Each scheduled sync adds a small random jitter (0–60 seconds) to avoid creating a predictable request pattern on the target server.

## Resume after interruption

### During initial sync

The observer persists the last completed page number after each page is stored. If the process is interrupted:

1. On restart, load the last completed page number.
2. Resume from `page = last_completed + 1`.
3. Continue normal pagination until exhaustion.

Since storage uses upsert semantics, there is no risk of duplicate data. The worst case is re-fetching the last partially-processed page — a single wasted request.

### During delta sync

Delta sync is short (1–3 pages typically). If interrupted:

1. On restart, the watermark has not been updated (it is written only after a successful cycle).
2. The next delta sync re-fetches from the old watermark — effectively retrying the interrupted cycle.
3. Upsert semantics prevent duplicates.

No special resume logic is needed for delta sync. The watermark-after-completion pattern makes it naturally idempotent.

## Error handling

### HTTP errors

| Error | Response |
|-------|----------|
| HTTP 429 (rate limited) | Read `Retry-After` header. Wait that duration, then retry the same page. |
| HTTP 5xx (server error) | Wait 60 seconds, retry the same page. After 3 consecutive failures on the same page, abort the sync cycle and log an error. The next scheduled cycle will retry. |
| HTTP 4xx (client error, not 429) | Log the error with the URL and status code. Abort the sync cycle. This likely indicates a configuration problem (wrong URL, invalid API key). |
| Network error (timeout, DNS, connection refused) | Same as 5xx — wait and retry up to 3 times. |

### Partial sync

A sync cycle that is aborted mid-way leaves the database in a consistent state:

- Topics fetched before the failure are stored (upsert is per-page).
- The watermark is not updated (written only on successful completion).
- The next cycle will re-fetch from the old watermark, covering everything that was missed.

### Idempotency

Every sync operation is idempotent. Running the same sync twice with the same data produces the same database state. This is guaranteed by `INSERT OR REPLACE` semantics in SQLite.

## Observability

### What to log

Each sync cycle logs:

| Event | Fields |
|-------|--------|
| Sync started | `type` (initial/delta), `watermark` (if delta) |
| Page fetched | `page`, `topics_count`, `elapsed` |
| Page stored | `page`, `upserted_count` |
| Rate limited | `page`, `retry_after_seconds` |
| Page error | `page`, `status_code`, `retry_attempt` |
| Sync completed | `type`, `pages_fetched`, `topics_upserted`, `new_watermark`, `duration` |
| Sync aborted | `type`, `page`, `reason`, `duration` |

### Health indicators

The observer exposes state that the existing `/api/v1/status` endpoint can surface:

- `last_synced_at` — timestamp of the last successful sync completion.
- `last_sync_duration` — how long the last successful cycle took.
- `last_sync_topics` — how many topics were upserted in the last cycle.
- `sync_state` — `idle`, `running`, or `error`.
- `initial_sync_progress` — during initial sync: `{current_page, estimated_total_pages}`.

An operator can tell sync is healthy when `last_synced_at` is recent and `sync_state` is `idle`.

## First test against a real server

The minimum step to validate the connection:

1. Configure `.env` with real Discourse credentials.
2. Run the observer with `page=0` only (fetch one page, store it, stop).
3. Verify: topics appear in SQLite, categories are resolved, timestamps parse correctly.
4. Check: the response contains `more_topics_url` (confirming pagination will work for subsequent pages).

This is a single HTTP request — zero risk to the target server.

## Configuration

All sync parameters are provided through deployment configuration (environment variables or config file):

| Parameter | Default | Description |
|-----------|---------|-------------|
| `SYNC_DELAY_SECONDS` | `20` | Delay between paginated requests |
| `SYNC_INTERVAL_WORK` | `15m` | Sync interval during working hours |
| `SYNC_INTERVAL_EVENING` | `30m` | Sync interval during evenings |
| `SYNC_INTERVAL_OFF` | `60m` | Sync interval during nights/weekends |
| `SYNC_WORK_HOURS_START` | `08` | Start of working hours (local time, 24h) |
| `SYNC_WORK_HOURS_END` | `18` | End of working hours (local time, 24h) |
| `SYNC_MAX_RETRIES` | `3` | Max consecutive retries per page on error |
| `SYNC_JITTER_SECONDS` | `60` | Max random jitter added to scheduled syncs |
