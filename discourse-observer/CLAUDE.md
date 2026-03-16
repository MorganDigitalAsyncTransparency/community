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

The frontend dashboard (`frontend/`) is actively implemented with mock data. It has five pages — Queue, Response Metrics, Distribution, SLO, Volume — covering UC-1 through UC-17. Specs, tests, and source files are in place for all implemented use cases. SLO thresholds are configured per tag in `config/sloThresholds.json` and tag/area groupings in `config/tagConfig.json` (both created from example files during setup).

The backend pipeline (Go) is in the foundation stage: module boundaries, architecture decisions, and directory structure are defined, but implementation has not started.

## Project-specific delivery details

When executing delivery phases from the repository-level CLAUDE.md:

- **Phase 5 (Impact scan):** Include `../mkdocs.yml` in every scan when specs are created, renamed, or removed.
- **Phase 7 (Rebase and PR):** After rebase, run `npm run lint` and `npm test` to confirm the branch is clean before creating the PR.
- **Phase 8 (CI):** Wait for all checks to pass, including CodeQL and any other configured checks.

## Key constraints

- Single Discourse forum per deployment — not multi-tenant.
- The target server is resource-constrained. Be careful about load, polling frequency, and incremental sync.
- Most forum activity happens during working hours — sync frequency can be optimized around that.
