# Scheduler

The scheduler drives the sync lifecycle: it runs an initial or delta sync on startup, repeats delta syncs on a configurable interval with jitter, detects low-activity windows, and shuts down gracefully when the server stops.

The scheduling approach is described in [sync-strategy.md](../../docs/sync-strategy.md) (Scheduling section) and decided in [ADR 0013](../../docs/decisions/0013-sync-strategy.md).

---

## Configuration

SC-1 — **Config struct.** A `SyncConfig` struct holds all scheduling parameters. It has no dependencies on other modules.

| Field | Env var | Type | Default |
|-------|---------|------|---------|
| `InitialDelay` | `SYNC_INITIAL_DELAY_SECONDS` | `time.Duration` | `20s` |
| `DeltaDelay` | `SYNC_DELTA_DELAY_SECONDS` | `time.Duration` | `2s` |
| `Interval` | `SYNC_INTERVAL` | `time.Duration` | `15m` |
| `LowActivityThreshold` | `SYNC_LOW_ACTIVITY_THRESHOLD` | `int` | `3` |
| `MaxRetries` | `SYNC_MAX_RETRIES` | `int` | `3` |
| `JitterMax` | `SYNC_JITTER_SECONDS` | `time.Duration` | `60s` |

SC-2 — **Config loading.** `LoadSyncConfig()` reads env vars and returns a `SyncConfig` with defaults for any unset variable. It is a pure function: reads environment, returns struct.

---

## Sync runner interface

SC-3 — **SyncRunner interface.** The scheduler defines a `SyncRunner` interface:

```go
type SyncRunner interface {
    Run(ctx context.Context) (observer.SyncResult, error)
    RunDeltaSync(ctx context.Context) (observer.SyncResult, error)
}
```

`*observer.Observer` satisfies this interface. Tests use a fake implementation. The scheduler never imports `discourse/` or `storage/`.

---

## Startup behavior

SC-4 — **Immediate first sync.** On `Start`, the scheduler runs one sync cycle immediately using `SyncRunner.Run` (which auto-detects initial vs delta). It does not wait for the interval before the first sync.

SC-5 — **Dev-mode skip.** If Discourse credentials are absent (`DISCOURSE_URL` is empty), the scheduler does not start. The server still runs for the API. A log message explains why sync is disabled.

---

## Interval loop

SC-6 — **Periodic delta sync.** After the first sync completes, the scheduler runs `SyncRunner.RunDeltaSync` every `Interval` plus a random jitter between 0 and `JitterMax`.

SC-7 — **Jitter.** Each scheduled sync adds a uniformly random duration in `[0, JitterMax)` to the base interval. This avoids creating a predictable request pattern on the target server.

---

## Concurrency guard

SC-8 — **No overlapping syncs.** The scheduler prevents concurrent sync cycles. If a sync is still running when the next interval fires, the scheduled sync is skipped. The guard uses a mutex or atomic flag — not a channel.

---

## Low-activity detection

SC-9 — **Consecutive zero-change tracking.** The scheduler counts consecutive delta syncs where `SyncResult.TopicsStored == 0`. The counter resets when a sync stores at least one topic.

SC-10 — **Low-activity logging.** When the consecutive zero-change count reaches `LowActivityThreshold`, the scheduler logs a low-activity event with the count. Detail sync triggering based on this detection is deferred to PR 5.

---

## Logging

SC-11 — **Sync lifecycle logging.** The scheduler logs events per the observability table in [sync-strategy.md](../../docs/sync-strategy.md):

| Event | Fields |
|-------|--------|
| Sync started | type, watermark (if delta) |
| Sync completed | type, pages fetched, topics upserted, new watermark, duration |
| Sync aborted | type, reason, duration |
| Low activity detected | consecutive zero count |

---

## Status exposure

SC-12 — **Thread-safe sync status.** The scheduler exposes operational state via a thread-safe struct:

| Field | Type | Description |
|-------|------|-------------|
| `State` | string | `"idle"`, `"running"`, or `"disabled"` |
| `LastDuration` | `time.Duration` | Duration of the last completed sync |
| `LastTopics` | int | Topics upserted in the last completed sync |
| `LastSyncedAt` | `*time.Time` | Timestamp of the last completed sync |

The API reads this state through a `SyncStateProvider` interface defined in the `api` package. The scheduler satisfies the interface. `main.go` wires them together via dependency injection.

---

## Graceful shutdown

SC-13 — **Context cancellation.** When the context passed to `Start` is canceled, the scheduler stops scheduling new syncs. If a sync is in progress, it waits for it to finish before `Start` returns.

---

## Boundary rules

SC-14 — **Module dependencies.** The scheduler module (`backend/scheduler/`) imports `config/` and `observer/` (for the `SyncResult` type only). It defines a `SyncRunner` interface — `*observer.Observer` satisfies it, but the scheduler never calls observer methods directly. It does not import `discourse/` or `storage/`.

The API module reads scheduler state through the `SyncStateProvider` interface — it does not import `scheduler/` directly.
