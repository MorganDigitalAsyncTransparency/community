# Detail Sync

Detail sync enriches topics with revision history — tag changes, category moves, and title edits — by fetching post revisions from the Discourse API during low-activity windows.

The approach is described in [sync-strategy.md](../../docs/sync-strategy.md) (Detail sync section) and decided in [ADR 0013](../../docs/decisions/0013-sync-strategy.md).

---

## Model types

DS-1 — **Topic detail response.** A `RawTopicDetail` type represents the relevant fields from `/t/{id}.json`: topic ID and a `post_stream` containing posts. Each post has an ID, post number, and version count.

DS-2 — **Revision response.** A `RawRevision` type represents a single revision from `/posts/{id}/revisions/{v}.json`. Fields: `created_at`, `title_changes` (previous/current strings), `tags_changes` (previous/current string slices), `category_id_changes` (previous/current ints). Field names match the Discourse API (notably `tags_changes`, not `tag_changes`).

DS-3 — **Topic event.** A `TopicEvent` type represents a stored event extracted from a revision: topic ID, event type (`tag_change`, `category_move`, `title_edit`), timestamp, and a detail string (JSON or human-readable description of the change).

---

## Client endpoints

DS-4 — **FetchTopicDetail.** `FetchTopicDetail(ctx, topicID) (*model.RawTopicDetail, error)` fetches `/t/{id}.json`. Uses the same retry/auth logic as existing methods. Respects the client's configured delay between requests.

DS-5 — **FetchPostRevision.** `FetchPostRevision(ctx, postID, version) (*model.RawRevision, error)` fetches `/posts/{post_id}/revisions/{version}.json`. Uses the same retry/auth logic as existing methods. Respects the client's configured delay between requests.

---

## Observer interface expansion

DS-6 — **FetchClient expansion.** The `FetchClient` interface adds `FetchTopicDetail` and `FetchPostRevision` methods matching the client signatures from DS-4 and DS-5.

DS-7 — **StorageBackend expansion.** The `StorageBackend` interface adds methods for storing topic events and tracking detail sync progress with revision versions:

- `SaveTopicEvents(ctx, events []model.TopicEvent) error`
- `SaveDetailSync(ctx, topicID int, lastRevision int, syncedAt time.Time) error`
- `TopicsNeedingDetailSync(ctx, limit int) ([]TopicDetailState, error)`
- `LoadTopicEvents(ctx, topicID int) ([]model.TopicEvent, error)`

`TopicDetailState` includes topic ID and last fetched revision version (0 if never synced), enabling delta revision fetching.

---

## Storage

DS-8 — **Topic events table.** A `topic_events` table stores extracted revision events:

| Column | Type | Description |
|--------|------|-------------|
| `id` | INTEGER PRIMARY KEY | Auto-increment |
| `topic_id` | INTEGER NOT NULL | References topics(id) |
| `event_type` | TEXT NOT NULL | `tag_change`, `category_move`, or `title_edit` |
| `happened_at` | TEXT NOT NULL | RFC 3339 timestamp from the revision |
| `detail` | TEXT NOT NULL | JSON description of the change |

DS-9 — **Detail sync tracking.** The existing `topic_detail_sync` table gains a `last_revision` column (INTEGER NOT NULL DEFAULT 0) to track the highest fetched revision version per topic. This enables delta fetching — only new revisions are fetched on subsequent syncs.

DS-10 — **Prioritization query.** `TopicsNeedingDetailSync` returns topics ordered by: (1) never detail-synced, (2) `bumped_at` newer than `synced_at` (something changed), (3) oldest `synced_at` first. Returns both topic ID and last fetched revision version.

---

## Observer — RunDetailSync

DS-11 — **RunDetailSync method.** `RunDetailSync(ctx context.Context) (SyncResult, error)` enriches topics with revision data:

1. Call `TopicsNeedingDetailSync(ctx, limit)` to get prioritized topics with their last fetched revision.
2. For each topic:
   a. Fetch `/t/{id}.json` to get the first post ID and its version count.
   b. If version > last fetched revision and version > 1: fetch only revisions from `(lastRevision + 1)` through `version` (or from 2 if never synced). Skip revisions already fetched.
   c. Extract tag change, category move, and title edit events from each revision.
   d. Store events via `SaveTopicEvents`.
   e. Call `SaveDetailSync(ctx, topicID, version, time.Now())` to record progress.
   f. Respect delay between each HTTP request (not just between topics).
3. Return `SyncResult` with mode `"detail"`.

DS-12 — **Interruptibility.** `RunDetailSync` checks `ctx.Done()` between topics. On cancellation, it returns what was completed so far. Completed topics remain marked — detail sync resumes from where it left off.

DS-13 — **No-revision topics.** If a topic's first post has version = 1, no revision fetches are needed. The topic is still marked as detail-synced (with `last_revision = 1`) so it is not re-selected until bumped.

DS-14 — **Deleted topic handling.** If `/t/{id}.json` returns 404, keep all stored history (events are never deleted) and mark the topic as skipped (`last_revision = -1`) so it is not re-selected for detail sync.

---

## Mock server

DS-15 — **Topic detail endpoint.** The mock server serves `GET /t/{id}.json` returning a `RawTopicDetail` with the first post having a plausible version count derived from the mock data (e.g., topics with tag changes get version > 1).

DS-16 — **Revision endpoint.** The mock server serves `GET /posts/{post_id}/revisions/{version}.json` returning a `RawRevision` with plausible tag, category, or title changes for at least some topics.

---

## Scheduler integration

DS-17 — **SyncRunner expansion.** The `SyncRunner` interface adds `RunDetailSync(ctx context.Context) (observer.SyncResult, error)`.

DS-18 — **Low-activity window detection.** The scheduler identifies low-activity windows using peak activity data from stored topics (via the heatmap/activity-by-hour pattern). The current UTC hour is compared against historical activity levels. When the current hour falls in a low-activity period, detail sync is triggered. Fallback: the existing zero-streak heuristic (SC-9) is used when insufficient historical data exists.

DS-19 — **Detail sync triggering.** When low activity is detected, the scheduler calls `RunDetailSync` with a context that is canceled when the next delta sync is due. This makes detail sync naturally interruptible — it runs during idle time and yields when regular sync resumes.

DS-20 — **Detail sync logging.** The scheduler logs detail sync events per the observability table in sync-strategy.md: detail sync started (topics queued), topic detail fetched (topic ID, revision count), detail sync completed (topics enriched, duration).

---

## Boundary rules

DS-21 — **Module dependencies.** Observer imports only `model/`. HTTP concerns (endpoints, auth, delay) stay in `discourse/`. Persistence (SQL, tables) stays in `storage/`. The scheduler triggers detail sync — the observer does not know about timing. The scheduler reads activity data through an interface, not by importing `domain/` or `storage/` directly.
