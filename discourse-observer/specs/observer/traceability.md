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

## Gaps

| Gap | Status |
|-----|--------|
| Detail sync (revision history for tag/category transitions) | Not yet specified — planned for PR 5 |
| Scheduling (delta sync interval, low-activity detection) | Not yet specified — planned for PR 4 |
