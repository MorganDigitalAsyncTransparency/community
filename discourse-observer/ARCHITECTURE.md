# Architecture

This document describes the current architecture boundaries of discourse-observer and the reasoning behind them.

## Overview

discourse-observer is organized into layers that separate concerns cleanly. Each layer has a single responsibility and communicates with adjacent layers through well-defined interfaces.

### Data flow

Data flow at runtime:

```text
                    ┌─────────────────┐
                    │ backend/scheduler/ │ — runs sync on interval with jitter
                    └────────┬────────┘
                             │ triggers sync cycle
                             ▼
Discourse Forum (external)
        │ HTTP
        ▼
  backend/discourse/      — fetches raw API data, handles auth and pagination
        │ raw API data
        ▼
  backend/observer/       — normalizes, detects changes, coordinates fetch and store
        │ model types
        ▼
  backend/storage/        — persists normalized topics to SQLite
```

API serving:

```text
  backend/storage/       — queries SQLite with time/tag filters (implements TopicReader)
        │ []model.Topic
        ▼
  backend/api/           — HTTP handlers, routing, filter parsing, JSON responses
        │ domain types
        ▼
  backend/domain/        — pure calculation functions (medians, bucketing, SLO, heatmap)
        │ model types
        ▼
  backend/model/         — shared domain types
```

Cross-cutting:

```text
  backend/model/   — shared domain types, used by all modules above
  backend/config/  — forum-specific configuration, provided at startup
  backend/mock/    — hardcoded topic fixtures for pipeline integration tests
```

### Dependency direction

The arrows above show *data flow*, not *import dependencies*. Imports follow dependency inversion:

- `model` has no imports. It is the innermost layer.
- `observer` imports only `model`. It defines interfaces (`FetchClient`, `StorageBackend`) for the adapters.
- `discourse` and `storage` import `model`. They implement the interfaces defined by `observer`. At runtime they are injected into the observer — the observer never imports them.
- `scheduler` imports `config` and `model`. It defines a `SyncRunner` interface that `*observer.Observer` satisfies at runtime. The scheduler never imports `observer`, `discourse`, or `storage` directly. It also exposes a thread-safe `SyncStatus` struct that the API reads through a `SyncStateProvider` interface.
- `domain` imports only `model`. It contains pure calculation functions with no framework or I/O dependencies.
- `api` imports `domain` and `model`. It defines the `TopicReader` and `SyncStateProvider` interfaces and handles HTTP concerns (routing, JSON, filter parsing). At runtime, `storage.SQLiteStore` is injected as the `TopicReader` implementation and the scheduler satisfies `SyncStateProvider`. Handlers delegate computation to `domain`.
- `storage` also implements `api.TopicReader` (via `QueryTopics`), satisfying the interface implicitly. The `api` package never imports `storage` — the dependency is inverted through the interface.
- `mock` imports only `model`. It provides hardcoded topic fixtures used by the mock Discourse server for pipeline integration tests.
- `config` has no imports. Config values are read at startup and passed into module constructors.

## Layer responsibilities

### backend/discourse/

Responsible for all communication with the Discourse API. This module knows how to authenticate, paginate, and fetch raw data from a single Discourse forum. It does **not** interpret the data — it returns it in a form close to the API response.

This isolation means that if the Discourse API changes, only this module needs to change.

### backend/observer/

Responsible for change detection, normalization, and coordinating the fetch-observe-store cycle. The observer defines interfaces for its dependencies (`FetchClient`, `StorageBackend`) and works entirely in terms of `model` types.

The observer supports three sync modes — initial (full crawl), delta (incremental from watermark), and detail (revision history enrichment) — selected automatically or triggered by the scheduler. Initial and delta sync paginate through topics page by page, normalizing and storing each page before advancing. Detail sync fetches per-topic revision history during low-activity windows, extracting tag change, category move, and title edit timestamps. See [initial-delta-sync spec](specs/observer/initial-delta-sync.md) and [detail-sync spec](specs/observer/detail-sync.md) for details.

The observer does not import `discourse` or `storage`. Those modules are injected at startup. This keeps the core logic independent of API details and persistence implementation.

### backend/scheduler/

Drives the sync lifecycle. Runs an initial or delta sync immediately on startup, then repeats delta syncs on a configurable interval with random jitter. Detects low-activity windows using peak activity data (with a zero-streak heuristic as fallback) and triggers detail sync during those windows. Exposes thread-safe sync status for the API.

The scheduler defines a `SyncRunner` interface that `*observer.Observer` satisfies. It does not import `observer`, `discourse`, or `storage` — those are wired together in `main.go`. See [scheduler spec](specs/observer/scheduler.md) for details.

### backend/model/

Contains the internal normalized types and domain concepts used throughout the project. These types are independent of the Discourse API shape and represent the project's own understanding of forum activity.

The model module has no dependencies on other modules.

### backend/config/

Holds forum-specific configuration and adaptation points. This is where a deployment specifies its forum URL, API credentials, polling intervals, and any forum-specific mappings (such as which categories to observe or which tags to track).

The config module has no imports. Config values are provided to other modules at startup — passed into constructors or init functions — rather than being imported directly by those modules.

### backend/storage/

Persists normalized topics in SQLite (decided in [ADR 0006](docs/decisions/0006-analytical-storage.md)). Serves two consumers: the observer writes topics via `StoreTopics` (implementing `observer.StorageBackend`), and the API reads topics via `QueryTopics` (implementing `api.TopicReader`). `QueryTopics` accepts time-range and tag filters, pushing them to SQL WHERE clauses. Raw append-only NDJSON files (decided in [ADR 0005](docs/decisions/0005-storage-format.md)) are planned for a future layer.

### backend/api/

HTTP handlers for all `/api/v1/` endpoints defined in the [API contract](specs/api/api-contract.md). Responsible for routing, query parameter parsing and validation, filter application, and JSON response encoding. Defines the `TopicReader` interface — the abstraction through which handlers load topics from the store. Each handler resolves filter parameters to `model.QueryOpts`, queries the store, then delegates computation to `backend/domain/`. Does not contain business logic.

### backend/domain/

Pure calculation functions implementing domain aggregates: medians, time bucketing, histogram distribution, tag rankings, SLO violation detection, compliance computation, and heatmap generation. These functions receive pre-filtered topic slices and return computed results. No HTTP, I/O, or framework dependencies.

### backend/mock/

Hardcoded topic fixtures providing a realistic dataset covering all endpoint scenarios (unreplied, resolved, stalled, untagged, multi-tag). Used by the mock Discourse server (`discourse/mockserver/`) for pipeline integration tests and by API contract tests (seeded into a temporary SQLite database). The API layer does not import `mock/` at runtime — it reads from SQLite.

## Terminology

These terms have specific meanings in this project. Other documentation uses them consistently with these definitions.

| Term | Meaning |
|------|---------|
| **Fetch** | Retrieve raw data from the Discourse API. Happens in `backend/discourse/`. |
| **Normalize** | Transform raw API data into internal `model` types. Happens in `backend/observer/`. |
| **Observation** | A structured record of what the observer saw — a snapshot of one or more topics at a point in time, including what changed since the previous observation. |
| **Sync** | A complete fetch-normalize-store cycle: poll the API for updates, produce observations, persist them. |
| **Poll** | Check the API for new or changed data. Polling is the mechanism; sync is the full cycle. |

## What is intentionally not included

### Frontend / Dashboard

A React/TypeScript frontend exists in `frontend/` and renders a multi-page dashboard (Queue, Response metrics, Distribution, SLO, Activity). Individual feature specs live in `specs/dashboard/`: queue-visibility, response-metrics, time-period-filter, tag-distribution, slo-monitoring, tag-area-filter, topic-intake (bucketing logic reused by volume charts), stalled-topics (on the Queue page), and peak-activity (on the Activity page). Cross-cutting component behavior is in `specs/dashboard/dashboard-components.md`. The frontend consumes domain aggregate endpoints from the backend API ([specs/api/api-contract.md](specs/api/api-contract.md)). All calculation logic (medians, bucketing, rankings) lives in the backend — the responsibility model is documented in [ADR 0012](docs/decisions/0012-api-responsibility-model.md).

### Backend API

The API contract is specified in [specs/api/api-contract.md](specs/api/api-contract.md). It defines domain aggregate endpoints that serve pre-computed data to the frontend and future consumers (MCP servers, CLI tools). The responsibility model — why the backend computes aggregates rather than serving raw data — is recorded in [ADR 0012](docs/decisions/0012-api-responsibility-model.md).

The API is implemented in `backend/api/` with domain calculations in `backend/domain/`. Handlers read topics from SQLite via the `TopicReader` interface, with time and tag filters pushed to SQL. The data pipeline (fetch → observe → store) populates the same SQLite database that the API reads from.

### Event / History model

The observation layer captures current state and changes, but a formal event sourcing or history model is not implemented yet. This is a likely future addition once the observation patterns are understood.

## Design decisions

Architecture decisions are recorded as ADRs in [docs/decisions/](docs/decisions/README.md). Consult them for context on why the architecture is shaped this way.
