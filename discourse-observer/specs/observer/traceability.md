# Observer — Traceability Matrix

This matrix shows how observer specs decompose into requirements and verification artifacts.

The observer module is infrastructure — it populates the database that all use cases read from. It does not map to individual use cases directly but is a prerequisite for all of them.

---

## Observer behavior

| Spec | Requirements | Verification |
|------|-------------|--------------|
| [observer-behavior.md](observer-behavior.md) | (high-level behavior description, no numbered requirements) | `backend/pipeline_test.go`, `backend/pipeline_report_test.go` |

---

## Sync metadata storage

| Spec | Requirements | Verification |
|------|-------------|--------------|
| [sync-metadata.md](sync-metadata.md) | R1 — mock bumped_at | `backend/pipeline_test.go` (all topics have non-nil last_activity_at) |
| | R2 — sync_state table | `backend/storage/sqlite_test.go` |
| | R3 — topic_detail_sync table | `backend/storage/sqlite_test.go` |
| | R4 — storage methods | `backend/storage/sqlite_test.go` |
| | R5 — idempotent migration | `backend/storage/sqlite_test.go` |
| | R6 — interface unchanged in PR 2 | Expanded in [initial-delta-sync.md](initial-delta-sync.md) |

---

## Initial and delta sync

| Spec | Requirements | Verification |
|------|-------------|--------------|
| [initial-delta-sync.md](initial-delta-sync.md) | R1 — FetchClient paginated | `backend/sync_test.go`: compile-time interface check |
| | R2 — StorageBackend expanded | `backend/sync_test.go`: compile-time interface check |
| | R3 — Client satisfies interface | `backend/sync_test.go`: compile-time interface check |
| | R4 — RunInitialSync | `backend/sync_test.go`: `TestInitialSyncEndToEnd`, `TestInitialSyncResume`, `TestWatermarkIsMaxBumpedAt` |
| | R5 — RunDeltaSync | `backend/sync_test.go`: `TestDeltaSyncEndToEnd`, `TestDeltaSyncStopsAtWatermark`, `TestDeltaSyncWithoutWatermarkFails` |
| | R6 — Run auto-detect | `backend/sync_test.go`: `TestRunAutoDetectsMode` |
| | R7 — SyncResult | `backend/sync_test.go`: `TestSyncResultFields` |
| | R8 — Mock server sort | `backend/sync_test.go`: `TestMockServerSortOrder` |
| | R9 — Boundary rules | `go build ./backend/...` (no cross-boundary imports) |

---

## Gaps

- **Detail sync** (PR 5) — not yet specified. Will add spec and requirements when implemented.
- **Scheduling** (PR 4) — not yet specified. Will add spec and requirements when implemented.
