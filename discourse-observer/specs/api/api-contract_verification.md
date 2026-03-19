# API Contract — Verification Strategy

This document defines how the API contract specification (`api-contract.md`) is verified.

## Automated verification

Each requirement in the API contract spec maps to one or more contract tests. These tests live in `backend/api/contract_test.go` (HTTP-level shape verification) and `backend/domain/*_test.go` (calculation correctness) and verify:

1. **Response structure** — Each endpoint returns the documented fields with correct types. Missing fields, extra fields, and wrong types are test failures.
2. **Filter behavior** — Each endpoint correctly applies (or ignores) `period`, `from`/`to`, and `tag` parameters as documented. Filter parsing and validation is tested in `backend/api/filters_test.go`.
3. **Sort order** — List endpoints return items in the documented default sort order.
4. **Empty dataset handling** (AC-5) — Filters that match no topics produce the normal response structure with zero counts or empty arrays, not 404.
5. **Error responses** (AC-7, AC-11) — Invalid filter values produce 400 responses with the documented error structure.
6. **Calculation correctness** — Aggregate endpoints (medians, compliance percentages, heatmap counts) produce values matching the same calculations implemented in the frontend's pure functions. Domain tests use the same fixtures and expected values as the frontend tests.

Test files:

| Test file | Verifies |
|-----------|----------|
| `backend/api/contract_test.go` | Response shapes for all 17 endpoints, content type, error format, empty data handling |
| `backend/api/filters_test.go` | Filter parameter parsing, validation, error cases |
| `backend/domain/median_test.go` | Median calculation (truncation, odd/even, empty) |
| `backend/domain/filter_test.go` | Period, date range, tag, unreplied, untagged filters |
| `backend/domain/queue_test.go` | Queue summary, stalled topic detection, boundary cases |
| `backend/domain/metrics_test.go` | Response metrics summary, answer rate |
| `backend/domain/distribution_test.go` | Histogram bucketing, boundary behavior |
| `backend/domain/trend_test.go` | Volume bucketing, median trends, granularity, MondayOf |
| `backend/domain/tagdist_test.go` | Tag rankings (volume, resolution, backlog), weekly backlog |
| `backend/domain/slo_test.go` | SLO violations, compliance, threshold selection |
| `backend/domain/heatmap_test.go` | 7x24 heatmap grid, day mapping, empty input |

## Manual verification (completed)

The following items were verified during API contract design:

- [x] Every dashboard view element (cards, tables, charts) has a corresponding API endpoint that provides its data.
- [x] Every use case (UC-1 through UC-20) is traceable to at least one endpoint in the traceability table.
- [x] Filter semantics match the existing frontend behavior (period scopes by `createdAt`, tag scopes by tag membership).
- [x] Response field names and types match the data shapes consumed by frontend components.
- [x] Endpoints that the frontend calls without period filtering (weekly backlog trend) are documented as period-exempt.
- [x] The configuration endpoint provides all fields needed for area navigation, SLO threshold display, and stalled-topic threshold display.
