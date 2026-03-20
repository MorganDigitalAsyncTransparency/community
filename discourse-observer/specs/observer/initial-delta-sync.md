# Initial and Delta Sync

This document specifies how the observer performs initial sync (full crawl) and delta sync (incremental from watermark).

**Related:** [sync-strategy.md](../../docs/sync-strategy.md) · [ADR 0013](../../docs/decisions/0013-sync-strategy.md) · [observer-behavior.md](observer-behavior.md) · [sync-metadata.md](sync-metadata.md)

## Problem

The observer has paginated fetching (PR 1) and sync metadata storage (PR 2), but they are not wired together. The observer's `Run()` method fetches all topics in one call and stores them — it has no concept of watermarks, page-by-page storage, resume, or incremental sync.

## Requirements

### R1 — FetchClient interface supports paginated fetching

The `FetchClient` interface in the observer package adds a paginated fetch method:

```go
FetchTopicsPages(ctx context.Context, startPage int, fn func(topics []model.RawTopic, page int) error) error
```

The method paginates from `startPage` and calls `fn` for each page. If `fn` returns an error, pagination stops and that error is returned. The observer does not import `discourse.PageConfig` — timing configuration is internal to the concrete client.

The existing `FetchTopics` method remains on the interface for backward compatibility.

### R2 — StorageBackend interface includes sync metadata methods

The `StorageBackend` interface expands to include:

| Method | Purpose |
|--------|---------|
| `SaveWatermark(ctx, time.Time)` | Persist high-water mark |
| `LoadWatermark(ctx)` → `*time.Time` | Load watermark or nil |
| `SaveLastPage(ctx, int)` | Persist last completed page |
| `LoadLastPage(ctx)` → `int` | Load last page or -1 |
| `ClearLastPage(ctx)` | Remove page progress |

`SQLiteStore` already implements all of these. This requirement promotes them to the interface.

### R3 — Discourse client satisfies the new FetchClient interface

The concrete `discourse.Client` satisfies the observer's `FetchTopicsPages(ctx, startPage, fn)` signature. Pagination timing (delay, retries) is configured at construction time and applied internally — the observer does not control or know about these details.

### R4 — RunInitialSync performs a full crawl

`RunInitialSync(ctx) (SyncResult, error)` performs:

1. Fetch categories → build category map.
2. Load last completed page (for resume). If -1, start from page 0.
3. Paginate via `FetchTopicsPages` from the resume page.
4. Per page: normalize, store, save page number, track `max(bumped_at)`.
5. After all pages: save watermark as `max(bumped_at)`, clear last completed page.

### R5 — RunDeltaSync fetches only recent changes

`RunDeltaSync(ctx) (SyncResult, error)` performs:

1. Load watermark. If nil, return an error — caller should run initial sync.
2. Fetch categories → build category map.
3. Paginate via `FetchTopicsPages` from page 0.
4. Per page: normalize, store, track `max(bumped_at)`. If all topics on the page have `bumped_at ≤ watermark`, stop pagination.
5. After pagination: save new watermark as `max(old watermark, max fetched bumped_at)`.

The stop condition uses a sentinel error returned from the callback, checked and swallowed by `RunDeltaSync`.

### R6 — Run auto-detects sync mode

`Run(ctx) (SyncResult, error)` checks for a stored watermark:

- No watermark → `RunInitialSync`
- Watermark exists → `RunDeltaSync`

### R7 — SyncResult reports what happened

Both sync methods return:

```go
type SyncResult struct {
    Mode         string
    PagesFetched int
    TopicsStored int
    NewWatermark *time.Time
    Duration     time.Duration
}
```

### R8 — Mock server serves topics sorted by bumped_at descending

The mock server sorts `rawTopics` by `BumpedAt` descending before serving, matching real Discourse `/latest.json` behavior. This is required for realistic watermark stop testing.

### R9 — Observer boundary rules preserved

- The observer imports only `model/`. It does not import `discourse/` or `storage/`.
- HTTP concerns (delays, retries) stay in `discourse/`.
- Persistence concerns (SQL, tables) stay in `storage/`.

## Verification

All tests are in `backend/sync_test.go` unless otherwise noted.

| Requirement | Test | Method |
|-------------|------|--------|
| R1 | `var _ observer.FetchClient = (*discourse.Client)(nil)` | Compile-time interface check |
| R2 | `var _ observer.StorageBackend = (*storage.SQLiteStore)(nil)` | Compile-time interface check |
| R3 | (same as R1) | Compile-time interface check |
| R4 | `TestInitialSyncEndToEnd` | Empty DB → RunInitialSync → 44 topics, watermark set, last page cleared |
| R4 | `TestInitialSyncResume` | Save last page = 5 → RunInitialSync → starts from page 6 |
| R4 | `TestWatermarkIsMaxBumpedAt` | Watermark equals max bumped_at across all topics |
| R5 | `TestDeltaSyncEndToEnd` | Initial sync → delta sync → watermark unchanged when no new data |
| R5 | `TestDeltaSyncStopsAtWatermark` | Set watermark mid-dataset → verify pagination stops early |
| R5 | `TestDeltaSyncWithoutWatermarkFails` | No watermark → RunDeltaSync returns error |
| R6 | `TestRunAutoDetectsMode` | Empty DB → Run() does initial → Run() again does delta |
| R7 | `TestSyncResultFields` | Mode, PagesFetched, TopicsStored, NewWatermark, Duration all populated |
| R8 | `TestMockServerSortOrder` | All 44 topics in bumped_at descending order |
| R9 | `go build ./backend/...` | Build succeeds with no cross-boundary imports |
