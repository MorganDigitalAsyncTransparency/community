# CLAUDE.md — discourse-observer

This file complements the repository-level [CLAUDE.md](../CLAUDE.md) with project-specific context for Claude Code.

## Project guidance

Read [AI_GUIDELINES.md](AI_GUIDELINES.md) for project-specific rules covering module boundaries, scope constraints, and coding expectations.

Read [ARCHITECTURE.md](ARCHITECTURE.md) for the layer diagram, dependency flow, and design rationale.

## Project-specific delivery details

When executing delivery phases from the repository-level CLAUDE.md:

- **Phase 7 (Rebase and PR):** After rebase, run `npm test` and `npm run lint:md` to confirm the branch is clean before creating the PR.
- **Phase 8 (CI):** Wait for all checks to pass, including CodeQL and any other configured checks.

## Key constraints

- Single Discourse forum per deployment — not multi-tenant.
- The target server is resource-constrained. Be careful about load, polling frequency, and incremental sync.
- Most forum activity happens during working hours — sync frequency can be optimized around that.
- No implementation code exists yet. The project is in foundation stage (specs, docs, ADRs).
