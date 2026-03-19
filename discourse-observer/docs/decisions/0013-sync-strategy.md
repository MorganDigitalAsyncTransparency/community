# 0013. Sync Strategy

**Status:** Accepted
**Date:** 2026-03-19

## Context

discourse-observer needs to fetch topic data from a real Discourse forum and keep its local SQLite database up to date. The target forum may contain 5 000+ topics. The target server is resource-constrained — aggressive fetching is not acceptable.

The current implementation ([backend/discourse/client.go](../../backend/discourse/client.go)) calls `/latest.json` once and returns whatever the first page contains (~30 topics). There is no pagination, no delta detection, and no scheduling. The observer runs a single atomic fetch→normalize→store cycle with upsert semantics (`INSERT OR REPLACE`).

This ADR decides how the observer will:

1. Perform an **initial sync** — fetching all existing topics into an empty database.
2. Perform **delta syncs** — efficiently fetching only what changed since the last sync.
3. **Throttle** requests to stay well within Discourse rate limits and respect the server's resource constraints.
4. **Resume** an interrupted sync without data loss or duplication.

### Discourse API constraints (researched)

| Aspect | Detail |
|--------|--------|
| Endpoint | `/latest.json?page=N` — returns ~30 topics per page, ordered by `bumped_at` descending |
| Pagination signal | Response includes `more_topics_url` when more pages exist; field is absent on the last page |
| Page depth | No documented hard limit — pagination continues until all topics are exhausted |
| Rate limits (default) | 200 req/min per IP, 50 req/10s per IP. Admin API keys: 60 req/min shared |
| Rate limit signal | HTTP 429 with `Retry-After` header |
| `bumped_at` | Updated on new reply, title edit, manual bump, some staff actions. Reliable indicator of "topic was modified" |
| Categories | `/categories.json` — not paginated, returns all categories in one request |
| Topic detail | `/t/{id}.json` — returns full topic data including post IDs, timestamps, and metadata not present in list views |
| Post revisions | `/posts/{post_id}/revisions/{version}.json` — returns diff for a specific revision. Includes title changes, body changes, tag changes, and category changes. Version numbering starts at 2 (version 1 is the original). Fields: `created_at`, `current_revision`, `last_revision`, `version_count`, `body_changes`, `title_changes`, `tags_changes` |
| Webhooks | Discourse can push `topic` events (including tag changes) to an external URL. Alternative to polling for real-time change detection. Requires configuring a webhook on the Discourse server |
| Bulk alternatives | None built-in. Data Explorer plugin is available on the target server |

### Key numbers

- 5 000 topics ÷ 30 per page = ~167 pages
- At 1 request every 20 seconds: ~56 minutes for full crawl (fits within the 1-hour budget)
- At 3 requests per minute: stays far below the 200 req/min default limit
- Delta sync of a quiet day (~50 changed topics): 2 pages, done in under a minute

## Alternatives Considered

### A. Full paginated crawl of /latest.json

Paginate `/latest.json?page=0,1,2,...` until `more_topics_url` disappears. Throttle to ~3 req/min. For delta sync, stop pagination when all topics on a page have `bumped_at` ≤ the stored watermark.

**Pros:** Uses a single well-documented endpoint. No server-side plugins needed. Natural ordering by `bumped_at` makes delta cutoff simple. Resume is straightforward — track the last page completed.

**Cons:** Initial sync of 5 000 topics takes ~1 hour. Cannot control page size (fixed at ~30).

### B. Search endpoint (/search.json)

Use the search API with date filters (`after:YYYY-MM-DD`) to find recently changed topics.

**Pros:** Could be more targeted for delta sync.

**Cons:** Search results are capped at 50 and paginate differently (cursor-based, max ~5 pages). Not suitable for initial sync. Search is computationally expensive on the server — worse for a resource-constrained target. Topic data in search results is incomplete compared to `/latest.json`.

### C. Topic-by-ID enumeration (/t/{id}.json)

Iterate through topic IDs sequentially, fetching each topic individually.

**Pros:** Full topic detail per request.

**Cons:** Requires one request per topic — 5 000 requests for initial sync. Deleted/non-existent IDs return 404, wasting requests. No way to know the highest ID without another query. Far more load on the server than paginated listing. Does not scale.

### D. Data Explorer plugin

Use the Data Explorer plugin (available on the target server) to run custom SQL queries via API.

**Pros:** Maximum flexibility — exact fields, exact filters, arbitrary page sizes.

**Cons:** Ties the observer to a specific plugin — deployments without Data Explorer cannot use it. Requires maintaining raw SQL queries that must track Discourse schema changes. The plugin's own rate limit is strict (2 req/10s). Adds coupling to Discourse internals that the standard API abstracts away.

## Decision

**Alternative A — full paginated crawl of `/latest.json`** with a `bumped_at` watermark for delta sync.

This is the only approach that works with a vanilla Discourse instance, does not require server-side changes, and keeps request volume low enough for resource-constrained servers.

### Core mechanism

- **Initial sync:** Paginate `/latest.json?page=0,1,2,...` until `more_topics_url` is absent. Throttle to ~3 requests per minute (~1 request every 20 seconds). Store topics via existing upsert. Track progress (last completed page) for resume.
- **Delta sync:** Same endpoint, same pagination. Stop when every topic on a page has `bumped_at` ≤ the stored high-water mark. On a typical day this means 1–3 pages. Uses a shorter delay between requests (configurable, default 2 seconds) since delta sync is brief and low-volume.
- **Detail sync:** During detected low-activity periods, fetch revision history for topics via `/t/{id}.json` (to discover post IDs and version counts) and `/posts/{post_id}/revisions/{v}.json` (to extract tag change timestamps, category move timestamps, and title edit timestamps). These transitions are not visible in `/latest.json` but are essential for understanding support workflows.
- **Watermark:** After each successful sync cycle, persist `max(bumped_at)` from all fetched topics as the high-water mark.
- **Categories:** Fetch `/categories.json` once per sync cycle (single request, not paginated).

### Throttling

Two speed tiers:

- **Initial sync and detail sync:** 20-second delay between requests (~3 req/min). These are long-running and must stay gentle.
- **Delta sync:** 2-second delay between requests (~30 req/min). Delta fetches 1–3 pages — the total request count is small even at higher speed.
- On HTTP 429: respect `Retry-After` header, then resume from the same page.
- No parallelism — sequential requests only.

### Resume

- Initial sync tracks the last successfully stored page number.
- On restart, resume from the next page rather than starting over.
- Since storage is upsert-based, re-fetching already-stored pages is safe (idempotent) but wasteful — so we avoid it.

### Why not the others

- **Search (B):** Result caps, incomplete data, and higher server load make it unsuitable for both initial and delta sync.
- **Topic-by-ID (C):** 5 000+ individual requests is unacceptable for a resource-constrained server.
- **Data Explorer (D):** Couples the observer to a plugin and Discourse's internal schema. Strict rate limit (2 req/10s). Not portable to deployments without the plugin.

## Consequences

**Positive:**

- Works with any vanilla Discourse instance — no plugins, no admin access beyond a read-only API key.
- Initial sync completes in ~1 hour with minimal server impact (~3 req/min).
- Delta sync is fast — typically 1–3 pages for a day's changes — and runs at higher speed since request volume is low.
- Detail sync fills in revision history and tag change timestamps during low-activity windows, without impacting the server during peak usage.
- Scheduling adapts to observed activity patterns rather than relying on hardcoded time windows.
- Resume after interruption avoids re-fetching completed work.
- Upsert storage means any sync is idempotent — safe to re-run.

**Negative:**

- Initial sync is slow (~1 hour). This is an intentional trade-off for server courtesy.
- Fixed page size (~30) is not configurable via the API.
- Topics that are never bumped (very old, no activity) will be on the last pages of the initial sync — but they will be fetched eventually.
- If a topic is bumped during initial sync, it may appear on an earlier page and a later page — harmless due to upsert, but the observer will normalize it twice.
- Detail sync via `/t/{id}.json` is one request per topic — acceptable at low-activity throttle rates but not suitable for bulk use during peak hours.
