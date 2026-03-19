# CLAUDE.md — discourse-observer

This file complements the repository-level [CLAUDE.md](../CLAUDE.md) with project-specific context for Claude Code.

## Project guidance

Read [AI_GUIDELINES.md](AI_GUIDELINES.md) for project-specific rules covering module boundaries, scope constraints, and coding expectations.

Read [ARCHITECTURE.md](ARCHITECTURE.md) for the layer diagram, dependency flow, and design rationale.

Read [docs/documentation-strategy.md](docs/documentation-strategy.md) for how specs, tests, and source files are organized and linked.

## Monorepo structure

This project lives inside a monorepo at `c:\code\community\discourse-observer\`. Shared files that belong to the monorepo root — one level up — include:

- **`../mkdocs.yml`** — MkDocs navigation for the published documentation site. It uses `docs_dir: discourse-observer`, so all nav paths are relative to this project folder. **When adding a new spec file, add a corresponding entry to the `nav:` block in `../mkdocs.yml`.** This file is easy to miss in Phase 5 impact scans because it sits outside the project directory — always include it when specs are created or renamed.

During Phase 5 impact scans, search from the monorepo root (`c:\code\community\`), not just from `discourse-observer/`.

## Current state

The frontend dashboard (`frontend/`) fetches all data from the Go backend API (`/api/v1/` endpoints). It has five pages — Queue, Response Metrics, Distribution, SLO, Activity — covering UC-1 through UC-20. The frontend is a thin rendering client: it receives pre-computed data and displays it without performing domain calculations. The API client layer lives in `frontend/src/api/` (types, client, endpoints). Tag configuration — areas, SLO thresholds, stalled-days settings, and closed-tag definitions — is unified in `config/tagConfig.json`. Response time distribution buckets are in `config/distributionBuckets.json`. Both are created from example files during setup.

The backend pipeline (Go) has module boundaries, architecture decisions, and directory structure defined. The HTTP API layer (`backend/api/`, `backend/domain/`, `backend/mock/`) is implemented with all 17 endpoints from the API contract serving mock data with real domain calculations. The backend serves as the single source of truth for all computation — filtering, bucketing, medians, SLO compliance, and heatmap generation. The data pipeline (fetch, observe, store) is not yet implemented.

## Project-specific delivery details

When executing delivery phases from the repository-level CLAUDE.md:

- **Phase 5 (Impact scan):** Include `../mkdocs.yml` in every scan when specs are created, renamed, or removed.
- **Phase 7 (Rebase and PR):** After rebase, run `npm run lint` and `npm test` to confirm the branch is clean before creating the PR.
- **Phase 8 (CI):** Wait for all checks to pass, including CodeQL and any other configured checks.

## Key constraints

- Single Discourse forum per deployment — not multi-tenant.
- The target server is resource-constrained. Be careful about load, polling frequency, and incremental sync.
- Most forum activity happens during working hours — sync frequency can be optimized around that.
