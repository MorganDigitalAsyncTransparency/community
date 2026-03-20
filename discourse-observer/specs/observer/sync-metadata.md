# Sync Metadata

This document specifies the sync metadata storage and the mock data fix that enables watermark-based sync.

**Related:** [sync-strategy.md](../../docs/sync-strategy.md) · [ADR 0013](../../docs/decisions/0013-sync-strategy.md) · [observer-behavior.md](observer-behavior.md)

## Problem

Two issues block watermark-based sync as designed in ADR 0013:

1. **Missing `bumped_at` in mock data.** The mock server sets `BumpedAt` and `LastPostedAt` from `LastActivityAt`, which is only populated for 7 of 44 topics. Real Discourse always populates `bumped_at`. Without it, `max(bumped_at)` is meaningless.

2. **No sync state persistence.** The storage layer has no tables or methods for watermarks, page progress, or detail sync tracking. The observer cannot resume an interrupted sync or perform delta sync.

## Requirements

### R1 — Mock server populates `bumped_at` for all topics

The `convertTopics` function must set `BumpedAt` and `LastPostedAt` to a non-nil value for every topic, using fallback logic that matches real Discourse behavior:

1. `LastActivityAt` (replied-open topics with explicit activity timestamp)
2. `ResolvedAt` (solved or self-closed topics — resolution is the last bump)
3. `FirstReplyAt` (replied topics without explicit LastActivityAt)
4. `CreatedAt` (unreplied topics — Discourse sets `bumped_at = created_at`)

All 44 mock topics must have non-nil `bumped_at` and `last_posted_at` in API responses.

### R2 — `sync_state` table stores watermark and page progress

A key-value table `sync_state (key TEXT PRIMARY KEY, value TEXT NOT NULL)` stores:

- `watermark` — RFC 3339 timestamp of `max(bumped_at)` from the last successful sync.
- `last_completed_page` — integer (as text) tracking initial sync progress.

### R3 — `topic_detail_sync` table tracks per-topic detail sync

A table `topic_detail_sync (topic_id INTEGER PRIMARY KEY, synced_at TEXT NOT NULL)` records when each topic was last detail-synced.

### R4 — Storage methods on SQLiteStore

Seven methods on `SQLiteStore` (not on the `StorageBackend` interface):

| Method | Behavior |
|--------|----------|
| `SaveWatermark(ctx, time.Time)` | Upsert watermark as RFC 3339 |
| `LoadWatermark(ctx)` → `*time.Time` | Return stored watermark or nil |
| `SaveLastPage(ctx, int)` | Upsert last completed page |
| `LoadLastPage(ctx)` → `int` | Return stored page or -1 if absent |
| `ClearLastPage(ctx)` | Delete last_completed_page entry |
| `SaveDetailSync(ctx, topicID, time.Time)` | Upsert detail sync timestamp |
| `TopicsNeedingDetailSync(ctx, limit)` → `[]int` | Return topic IDs needing detail sync: never-synced first, then stale (oldest `synced_at`), limited |

### R5 — Migration is additive and idempotent

Both new tables are created in the existing `migrate` function using `CREATE TABLE IF NOT EXISTS`. The `topics` table is unchanged.

### R6 — StorageBackend interface unchanged in this PR

The `StorageBackend` interface in `observer.go` is not expanded in this PR. New methods are concrete on `SQLiteStore` only. Interface expansion happens in the [initial-delta-sync](initial-delta-sync.md) PR that wires watermarks into the observer.

## Verification

| Requirement | Method |
|-------------|--------|
| R1 | Automated: test all 44 mock server topics have non-nil `bumped_at` and `last_posted_at` |
| R1 | Automated: pipeline report shows 0 null `last_activity_at` |
| R2–R5 | Automated: storage round-trip tests in `backend/storage/sqlite_test.go` |
| R6 | Manual: verify `StorageBackend` interface is unchanged in this PR (expanded in [initial-delta-sync](initial-delta-sync.md)) |
