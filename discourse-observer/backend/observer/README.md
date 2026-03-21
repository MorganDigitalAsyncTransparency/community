# observer/

This module is responsible for turning raw Discourse data into structured observations.

## Responsibility

The observer sits between the raw API layer and the rest of the system. It:

- Receives raw data from `discourse/`
- Normalizes it into internal types defined in `model/`
- Detects meaningful changes by comparing current data with previous observations
- Produces observation records that capture what changed, when, and in what context
- Coordinates fetch operations (scheduled sync, on-demand, backfill)

## Boundaries

This module does **not**:

- Call the Discourse API directly (it receives data from `discourse/`)
- Define its own data types (it uses types from `model/`)
- Persist data directly (it passes observations to `storage/`)
- Perform analytics or derive higher-level events

## Current implementation

`Observer` in `observer.go` coordinates sync cycles using two interfaces: `FetchClient` (implemented by `discourse/`) and `StorageBackend` (implemented by `storage/`). It supports three sync modes:

- **Initial sync** (`RunInitialSync`) — full crawl with page-by-page resume. Auto-selected when no watermark exists.
- **Delta sync** (`RunDeltaSync`) — incremental from a stored watermark. Fetches only topics changed since last sync.
- **Detail sync** (`RunDetailSync` in `detail_sync.go`) — fetches post revision history per topic during low-activity windows. Extracts tag change, category move, and title edit timestamps from `/t/{id}.json` and `/posts/{id}/revisions/{v}.json`. Uses delta revision fetching (tracks `last_revision` per topic, only fetches new revisions). Interruptible between topics via context cancellation. Handles deleted topics (404) by preserving history and marking as skipped.

`Run()` auto-detects initial vs delta based on whether a watermark exists. Detail sync is triggered by the scheduler, not by `Run()`.

`Normalize` maps `model.RawTopic` fields to domain types, deriving outcome (solved/self-closed/open) from Discourse flags and constructing topic URLs.

### Key types

| Type | Where | Purpose |
|------|-------|---------|
| `FetchClient` | `observer.go` | Interface for Discourse API calls (topics, categories, topic detail, revisions) |
| `StorageBackend` | `observer.go` | Interface for persistence (topics, watermarks, detail sync tracking, events) |
| `SyncResult` | `observer.go` | Return value from all sync methods (mode, pages, topics, duration) |
| `model.TopicDetailState` | `model/revision.go` | Topic ID + last fetched revision version, for delta revision fetching |
| `model.TopicEvent` | `model/revision.go` | Extracted event: topic ID, event type, timestamp, detail JSON |

## Design expectations

- Observation logic should be composed of pure transformation functions where possible
- Change detection should be deterministic and testable with known inputs
- The module should not assume which categories, tags, or workflows matter — that comes from `config/`
- Observation operations (sync, fetch, backfill) should be composable and independently testable
