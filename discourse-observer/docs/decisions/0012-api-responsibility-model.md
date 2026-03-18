# 12. API Responsibility Model

**Status:** Proposed
**Date:** 2026-03-18

## Context

The frontend dashboard is fully implemented with mock data. The backend pipeline (Go) has defined module boundaries but no implementation yet. The next step is defining an HTTP API that the backend exposes and the frontend consumes.

This decision determines **where computation happens** — which calculations are performed by the backend versus the frontend, and how the API is structured. The choice affects:

- **Frontend replaceability.** If a different frontend or a non-browser consumer (MCP server, CLI tool, notification service) needs the same data, can it get it without reimplementing calculation logic?
- **Evolvability.** New views, metrics, and breakdowns will be added over time. How much backend work does each new view require?
- **Complexity distribution.** Where does business logic live, and how much coupling exists between the API shape and any specific consumer?

The frontend currently performs all calculations: medians, bucketing, rankings, heatmaps, SLO compliance, stalled-topic detection. These are pure functions operating on arrays of `Topic` objects. The question is whether the API should return raw topic data and let consumers calculate, return pre-computed view-ready aggregates, or something in between.

An additional consideration: the project plans to expose data via an MCP server, enabling AI assistants and other tools to query support metrics. MCP consumers will ask varied questions — not just the ones today's dashboard answers.

### Relevant prior decisions

- [ADR 0005](0005-storage-format.md): Raw data is stored as NDJSON files.
- [ADR 0006](0006-analytical-storage.md): Derived data lives in SQLite, queryable for dashboard requests. The analytical store already implies backend-side computation.

## Alternatives Considered

### A. Raw data API

The backend serves normalized topic lists. Consumers perform all calculations.

**Endpoints:** A single `/topics` endpoint (or a small number of list endpoints) returning `Topic[]` with filtering. Consumers compute medians, rankings, heatmaps, SLO checks, and bucketing themselves.

**Trade-offs:**
- Simplest backend — essentially a filtered data dump.
- Frontend keeps its existing calculation logic unchanged.
- Every new consumer must reimplement or import the full calculation library. An MCP server, a CLI tool, or a replacement frontend would each need to duplicate this logic.
- Large payload sizes as the topic set grows. Every request transfers all matching topics, even when the consumer only needs a single aggregate number.
- No benefit from the SQLite analytical store (ADR 0006), which was specifically chosen to pre-compute derived data for dashboard queries.
- Contradicts the stated goal of frontend replaceability.

### B. View-bound aggregate API

The backend computes and returns exactly what each dashboard view renders. One endpoint per view section, returning pre-formatted data.

**Endpoints:** `/queue/summary`, `/metrics/cards`, `/metrics/volume-chart`, `/metrics/first-reply-trend`, `/slo/violations`, `/activity/heatmap`, etc. Each returns view-ready data including formatted strings.

**Trade-offs:**
- Frontend becomes a pure rendering layer — easy to replace.
- Backend knows what every view needs, so payloads are minimal.
- Tight coupling: every new view, layout change, or metric tweak requires a backend change. The API becomes a mirror of the UI.
- MCP consumers must use endpoints designed around dashboard cards and charts, not around the questions they want to ask.
- Poor evolvability — the API cannot answer questions it was not designed to render.

### C. Domain aggregate API

The backend computes domain-meaningful aggregates (medians, rankings, compliance rates, distributions) without coupling to specific views. Endpoints are organized around data concepts (queue state, response metrics, tag distribution, SLO status, activity patterns), not around UI components.

**Endpoints:** `/queue/summary`, `/queue/unreplied`, `/metrics/summary`, `/metrics/volume`, `/distribution/volume`, `/slo/compliance`, `/activity/heatmap`, etc. Each returns computed data in domain terms (milliseconds, counts, percentages) rather than formatted display strings.

**Trade-offs:**
- Frontend is a thin rendering layer — easy to replace.
- MCP server can consume the same endpoints and compose answers.
- Endpoints are stable across UI changes — a layout rework does not change the API.
- Values are in machine-readable units (milliseconds, integers) — consumers format for their medium.
- New views that combine existing aggregates differently need no backend changes.
- More backend complexity than option A — the backend must implement all calculation logic.
- Adding a genuinely new metric (not a new view of existing data) requires a backend change.
- The endpoint set will grow as new aggregate concepts are added, but individual endpoints remain stable.

### D. Parametric query API

The backend exposes a flexible query layer where consumers specify dimensions, aggregations, filters, and groupings via query parameters. Think of it as an analytics query builder over the topic dataset.

**Example:** `GET /api/v1/query?group_by=tag&aggregate=median_resolution&period=30d` returns median resolution grouped by tag. `GET /api/v1/query?group_by=day_of_week,hour&aggregate=count&period=7d` returns the heatmap.

**Trade-offs:**
- Maximum flexibility — consumers can ask questions not anticipated at design time.
- A single endpoint (or very few) covers all current and future needs.
- Truly consumer-agnostic — MCP, dashboards, scripts all use the same query language.
- Significant backend complexity to implement a correct, performant, and safe query engine.
- Query validation becomes a non-trivial problem — what combinations are valid?
- Risk of building a bespoke analytics platform. The project has ~6k topics. The query engine would be more complex than the domain logic it replaces.
- Harder to document, test, and reason about than fixed endpoints.
- Performance optimization is harder — each query is unpredictable, making indexing and caching strategies unclear.

### E. Hybrid: domain aggregates with parametric extension

Combine options C and D. Fixed domain endpoints serve the known, stable aggregates. A parametric endpoint (or a small number of them) handles ad-hoc queries for MCP and future needs.

**Trade-offs:**
- Known views are fast and well-tested through fixed endpoints.
- Parametric endpoint provides flexibility for MCP and unforeseen needs.
- Two API styles to maintain, document, and test.
- The parametric endpoint may never be built if fixed endpoints cover all real needs — speculative complexity.
- Risk of scope creep: the parametric layer keeps growing until it becomes the primary API.

## Decision

*Pending review — this ADR is proposed, not yet accepted.*

**Recommended: Option C — Domain aggregate API.**

Reasoning:

1. **ADR 0006 already commits to backend computation.** The SQLite analytical store exists specifically to serve pre-computed derived data. Option A would waste this infrastructure. Option C is the natural API surface over that store.

2. **Frontend replaceability is a stated goal.** Moving calculations to the backend means any consumer — a new frontend, an MCP server, a CLI tool — can access metrics without reimplementing domain logic. Options C, D, and E all achieve this; option C does it with the least complexity.

3. **Evolvability without over-engineering.** Domain aggregates are stable across UI changes. When genuinely new metrics are needed, adding an endpoint is a bounded, well-understood task. Option D's flexibility sounds appealing but trades predictable, bounded work for unbounded query-engine complexity.

4. **MCP compatibility.** The domain endpoints are already question-shaped: "What is the queue state?", "What are SLO violations?", "How are topics distributed by tag?" An MCP server wraps these naturally. If MCP needs a query that no endpoint covers, the response is to add a domain endpoint — not to build a query engine.

5. **Parametric extension is a future option, not a current need.** If the fixed endpoints prove insufficient for MCP or other consumers, option E can be adopted later by adding a parametric layer alongside the existing endpoints. This is an additive change that does not invalidate option C.

The API returns values in machine-readable units (milliseconds for durations, integers for counts, percentages as whole numbers). Consumers handle formatting for their display medium. This keeps the API consumer-agnostic while keeping payloads precise.

## Consequences

**Positive:**

- Calculations live in one place (backend), eliminating duplication across consumers.
- Frontend becomes a rendering layer that can be replaced without losing business logic.
- MCP server can expose the same endpoints as tools, composing answers from domain aggregates.
- Each endpoint is independently testable, cacheable, and documentable.
- Aligns with ADR 0006's analytical store design — endpoints map naturally to SQLite queries.
- Adding new views that reuse existing aggregates requires no backend changes.

**Negative:**

- All current frontend calculation logic must be reimplemented in Go on the backend.
- Adding a genuinely new metric type requires backend work (new endpoint, new SQLite query, new tests).
- The backend is a larger, more complex service than a simple data proxy.
- If MCP consumers need highly flexible queries, the fixed endpoint set may prove limiting — at which point the parametric extension (option E) becomes necessary.
