# Observer — Traceability Matrix

This matrix shows how use cases decompose into observer specs, requirements, and verification artifacts.

The observer module is the data pipeline — it fetches, normalizes, and stores the topic data that all use cases read from. Every use case (UC-1 through UC-20) depends on the observer having populated the database. The mapping below groups by observer responsibility rather than by individual use case, since all use cases share the same dependency on the sync pipeline.

---

## Data pipeline (fetch → normalize → store)

| Use case | Spec | Requirements | Verification |
|----------|------|-------------|--------------|
| UC-1 through UC-20 (all) | [observer-behavior.md](observer-behavior.md) | (high-level behavior, no numbered requirements) | `backend/pipeline_test.go`, `backend/pipeline_report_test.go` |

---

## Sync metadata storage

| Use case | Spec | Requirements | Verification |
|----------|------|-------------|--------------|
| (infrastructure) | [sync-metadata.md](sync-metadata.md) | R1 — mock bumped_at | `backend/pipeline_test.go` |
| | | R2 — sync_state table | `backend/storage/sqlite_test.go` |
| | | R3 — topic_detail_sync table | `backend/storage/sqlite_test.go` |
| | | R4 — storage methods | `backend/storage/sqlite_test.go` |
| | | R5 — idempotent migration | `backend/storage/sqlite_test.go` |
| | | R6 — interface unchanged in PR 2 | Expanded in [initial-delta-sync.md](initial-delta-sync.md) |

---

## Initial and delta sync

| Use case | Spec | Requirements | Verification |
|----------|------|-------------|--------------|
| UC-1 through UC-20 (all) | [initial-delta-sync.md](initial-delta-sync.md) | R1 — FetchClient paginated | `backend/sync_test.go`: compile-time check |
| | | R2 — StorageBackend expanded | `backend/sync_test.go`: compile-time check |
| | | R3 — Client satisfies interface | `backend/sync_test.go`: compile-time check |
| | | R4 — RunInitialSync | `backend/sync_test.go`: `TestInitialSyncEndToEnd`, `TestInitialSyncResume`, `TestWatermarkIsMaxBumpedAt` |
| | | R5 — RunDeltaSync | `backend/sync_test.go`: `TestDeltaSyncEndToEnd`, `TestDeltaSyncStopsAtWatermark`, `TestDeltaSyncWithoutWatermarkFails` |
| | | R6 — Run auto-detect | `backend/sync_test.go`: `TestRunAutoDetectsMode` |
| | | R7 — SyncResult | `backend/sync_test.go`: `TestSyncResultFields` |
| | | R8 — Mock server sort | `backend/sync_test.go`: `TestMockServerSortOrder` |
| | | R9 — Boundary rules | `go build ./backend/...` |

---

## Scheduling

| Use case | Spec | Requirements | Verification |
|----------|------|-------------|--------------|
| UC-1 through UC-20 (all) | [scheduler.md](scheduler.md) | SC-1 — Config struct | `backend/scheduler/scheduler_acceptance_test.go`: `TestSyncConfigDefaults` |
| | | SC-2 — Config loading | `backend/scheduler/scheduler_acceptance_test.go`: `TestSyncConfigFromEnv` |
| | | SC-3 — SyncRunner interface | `backend/scheduler/scheduler_acceptance_test.go`: compile-time check |
| | | SC-4 — Immediate first sync | `backend/scheduler/scheduler_acceptance_test.go`: `TestSchedulerRunsImmediately` |
| | | SC-5 — Dev-mode skip | `backend/scheduler/scheduler_acceptance_test.go`: `TestSchedulerSkipsWithoutCredentials` |
| | | SC-6 — Periodic delta sync | `backend/scheduler/scheduler_acceptance_test.go`: `TestSchedulerRunsOnInterval` |
| | | SC-7 — Jitter | `backend/scheduler/scheduler_acceptance_test.go`: `TestSchedulerJitter` |
| | | SC-8 — No overlapping syncs | `backend/scheduler/scheduler_acceptance_test.go`: `TestSchedulerConcurrencyGuard` |
| | | SC-9 — Consecutive zero-change tracking | `backend/scheduler/scheduler_acceptance_test.go`: `TestSchedulerLowActivityDetection` |
| | | SC-10 — Low-activity logging | `backend/scheduler/scheduler_acceptance_test.go`: `TestSchedulerLowActivityDetection` |
| | | SC-11 — Sync lifecycle logging | `backend/scheduler/scheduler_acceptance_test.go`: (covered by lifecycle tests) |
| | | SC-12 — Thread-safe sync status | `backend/scheduler/scheduler_acceptance_test.go`: `TestStatusReflectsSyncState` |
| | | SC-13 — Graceful shutdown | `backend/scheduler/scheduler_acceptance_test.go`: `TestSchedulerGracefulShutdown` |
| | | SC-14 — Module dependencies | `go build ./backend/...` |

---

## Mock server service

| Use case | Spec | Requirements | Verification |
|----------|------|-------------|--------------|
| (infrastructure) | [mock-server-service.md](mock-server-service.md) | MS-1 — Exported handler | Existing mock server tests pass unchanged |
| | | MS-2 — Standalone entrypoint | `go build ./backend/cmd/mockserver` |
| | | MS-3 — Docker service | `docker compose config` validates |
| | | MS-4 — Dockerfile | `docker compose build mockserver` succeeds |
| | | MS-5 — Dev-mode detection | Backend logs "sync scheduler started" with mock URL |
| | | MS-6 — Environment defaults | `.env.example` inspection |
| | | MS-7 — Backward compatibility | `make start` seeds then scheduler syncs |

---

## Detail sync

| Use case | Spec | Requirements | Verification |
|----------|------|-------------|--------------|
| UC-1 through UC-20 (all) | [detail-sync.md](detail-sync.md) | DS-1 — Topic detail response | `backend/observer/detail_sync_test.go`: `TestFetchTopicDetail` |
| | | DS-2 — Revision response | `backend/observer/detail_sync_test.go`: `TestFetchPostRevision` |
| | | DS-3 — Topic event | `backend/observer/detail_sync_test.go`: `TestDetailSyncEndToEnd` |
| | | DS-4 — FetchTopicDetail | `backend/discourse/client_test.go`: `TestFetchTopicDetail` |
| | | DS-5 — FetchPostRevision | `backend/discourse/client_test.go`: `TestFetchPostRevision` |
| | | DS-6 — FetchClient expansion | `backend/observer/detail_sync_test.go`: compile-time check |
| | | DS-7 — StorageBackend expansion | `backend/observer/detail_sync_test.go`: compile-time check |
| | | DS-8 — Topic events table | `backend/storage/sqlite_test.go`: `TestTopicEventsStorage` |
| | | DS-9 — Detail sync tracking | `backend/storage/sqlite_test.go`: `TestDetailSyncTracking` |
| | | DS-10 — Prioritization query | `backend/observer/detail_sync_test.go`: `TestDetailSyncPrioritization` |
| | | DS-11 — RunDetailSync | `backend/observer/detail_sync_test.go`: `TestDetailSyncEndToEnd` |
| | | DS-12 — Interruptibility | `backend/observer/detail_sync_test.go`: `TestDetailSyncInterruptible` |
| | | DS-13 — No-revision topics | `backend/observer/detail_sync_test.go`: `TestDetailSyncNoRevisions` |
| | | DS-14 — Deleted topic handling | `backend/observer/detail_sync_test.go`: `TestDetailSyncDeletedTopic` |
| | | DS-15 — Mock topic detail endpoint | `backend/observer/detail_sync_test.go`: `TestMockServerTopicDetail` |
| | | DS-16 — Mock revision endpoint | `backend/observer/detail_sync_test.go`: `TestMockServerRevisions` |
| | | DS-17 — SyncRunner expansion | `backend/scheduler/scheduler_acceptance_test.go`: compile-time check |
| | | DS-18 — Low-activity window detection | `backend/scheduler/scheduler_acceptance_test.go`: `TestSchedulerLowActivityWindow` |
| | | DS-19 — Detail sync triggering | `backend/scheduler/scheduler_acceptance_test.go`: `TestSchedulerTriggersDetailSync` |
| | | DS-20 — Detail sync logging | `backend/scheduler/scheduler_acceptance_test.go`: `TestSchedulerTriggersDetailSync` |
| | | DS-21 — Module dependencies | `go build ./backend/...` |

---

## Sync error logging

| Use case | Spec | Requirements | Verification |
|----------|------|-------------|--------------|
| (operational) | [sync-error-logging.md](sync-error-logging.md) | SE-1 — Error field on SyncLogEntry | `backend/scheduler/scheduler_acceptance_test.go`: `TestSyncErrorSavedToLog` |
| | | SE-2 — Scheduler saves error entry | `backend/scheduler/scheduler_acceptance_test.go`: `TestSyncErrorSavedToLog` |
| | | SE-3 — SQLite error column | `backend/storage/sqlite_test.go`: `TestSyncLogErrorColumn` |
| | | SE-4 — Endpoint includes error | `backend/api/contract_test.go`: `TestSyncLogErrorEntry` |
| | | SE-5 — Frontend red styling | Manual verification |
| | | SE-6 — Error retention rules | `backend/storage/sqlite_test.go`: `TestSyncLogErrorRetention` |
| | | SE-7 — Backward compatibility | `backend/storage/sqlite_test.go`: `TestSyncLogErrorColumn` |

---

## Gaps

| Gap | Status |
|-----|--------|
| (none) | All observer responsibilities are specified |
