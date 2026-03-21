# Sync Error Logging

When a sync cycle fails (HTTP errors, rate limiting, network errors), the failure must be visible in the sync log — the same place operators already look for sync history. Currently, failed syncs are only logged to stdout and leave no trace in the sync-log endpoint.

This spec adds an error field to sync log entries so failures appear alongside successful syncs in the sync-log UI.

Decided as part of the observability rollout. Related: [sync-strategy.md](../../docs/sync-strategy.md) observability section, [api-contract.md](../api/api-contract.md) AC-33.

---

## Requirements

Requirements use the prefix **SE** (Sync Error).

**SE-1.** `SyncLogEntry` includes an `Error` field (string). Empty string means success; non-empty means the sync failed with that error message.

**SE-2.** When a sync cycle fails, the scheduler saves a `SyncLogEntry` with the error message, mode, and duration. Pages and topics reflect progress made before the failure.

**SE-3.** The `sync_log` SQLite table includes an `error` column (TEXT, NOT NULL, default empty string). The migration is additive — existing rows get an empty error field.

**SE-4.** The `/api/v1/sync-log` endpoint includes an `error` field (string) in each entry. Empty string for successful syncs, error message for failures.

**SE-5.** The frontend sync-log page renders error entries with a red left border and displays the error message.

**SE-6.** Error entries follow the same retention rules as normal entries (20 per mode, no-change deduplication does not apply to error entries).

**SE-7.** Backward compatibility: existing sync log entries (before migration) appear with an empty error field. The status endpoint (AC-28) is unchanged.
