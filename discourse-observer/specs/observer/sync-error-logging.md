# Sync Error Logging

When a sync cycle fails (HTTP errors, rate limiting, network errors), the failure must be visible in the sync log — the same place operators already look for sync history. Currently, failed syncs are only logged to stdout and leave no trace in the sync-log endpoint.

This spec covers error visibility in the sync log, retry state in the progress indicator, and related observability improvements.

Decided as part of the observability rollout. Related: [sync-strategy.md](../../docs/sync-strategy.md) observability section, [api-contract.md](../api/api-contract.md) AC-33.

---

## Requirements

Requirements use the prefix **SE** (Sync Error).

### Error entries in sync log

**SE-1.** `SyncLogEntry` includes an `Error` field (string). Empty string means success; non-empty means the sync failed with that error message.

**SE-2.** When a sync cycle fails, the scheduler saves a `SyncLogEntry` with the error message, mode, and duration. Pages and topics reflect progress made before the failure.

**SE-3.** The `sync_log` SQLite table includes an `error` column (TEXT, NOT NULL, default empty string). The migration is additive — existing rows get an empty error field.

**SE-4.** The `/api/v1/sync-log` endpoint includes an `error` field (string) in each entry. Empty string for successful syncs, error message for failures.

**SE-5.** The frontend sync-log page renders error entries with a red left border and displays the error message.

**SE-6.** Error entries follow the same retention rules as normal entries (20 per mode, no-change deduplication does not apply to error entries).

**SE-7.** Backward compatibility: existing sync log entries (before migration) appear with an empty error field. The status endpoint (AC-28) is unchanged.

### Retry visibility in progress

**SE-8.** `SyncProgress` includes `RetryAttempt` (int) and `RetryReason` (string). Zero attempt means no retry in progress.

**SE-9.** The client reports retries via a callback (`RetryFunc`). The callback is wired to `SyncStatus.SetRetry` at startup, without crossing module boundaries.

**SE-10.** The `/api/v1/sync-log` progress object includes `retryAttempt` and `retryReason` fields.

**SE-11.** The frontend progress row shows retry state with a red border, the attempt number, and the reason. The pulse animation stops during retries.

### Progress mode

**SE-12.** `runSync` receives the sync mode and sets it on `SyncProgress` at creation, so the progress row shows the correct mode immediately — not after the first page callback.

### Detail sync efficiency

**SE-13.** `TopicsNeedingDetailSync` excludes topics where `synced_at >= last_activity_at`. Already up-to-date topics are not re-fetched.

**SE-14.** The scheduler skips detail sync when the preceding delta sync failed.
