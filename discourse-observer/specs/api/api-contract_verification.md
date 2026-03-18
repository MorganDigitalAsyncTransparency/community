# API Contract — Verification Strategy

This document defines how the API contract specification (`api-contract.md`) is verified.

## Automated verification (when backend is implemented)

Each requirement in the API contract spec maps to one or more contract tests. These tests will live in `tests/api/` and verify:

1. **Response structure** — Each endpoint returns the documented fields with correct types. Missing fields, extra fields, and wrong types are test failures.
2. **Filter behavior** — Each endpoint correctly applies (or ignores) `period`, `from`/`to`, and `tag` parameters as documented.
3. **Sort order** — List endpoints return items in the documented default sort order.
4. **Empty dataset handling** (AC-5) — Filters that match no topics produce the normal response structure with zero counts or empty arrays, not 404.
5. **Error responses** (AC-7, AC-11) — Invalid filter values produce 400 responses with the documented error structure.
6. **Calculation correctness** — Aggregate endpoints (medians, compliance percentages, heatmap counts) produce values matching the same calculations currently implemented in the frontend's pure functions.

Test naming convention: `api-contract.contract.test.ts` (or `api_contract_contract_test.go` for Go tests).

## Manual verification (current phase)

Since the backend API does not exist yet, the specification is verified through review:

- [ ] Every dashboard view element (cards, tables, charts) has a corresponding API endpoint that provides its data.
- [ ] Every use case (UC-1 through UC-20) is traceable to at least one endpoint in the traceability table.
- [ ] Filter semantics match the existing frontend behavior (period scopes by `createdAt`, tag scopes by tag membership).
- [ ] Response field names and types match the data shapes consumed by frontend components.
- [ ] Endpoints that the frontend calls without period filtering (weekly backlog trend) are documented as period-exempt.
- [ ] The configuration endpoint provides all fields needed for area navigation, SLO threshold display, and stalled-topic threshold display.
