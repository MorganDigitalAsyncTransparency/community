# 5. Raw Data Storage Format

**Status:** Proposed
**Date:** 2026-03-13

## Context

discourse-observer collects data from a Discourse forum by polling the API periodically. Between runs — and eventually on a server running continuously — the collected data must be persisted so that the observer can resume without re-fetching already collected data from Discourse.

The stored observations are also consumed by other parts of discourse-observer: an analysis layer that derives workflow events, and ultimately the frontend that presents dashboards. Key analyses include measuring how long a topic has been assigned to a given team and how long a user has been waiting — both of which require precise timestamps for each state transition.

The system runs in two stages:

- **Now:** Runs locally, started manually, not always running.
- **Later:** Runs on a server 24/7 with continuous polling.

The storage format must serve both stages without requiring a migration.

The primary motivation for observation is workflow visibility: understanding how support topics move through categories and tags over time. This requires not just snapshots of current state, but the full history of changes with timestamps. Revision data from the Discourse API is the authoritative source for this history.

This ADR refines the storage decision in ADR 0002, which stated SQLite as the storage choice without distinguishing between raw and analytical data. It covers only the raw data layer — the persistent record of what was fetched from Discourse. Derived analytical data (computed events, aggregations, metrics for dashboards) is a separate concern to be addressed in a future ADR.

## Alternatives Considered

### SQLite

An embedded relational database stored as a single file. No separate server process. Supports SQL queries, transactions, and concurrent reads.

Fits both the local and server scenarios. Allows the observer to query its own state efficiently (e.g., find topics updated since the last sync). However, raw Discourse data does not map naturally to a relational schema — nested structures like tag lists and revision diffs require either flattening or JSON columns, and schema evolution as new fields are added becomes a migration burden. SQLite is better suited to structured, queryable derived data than to raw append-only observation records.

### CSV files

One file per entity type, rows appended on each sync. Human-readable and directly importable into spreadsheet tools (aligns with an existing xlsx proof-of-concept).

CSV is flat. Tag lists and revision diffs cannot be represented naturally without flattening into multiple columns or separate files with join keys. Schema evolution (adding fields) is fragile. Concurrent appends on a 24/7 server introduce race conditions without additional locking.

### NDJSON (newline-delimited JSON)

One file per entity type. Each line is a complete JSON object representing one observation or revision. New data is appended — existing lines are never modified.

JSON is the native format of the Discourse API, which minimizes transformation. Nested structures (tag lists, revision diffs with from/to values) are represented naturally. Each line is independently parseable, making the files streamable and easy to inspect with standard tools (`jq`, text editors). Append-only writes are safe and simple. Files survive process crashes without corruption of existing lines.

The format is language-agnostic, making it straightforward for the analysis layer to consume.

### Flat JSON (one file per sync run)

One JSON file per sync run, containing all observations from that run.

Simple to produce, but requires reading all files to reconstruct history. No natural resume point. Large sync runs produce large files that are expensive to parse incrementally.

## Decision

Use **NDJSON files, one per entity type**, as the persistent storage format for discourse-observer.

Files:

- `topics.ndjson` — one line per detected topic change (not one line per poll); deleted topics are recorded as a tombstone line with a `deleted_at` field (requires active deletion detection — see below)
- `revisions.ndjson` — one line per fetched revision
- `categories.ndjson` — one line per detected category change
- `tags.ndjson` — one line per detected tag change

Each line is a complete, self-contained JSON object including an `observed_at` timestamp added by the observer. Existing lines are never modified. New observations are appended.

On startup, the observer reads the tail of the relevant files to determine where it left off. For topics, the resume point is the latest `bumped_at` seen. For revisions, it is the highest revision number seen per post. For categories, it is the latest `updated_at` seen. Tags have no reliable timestamp in the Discourse API and are re-fetched in full on each sync; a new line is appended only when a change is detected by comparing against the last observed state.

Deletion detection: the normal sync polls `/latest.json` and only sees active topics. To detect deletions, the observer must periodically verify that previously seen topics are still accessible. When a topic returns a 404, a tombstone line is appended to `topics.ndjson`.

Deduplication keys: if the process restarts mid-sync, duplicate lines may appear. Consumers should deduplicate on `(id, bumped_at)` for topics, `(post_id, revision_number)` for revisions, and `(id, updated_at)` for categories. Tags are identified by name; deduplication key is `(name, observed_at)`.

discourse-observer does not embed a database. Storage is transparent files that other parts of the system can read directly.

## Consequences

**Positive:**

- No database process or schema to manage in discourse-observer
- Matches Discourse API's native JSON format — minimal transformation
- Naturally handles nested structures (tag lists, revision diffs)
- Append-only writes are simple, crash-safe, and require no locking for single-writer use
- Files are human-readable and inspectable with standard tools
- Readable by all parts of the system without coupling to a specific query layer
- Works identically in local and server deployment

**Negative:**

- No built-in query support — the analysis layer must scan or index files itself
- Resume logic requires reading file tails on startup — as files grow this becomes slower; a separate state file may be needed to avoid scanning large files
- Duplicate lines are possible if the process crashes mid-sync and restarts; the analysis layer must tolerate or deduplicate them
- File size grows indefinitely — rotation or archival strategy will be needed eventually
- Concurrent writes from multiple observer processes would require coordination (not a concern initially, single-writer assumed)
