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
         ┌─────────────────┼─────────────────┐
         ▼                 ▼                  ▼
   /categories.json   /latest.json       /t/{id}.json
     (1 req)           ?page=N            (detail)
                      (delta/initial)   (low-activity)
         │                 │                  │
         └─────────────────┼──────────────────┘
                           ▼
                     Store (upsert)
```

Three sync modes serve different purposes:

| Mode | Purpose | Speed | When |
|------|---------|-------|------|
| **Initial** | Populate an empty database | ~3 req/min | First run |
| **Delta** | Fetch recent changes | ~30 req/min | Every 15 min during operation |
| **Detail** | Fetch revision history (tag changes, category moves, title edits) | ~3 req/min | Automatically during low activity |

Every sync cycle follows the same base sequence:

1. Fetch categories (single request).
2. Paginate `/latest.json` — either to exhaustion (initial) or to the watermark (delta).
3. Normalize and upsert each page of topics into SQLite.
4. Update the high-water mark.

## Initial sync

On first run the database is empty and there is no stored watermark. The observer performs a full crawl.

### Initial sync flow

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

### Delta sync flow

1. Load the stored high-water mark (`last_bumped_at`).
2. Fetch `/categories.json` → update category map.
3. Set `page = 0`.
4. Fetch `/latest.json?page={page}`.
5. Normalize and upsert topics from the page.
6. Check: are all topics on this page older than or equal to `last_bumped_at`?
   - **Yes** → this page contains no new changes. Stop pagination.
   - **No** → wait the delta delay (default 2 seconds), increment `page`, go to step 4.
7. Store `max(bumped_at)` from all fetched topics as the new high-water mark.

### Speed

Delta sync uses a shorter delay between requests (default 2 seconds, ~30 req/min) compared to initial sync. This is safe because delta sync fetches very few pages — the total request count stays low even at higher speed.

### Why this works

`/latest.json` is ordered by `bumped_at` descending. Once the observer reaches a page where every topic has `bumped_at ≤ last_bumped_at`, all subsequent pages contain only topics already seen in their current state. Re-fetching them would be safe (upsert) but wasteful.

### Typical delta volume

On a moderately active forum (~50 topic changes per working day), delta sync fetches 1–3 pages — completing in under a minute even with throttling.

## Detail sync

`/latest.json` returns summary data per topic — enough for tracking activity, but not enough for understanding *when* tag changes, category moves, or title edits happened. That detail lives in the post revision history.

### Why revisions matter

A topic's current state (tags, category, title) is visible in `/latest.json`. But the *transitions* — when a tag was added, when a topic was moved between categories, when the title was edited — are only visible through revision history. These transitions are essential for understanding support workflows: how topics move through triage, escalation, and resolution.

### Data sources

| Endpoint | Data | Usage |
|----------|------|-------|
| `/t/{id}.json` | Full topic metadata, post IDs, `version` count per post | Identify which topics have revisions to fetch |
| `/posts/{post_id}/revisions/{version}.json` | Per-revision diff: `title_changes`, `tags_changes`, `category_id` changes, `body_changes`, `created_at` | Extract when each change happened |

Revision version numbering starts at 2 (version 1 is the original post). The `last_revision` field indicates the highest version available.

### Detail sync flow

1. Wait for the scheduler to detect a low-activity window (see [Scheduling](#scheduling)).
2. Select topics that need detail enrichment — either never detail-synced, or where `bumped_at` is newer than the last detail sync.
3. For each selected topic:
   a. Fetch `/t/{id}.json` to get the first post ID and its `version` count.
   b. If `version > last fetched revision` and `version > 1`: fetch only new revisions — from `(last fetched revision + 1)` through `version`, or from 2 if never synced. This avoids re-reading revisions already stored.
   c. Extract and store: tag change timestamps, category move timestamps, title change timestamps.
   d. Mark the topic as detail-synced with the current timestamp and the highest fetched revision version.
   e. Wait the configured delay (default 20 seconds) between each HTTP request.
   f. If `/t/{id}.json` returns 404 (deleted topic): keep all stored history and mark as skipped so it is not re-selected.
4. Stop when the activity window ends or all selected topics are enriched.

### Prioritization

Topics are selected for detail sync in this order:

1. Topics that have never been detail-synced (new topics from delta sync).
2. Topics where `bumped_at` is newer than the last detail sync (something changed).
3. Oldest detail-synced topics first (staleness-based refresh).

### Interruptibility

Detail sync is interruptible at any point — each topic is stored independently. If activity picks up or the process restarts, detail sync simply resumes from where it left off next time a low-activity window occurs.

## Scheduling

### Default interval

Delta sync runs every 15 minutes while the server is running. This is the base interval.

### Learned activity patterns

Rather than relying on hardcoded time windows, the scheduler identifies low-activity windows using peak activity data derived from stored topics. The current UTC day-of-week and hour are compared against historical topic creation patterns (the same heatmap data used by the Activity page). Hours with activity below a threshold relative to the peak are considered low-activity windows suitable for detail sync.

Until enough data is collected (first few days of operation), the scheduler uses a simple fallback heuristic: if the last N delta syncs each returned 0 changed topics, assume low activity.

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
| Detail sync started | `topics_queued` |
| Topic detail fetched | `topic_id`, `revisions_count`, `elapsed` |
| Detail sync completed | `topics_enriched`, `duration` |
| Low activity detected | `consecutive_zero_syncs` |

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
| `SYNC_INITIAL_DELAY_SECONDS` | `20` | Delay between requests during initial and detail sync |
| `SYNC_DELTA_DELAY_SECONDS` | `2` | Delay between requests during delta sync |
| `SYNC_INTERVAL` | `15m` | Base interval between delta syncs |
| `SYNC_LOW_ACTIVITY_THRESHOLD` | `3` | Consecutive zero-change syncs before declaring low activity |
| `SYNC_MAX_RETRIES` | `3` | Max consecutive retries per page on error |
| `SYNC_JITTER_SECONDS` | `60` | Max random jitter added to scheduled syncs |
