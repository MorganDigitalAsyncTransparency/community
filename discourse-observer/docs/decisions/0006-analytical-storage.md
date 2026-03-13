# 6. Analytical Storage

**Status:** Proposed
**Date:** 2026-03-13

## Context

ADR 0005 establishes NDJSON files as the raw data layer — the persistent, append-only record of what was fetched from Discourse. Raw data is the source of truth but is not queryable: it requires scanning files and applying logic to extract meaning.

The reporting requirements ([specs/reporting-requirements.md](../../specs/reporting-requirements.md)) define the concrete questions the system must answer. These questions drive what the analytical store must contain and what query patterns it must support:

- Which topics have been open the longest without a reply?
- What is the median time from topic creation to first reply, and to resolution?
- How many topics are solved versus self-closed per period, and is that ratio changing over time?
- Which monitored tags have the highest volume or the longest average handling time?
- Which topics are late or stalled relative to configured SLA thresholds?
- How many topics currently carry no tag at all?

Answering these questions requires derived data that does not exist directly in the NDJSON files: tag change events with timestamps, computed reply times and resolution times, per-topic outcome classifications (solved, self-closed, open), and precomputed aggregations scoped to configurable time horizons (7 days, 30 days, 1 year, all time).

This data is consumed by the frontend (a React SPA) via a backend API. Queries must be fast enough for interactive dashboard use.

The system runs in two stages:

- **Now:** Runs locally, started manually, not always running.
- **Later:** Runs on a server 24/7 with continuous polling.

Analytical data must not be recomputed from scratch on every run. As NDJSON files grow, a full replay becomes increasingly expensive. The system must support **incremental computation**: knowing how far it has processed and continuing from that point.

A key design constraint is that the analytical store is a **re-derivable cache**. It can be deleted and rebuilt from NDJSON at any time. This constraint simplifies reasoning: it means no authoritative data ever lives only in the analytical store.

This ADR completes the storage picture begun in ADR 0002 (which named SQLite without distinguishing raw from analytical data) and refined in ADR 0005 (which reserved SQLite for structured derived data).

## Alternatives Considered

### In-memory recompute on startup

No persistent analytical store. On each startup, all NDJSON is read and analysis is rerun in memory.

Simple: no extra dependencies, no schema to manage. Works acceptably when data is small. Fails as data grows: startup becomes slow, and a 24/7 server that crashes and restarts must replay the full history before serving any requests. Incremental computation is not possible — there is no persistent record of where processing stopped.

Rejected because it does not work in the server stage.

### Derived NDJSON files

Mirror the raw data pattern: one NDJSON file per derived event type (e.g., `category-events.ndjson`, `wait-measurements.ndjson`). Analytical outputs are appended just like raw observations.

Consistent with ADR 0005's pattern and requires no new dependencies. However, NDJSON provides no query support — the backend must scan files for every dashboard request. Aggregations and trends require reading full history. Incremental computation is hard to implement cleanly: there is no natural place to record how far into the raw NDJSON processing has reached. Response times for frontend queries would be unpredictable.

Rejected because it suits raw observations, not queryable derived data.

### DuckDB

An embedded OLAP database that can query NDJSON files directly without import. Strong analytical SQL support and a file-based model.

The direct NDJSON querying is compelling, but DuckDB is a heavier dependency with lower Node.js ecosystem maturity than SQLite. At this data volume (~6k topics, ~1000 posts/day), its analytical power is not needed. The added complexity is not justified.

Rejected as disproportionate for the expected scale.

### SQLite

An embedded relational database stored as a single file. No separate process. Supports SQL queries, indexes, transactions, and upserts.

ADR 0005 explicitly reserved SQLite for derived structured data: "SQLite is better suited to structured, queryable derived data than to raw append-only observation records." The expected data volume is well within SQLite's range. SQL queries with indexes provide the response times the frontend needs. Incremental computation is enabled by a watermark table in the database itself — the analysis layer records how far it has processed in the NDJSON files, and resumes from that point on each run. Schema migration is handled with numbered SQL files run at startup (consistent with the project's tooling approach from ADR 0004).

## Decision

Use **SQLite** (`analytics.db`) as the analytical store for discourse-observer.

The database contains:

- **Derived events** — one row per extracted state transition (tag change, solved, self-closed), each with a precise timestamp sourced from revision data
- **Time measurements** — computed durations (time to first reply, time to resolution), stored as rows referencing the topic and the events that bound them
- **Aggregations and trend snapshots** — precomputed summaries updated incrementally, used to serve dashboard queries without full-table scans
- **Processing watermark** — a metadata table recording how far into each NDJSON file analysis has reached (byte offset or line count), enabling safe incremental computation

The analytical store is a re-derivable cache. It must be possible to delete `analytics.db` and rebuild it from the NDJSON files. No authoritative data lives only in the analytical store.

Schema evolution is handled by numbered SQL migration files. On startup, the system checks the current schema version against the applied version recorded in the database and runs any outstanding migrations.

Raw NDJSON files and the analytical SQLite database are the only two persistent storage mechanisms in the system. They have distinct responsibilities: NDJSON is the record of what was observed; SQLite is the record of what was computed.

## Consequences

**Positive:**

- SQL queries with indexes serve dashboard requests efficiently without application-level scanning
- Incremental computation is reliable: the watermark table records exactly where processing stopped, and analysis resumes from that point after any restart
- Single embedded file, no separate process — consistent with ADR 0002's lightweight philosophy
- Re-derivable from NDJSON: the database can be deleted and rebuilt at any time without data loss
- Clear separation of concerns: raw observations in NDJSON, derived meaning in SQLite

**Negative:**

- Schema must evolve alongside analytical requirements — migrations are required as new event types or metrics are added
- Rebuilding from scratch grows slower as NDJSON files accumulate; for long-running deployments, a full rebuild will eventually take meaningful time
- Two persistent stores to manage and back up (NDJSON files and `analytics.db`)
- SQLite write performance degrades under concurrent writers; the system must ensure only one analysis process writes to `analytics.db` at a time (single-writer assumption, consistent with ADR 0005)
