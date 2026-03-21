---
applyTo: "discourse-observer/**"
---

# AGENT.md — discourse-observer

This file complements the repository-level [AGENT.md](../AGENT.md) with project-specific context for AI agents working in this subproject.

## Project guidance

Read [AI_GUIDELINES.md](AI_GUIDELINES.md) for project-specific rules covering module boundaries, scope constraints, and coding expectations.

Read [ARCHITECTURE.md](ARCHITECTURE.md) for the layer diagram, dependency flow, and design rationale.

Read [docs/documentation-strategy.md](docs/documentation-strategy.md) for how specs, tests, and source files are organized and linked.

## Monorepo structure

This project lives inside a monorepo at `c:\code\community\discourse-observer\`. Documentation config (`mkdocs.yml`, `requirements.txt`) lives in `docs/`. **When adding a new spec file, add a corresponding entry to the `nav:` block in `docs/mkdocs.yml`.**

## Current state

The frontend dashboard (`frontend/`) fetches all data from the Go backend API (`/api/v1/` endpoints). It has six pages — Queue, Response Metrics, Distribution, SLO, Activity, and Sync Log (accessible from the sidebar's About section) — covering UC-1 through UC-20 plus operational sync visibility. The frontend is a thin rendering client: it receives pre-computed data and displays it without performing domain calculations. The API client layer lives in `frontend/src/api/` (types, client, endpoints). Tag configuration — areas, SLO thresholds, stalled-days settings, and closed-tag definitions — is unified in `config/tagConfig.json`. Response time distribution buckets are in `config/distributionBuckets.json`. Both are created from example files during setup.

The backend (Go) has the HTTP API layer (`backend/api/`, `backend/domain/`) implemented with all 18 endpoints from the API contract (including the sync-log endpoint AC-33). Handlers read topics from SQLite via the `TopicReader` interface (defined in `api/`, implemented by `storage.SQLiteStore`). Time and tag filters are pushed to SQL WHERE clauses; remaining filtering and all computation (medians, bucketing, SLO compliance, heatmap) happen in `backend/domain/`. The data pipeline is implemented and tested: `backend/discourse/` fetches from a Discourse-compatible API, `backend/observer/` normalizes raw topics to domain types, and `backend/storage/` persists to the same SQLite database that the API reads from. A mock Discourse server (`backend/discourse/mockserver/`) enables end-to-end testing without a real forum and runs as a docker-compose service in dev mode so the full sync pipeline works out of the box. `backend/mock/` provides hardcoded topic fixtures used by the mock server and contract tests (seeded into a temporary SQLite database); it is not imported at runtime.

## Project-specific delivery details

When executing delivery phases from the repository-level AGENT.md:

- **Phase 5 (Impact scan):** Include `docs/mkdocs.yml` in every scan when specs are created, renamed, or removed.
- **Phase 7 (Rebase and PR):** After rebase, run `npm run lint` and `npm test` to confirm the branch is clean before creating the PR.
- **Phase 8 (CI):** Wait for all checks to pass, including CodeQL and any other configured checks.

## Key constraints

- Single Discourse forum per deployment — not multi-tenant.
- The target server is resource-constrained. Be careful about load, polling frequency, and incremental sync.
- Most forum activity happens during working hours — sync frequency can be optimized around that.
